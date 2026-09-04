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

// watchBackups preserves the startup baseline, then retries new/changed saves
// until validation and the existing copy operation succeed. One goroutine owns
// this state; there are no per-file timers or background copy jobs.
func (m *BackupManager) watchBackups(identifier string, handled map[string]saveIdentity) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.config.WaitTime)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := pollBackups(m.ctx, m.config.BackupDir, handled, m.handleNewBackup); err != nil && m.ctx.Err() == nil {
				logger.Backup.Warnf("%s Autosave scan failed: %v", identifier, err)
			}
		}
	}
}

func pollBackups(ctx context.Context, root string, handled map[string]saveIdentity, copyBackup func(string) error) error {
	files, err := scanBackupFiles(ctx, root)
	if err != nil {
		return err
	}
	for path := range handled {
		if _, exists := files[path]; !exists {
			delete(handled, path)
		}
	}
	for path, identity := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if previous, ok := handled[path]; ok && previous == identity {
			continue
		}
		if err := validateBackupSave(ctx, path, identity); err != nil {
			logger.Backup.Debugf("Autosave not ready, will retry %s: %v", path, err)
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := copyBackup(path); err != nil {
			logger.Backup.Warnf("Autosave copy failed, will retry %s: %v", path, err)
			continue
		}
		handled[path] = identity
	}
	return nil
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
	decoder := xml.NewDecoder(&contextReader{ctx: ctx, reader: reader})
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
