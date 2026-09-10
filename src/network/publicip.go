package network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ResolveAdvertisedIP follows the same public-IP lookup used by the game when
// it registers its session. An explicit advertiser override always wins.
func ResolveAdvertisedIP(address string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" || strings.EqualFold(address, "auto") {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("https://api.ipify.org")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("public IP lookup returned HTTP %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		ip := net.ParseIP(strings.TrimSpace(string(body)))
		if ip == nil || ip.To4() == nil {
			return "", errors.New("public IP lookup did not return an IPv4 address")
		}
		return ip.To4().String(), nil
	}

	if ip := net.ParseIP(address); ip != nil {
		if ip.To4() != nil {
			return ip.To4().String(), nil
		}
		return "", errors.New("IPv6 addresses are not supported for advertiser override")
	}

	ips, err := net.LookupIP(address)
	if err != nil {
		return "", err
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.To4().String(), nil
		}
	}
	return "", errors.New("unable to resolve IP from advertiser override")
}
