//helpers.go

package config

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func getString(jsonVal, envKey, defaultVal string) string {
	if jsonVal != "" {
		return jsonVal
	}
	if envVal := os.Getenv(envKey); envVal != "" {
		return envVal
	}
	return defaultVal
}

func getStringSlice(jsonValue []string, envKey string, fallback []string) []string {
	if len(jsonValue) > 0 {
		return jsonValue
	}
	if envValue := os.Getenv(envKey); envValue != "" {
		// Split the environment variable by commas, trim whitespace
		parts := strings.Split(envValue, ",")
		var result []string
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return fallback
}

func getInt(jsonVal int, envKey string, defaultVal int) int {
	if jsonVal != 0 {
		return jsonVal
	}
	if envVal := os.Getenv(envKey); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			return val
		}
	}
	return defaultVal
}

// getOptionalInt preserves an explicitly configured zero while still allowing
// omitted values to fall back to the environment or the default.
func getOptionalInt(jsonVal *int, envKey string, defaultVal int) int {
	if jsonVal != nil {
		return *jsonVal
	}
	if envVal := os.Getenv(envKey); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			return val
		}
	}
	return defaultVal
}

func intPointer(value int) *int {
	return &value
}

func getBool(jsonVal *bool, envKey string, defaultVal bool) bool {
	if jsonVal != nil {
		return *jsonVal
	}
	if envVal := os.Getenv(envKey); envVal != "" {
		if val, err := strconv.ParseBool(envVal); err == nil {
			return val
		}
	}
	return defaultVal
}

// getUsers retrieves a map[string]string with JSON -> env -> default hierarchy
func getUsers(jsonValue map[string]string, envKey string, defaultValue map[string]string) map[string]string {
	if jsonValue != nil {
		return jsonValue
	}
	if envValue := os.Getenv(envKey); envValue != "" {
		// Expect env var as "user1:hash1,user2:hash2"
		users := make(map[string]string)
		pairs := strings.Split(envValue, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				users[parts[0]] = parts[1]
			}
		}
		if len(users) > 0 {
			return users
		}
	}
	return defaultValue
}

func getDefaultExePath() string {
	if runtime.GOOS == "windows" {
		return "./rocketstation_DedicatedServer.exe"
	}
	return "./rocketstation_DedicatedServer.x86_64"
}

func generateJwtKey() string {

	// ensure we return JwtKey if it's set
	if len(JwtKey) > 0 {
		return JwtKey
	}
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		fmt.Println("Failed to generate JWT key, using fallback")
		return "i-am-a-fallback-32-byte-secret-key!!"
	}
	return base64.RawURLEncoding.EncodeToString(key)
}

func SetV1UIFS(v1uiFS embed.FS) {
	V1UIFS = v1uiFS
}

func GetV1UIFS() embed.FS {
	return V1UIFS
}
