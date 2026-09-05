package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/JacksonTheMaster/StationeersServerUI/v5/src/config"
)

func TestDownloadReplacesStaleCompletedFile(t *testing.T) {
	originalClient := downloadHTTPClient
	t.Cleanup(func() { downloadHTTPClient = originalClient })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("new executable"))
	}))
	defer server.Close()
	downloadHTTPClient = server.Client()

	t.Chdir(t.TempDir())
	filename := "StationeersServerControlv6.0.0.exe"
	if err := os.WriteFile(filename, []byte("stale executable"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := downloadNewExecutable(filename, server.URL); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new executable" {
		t.Fatalf("downloaded data = %q", data)
	}
}

func TestParseVersion(t *testing.T) {
	tests := map[string]Version{
		"5.14.1":           {Major: 5, Minor: 14, Patch: 1},
		"v6.0.0":           {Major: 6, Minor: 0, Patch: 0},
		"v6.0.0-rc.1":      {Major: 6, Minor: 0, Patch: 0},
		"v12.345.678-beta": {Major: 12, Minor: 345, Patch: 678},
	}
	for input, expected := range tests {
		actual, err := parseVersion(input)
		if err != nil {
			t.Fatalf("parseVersion(%q): %v", input, err)
		}
		if actual != expected {
			t.Fatalf("parseVersion(%q) = %#v, want %#v", input, actual, expected)
		}
	}
	if _, err := parseVersion("nightly"); err == nil {
		t.Fatal("accepted a tag without a semantic version")
	}
}

func TestMajorUpdateNeedsPersistentOrOneTimeApproval(t *testing.T) {
	original := config.AllowMajorUpdates
	t.Cleanup(func() { config.AllowMajorUpdates = original })
	config.AllowMajorUpdates = false

	current := Version{Major: 5, Minor: 14, Patch: 1}
	latest := Version{Major: 6, Minor: 0, Patch: 0}
	if reason, allowed := shouldUpdate(current, latest, true, false); allowed || reason != "major-update" {
		t.Fatalf("unapproved major update = %q, %t", reason, allowed)
	}
	if reason, allowed := shouldUpdate(current, latest, true, true); !allowed || reason != "" {
		t.Fatalf("one-time approved major update = %q, %t", reason, allowed)
	}

	config.AllowMajorUpdates = true
	if reason, allowed := shouldUpdate(current, latest, true, false); !allowed || reason != "" {
		t.Fatalf("persistently enabled major update = %q, %t", reason, allowed)
	}
}

func TestLatestReleaseRespectsReleaseChannel(t *testing.T) {
	originalURL := githubReleasesURL
	originalClient := releaseHTTPClient
	originalPrereleases := config.AllowPrereleaseUpdates
	t.Cleanup(func() {
		githubReleasesURL = originalURL
		releaseHTTPClient = originalClient
		config.AllowPrereleaseUpdates = originalPrereleases
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"tag_name":"v6.1.0","prerelease":false,"draft":true},
			{"tag_name":"v6.0.1-rc.1","prerelease":true,"draft":false},
			{"tag_name":"v5.14.1","prerelease":false,"draft":false},
			{"tag_name":"v5.13.2","prerelease":false,"draft":false}
		]`))
	}))
	defer server.Close()
	githubReleasesURL = server.URL
	releaseHTTPClient = server.Client()

	config.AllowPrereleaseUpdates = false
	release, err := getLatestRelease()
	if err != nil {
		t.Fatal(err)
	}
	if release.TagName != "v5.14.1" {
		t.Fatalf("stable release = %q", release.TagName)
	}

	config.AllowPrereleaseUpdates = true
	release, err = getLatestRelease()
	if err != nil {
		t.Fatal(err)
	}
	if release.TagName != "v6.0.1-rc.1" {
		t.Fatalf("prerelease channel release = %q", release.TagName)
	}
}

func TestCheckFindsMajorUpdateWithoutApplyingIt(t *testing.T) {
	originalURL := githubReleasesURL
	originalClient := releaseHTTPClient
	originalVersion := config.Version
	originalBranch := config.Branch
	originalUpdates := config.IsUpdateEnabled
	originalPrereleases := config.AllowPrereleaseUpdates
	t.Cleanup(func() {
		githubReleasesURL = originalURL
		releaseHTTPClient = originalClient
		config.Version = originalVersion
		config.Branch = originalBranch
		config.IsUpdateEnabled = originalUpdates
		config.AllowPrereleaseUpdates = originalPrereleases
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"tag_name":"v6.0.0","prerelease":false,"draft":false,"assets":[]}]`))
	}))
	defer server.Close()
	githubReleasesURL = server.URL
	releaseHTTPClient = server.Client()
	config.Version = "5.14.1"
	config.Branch = "release"
	config.IsUpdateEnabled = true
	config.AllowPrereleaseUpdates = false

	err, version := CheckForUpdates()
	if err != nil {
		t.Fatal(err)
	}
	if version != "v6.0.0" {
		t.Fatalf("available version = %q", version)
	}
	if !IsMajorUpdate(version) {
		t.Fatal("major update was not identified")
	}
}

func TestApprovedVersionCannotChangeDuringApply(t *testing.T) {
	originalURL := githubReleasesURL
	originalClient := releaseHTTPClient
	originalVersion := config.Version
	originalBranch := config.Branch
	originalUpdates := config.IsUpdateEnabled
	t.Cleanup(func() {
		githubReleasesURL = originalURL
		releaseHTTPClient = originalClient
		config.Version = originalVersion
		config.Branch = originalBranch
		config.IsUpdateEnabled = originalUpdates
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"tag_name":"v6.0.1","prerelease":false,"draft":false,"assets":[]}]`))
	}))
	defer server.Close()
	githubReleasesURL = server.URL
	releaseHTTPClient = server.Client()
	config.Version = "5.14.1"
	config.Branch = "release"
	config.IsUpdateEnabled = true

	err, version := ApplyVersion("v6.0.0", true)
	if err == nil {
		t.Fatal("applied a release other than the one that was approved")
	}
	if version != "v6.0.1" {
		t.Fatalf("changed version = %q", version)
	}
}
