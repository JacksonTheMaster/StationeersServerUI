package backupmgr

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// These tests exercise running managers and real filesystem writes. They are
// opt-in because the real-save and default-interval runs take several minutes.
func requireBackupSoak(t *testing.T) {
	t.Helper()
	if os.Getenv("SSUI_BM_STRESS") != "1" {
		t.Skip("set SSUI_BM_STRESS=1 to run backup integration stress tests")
	}
}

func waitForBackupCondition(t *testing.T, timeout time.Duration, what string, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func backupIsReady(m *BackupManager, name string) bool {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()
	record, exists := m.records[name]
	return exists && analysisReady(record)
}

func startSoakManager(t *testing.T, cfg BackupConfig, old *BackupManager) *BackupManager {
	t.Helper()
	if old != nil {
		old.Shutdown()
	}
	m := NewBackupManager(cfg)
	inheritDetectorState(m, old)
	if err := m.Start("soak"); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestBackupSoakRotationReloadAndHTTP(t *testing.T) {
	requireBackupSoak(t)
	data, err := os.ReadFile(analysisFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir(), WaitTime: 20 * time.Millisecond,
		RetentionPolicy: RetentionPolicy{KeepNewestCount: 1000, CleanupInterval: time.Hour}}
	m := startSoakManager(t, cfg, nil)
	defer func() { m.Shutdown() }()
	handler := &HTTPHandler{manager: m}
	server := httptest.NewServer(http.HandlerFunc(handler.ListBackupsHandler))
	defer server.Close()
	client := &http.Client{Timeout: 2 * time.Second}
	stop, done := make(chan struct{}), make(chan struct{})
	requests := atomic.Int64{}
	failures := make(chan error, 1)
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
			}
			response, err := client.Get(server.URL + "?include=summary")
			if err != nil {
				failures <- err
				return
			}
			var rows []backupListResponse
			err = json.NewDecoder(response.Body).Decode(&rows)
			response.Body.Close()
			if response.StatusCode != http.StatusOK || err != nil {
				failures <- fmt.Errorf("list: status %d, decode %v", response.StatusCode, err)
				return
			}
			seen := make(map[string]bool, len(rows))
			for _, row := range rows {
				if seen[row.SaveFile] || row.Summary != nil && row.Summary.DaysPlayed != 67 {
					failures <- fmt.Errorf("inconsistent list row: %+v", row)
					return
				}
				seen[row.SaveFile] = true
			}
			requests.Add(1)
			time.Sleep(2 * time.Millisecond)
		}
	}()
	defer func() { close(stop); <-done }()

	const count = 60
	names := make([]string, 0, count)
	stamp := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	reloads := 0
	for i := 0; i < count; i++ {
		name := stamp.Add(time.Duration(i) * 5 * time.Minute).Format("020106_150405_auto.save")
		names = append(names, name)
		if i >= 5 {
			if err := os.Remove(filepath.Join(cfg.BackupDir, names[i-5])); err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(cfg.BackupDir, name)
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(data[:len(data)/2]); err != nil {
			file.Close()
			t.Fatal(err)
		}
		// Leave a stable but incomplete ZIP through multiple actual polls.
		time.Sleep(3 * cfg.WaitTime)
		if _, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name)); !os.IsNotExist(err) {
			file.Close()
			t.Fatalf("published an incomplete save: %v", err)
		}
		if i%7 == 3 {
			m = startSoakManager(t, cfg, m)
			handler.mu.Lock()
			handler.manager = m
			handler.mu.Unlock()
			reloads++
		}
		if _, err := file.Write(data[len(data)/2:]); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		waitForBackupCondition(t, 5*time.Second, name, func() bool { return backupIsReady(m, name) })
		if i%13 == 8 {
			// A missing expected copy must be repaired by the running detector.
			if err := os.Remove(filepath.Join(cfg.SafeBackupDir, name)); err != nil {
				t.Fatal(err)
			}
			waitForBackupCondition(t, 5*time.Second, "archive repair", func() bool {
				_, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name))
				return err == nil && backupIsReady(m, name)
			})
		}
		if i == 20 || i == 40 {
			root := cfg.SafeBackupDir
			if i == 40 {
				root = cfg.BackupDir
			}
			// Windows cannot rename a directory while a manifest handle is open.
			// Wait for a gap so the storage outage is actually injected there too.
			waitForBackupCondition(t, 2*time.Second, "storage outage injection", func() bool {
				return os.Rename(root, root+".offline") == nil
			})
			time.Sleep(3 * cfg.WaitTime)
			if err := os.Rename(root+".offline", root); err != nil {
				t.Fatal(err)
			}
		}
		select {
		case err := <-failures:
			t.Fatal(err)
		default:
		}
	}
	// Verify every archive, including originals that the simulated game rotated out.
	for _, name := range names {
		archived, err := os.ReadFile(filepath.Join(cfg.SafeBackupDir, name))
		if err != nil || sha256.Sum256(archived) != sha256.Sum256(data) {
			t.Fatalf("lost or changed archive %s: %v", name, err)
		}
		if !backupIsReady(m, name) {
			t.Fatalf("archive was never analyzed: %s", name)
		}
	}
	// Rebuild a broken manifest using only files, not inherited inventory state.
	m.Shutdown()
	if err := os.WriteFile(filepath.Join(cfg.SafeBackupDir, manifestFilename), []byte("{interrupted"), 0600); err != nil {
		t.Fatal(err)
	}
	m = startSoakManager(t, cfg, nil)
	handler.mu.Lock()
	handler.manager = m
	handler.mu.Unlock()
	waitForBackupCondition(t, 10*time.Second, "complete manifest rebuild", func() bool {
		for _, name := range names {
			if !backupIsReady(m, name) {
				return false
			}
		}
		return true
	})
	// Exercise retention through its public entry point with HTTP reads in flight.
	cfg.RetentionPolicy.KeepNewestCount = 3
	m = startSoakManager(t, cfg, m)
	handler.mu.Lock()
	handler.manager = m
	handler.mu.Unlock()
	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}
	m = startSoakManager(t, cfg, m)
	handler.mu.Lock()
	handler.manager = m
	handler.mu.Unlock()
	time.Sleep(5 * cfg.WaitTime)
	saves, err := m.ListBackups(0)
	if err != nil || len(saves) != 3 {
		t.Fatalf("retention resurrected archives: %d, %v", len(saves), err)
	}
	m.Shutdown()
	manifest, err := readManifest(filepath.Join(cfg.SafeBackupDir, manifestFilename))
	if err != nil || len(manifest.Backups) != 3 {
		t.Fatalf("final manifest disagrees with retained files: %v", err)
	}
	for name := range manifest.Backups {
		if _, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name)); err != nil {
			t.Fatal(err)
		}
	}
	if len(m.handled) > 5 || len(m.retired) > 5 || len(m.observed) > 5 {
		t.Fatal("source state grew beyond the five live autosaves")
	}
	select {
	case err := <-failures:
		t.Fatal(err)
	default:
	}
	t.Logf("%d saves, %d reloads during partial writes, %d HTTP reads, repairs, two storage outages, manifest rebuild and retention passed", count, reloads, requests.Load())
}

func TestBackupProcessHelper(t *testing.T) {
	root := os.Getenv("SSUI_BM_CHILD_ROOT")
	if root == "" {
		t.Skip("subprocess helper")
	}
	m := NewBackupManager(BackupConfig{BackupDir: filepath.Join(root, "source"), SafeBackupDir: filepath.Join(root, "safe"), WaitTime: 20 * time.Millisecond})
	if err := m.Start("crash-child"); err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()
	// Only the parent test terminates this dedicated process.
	<-m.ctx.Done()
}

func TestBackupSoakBackfill4000WithLiveAutosaves(t *testing.T) {
	requireBackupSoak(t)
	data, err := os.ReadFile(analysisFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir(), WaitTime: 20 * time.Millisecond,
		RetentionPolicy: RetentionPolicy{KeepNewestCount: 5000, CleanupInterval: time.Hour}}
	for i := 0; i < 4000; i++ {
		if err := os.WriteFile(filepath.Join(cfg.SafeBackupDir, fmt.Sprintf("old-%04d.save", i)), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	started := time.Now()
	m := startSoakManager(t, cfg, nil)
	defer m.Shutdown()
	var names []string
	const count = 30
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("live-%04d.save", i)
		names = append(names, name)
		if i >= 5 {
			if err := os.Remove(filepath.Join(cfg.BackupDir, names[i-5])); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(cfg.BackupDir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
		// The producer never waits for handling; the game rotates each source
		// after five intervals while the worker is also draining 4000 archives.
		time.Sleep(150 * time.Millisecond)
		listed := time.Now()
		if _, err := m.ListBackups(0); err != nil {
			t.Fatal(err)
		}
		if time.Since(listed) > time.Second {
			t.Fatal("list blocked behind the analysis backlog")
		}
	}
	waitForBackupCondition(t, 90*time.Second, "4000 backfill items and live saves", func() bool {
		m.stateMu.RLock()
		defer m.stateMu.RUnlock()
		if len(m.records) != 4000+count {
			return false
		}
		for _, record := range m.records {
			if !analysisReady(record) {
				return false
			}
		}
		return true
	})
	m.Shutdown()
	manifest, err := readManifest(filepath.Join(cfg.SafeBackupDir, manifestFilename))
	if err != nil || len(manifest.Backups) != 4000+count {
		t.Fatalf("incomplete final manifest: %v", err)
	}
	for _, name := range names {
		archived, err := os.ReadFile(filepath.Join(cfg.SafeBackupDir, name))
		if err != nil || sha256.Sum256(archived) != sha256.Sum256(data) {
			t.Fatalf("live save lost during backfill: %s: %v", name, err)
		}
	}
	t.Logf("4000 archives backfilled, %d independently rotating live saves preserved in %s", count, time.Since(started))
}

func TestBackupSoakReloadDuringHandling(t *testing.T) {
	requireBackupSoak(t)
	path := os.Getenv("SSUI_ANALYSIS_TEST_SAVE")
	if path == "" {
		t.Skip("set SSUI_ANALYSIS_TEST_SAVE to a large real save")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir(), WaitTime: 20 * time.Millisecond}
	name := "040926_120000_auto.save"
	source := filepath.Join(cfg.BackupDir, name)
	if err := os.WriteFile(source, data, 0600); err != nil {
		t.Fatal(err)
	}
	m := startSoakManager(t, cfg, nil)
	defer func() { m.Shutdown() }()
	waitForBackupCondition(t, 60*time.Second, "complete temporary copy before publication", func() bool {
		paths, _ := filepath.Glob(filepath.Join(cfg.SafeBackupDir, ".backup-*.tmp"))
		for _, temp := range paths {
			stat, err := os.Stat(temp)
			if err == nil && stat.Size() == int64(len(data)) {
				return true
			}
		}
		return false
	})
	if _, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name)); !os.IsNotExist(err) {
		t.Fatalf("missed handling reload window: %v", err)
	}
	// Simulate game rotation once the full copy exists. Windows may briefly
	// refuse removal until the copy's input handle has closed.
	waitForBackupCondition(t, 2*time.Second, "source rotation", func() bool { return os.Remove(source) == nil })
	started := time.Now()
	m.Shutdown()
	elapsed := time.Since(started)
	archived, err := os.ReadFile(filepath.Join(cfg.SafeBackupDir, name))
	if err != nil || sha256.Sum256(archived) != sha256.Sum256(data) {
		t.Fatalf("reload lost the only completed copy: %v", err)
	}
	m = startSoakManager(t, cfg, m)
	waitForBackupCondition(t, 90*time.Second, "resumed deep analysis", func() bool { return backupIsReady(m, name) })
	t.Logf("reload during handling finished in %s; rotated source was preserved and analysis resumed", elapsed)
}

func TestBackupSoakManifestWriteFailure(t *testing.T) {
	requireBackupSoak(t)
	cfg := BackupConfig{BackupDir: t.TempDir(), SafeBackupDir: t.TempDir(), WaitTime: 20 * time.Millisecond}
	m := startSoakManager(t, cfg, nil)
	defer m.Shutdown()
	manifestPath := filepath.Join(cfg.SafeBackupDir, manifestFilename)
	// Block publication of the manifest, not the archive. A writable archive
	// must survive this failure, and retention must not delete it meanwhile.
	if err := os.Mkdir(manifestPath, 0755); err != nil {
		t.Fatal(err)
	}
	name := "040926_120000_auto.save"
	if err := copyFile(analysisFixture(t), filepath.Join(cfg.BackupDir, name)); err != nil {
		t.Fatal(err)
	}
	waitForBackupCondition(t, 5*time.Second, "handling with unwritable manifest", func() bool { return backupIsReady(m, name) })
	if err := m.Cleanup(); err == nil {
		t.Fatal("retention ignored a manifest write failure")
	}
	if _, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name)); err != nil {
		t.Fatalf("lost archive when its manifest could not be written: %v", err)
	}
	if err := os.Remove(manifestPath); err != nil {
		t.Fatal(err)
	}
	waitForBackupCondition(t, 5*time.Second, "automatic manifest write retry", func() bool {
		manifest, err := readManifest(manifestPath)
		return err == nil && analysisReady(manifest.Backups[name])
	})
	t.Log("handling survived a manifest write failure; retention waited and the background writer recovered")
}

func TestBackupSoakProcessCrash(t *testing.T) {
	requireBackupSoak(t)
	realSave := os.Getenv("SSUI_ANALYSIS_TEST_SAVE")
	if realSave == "" {
		t.Skip("set SSUI_ANALYSIS_TEST_SAVE to a large real save for crash testing")
	}
	data, err := os.ReadFile(realSave)
	if err != nil {
		t.Fatal(err)
	}
	for _, stage := range []string{"temporary copy", "published archive"} {
		t.Run(stage, func(t *testing.T) {
			root := t.TempDir()
			cfg := BackupConfig{BackupDir: filepath.Join(root, "source"), SafeBackupDir: filepath.Join(root, "safe"), WaitTime: 20 * time.Millisecond}
			for _, dir := range []string{cfg.BackupDir, cfg.SafeBackupDir} {
				if err := os.Mkdir(dir, 0755); err != nil {
					t.Fatal(err)
				}
			}
			name := "040926_120000_auto.save"
			if err := os.WriteFile(filepath.Join(cfg.BackupDir, name), data, 0600); err != nil {
				t.Fatal(err)
			}
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			command := exec.Command(executable, "-test.run=^TestBackupProcessHelper$", "-test.timeout=5m")
			command.Env = append(os.Environ(), "SSUI_BM_CHILD_ROOT="+root)
			var output bytes.Buffer
			command.Stdout, command.Stderr = &output, &output
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if command.ProcessState == nil {
					_ = command.Process.Kill()
					_ = command.Wait()
				}
			}()
			waitForBackupCondition(t, 90*time.Second, stage, func() bool {
				if stage == "published archive" {
					_, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name))
					return err == nil
				}
				temps, _ := filepath.Glob(filepath.Join(cfg.SafeBackupDir, ".backup-*.tmp"))
				return len(temps) != 0
			})
			if err := command.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err := command.Wait(); err == nil {
				t.Fatal("child was not interrupted")
			}
			if stage == "temporary copy" {
				if _, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name)); !os.IsNotExist(err) {
					t.Fatalf("missed the pre-publication crash window: %v", err)
				}
			}
			m := startSoakManager(t, cfg, nil)
			defer m.Shutdown()
			waitForBackupCondition(t, 90*time.Second, "recovery after process kill", func() bool { return backupIsReady(m, name) })
			m.Shutdown()
			archived, err := os.ReadFile(filepath.Join(cfg.SafeBackupDir, name))
			if err != nil || sha256.Sum256(archived) != sha256.Sum256(data) {
				t.Fatalf("crash recovery changed or lost the save: %v", err)
			}
			manifest, err := readManifest(filepath.Join(cfg.SafeBackupDir, manifestFilename))
			if err != nil || len(manifest.Backups) != 1 || !analysisReady(manifest.Backups[name]) {
				t.Fatalf("crash recovery left inconsistent metadata: %v", err)
			}
			t.Logf("killed child PID %d during %s; archive and metadata recovered", command.Process.Pid, stage)
		})
	}
}

func TestBackupSoakDefaultIntervalAndRestore(t *testing.T) {
	requireBackupSoak(t)
	path := os.Getenv("SSUI_ANALYSIS_TEST_SAVE")
	if path == "" {
		t.Skip("set SSUI_ANALYSIS_TEST_SAVE for the 45-second real-save test")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before := sha256.Sum256(data)
	t.Chdir(t.TempDir()) // Restore writes a HEAD save relative to the working directory.
	cfg := BackupConfig{WorldName: "soak-restored", BackupDir: t.TempDir(), SafeBackupDir: t.TempDir()}
	name := "040926_120000_auto.save"
	source := filepath.Join(cfg.BackupDir, name)
	if err := os.WriteFile(source, data, 0600); err != nil {
		t.Fatal(err)
	}
	m := startSoakManager(t, cfg, nil)
	defer func() { m.Shutdown() }()
	var observed time.Time
	waitForBackupCondition(t, 5*time.Second, "first observation", func() bool {
		m.stateMu.RLock()
		defer m.stateMu.RUnlock()
		entry, exists := m.observed[name]
		observed = entry.since
		return exists
	})
	// Reload during the real wait, without changing the production interval.
	start := time.Now()
	m = startSoakManager(t, cfg, m)
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("reload waited for polling: %s", elapsed)
	}
	if _, err := os.Stat(filepath.Join(cfg.SafeBackupDir, name)); !os.IsNotExist(err) {
		t.Fatalf("archive appeared before the second observation: %v", err)
	}
	waitForBackupCondition(t, 150*time.Second, "default-interval handling", func() bool { return backupIsReady(m, name) })
	if time.Since(observed) < defaultWaitTime {
		t.Fatal("handling skipped the real 45-second stability interval")
	}
	download, err := m.GetBackupFileData(0)
	if err != nil || sha256.Sum256(download.Data) != before {
		t.Fatalf("download differs from original: %v", err)
	}
	if err := m.RestoreBackup(0); err != nil {
		t.Fatal(err)
	}
	restored := filepath.Join("saves", cfg.WorldName, cfg.WorldName+".save")
	identity, err := identifySave(restored)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateBackupSave(context.Background(), restored, identity); err != nil {
		t.Fatalf("restored archive failed structural validation: %v", err)
	}
	result, err := AnalyzeSave(context.Background(), restored)
	if err != nil {
		t.Fatal(err)
	}
	m.stateMu.RLock()
	original := m.records[name].Analysis
	m.stateMu.RUnlock()
	if result.Players != original.Players || result.Furnaces != original.Furnaces || result.Things != original.Things || result.DaysPlayed != original.DaysPlayed {
		t.Fatalf("restore changed world statistics: %+v versus %+v", result, original)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.New()
	_, hashErr := io.Copy(digest, file)
	file.Close()
	if hashErr != nil || !bytes.Equal(digest.Sum(nil), before[:]) {
		t.Fatal("the original test fixture changed")
	}
	t.Logf("real 45-second poll, reload, handling, download and restore passed: %d things, %d players", result.Things, result.Players)
}
