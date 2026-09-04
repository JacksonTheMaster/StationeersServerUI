package backupmgr

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// New clients select a stable path. The numeric index remains supported for older clients.
func selectBackup(m *BackupManager, index int, saveFile string) (BackupSaveFile, error) {
	saves, err := m.getBackupSaveFiles()
	if err != nil {
		return BackupSaveFile{}, err
	}
	if saveFile != "" {
		for _, save := range saves {
			if save.SaveFile == saveFile {
				return save, nil
			}
		}
		return BackupSaveFile{}, fmt.Errorf("selected backup is no longer available")
	}
	if index < 0 || index >= len(saves) {
		return BackupSaveFile{}, fmt.Errorf("backup index %d out of range (0-%d)", index, len(saves)-1)
	}
	return saves[index], nil
}

// Unknown archives are visible immediately; metadata arrives in the background.
func (m *BackupManager) getBackupSaveFiles() ([]BackupSaveFile, error) {
	if err := loadInventory(m); err != nil {
		return nil, err
	}
	m.stateMu.RLock()
	saves := make([]BackupSaveFile, 0, len(m.records))
	for name, record := range m.records {
		savedAt := recordTimestamp(name, record)
		saves = append(saves, BackupSaveFile{SaveFile: filepath.Join(m.config.SafeBackupDir, filepath.FromSlash(name)), SaveTime: savedAt, Summary: record.Analysis.SaveSummary, SummaryReady: record.SummaryReady})
	}
	m.stateMu.RUnlock()
	sort.Slice(saves, func(i, j int) bool {
		if saves[i].SaveTime.Equal(saves[j].SaveTime) {
			return saves[i].SaveFile < saves[j].SaveFile
		}
		return saves[i].SaveTime.Before(saves[j].SaveTime)
	})
	for i := range saves {
		saves[i].Index = i
	}
	return saves, nil
}

func recordTimestamp(name string, record backupRecord) time.Time {
	if record.SummaryReady {
		return record.Analysis.SavedAt
	}
	return backupTimestamp(name, record.ModifiedNS)
}

func backupTimestamp(name string, modifiedNS int64) time.Time {
	stamp := strings.TrimSuffix(filepath.Base(name), "_auto.save")
	if parsed, err := time.ParseInLocation("020106_150405", stamp, time.Local); err == nil {
		return parsed
	}
	return time.Unix(0, modifiedNS)
}
