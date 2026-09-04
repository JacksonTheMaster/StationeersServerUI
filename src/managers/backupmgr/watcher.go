package backupmgr

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/logger"
)

// scanBackupFiles takes a metadata-only snapshot. Do not interpret an incomplete
// scan (for example a disconnected mount) as files having disappeared.
func scanBackupFiles(ctx context.Context, root string) (map[string]saveIdentity, error) {
	files := make(map[string]saveIdentity)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err != nil {
			return err
		}
		if entry.IsDir() || !isValidBackupFile(entry.Name()) || entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		identity, err := identifySave(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil // A file can disappear between enumeration and stat.
		}
		if err != nil {
			return err
		}
		files[path] = identity
		return nil
	})
	return files, err
}

type saveObservation struct {
	identity saveIdentity
	since    time.Time
}

func watchBackups(m *BackupManager) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.config.WaitTime)
	defer ticker.Stop()
	for {
		if err := pollBackups(m, time.Now()); err != nil && m.ctx.Err() == nil && !errors.Is(err, os.ErrNotExist) {
			logger.Backup.Warnf("%s Autosave scan failed: %v", m.config.Identifier, err)
		}
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// An unsuccessful directory read must not erase observations or processed names.
func pollBackups(m *BackupManager, now time.Time) error {
	files, err := scanBackupFiles(m.ctx, m.config.BackupDir)
	if err != nil {
		return err
	}
	current := make(map[string]saveIdentity, len(files))
	for path, identity := range files {
		name, err := backupName(m.config.BackupDir, path)
		if err != nil {
			return err
		}
		current[name] = identity
	}
	m.stateMu.Lock()
	for name := range m.handled {
		if _, exists := current[name]; !exists {
			delete(m.handled, name)
			m.revision++
		}
	}
	for name := range m.observed {
		if _, exists := current[name]; !exists {
			delete(m.observed, name)
			delete(m.pending, name)
		}
	}
	for name, identity := range current {
		_, archived := m.records[name]
		if archived && !m.handled[name] {
			m.handled[name] = true
			m.revision++
		}
		if archived || m.handled[name] {
			delete(m.observed, name)
			delete(m.pending, name)
			continue
		}
		previous, exists := m.observed[name]
		if !exists || previous.identity != identity {
			m.observed[name] = saveObservation{identity: identity, since: now}
			delete(m.pending, name)
			continue
		}
		if now.Sub(previous.since) >= m.config.WaitTime {
			m.pending[name] = identity
		}
	}
	m.stateMu.Unlock()
	select {
	case m.wake <- struct{}{}:
	default:
	}
	return nil
}

// Called after old.Shutdown. Only identical source and destination folders share state.
func inheritDetectorState(next, old *BackupManager) {
	if old == nil || filepath.Clean(next.config.BackupDir) != filepath.Clean(old.config.BackupDir) ||
		filepath.Clean(next.config.SafeBackupDir) != filepath.Clean(old.config.SafeBackupDir) {
		return
	}
	old.stateMu.RLock()
	defer old.stateMu.RUnlock()
	for name, observation := range old.observed {
		next.observed[name] = observation
	}
	for name, handled := range old.handled {
		next.handled[name] = handled
	}
}

// Validation reads both XML members to EOF, checking XML structure and ZIP CRCs.
// This establishes that the observed archive is complete, not that another
// process can never write to the source again. Size/mtime changes invalidate it.
func validateBackupSave(ctx context.Context, path string, expected saveIdentity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	archive, stat, err := openSaveArchive(path)
	if err != nil {
		return err
	}
	defer archive.Close()
	if (saveIdentity{size: stat.Size(), modifiedNS: stat.ModTime().UnixNano()}) != expected {
		return fmt.Errorf("save changed before validation")
	}
	for _, member := range []struct {
		name  string
		root  string
		limit uint64
	}{
		{worldMetaFilename, "WorldMetaData", maxWorldMetaSize},
		{worldFilename, "WorldData", maxWorldSize},
	} {
		var file *zip.File
		for _, candidate := range archive.File {
			if candidate.Name == member.name {
				if file != nil {
					return fmt.Errorf("duplicate %s", member.name)
				}
				file = candidate
			}
		}
		if file == nil || file.UncompressedSize64 > member.limit {
			return fmt.Errorf("missing or oversized %s", member.name)
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		err = validateSaveXML(ctx, io.LimitReader(reader, int64(member.limit)+1), member.root)
		closeErr := reader.Close()
		if err != nil {
			return fmt.Errorf("validate %s: %w", member.name, err)
		}
		if closeErr != nil {
			return closeErr
		}
	}
	after, err := identifySave(path)
	if err != nil {
		return err
	}
	if after != expected {
		return fmt.Errorf("save changed during validation")
	}
	return ctx.Err()
}

func validateSaveXML(ctx context.Context, reader io.Reader, root string) error {
	decoder := xml.NewDecoder(&saveXMLReader{reader: &contextReader{ctx: ctx, reader: reader}})
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			if roots != 1 || depth != 0 {
				return fmt.Errorf("missing or incomplete %s root", root)
			}
			return nil
		}
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots != 1 || token.Name.Local != root {
					return fmt.Errorf("expected one %s root", root)
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.TrimSpace(token)) != 0 {
				return fmt.Errorf("text outside %s root", root)
			}
		}
	}
}
