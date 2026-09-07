package web

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestDiscordHubConfigTemplate(t *testing.T) {
	root := filepath.Join("..", "..", "SSUI", "onboard_bundled")
	tmpl, err := template.ParseFiles(filepath.Join(root, "ui", "config.html"))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, ConfigTemplateData{CanViewSettings: true, DiscordAdminRoleID: "123456789012345678", StatusPanelChannelID: "987654321098765432"}); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, expected := range []string{`name="discordAdminRoleID"`, `value="123456789012345678"`, `name="statusPanelChannelID"`, `value="987654321098765432"`} {
		if !strings.Contains(html, expected) {
			t.Fatalf("missing config markup %s", expected)
		}
	}
	for _, obsolete := range []string{`name="controlChannelID"`, `name="controlPanelChannelID"`} {
		if strings.Contains(html, obsolete) {
			t.Fatalf("obsolete field %s", obsolete)
		}
	}

	output.Reset()
	if err := tmpl.Execute(&output, ConfigTemplateData{DiscordToken: "should-not-be-rendered"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "should-not-be-rendered") {
		t.Fatal("config values were rendered without settings.view")
	}
	for _, language := range []string{"en-US", "de-DE", "sv-SE"} {
		data, err := os.ReadFile(filepath.Join(root, "localization", language+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if !json.Valid(data) {
			t.Fatalf("invalid %s localization", language)
		}
		for _, key := range []string{"UIText_DiscordAdminRole", "UIText_DiscordAdminRoleInfo", "UIText_StatusPanelChannelInfo"} {
			if !bytes.Contains(data, []byte(`"`+key+`"`)) {
				t.Fatalf("missing %s in %s", key, language)
			}
		}
	}
}
