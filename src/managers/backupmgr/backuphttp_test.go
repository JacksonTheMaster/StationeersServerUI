package backupmgr

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
)

func TestListBackupsSummaryIsOptIn(t *testing.T) {
	path := analysisFixture(t)
	handler := &HTTPHandler{manager: NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})}

	primeBackupInventory(t, handler.manager)

	legacyRequest := httptest.NewRequest(http.MethodGet, "/api/v2/backups?limit=5", nil)
	legacyResponse := httptest.NewRecorder()
	handler.ListBackupsHandler(legacyResponse, legacyRequest)
	if legacyResponse.Code != http.StatusOK {
		t.Fatalf("legacy response status %d: %s", legacyResponse.Code, legacyResponse.Body.String())
	}
	if bytes.Contains(legacyResponse.Body.Bytes(), []byte(`"Summary"`)) {
		t.Fatalf("legacy response unexpectedly contains Summary: %s", legacyResponse.Body.String())
	}
	var legacyRows []map[string]any
	if err := json.NewDecoder(bytes.NewReader(legacyResponse.Body.Bytes())).Decode(&legacyRows); err != nil {
		t.Fatal(err)
	}
	if len(legacyRows) != 1 || len(legacyRows[0]) != 3 {
		t.Fatalf("legacy response shape changed: %#v", legacyRows)
	}
	for _, key := range []string{"Index", "SaveFile", "SaveTime"} {
		if _, exists := legacyRows[0][key]; !exists {
			t.Fatalf("legacy response is missing %s: %#v", key, legacyRows[0])
		}
	}

	summaryRequest := httptest.NewRequest(http.MethodGet, "/api/v2/backups?limit=5&include=summary", nil)
	summaryResponse := httptest.NewRecorder()
	handler.ListBackupsHandler(summaryResponse, summaryRequest)
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("summary response status %d: %s", summaryResponse.Code, summaryResponse.Body.String())
	}
	if !bytes.Contains(summaryResponse.Body.Bytes(), []byte(`"Summary"`)) ||
		!bytes.Contains(summaryResponse.Body.Bytes(), []byte(`"daysPlayed":67`)) {
		t.Fatalf("summary response is missing metadata: %s", summaryResponse.Body.String())
	}
}

func TestBackupHTTPSelectionUsesStableFile(t *testing.T) {
	path := analysisFixture(t)
	m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})
	handler := &HTTPHandler{manager: m}
	query := url.Values{"index": {"999"}, "file": {path}}
	request := httptest.NewRequest(http.MethodGet, "/api/v2/backups/analyze?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	handler.AnalyzeBackupHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("stable selection depended on the old index: %d %s", response.Code, response.Body.String())
	}
	body, err := json.Marshal(DownloadBackupRequest{Index: 999, SaveFile: path})
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	handler.DownloadBackupHandler(response, httptest.NewRequest(http.MethodPost, "/api/v2/backups/download", bytes.NewReader(body)))
	if response.Code != http.StatusOK || !bytes.HasPrefix(response.Body.Bytes(), []byte("PK")) {
		t.Fatalf("stable download: %d", response.Code)
	}
	if _, err := getBackupFileData(m, 0, filepath.Join(filepath.Dir(path), "..", "outside.save")); err == nil {
		t.Fatal("accepted a file outside the inventory")
	}
}

func TestListPendingAnalysisDoesNotInventSummary(t *testing.T) {
	path := analysisFixture(t)
	handler := &HTTPHandler{manager: NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})}
	response := httptest.NewRecorder()
	handler.ListBackupsHandler(response, httptest.NewRequest(http.MethodGet, "/api/v2/backups?include=summary", nil))
	if response.Code != http.StatusOK || bytes.Contains(response.Body.Bytes(), []byte(`"Summary"`)) {
		t.Fatalf("pending archive returned fabricated statistics: %s", response.Body.String())
	}
}

func TestAnalyzeBackupHandler(t *testing.T) {
	path := analysisFixture(t)
	manager := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})
	handler := &HTTPHandler{manager: manager}
	request := httptest.NewRequest(http.MethodGet, "/api/v2/backups/analyze?index=0", nil)
	response := httptest.NewRecorder()

	handler.AnalyzeBackupHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	var analysis SaveAnalysis
	if err := json.NewDecoder(response.Body).Decode(&analysis); err != nil {
		t.Fatal(err)
	}
	if analysis.DaysPlayed != 67 || analysis.Players != 3 || analysis.Furnaces != 2 {
		t.Fatalf("unexpected analysis response: %+v", analysis)
	}
}

func TestAnalyzeBackupHandlerRejectsInvalidIndex(t *testing.T) {
	handler := &HTTPHandler{manager: NewBackupManager(BackupConfig{})}
	request := httptest.NewRequest(http.MethodGet, "/api/v2/backups/analyze?index=nope", nil)
	response := httptest.NewRecorder()

	handler.AnalyzeBackupHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
}
