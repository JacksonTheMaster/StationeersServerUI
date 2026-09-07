package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const defaultBodyLimit = 1 << 20

type Envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func WriteData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteErrorDetails(w, status, code, message, nil)
}

func WriteErrorDetails(w http.ResponseWriter, status int, code, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Error: &APIError{Code: code, Message: message, Details: details}})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	return DecodeJSONLimit(w, r, target, defaultBodyLimit)
}

func DecodeJSONLimit(w http.ResponseWriter, r *http.Request, target any, limit int64) error {
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		return errors.New("Content-Type must be application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errors.New("invalid JSON body")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func Method(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
			return
		}
		next(w, r)
	}
}

// JSONBoundary gives old handlers the v3 response contract while they are
// moved over one by one. Downloads and event streams don't go through here.
func JSONBoundary(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buffer := &responseBuffer{header: make(http.Header), status: http.StatusOK}
		next(buffer, r)
		copyHeaders(w.Header(), buffer.header)
		if buffer.status == http.StatusNoContent {
			w.WriteHeader(buffer.status)
			return
		}
		body := bytes.TrimSpace(buffer.body.Bytes())
		if buffer.status >= http.StatusBadRequest {
			WriteError(w, buffer.status, "request_failed", errorMessage(body, http.StatusText(buffer.status)))
			return
		}
		if len(body) == 0 {
			WriteData(w, buffer.status, map[string]string{"message": http.StatusText(buffer.status)})
			return
		}
		var value any
		if json.Unmarshal(body, &value) == nil {
			if object, ok := value.(map[string]any); ok && (object["data"] != nil || object["error"] != nil) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(buffer.status)
				_, _ = w.Write(append(body, '\n'))
				return
			}
			WriteData(w, buffer.status, value)
			return
		}
		WriteData(w, buffer.status, map[string]string{"message": string(body)})
	}
}

type responseBuffer struct {
	header http.Header
	body   bytes.Buffer
	status int
	wrote  bool
}

func (buffer *responseBuffer) Header() http.Header {
	return buffer.header
}

func (buffer *responseBuffer) WriteHeader(status int) {
	if buffer.wrote {
		return
	}
	buffer.wrote = true
	buffer.status = status
}

func (buffer *responseBuffer) Write(data []byte) (int, error) {
	if !buffer.wrote {
		buffer.WriteHeader(http.StatusOK)
	}
	return buffer.body.Write(data)
}

func copyHeaders(destination, source http.Header) {
	for key, values := range source {
		if strings.EqualFold(key, "Content-Length") || strings.EqualFold(key, "Content-Type") {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

func errorMessage(body []byte, fallback string) string {
	if len(body) == 0 {
		return fallback
	}
	var value map[string]any
	if json.Unmarshal(body, &value) == nil {
		if message, ok := value["error"].(string); ok && strings.TrimSpace(message) != "" {
			return message
		}
		if message, ok := value["message"].(string); ok && strings.TrimSpace(message) != "" {
			return message
		}
	}
	return string(body)
}
