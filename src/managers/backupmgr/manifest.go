package backupmgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"time"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/logger"
)

const manifestFilename = "backup-meta-manifest.ssui"
const manifestVersion = 1
const analysisVersion = 1
const maxManifestSize = 64 << 20

// Only aggregate values belong here. Never retain world XML or individual things.
type backupRecord struct {
	Size         int64        `json:"size"`
	ModifiedNS   int64        `json:"mtime"`
	Analysis     SaveAnalysis `json:"analysis"`
	SummaryReady bool         `json:"summaryReady,omitempty"`
	ScanVersion  int          `json:"scanVersion,omitempty"`
	ScanError    string       `json:"scanError,omitempty"`
}

type backupManifest struct {
	Version int                     `json:"version"`
	Backups map[string]backupRecord `json:"backups"`
	Handled map[string]bool         `json:"handled,omitempty"`
	Retired map[string]bool         `json:"retired,omitempty"`
}

func recordIdentity(record backupRecord) saveIdentity {
	return saveIdentity{size: record.Size, modifiedNS: record.ModifiedNS}
}

func analysisReady(record backupRecord) bool {
	return record.SummaryReady && record.ScanVersion == analysisVersion && record.ScanError == ""
}

func backupName(root, path string) (string, error) {
	name, err := filepath.Rel(root, path)
	if err != nil || !filepath.IsLocal(name) || name == "." {
		return "", fmt.Errorf("backup is outside its folder: %s", path)
	}
	return filepath.ToSlash(name), nil
}

var errInvalidManifest = errors.New("invalid backup manifest")

func readManifest(path string) (backupManifest, error) {
	var manifest backupManifest
	file, err := os.Open(path)
	if err != nil {
		return manifest, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxManifestSize+1))
	// Storage errors are not evidence of corrupt JSON. Leave the file in place.
	if err != nil {
		return manifest, err
	}
	if len(data) > maxManifestSize {
		return manifest, fmt.Errorf("%w: exceeds %d bytes", errInvalidManifest, maxManifestSize)
	}
	err = json.Unmarshal(data, &manifest)
	// A newer format may also change the types of existing fields.
	if manifest.Version > 0 && manifest.Version != manifestVersion {
		return backupManifest{Version: manifest.Version}, nil
	}
	if err != nil {
		return manifest, fmt.Errorf("%w: %v", errInvalidManifest, err)
	}
	if manifest.Version < 1 || manifest.Backups == nil {
		return manifest, fmt.Errorf("%w: missing version or backup inventory", errInvalidManifest)
	}
	return manifest, nil
}

// Startup reconciles names and file identities, not the contents of thousands of ZIPs.
func loadInventory(m *BackupManager) error {
	m.loadMu.Lock()
	defer m.loadMu.Unlock()
	m.stateMu.RLock()
	loaded := m.loaded
	m.stateMu.RUnlock()
	if loaded {
		return nil
	}
	started := time.Now()
	logger.Backup.Debugf("%s Loading backup inventory from %q", m.config.Identifier, m.config.SafeBackupDir)
	files, err := scanBackupFiles(m.ctx, m.config.SafeBackupDir)
	if err != nil {
		return err
	}
	path := filepath.Join(m.config.SafeBackupDir, manifestFilename)
	manifest, err := readManifest(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, errInvalidManifest) {
		return fmt.Errorf("read backup manifest: %w", err)
	}
	if errors.Is(err, errInvalidManifest) {
		// Preserve the broken file for diagnosis. A failed rename must not lead to an overwrite.
		quarantine := path + ".invalid-" + time.Now().UTC().Format("20060102T150405.000000000")
		if renameErr := os.Rename(path, quarantine); renameErr != nil {
			return fmt.Errorf("read manifest: %v; preserve invalid manifest: %w", err, renameErr)
		}
		logger.Backup.Warnf("Preserved unreadable backup manifest at %s: %v", quarantine, err)
		manifest = backupManifest{}
	}
	if manifest.Version != 0 && manifest.Version != manifestVersion {
		return fmt.Errorf("unsupported backup manifest version %d; leaving it untouched", manifest.Version)
	}
	records := make(map[string]backupRecord, len(files))
	analyzed := 0
	for path, identity := range files {
		name, err := backupName(m.config.SafeBackupDir, path)
		if err != nil {
			return err
		}
		record, exists := manifest.Backups[name]
		if !exists || recordIdentity(record) != identity {
			record = backupRecord{Size: identity.size, ModifiedNS: identity.modifiedNS}
		}
		records[name] = record
		if analysisReady(record) {
			analyzed++
		}
	}
	handled := make(map[string]bool)
	for name, done := range manifest.Handled {
		if done && filepath.IsLocal(filepath.FromSlash(name)) && isValidBackupFile(name) {
			handled[name] = true
		}
	}
	retired := make(map[string]bool)
	for name, done := range manifest.Retired {
		if done && filepath.IsLocal(filepath.FromSlash(name)) && isValidBackupFile(name) {
			retired[name] = true
		}
	}
	m.stateMu.Lock()
	m.records = records
	for name := range m.handled {
		handled[name] = true
	}
	for name := range m.retired {
		retired[name] = true
	}
	// A manifest entry means the archive was expected to exist. Unlike a
	// retention deletion, its disappearance must not suppress another copy.
	for name := range manifest.Backups {
		if _, exists := records[name]; !exists && !retired[name] {
			delete(handled, name)
		}
	}
	m.handled = handled
	m.retired = retired
	m.loaded = true
	if manifest.Version != manifestVersion || !maps.Equal(records, manifest.Backups) ||
		!maps.Equal(handled, manifest.Handled) || !maps.Equal(retired, manifest.Retired) {
		m.revision++
	}
	m.stateMu.Unlock()
	logger.Backup.Debugf("%s Backup inventory loaded: %d archives, %d analyses reused, %d awaiting analysis (%s)", m.config.Identifier, len(files), analyzed, len(files)-analyzed, time.Since(started).Round(time.Millisecond))
	return nil
}

// A single writer batches backfill updates. The old file remains intact until rename.
func saveManifest(m *BackupManager) error {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	m.stateMu.RLock()
	if !m.loaded || m.revision == m.savedRevision {
		m.stateMu.RUnlock()
		return nil
	}
	revision := m.revision
	manifest := backupManifest{Version: manifestVersion, Backups: maps.Clone(m.records), Handled: maps.Clone(m.handled), Retired: maps.Clone(m.retired)}
	m.stateMu.RUnlock()
	file, err := os.CreateTemp(m.config.SafeBackupDir, ".backup-manifest-*.tmp")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	err = json.NewEncoder(file).Encode(manifest)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(temp, filepath.Join(m.config.SafeBackupDir, manifestFilename)); err != nil {
		return err
	}
	m.stateMu.Lock()
	m.savedRevision = revision
	m.stateMu.Unlock()
	logger.Backup.Debugf("%s Backup manifest saved: %d archives, revision %d", m.config.Identifier, len(manifest.Backups), revision)
	return nil
}

func persistInventory(m *BackupManager) {
	defer m.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := saveManifest(m); err != nil {
				logger.Backup.Warnf("Could not save backup manifest: %v", err)
			}
		}
	}
}
