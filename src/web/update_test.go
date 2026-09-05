package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/setup/update"
)

func TestCheckUpdateHandlerReturnsMajorUpdateState(t *testing.T) {
	originalVersion := config.Version
	t.Cleanup(func() { config.Version = originalVersion })
	config.Version = "5.14.1"
	update.SetCheckResult(nil, "v6.0.0")

	request := httptest.NewRequest(http.MethodGet, "/api/v2/update/check", nil)
	response := httptest.NewRecorder()
	CheckUpdateHandler(response, request)

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["updateAvailable"] != "true" || body["majorUpdate"] != "true" || body["version"] != "v6.0.0" {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestMajorUpdateRequiresExplicitRequestApproval(t *testing.T) {
	originalVersion := config.Version
	t.Cleanup(func() { config.Version = originalVersion })
	config.Version = "5.14.1"
	update.SetCheckResult(nil, "v6.0.0")

	request := httptest.NewRequest(http.MethodPost, "/api/v2/update/trigger", bytes.NewBufferString(`{"allowUpdate":true,"version":"v6.0.0"}`))
	response := httptest.NewRecorder()
	TriggerUpdateHandler(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	// A rejected request must not leave the updater locked.
	if !update.TryStartOperation() {
		t.Fatal("update operation remained locked after rejected approval")
	}
	update.FinishOperation()
}

func TestMalformedUpdateRequestIsRejected(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v2/update/trigger", bytes.NewBufferString(`{`))
	response := httptest.NewRecorder()
	TriggerUpdateHandler(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestChangedUpdateNeedsFreshApproval(t *testing.T) {
	originalVersion := config.Version
	t.Cleanup(func() { config.Version = originalVersion })
	config.Version = "5.14.1"
	update.SetCheckResult(nil, "v6.0.1")

	request := httptest.NewRequest(http.MethodPost, "/api/v2/update/trigger", bytes.NewBufferString(`{"allowUpdate":true,"allowMajorUpdate":true,"version":"v6.0.0"}`))
	response := httptest.NewRecorder()
	TriggerUpdateHandler(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !update.TryStartOperation() {
		t.Fatal("update operation remained locked after stale approval")
	}
	update.FinishOperation()
}
