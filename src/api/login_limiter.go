package api

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/core/security"
)

const (
	loginFailureLimit = 5
	loginWindow       = 15 * time.Minute
	loginBlock        = 15 * time.Minute
	maxLoginKeys      = 4096
)

type loginAttempt struct {
	Failures     int
	FirstFailure time.Time
	LastFailure  time.Time
	BlockedUntil time.Time
}

var loginAttemptsMu sync.Mutex
var loginAttempts = make(map[string]loginAttempt)

func authenticateLogin(w http.ResponseWriter, r *http.Request, username, password string) (security.User, bool) {
	now := time.Now()
	key := loginAttemptKey(r, username)
	if retry := loginRetryAfter(key, now); retry > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int((retry+time.Second-1)/time.Second)))
		WriteError(w, http.StatusTooManyRequests, "login_throttled", "Too many login attempts; try again later")
		return security.User{}, false
	}
	user, err := security.AuthenticateUser(username, password)
	if err != nil {
		recordLoginFailure(key, now)
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password")
		return security.User{}, false
	}
	clearLoginFailures(key)
	return user, true
}

func loginAttemptKey(r *http.Request, username string) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return host + "|" + security.NormalizeUsername(username)
}

func loginRetryAfter(key string, now time.Time) time.Duration {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	attempt, ok := loginAttempts[key]
	if !ok || !attempt.BlockedUntil.After(now) {
		return 0
	}
	return attempt.BlockedUntil.Sub(now)
}

func recordLoginFailure(key string, now time.Time) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	attempt, ok := loginAttempts[key]
	if !ok || now.Sub(attempt.FirstFailure) > loginWindow {
		attempt = loginAttempt{FirstFailure: now}
	}
	attempt.Failures++
	attempt.LastFailure = now
	if attempt.Failures >= loginFailureLimit {
		attempt.BlockedUntil = now.Add(loginBlock)
	}
	if !ok && len(loginAttempts) >= maxLoginKeys {
		removeOldestLoginAttempt()
	}
	loginAttempts[key] = attempt
}

func clearLoginFailures(key string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	delete(loginAttempts, key)
}

func removeOldestLoginAttempt() {
	var oldestKey string
	var oldest time.Time
	for key, attempt := range loginAttempts {
		if oldestKey == "" || attempt.LastFailure.Before(oldest) {
			oldestKey = key
			oldest = attempt.LastFailure
		}
	}
	delete(loginAttempts, oldestKey)
}
