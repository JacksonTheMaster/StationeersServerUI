package connectivity

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/network"
)

const (
	checkEndpoint       = "https://jxsn.dev/api/v1/connectivity/check"
	checkTimeout        = 18 * time.Second
	nonceSize           = 16
	maxIPv4UDPPayload   = 1472
	probeMagic          = "SSUI-CONNECTIVITY-V1"
	probeAckMagic       = "SSUI-CONNECTIVITY-ACK-V1"
	stateNotChecked     = "not_checked"
	stateChecking       = "checking"
	stateDisabled       = "disabled"
	stateSkippedRunning = "skipped_running"
	stateReachable      = "reachable"
	statePartial        = "partial"
	stateUnreachable    = "unreachable"
	stateBindFailed     = "bind_failed"
	stateServiceError   = "service_unavailable"
	stateConfigError    = "configuration_error"
)

var probeSizes = []int{1000, 1200, 1400}

var (
	checkMu  sync.Mutex
	statusMu sync.RWMutex
	status   = Status{Enabled: true, State: stateNotChecked}
)

var ErrCheckInProgress = errors.New("connectivity check already in progress")

type ProbeResult struct {
	Size          int   `json:"size"`
	Received      bool  `json:"received"`
	LatencyMillis int64 `json:"latencyMs,omitempty"`
}

type Status struct {
	Enabled   bool          `json:"enabled"`
	State     string        `json:"state"`
	Address   string        `json:"address,omitempty"`
	Port      int           `json:"port,omitempty"`
	Probes    []ProbeResult `json:"probes,omitempty"`
	CheckedAt *time.Time    `json:"checkedAt,omitempty"`
	Message   string        `json:"message,omitempty"`
	Error     string        `json:"error,omitempty"`
}

type checkRequest struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Nonce   string `json:"nonce"`
	Sizes   []int  `json:"sizes"`
}

type checkResponse struct {
	State     string        `json:"state"`
	Address   string        `json:"address"`
	Port      int           `json:"port"`
	Probes    []ProbeResult `json:"probes"`
	Message   string        `json:"message"`
	CheckedAt *time.Time    `json:"checkedAt"`
}

func StatusSnapshot() Status {
	statusMu.RLock()
	defer statusMu.RUnlock()

	current := status
	current.Enabled = config.GetConnectivityCheckEnabled()
	if !current.Enabled {
		current.State = stateDisabled
		current.Message = "External connectivity checks are disabled."
	}
	current.Probes = append([]ProbeResult(nil), status.Probes...)
	return current
}

func RunStartupCheck() {
	if !config.GetConnectivityCheckEnabled() {
		setStatus(Status{Enabled: false, State: stateDisabled, Message: "External connectivity checks are disabled."})
		logger.Advertiser.Debugf("External connectivity check disabled by configuration")
		return
	}

	logger.Advertiser.Debugf("Starting external connectivity check")
	result, err := Check()
	if err != nil {
		logger.Advertiser.Debugf("External connectivity check did not run: %v", err)
		return
	}
	logger.Advertiser.Debugf("External connectivity check finished: %s", result.State)
}

func Check() (Status, error) {
	if !config.GetConnectivityCheckEnabled() {
		result := Status{Enabled: false, State: stateDisabled, Message: "External connectivity checks are disabled."}
		setStatus(result)
		logger.Advertiser.Debugf("External connectivity check skipped: disabled by configuration")
		return result, nil
	}
	if !checkMu.TryLock() {
		logger.Advertiser.Debugf("External connectivity check skipped: another check is already running")
		return StatusSnapshot(), ErrCheckInProgress
	}
	defer checkMu.Unlock()

	setStatus(Status{Enabled: true, State: stateChecking, Message: "Checking external connectivity..."})
	if gamemgr.InternalIsServerRunning() {
		result := Status{Enabled: true, State: stateSkippedRunning, Message: "The game server must be stopped before a connectivity check can run."}
		setStatus(result)
		logger.Advertiser.Debugf("External connectivity check skipped: game server is running")
		return result, nil
	}

	result := runCheck()
	setStatus(result)
	logResult(result)
	return result, nil
}

func runCheck() Status {
	port, err := strconv.Atoi(strings.TrimSpace(config.GetGamePort()))
	if err != nil || port < 1 || port > 65535 {
		logger.Advertiser.Debugf("External connectivity check failed: invalid game port %q", config.GetGamePort())
		return Status{Enabled: true, State: stateConfigError, Message: "The configured game port is invalid.", Error: "invalid game port"}
	}

	logger.Advertiser.Debugf("External connectivity check reserving UDP port %d", port)
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		logger.Advertiser.Debugf("External connectivity check could not reserve UDP port %d: %v", port, err)
		return Status{Enabled: true, State: stateBindFailed, Port: port, Message: "The game port is already in use or cannot be opened.", Error: err.Error()}
	}
	defer conn.Close()
	logger.Advertiser.Debugf("External connectivity check reserved UDP port %d", port)

	nonce, err := newNonce()
	if err != nil {
		logger.Advertiser.Debugf("External connectivity check could not create a probe nonce: %v", err)
		return Status{Enabled: true, State: stateServiceError, Port: port, Message: "Could not prepare the connectivity check.", Error: err.Error()}
	}

	address, err := network.ResolveAdvertisedIP(config.GetAdvertiserOverride())
	if err != nil {
		logger.Advertiser.Debugf("External connectivity check could not resolve the advertised address: %v", err)
		return Status{Enabled: true, State: stateConfigError, Port: port, Message: "SSUI could not determine the public server address.", Error: err.Error()}
	}
	logger.Advertiser.Debugf("External connectivity check target resolved: %s:%d", address, port)

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()
	go serveProbeReplies(ctx, conn, nonce)

	logger.Advertiser.Debugf("External connectivity check requesting %d UDP probes", len(probeSizes))
	response, err := requestRemoteCheck(ctx, checkRequest{
		Address: address,
		Port:    port,
		Nonce:   nonce,
		Sizes:   append([]int(nil), probeSizes...),
	})
	if err != nil {
		logger.Advertiser.Debugf("External connectivity check service request failed: %v", err)
		return Status{Enabled: true, State: stateServiceError, Address: address, Port: port, Message: "The external connectivity check service could not be reached.", Error: err.Error()}
	}
	logger.Advertiser.Debugf("External connectivity check service responded: state=%s address=%s:%d", response.State, response.Address, response.Port)

	state := normalizeState(response.State, response.Probes)
	message := response.Message
	if message == "" {
		message = messageForState(state)
	}
	return Status{
		Enabled:   true,
		State:     state,
		Address:   address,
		Port:      port,
		Probes:    response.Probes,
		CheckedAt: response.CheckedAt,
		Message:   message,
	}
}

func requestRemoteCheck(ctx context.Context, request checkRequest) (checkResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return checkResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, checkEndpoint, strings.NewReader(string(body)))
	if err != nil {
		return checkResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: checkTimeout}).Do(req)
	if err != nil {
		return checkResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return checkResponse{}, fmt.Errorf("connectivity service returned HTTP %d", resp.StatusCode)
	}

	var response checkResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return checkResponse{}, err
	}
	return response, nil
}

func serveProbeReplies(ctx context.Context, conn *net.UDPConn, nonce string) {
	buffer := make([]byte, maxIPv4UDPPayload+64)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		length, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
			return
		}

		size, ok := parseProbe(buffer[:length], nonce)
		if !ok {
			continue
		}

		logger.Advertiser.Debugf("External connectivity probe received: %d bytes from %s", size, remote.String())
		if _, err := conn.WriteToUDP(buildProbeAck(nonce, size), remote); err != nil {
			logger.Advertiser.Debugf("External connectivity probe ACK failed for %d bytes to %s: %v", size, remote.String(), err)
			continue
		}
		logger.Advertiser.Debugf("External connectivity probe ACK sent: %d bytes to %s", size, remote.String())
	}
}

func logResult(result Status) {
	logger.Advertiser.Debugf("External connectivity result: state=%s address=%s port=%d message=%q", result.State, result.Address, result.Port, result.Message)
	if result.Error != "" {
		logger.Advertiser.Debugf("External connectivity result error: %s", result.Error)
	}
	for _, probe := range result.Probes {
		logger.Advertiser.Debugf("External connectivity probe result: size=%d received=%t latency=%dms", probe.Size, probe.Received, probe.LatencyMillis)
	}
}

func parseProbe(packet []byte, nonce string) (int, bool) {
	headerSize := len(probeMagic) + nonceSize*2 + 2
	if len(packet) < headerSize || string(packet[:len(probeMagic)]) != probeMagic {
		return 0, false
	}
	nonceStart := len(probeMagic)
	nonceEnd := nonceStart + nonceSize*2
	if string(packet[nonceStart:nonceEnd]) != nonce {
		return 0, false
	}
	size := int(binary.BigEndian.Uint16(packet[nonceEnd : nonceEnd+2]))
	return size, size == len(packet) && isProbeSize(size)
}

func buildProbeAck(nonce string, size int) []byte {
	packet := make([]byte, len(probeAckMagic)+len(nonce)+2)
	copy(packet, probeAckMagic)
	copy(packet[len(probeAckMagic):], nonce)
	binary.BigEndian.PutUint16(packet[len(probeAckMagic)+len(nonce):], uint16(size))
	return packet
}

func newNonce() (string, error) {
	value := make([]byte, nonceSize)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func isProbeSize(size int) bool {
	for _, allowed := range probeSizes {
		if allowed == size {
			return true
		}
	}
	return false
}

func normalizeState(state string, probes []ProbeResult) string {
	switch state {
	case stateReachable, statePartial, stateUnreachable:
		return state
	}
	received := 0
	for _, probe := range probes {
		if probe.Received {
			received++
		}
	}
	if received == len(probeSizes) {
		return stateReachable
	}
	if received > 0 {
		return statePartial
	}
	return stateUnreachable
}

func messageForState(state string) string {
	switch state {
	case stateReachable:
		return "Your gameserver accepted all external UDP connectivity probes."
	case statePartial:
		return "Your gameserver is reachable, but some UDP packet sizes were lost. Check your MTU and firewall settings."
	case stateUnreachable:
		return "Your gameserver cannot be reached from the internet. Check port forwarding and firewall rules."
	default:
		return "External connectivity check finished."
	}
}

func setStatus(next Status) {
	statusMu.Lock()
	defer statusMu.Unlock()
	status = next
}
