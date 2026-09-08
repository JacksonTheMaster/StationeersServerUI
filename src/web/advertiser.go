package web

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/api"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config/configchanger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
)

type advertiserOverrideRequest struct {
	Mode  string `json:"mode"`
	Value string `json:"value"`
}

var dnsHostnamePattern = regexp.MustCompile(`(?i)^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func normalizeAdvertiserOverride(mode, value string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	value = strings.TrimSpace(value)

	switch mode {
	case "disabled":
		return "", nil
	case "auto":
		return "auto", nil
	case "ipv4":
		ip := net.ParseIP(value)
		if ip == nil || ip.To4() == nil {
			return "", fmt.Errorf("enter a valid IPv4 address")
		}
		return ip.To4().String(), nil
	case "dns":
		if len(value) == 0 || len(value) > 253 || !dnsHostnamePattern.MatchString(value) || strings.Contains(value, "..") {
			return "", fmt.Errorf("enter a valid DNS hostname")
		}
		return strings.ToLower(value), nil
	default:
		return "", fmt.Errorf("invalid advertiser mode")
	}
}

func SaveAdvertiserOverrideHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	var request advertiserOverrideRequest
	if err := api.DecodeJSONLimit(w, r, &request, 16<<10); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	override, err := normalizeAdvertiserOverride(request.Mode, request.Value)
	if err != nil {
		api.WriteError(w, http.StatusUnprocessableEntity, "invalid_advertiser", err.Error())
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "config_load_failed", "Failed to load configuration")
		return
	}
	cfg.AdvertiserOverride = override
	if override != "" {
		nativeAdvertisementDisabled := false
		cfg.ServerVisible = &nativeAdvertisementDisabled
	}
	if err := configchanger.SaveConfig(cfg, false); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "config_save_failed", "Failed to save configuration")
		return
	}

	api.WriteData(w, http.StatusAccepted, api.RestartReceipt{
		State: "restarting", Message: "Advertiser configuration saved. SSUI is restarting",
	})

	go func() {
		time.Sleep(750 * time.Millisecond)
		logger.Advertiser.Info("Advertiser configuration changed; restarting SSUI")
		update.RestartMySelf()
	}()
}
