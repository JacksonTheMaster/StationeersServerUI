package backupmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupNamesRejectAliases(t *testing.T) {
	for _, name := range []string{"", "0", "675", "../a.save", "/a.save", "./a.save", "a/../b.save", "a//b.save", `a\b.save`, `C:\a.save`, "a:b.save", "a\x00.save", "a\n.save", "a\x7f.save", "a.zip", "a.save/"} {
		t.Run(name, func(t *testing.T) {
			m := NewBackupManager(BackupConfig{SafeBackupDir: t.TempDir()})
			if err := CheckBackupAvailable(m, name); !errors.Is(err, ErrInvalidBackupName) {
				t.Fatalf("availability accepted %q: %v", name, err)
			}
			if _, err := m.GetBackupFileData(name); !errors.Is(err, ErrInvalidBackupName) {
				t.Fatalf("download accepted %q: %v", name, err)
			}
			if _, err := m.AnalyzeBackup(context.Background(), name); !errors.Is(err, ErrInvalidBackupName) {
				t.Fatalf("analysis accepted %q: %v", name, err)
			}
			if err := m.RestoreBackup(name); !errors.Is(err, ErrInvalidBackupName) {
				t.Fatalf("restore accepted %q: %v", name, err)
			}
		})
	}
}

func TestBackupSelectionSurvivesReorderingAndRemoval(t *testing.T) {
	safe := t.TempDir()
	names := []string{"older.save", "nested/shared.save", "other/shared.save"}
	for i, name := range names {
		path := filepath.Join(safe, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, bytes.Repeat([]byte{byte(i)}, i+1), 0600); err != nil {
			t.Fatal(err)
		}
	}
	m := NewBackupManager(BackupConfig{SafeBackupDir: safe})
	if err := loadInventory(m); err != nil {
		t.Fatal(err)
	}
	// Backfill changes sort order when the actual world_meta timestamp arrives.
	for round := 0; round < 2; round++ {
		m.stateMu.Lock()
		for i, name := range names {
			record := m.records[name]
			record.SummaryReady = true
			record.Analysis.SavedAt = time.Unix(int64(i*(1-2*round)), 0)
			m.records[name] = record
		}
		m.stateMu.Unlock()
		list, err := m.ListBackups(0)
		if err != nil || len(list) != 3 || list[0].Name != names[2-2*round] {
			t.Fatalf("list did not reorder: %+v, %v", list, err)
		}
		for i, name := range names {
			file, err := m.GetBackupFileData(name)
			if err != nil || !bytes.Equal(file.Data, bytes.Repeat([]byte{byte(i)}, i+1)) {
				t.Fatalf("%q selected different bytes after reordering: %v", name, err)
			}
		}
	}
	if err := deleteBackup(m, BackupSaveFile{Name: names[0]}); err != nil {
		t.Fatal(err)
	}
	file, err := m.GetBackupFileData(names[1])
	if err != nil || !bytes.Equal(file.Data, []byte{1, 1}) {
		t.Fatalf("retention shifted the selection: %v", err)
	}
	if _, err := m.GetBackupFileData(names[0]); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted selection fell back to another backup: %v", err)
	}
	// An external deletion must also fail while its record still exists in RAM.
	if err := os.Remove(filepath.Join(safe, filepath.FromSlash(names[1]))); err != nil {
		t.Fatal(err)
	}
	if _, err := m.GetBackupFileData(names[1]); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing archive fell back to another backup: %v", err)
	}
	if err := m.RestoreBackup(names[1]); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("restore accepted missing archive: %v", err)
	}
}

func TestBackupHTTPRejectsOldAndAmbiguousSelections(t *testing.T) {
	path := analysisFixture(t)
	h := &HTTPHandler{manager: NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})}
	name := filepath.Base(path)
	for _, query := range []string{"", "index=0", "file=" + url.QueryEscape(path), "name=" + name + "&index=0", "name=" + name + "&name=" + name, "name=../analysis.save", "name=" + name + "&bad=%zz", "name=" + name + ";index=0"} {
		for _, handle := range []http.HandlerFunc{h.AnalyzeBackupHandler, h.RestoreBackupHandler} {
			response := httptest.NewRecorder()
			handle(response, httptest.NewRequest(http.MethodGet, "/?"+query, nil))
			if response.Code != http.StatusBadRequest {
				t.Fatalf("query %q: got %d: %s", query, response.Code, response.Body.String())
			}
		}
	}
	for _, body := range []string{`{}`, `null`, `{"index":0}`, `{"saveFile":"analysis.save"}`, `{"name":"analysis.save","index":0}`, `{"name":0}`, `{"name":"../analysis.save"}`, `{"name":"analysis.save"} {}`, `{"name":"analysis.save"} garbage`} {
		response := httptest.NewRecorder()
		h.DownloadBackupHandler(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %s: got %d: %s", body, response.Code, response.Body.String())
		}
	}
	// Restore must return here, before attempting to stop the game server.
	for _, handle := range []http.HandlerFunc{h.AnalyzeBackupHandler, h.RestoreBackupHandler} {
		response := httptest.NewRecorder()
		handle(response, httptest.NewRequest(http.MethodGet, "/?name=missing.save", nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("missing selection: %d %s", response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	h.DownloadBackupHandler(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"missing.save"}`)))
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing download: %d %s", response.Code, response.Body.String())
	}
}

func TestBackupHTTPNameRoundTrip(t *testing.T) {
	safe := t.TempDir()
	name := "Europa & Mars/ä + # 040926_112440_auto.save"
	path := filepath.Join(safe, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(analysisFixture(t), path); err != nil {
		t.Fatal(err)
	}
	h := &HTTPHandler{manager: NewBackupManager(BackupConfig{SafeBackupDir: safe})}
	response := httptest.NewRecorder()
	h.ListBackupsHandler(response, httptest.NewRequest(http.MethodGet, "/", nil))
	var rows []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || len(rows[0]) != 2 || rows[0]["name"] != name || rows[0]["saveTime"] == nil {
		t.Fatalf("list leaked paths or indices: %#v", rows)
	}
	response = httptest.NewRecorder()
	h.AnalyzeBackupHandler(response, httptest.NewRequest(http.MethodGet, "/?"+url.Values{"name": {name}}.Encode(), nil))
	if response.Code != http.StatusOK {
		t.Fatalf("encoded name: %d %s", response.Code, response.Body.String())
	}
	body, _ := json.Marshal(DownloadBackupRequest{Name: name})
	response = httptest.NewRecorder()
	h.DownloadBackupHandler(response, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)))
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), want) {
		t.Fatalf("name round-trip downloaded wrong bytes: %d", response.Code)
	}
}

func TestBackupSelectionRejectsReplacedArchive(t *testing.T) {
	for _, symlink := range []bool{false, true} {
		t.Run(map[bool]string{false: "changed", true: "symlink outside"}[symlink], func(t *testing.T) {
			path := analysisFixture(t)
			m := NewBackupManager(BackupConfig{SafeBackupDir: filepath.Dir(path)})
			if err := loadInventory(m); err != nil {
				t.Fatal(err)
			}
			if symlink {
				outside := analysisFixture(t)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			} else if err := os.WriteFile(path, []byte("replacement"), 0600); err != nil {
				t.Fatal(err)
			}
			name := filepath.Base(path)
			if err := CheckBackupAvailable(m, name); err == nil {
				t.Fatal("preflight accepted a replacement")
			}
			if _, err := m.GetBackupFileData(name); err == nil {
				t.Fatal("download accepted a replacement")
			}
			if err := m.RestoreBackup(name); err == nil {
				t.Fatal("restore accepted a replacement")
			}
			if _, err := m.AnalyzeBackup(context.Background(), name); err == nil {
				t.Fatal("deep scan opened a replacement")
			}
		})
	}
}
