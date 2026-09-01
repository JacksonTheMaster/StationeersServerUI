package web

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestIndexTemplateRendersBackupSummaryLocalization(t *testing.T) {
	path := filepath.Join("..", "..", "UIMod", "onboard_bundled", "ui", "index.html")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("index.html").Parse(string(contents))
	if err != nil {
		t.Fatal(err)
	}

	var rendered bytes.Buffer
	data := IndexTemplateData{
		UIText_BackupDaysPlayed:  "Days played",
		UIText_BackupGameVersion: "Game version",
	}
	if err := tmpl.Execute(&rendered, data); err != nil {
		t.Fatal(err)
	}
	output := rendered.String()
	if !strings.Contains(output, `data-days-played="Days played"`) ||
		!strings.Contains(output, `data-game-version="Game version"`) {
		t.Fatal("backup summary localization was not rendered into the home panel")
	}
}
