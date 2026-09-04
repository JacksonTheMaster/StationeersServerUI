package backupmgr

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

// Use the logger's dashboard sink so these tests exercise level filtering and
// actual messages without adding a second logging interface to the manager.
func captureBackupLogs(t *testing.T, level int, run func(string)) string {
	t.Helper()
	identifier := "[" + t.Name() + "]"
	config.ConfigMu.Lock()
	oldLevel, oldFilters, oldFile := config.LogLevel, config.SubsystemFilters, config.CreateSSUILogFile
	config.LogLevel, config.SubsystemFilters, config.CreateSSUILogFile = level, nil, false
	config.ConfigMu.Unlock()
	var mu sync.Mutex
	var messages strings.Builder
	flushed := make(chan struct{}, 1)
	marker := identifier + " log capture complete"
	logger.RegisterDashboardHooks(func() bool { return true }, func(line string) {
		if !strings.Contains(line, identifier) {
			return
		}
		mu.Lock()
		messages.WriteString(line)
		mu.Unlock()
		if strings.Contains(line, marker) {
			flushed <- struct{}{}
		}
	})
	t.Cleanup(func() {
		logger.RegisterDashboardHooks(nil, nil)
		config.ConfigMu.Lock()
		config.LogLevel, config.SubsystemFilters, config.CreateSSUILogFile = oldLevel, oldFilters, oldFile
		config.ConfigMu.Unlock()
	})
	run(identifier)
	// The logger is asynchronous. A marker in the same queue gives us a barrier.
	logger.Backup.Error(marker)
	select {
	case <-flushed:
	case <-time.After(5 * time.Second):
		t.Fatal("logger did not drain")
	}
	mu.Lock()
	defer mu.Unlock()
	return messages.String()
}

func TestBackupHandlingLogs(t *testing.T) {
	for _, level := range []int{logger.DEBUG, logger.INFO} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			logs := captureBackupLogs(t, level, func(identifier string) {
				m := NewBackupManager(BackupConfig{Identifier: identifier, BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()})
				defer m.Shutdown()
				name := "040926_112440_auto.save"
				if err := copyFile(analysisFixture(t), filepath.Join(m.config.BackupDir, name)); err != nil {
					t.Fatal(err)
				}
				if err := loadInventory(m); err != nil {
					t.Fatal(err)
				}
				if err := pollBackups(m); err != nil {
					t.Fatal(err)
				}
				if err := pollBackupsAt(m, time.Now().Add(m.config.WaitTime)); err != nil {
					t.Fatal(err)
				}
				ready, identity := nextAutosave(m)
				if ready != name {
					t.Fatalf("expected ready backup %q, got %q", name, ready)
				}
				for range 2 {
					if err := handleBackup(m, name, identity); err != nil {
						t.Fatal(err)
					}
				}
				// A rejected source must never produce a success message.
				if err := handleBackup(m, "missing.save", identity); err == nil {
					t.Fatal("missing source was accepted")
				}
			})
			if strings.Count(logs, "Backup handled:") != 1 || !strings.Contains(logs, "copied and analyzed") {
				t.Fatalf("expected exactly one handling success:\n%s", logs)
			}
			if !strings.Contains(logs, "[BACKUP/INFO]") {
				t.Fatalf("handling success is not info-level:\n%s", logs)
			}
			if level == logger.INFO {
				if strings.Contains(logs, "[BACKUP/DEBUG]") {
					t.Fatalf("debug output leaked into info level:\n%s", logs)
				}
				return
			}
			for _, expected := range []string{"Backup inventory loaded:", "Autosave detected:", "Autosave poll:", "Autosave stable:", "validating source", "Copying backup", "Scanning copied backup", "Backup manifest saved:"} {
				if !strings.Contains(logs, expected) {
					t.Errorf("missing debug step %q:\n%s", expected, logs)
				}
			}
		})
	}
}

func TestBackupBackfillLogsOnlyActualAnalysis(t *testing.T) {
	logs := captureBackupLogs(t, logger.DEBUG, func(identifier string) {
		path := analysisFixture(t)
		m := NewBackupManager(BackupConfig{Identifier: identifier, SafeBackupDir: filepath.Dir(path)})
		defer m.Shutdown()
		if err := loadInventory(m); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			if _, err := analyzeBackup(context.Background(), m, filepath.Base(path)); err != nil {
				t.Fatal(err)
			}
		}
	})
	if strings.Count(logs, "Analyzing archived backup") != 1 || strings.Count(logs, "Archive analysis complete:") != 1 {
		t.Fatalf("expected one analysis, with no log spam for cached results:\n%s", logs)
	}
	if strings.Contains(logs, "Backup handled:") {
		t.Fatalf("backfill claimed to archive a new backup:\n%s", logs)
	}
}

func TestBackupLogsPendingDeepAnalysis(t *testing.T) {
	logs := captureBackupLogs(t, logger.INFO, func(identifier string) {
		// Valid XML, but a tracked value too large for the deep scanner.
		path := writeAnalysisSave(t, "<WorldMetaData/>", "<WorldData><Things><ThingSaveData><PrefabName>"+strings.Repeat("x", 128*1024)+"</PrefabName></ThingSaveData></Things></WorldData>")
		m := NewBackupManager(BackupConfig{Identifier: identifier, BackupDir: filepath.Dir(path), SafeBackupDir: t.TempDir()})
		defer m.Shutdown()
		if err := m.handleNewBackup(path); err != nil {
			t.Fatal(err)
		}
		if analysisReady(m.records[filepath.Base(path)]) {
			t.Fatal("expected deep analysis to remain pending")
		}
	})
	if !strings.Contains(logs, "Backup archived:") || !strings.Contains(logs, "deep analysis pending") || strings.Contains(logs, "Backup handled:") {
		t.Fatalf("pending analysis was not distinguished from complete handling:\n%s", logs)
	}
}
