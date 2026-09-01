package web

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupPageTemplateRendersLocalizedWorkspace(t *testing.T) {
	path := filepath.Join("..", "..", "UIMod", "onboard_bundled", "ui", "backups.html")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("backups.html").Parse(string(contents))
	if err != nil {
		t.Fatal(err)
	}

	var rendered bytes.Buffer
	data := backupPageTemplateData{
		UITextBackupManager:            "Backup Manager",
		UITextBackupManagerDescription: "Inspect and restore backups.",
		UITextBackupTechnicalDetails:   "Players & state",
		UITextBackupInfrastructure:     "World infrastructure",
		UITextBackupHistory:            "Backup history",
		UITextBackToDashboard:          "Back to dashboard",
		UITextRefresh:                  "Refresh",
		UITextBackupLast:               "Last",
		UITextBackupAll:                "All backups",
	}
	if err := tmpl.Execute(&rendered, data); err != nil {
		t.Fatal(err)
	}
	output := rendered.String()
	for _, expected := range []string{
		`<h1 id="backup-page-title">Backup Manager</h1>`,
		`<p>Inspect and restore backups.</p>`,
		`data-technical-details="Players &amp; state"`,
		`data-infrastructure="World infrastructure"`,
		`<span>Back to dashboard</span>`,
		`aria-label="Backup history"`,
		`<span aria-hidden="true">↻</span>Refresh`,
		`<option value="20" selected>Last 20</option>`,
		`<option value="">All backups</option>`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("rendered backup workspace does not contain %q", expected)
		}
	}
}
