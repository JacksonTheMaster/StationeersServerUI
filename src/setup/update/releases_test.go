package update

import "testing"

func TestVersionOrder(t *testing.T) {
	ordered := []string{"v6.0.0-alpha", "v6.0.0-alpha.1", "v6.0.0-beta", "v6.0.0-rc.1", "v6.0.0-rc.2", "v6.0.0", "v6.0.1"}
	for i := 1; i < len(ordered); i++ {
		previous, err := parseVersion(ordered[i-1])
		if err != nil {
			t.Fatal(err)
		}
		current, err := parseVersion(ordered[i])
		if err != nil {
			t.Fatal(err)
		}
		if compareVersions(previous, current) >= 0 {
			t.Fatalf("expected %s before %s", ordered[i-1], ordered[i])
		}
	}
}

func TestParseVersionRejectsLooseTags(t *testing.T) {
	for _, value := range []string{"6.0", "6.0.0.1", "v6.0.0-rc.01", "release-6.0.0"} {
		if _, err := parseVersion(value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestPolicyKeepsCompatiblePatchVisible(t *testing.T) {
	current, _ := parseVersion("v5.14.0")
	releases := []release{testRelease("v6.0.0-rc.1", true), testRelease("v5.14.1", false)}
	status := buildStatus(Status{CurrentVersion: "v5.14.0", Candidates: []Candidate{}}, current, releases, false, false)
	if status.CompatibleStable == nil || status.CompatibleStable.Version != "v5.14.1" {
		t.Fatalf("compatible update was hidden: %#v", status.CompatibleStable)
	}
	if status.LatestMajor == nil || status.LatestMajor.Version != "v6.0.0-rc.1" {
		t.Fatalf("major prerelease was not reported: %#v", status.LatestMajor)
	}
	if status.Automatic == nil || status.Automatic.Version != "v5.14.1" {
		t.Fatalf("wrong automatic candidate: %#v", status.Automatic)
	}
}

func TestRCNeedsBothAutomaticFlags(t *testing.T) {
	current, _ := parseVersion("v5.14.1")
	releases := []release{testRelease("v6.0.0-rc.1", true)}
	base := Status{CurrentVersion: "v5.14.1", Candidates: []Candidate{}}
	if status := buildStatus(base, current, releases, true, false); status.Automatic != nil {
		t.Fatal("prerelease became automatic without prerelease flag")
	}
	if status := buildStatus(base, current, releases, false, true); status.Automatic != nil {
		t.Fatal("major became automatic without major flag")
	}
	if status := buildStatus(base, current, releases, true, true); status.Automatic == nil {
		t.Fatal("candidate was not automatic with both flags")
	}
}

func testRelease(tag string, prerelease bool) release {
	version, err := parseVersion(tag)
	if err != nil {
		panic(err)
	}
	return release{Tag: tag, Version: version, Prerelease: prerelease, Assets: []githubAsset{{Name: expectedExecutable(tag), URL: "https://example.invalid/update"}}}
}
