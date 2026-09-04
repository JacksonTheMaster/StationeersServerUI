package backupmgr

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeBackupSave(t *testing.T, dir, name string, saveTime time.Time) string {
	t.Helper()

	path := filepath.Join(dir, name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	writer := zip.NewWriter(file)
	meta, err := writer.Create("world_meta.xml")
	if err != nil {
		t.Fatal(err)
	}
	filetime := saveTime.UnixNano()/100 + filetimeEpochOffset
	if _, err := fmt.Fprintf(meta, "<WorldMetaData><DateTime>%d</DateTime></WorldMetaData>", filetime); err != nil {
		t.Fatal(err)
	}
	world, err := writer.Create(worldFilename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.Write([]byte("<WorldData />")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	return path
}

func assertFileExists(t *testing.T, path string, want bool) {
	t.Helper()
	_, err := os.Stat(path)
	if want && err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
	if !want && !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat error: %v", path, err)
	}
}

func TestCleanSafeBackupDirAllowsAllRetentionRulesDisabled(t *testing.T) {
	dir := t.TempDir()
	first := writeBackupSave(t, dir, "first.save", time.Now().Add(-time.Hour))
	second := writeBackupSave(t, dir, "second.save", time.Now().Add(-2*time.Hour))
	m := NewBackupManager(BackupConfig{
		SafeBackupDir: dir,
		RetentionPolicy: RetentionPolicy{
			KeepNewestCount: 0,
		},
	})

	primeBackupInventory(t, m)
	if err := m.cleanSafeBackupDir(); err != nil {
		t.Fatal(err)
	}
	assertFileExists(t, first, false)
	assertFileExists(t, second, false)
}

func TestCleanSafeBackupDirKeepsConfiguredRestorePoints(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	newest := writeBackupSave(t, dir, "newest.save", now)
	daily := writeBackupSave(t, dir, "daily.save", startOfCalendarDay(now).Add(-12*time.Hour))
	weekly := writeBackupSave(t, dir, "weekly.save", startOfISOWeek(now).AddDate(0, 0, -7).Add(12*time.Hour))
	monthly := writeBackupSave(t, dir, "monthly.save", time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, now.Location()).AddDate(0, -2, 0))
	expired := writeBackupSave(t, dir, "expired.save", now.AddDate(-1, 0, 0))
	m := NewBackupManager(BackupConfig{
		SafeBackupDir: dir,
		RetentionPolicy: RetentionPolicy{
			KeepNewestCount: 1,
			DailyDays:       2,
			WeeklyWeeks:     3,
			MonthlyMonths:   3,
		},
	})

	primeBackupInventory(t, m)
	if err := m.cleanSafeBackupDir(); err != nil {
		t.Fatal(err)
	}
	assertFileExists(t, newest, true)
	assertFileExists(t, daily, true)
	assertFileExists(t, weekly, true)
	assertFileExists(t, monthly, true)
	assertFileExists(t, expired, false)
}

func TestCleanSafeBackupDirDoesNotKeepDuplicateRestorePoints(t *testing.T) {
	dir := t.TempDir()
	anchor := startOfCalendarDay(time.Now()).Add(12 * time.Hour)
	for i := 0; i < 4; i++ {
		writeBackupSave(t, dir, fmt.Sprintf("same-day-%d.save", i), anchor.Add(-time.Duration(i)*time.Minute))
	}
	m := NewBackupManager(BackupConfig{
		SafeBackupDir: dir,
		RetentionPolicy: RetentionPolicy{
			DailyDays:     1,
			WeeklyWeeks:   1,
			MonthlyMonths: 1,
		},
	})

	primeBackupInventory(t, m)
	if err := m.cleanSafeBackupDir(); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("kept %d backups for one covered day, want 1", len(files))
	}
}

func TestCleanupUsesLocalCalendarForUTCMetadata(t *testing.T) {
	location := time.FixedZone("UTC+2", 2*60*60)
	now := time.Date(2026, time.September, 2, 0, 30, 0, 0, location)
	newestUTC := now.UTC()
	previousLocalDayUTC := time.Date(2026, time.September, 1, 12, 0, 0, 0, location).UTC()

	var daily time.Time
	updateRetentionTrackers(newestUTC.In(location), &daily, new(time.Time), new(time.Time))
	if sameCalendarDay(previousLocalDayUTC.In(location), daily) {
		t.Fatal("UTC metadata collapsed backups from different local calendar days")
	}
}

func TestRetentionWindowsUseCalendarUnits(t *testing.T) {
	now := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

	if !withinDailyWindow(time.Date(2026, time.February, 28, 0, 1, 0, 0, time.UTC), now, 2) {
		t.Fatal("daily window did not include the previous calendar day")
	}
	if withinDailyWindow(time.Date(2026, time.February, 27, 23, 59, 0, 0, time.UTC), now, 2) {
		t.Fatal("daily window included a day outside the configured window")
	}
	if !withinMonthlyWindow(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), now, 3) {
		t.Fatal("monthly window did not include the third calendar month")
	}
	if withinMonthlyWindow(time.Date(2025, time.December, 31, 23, 59, 0, 0, time.UTC), now, 3) {
		t.Fatal("monthly window included a month outside the configured window")
	}

	weekNow := time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
	if !withinWeeklyWindow(time.Date(2026, time.March, 9, 0, 0, 0, 0, time.UTC), weekNow, 2) {
		t.Fatal("weekly window did not include the previous ISO calendar week")
	}
	if withinWeeklyWindow(time.Date(2026, time.March, 8, 23, 59, 0, 0, time.UTC), weekNow, 2) {
		t.Fatal("weekly window included a week outside the configured window")
	}
}

func TestCleanBackupDirOnlyRemovesOldSaveFiles(t *testing.T) {
	dir := t.TempDir()
	oldSave := filepath.Join(dir, "old.save")
	oldOther := filepath.Join(dir, "keep-me.txt")
	recentSave := filepath.Join(dir, "recent.save")
	for _, path := range []string{oldSave, oldOther, recentSave} {
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldSave, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(oldOther, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	m := NewBackupManager(BackupConfig{BackupDir: dir})
	m.handled["old.save"] = true
	if err := m.cleanBackupDir(); err != nil {
		t.Fatal(err)
	}
	assertFileExists(t, oldSave, false)
	assertFileExists(t, oldOther, true)
	assertFileExists(t, recentSave, true)
}

func TestCleanupKeepsAutosavesWhenSafeBackupDirIsUnavailable(t *testing.T) {
	backupDir := t.TempDir()
	oldSave := filepath.Join(backupDir, "old.save")
	if err := os.WriteFile(oldSave, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldSave, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	m := NewBackupManager(BackupConfig{
		BackupDir:     backupDir,
		SafeBackupDir: filepath.Join(t.TempDir(), "missing"),
	})
	if err := m.Cleanup(); err == nil {
		t.Fatal("Cleanup() returned nil with an unavailable safe backup directory")
	}
	assertFileExists(t, oldSave, true)
}

func TestNewBackupManagerUsesFixedScanInterval(t *testing.T) {
	m := NewBackupManager(BackupConfig{})
	if m.config.WaitTime != 45*time.Second {
		t.Fatalf("WaitTime = %s, want 45s", m.config.WaitTime)
	}
}
