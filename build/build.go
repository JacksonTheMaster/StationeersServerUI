//go:build ignore

// Run from the repository root with `go run build/build.go`.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)

type version struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

type platform struct {
	os   string
	arch string
}

func main() {
	requestedVersion := flag.String("version", "", "version embedded in the binaries (for example 6.0.0-rc.1)")
	target := flag.String("target", "", "build target: all, windows, or linux")
	legacyAssets := flag.Bool("legacy-assets", false, "also create the old StationeersServerControl asset names")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	buildVersion := strings.TrimPrefix(strings.TrimSpace(*requestedVersion), "v")
	buildTarget := strings.ToLower(strings.TrimSpace(*target))
	interactive := buildVersion == "" && buildTarget == ""

	if interactive {
		var err error
		buildVersion, err = chooseVersion(reader, config.Version)
		if err != nil {
			log.Fatal(err)
		}
		buildTarget, err = chooseTarget(reader)
		if err != nil {
			log.Fatal(err)
		}
		*legacyAssets = askYesNo(reader, "Create legacy updater assets?", false)
	} else {
		if buildVersion == "" {
			buildVersion = strings.TrimPrefix(config.Version, "v")
		}
		if buildTarget == "" {
			buildTarget = "all"
		}
	}

	if _, err := parseVersion(buildVersion); err != nil {
		log.Fatalf("Invalid build version: %v", err)
	}
	platforms, err := selectedPlatforms(buildTarget)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\nSSUI build %s%s\n", colorCyan, buildVersion, colorReset)
	if err := os.MkdirAll("build", 0755); err != nil {
		log.Fatal(err)
	}
	cleanupOldExecutables(buildVersion)
	for _, item := range platforms {
		if err := build(item, buildVersion); err != nil {
			log.Fatal(err)
		}
	}
	if *legacyAssets {
		if err := createLegacyAssets(platforms, buildVersion); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("%s\nBuild complete.%s\n", colorGreen, colorReset)
}

func chooseVersion(reader *bufio.Reader, current string) (string, error) {
	parsed, err := parseVersion(strings.TrimPrefix(current, "v"))
	if err != nil {
		return "", fmt.Errorf("current version %q is invalid: %w", current, err)
	}
	fmt.Printf("\nCurrent source version: %s\n", current)
	fmt.Println("  1) Build current version")
	fmt.Println("  2) Next patch")
	fmt.Println("  3) Next minor")
	fmt.Println("  4) Next major")
	fmt.Println("  5) Prerelease for current version")
	fmt.Println("  6) Enter a version")

	switch ask(reader, "Version", "1") {
	case "1":
		return parsed.String(), nil
	case "2":
		parsed.patch++
		parsed.prerelease = ""
	case "3":
		parsed.minor++
		parsed.patch = 0
		parsed.prerelease = ""
	case "4":
		parsed.major++
		parsed.minor = 0
		parsed.patch = 0
		parsed.prerelease = ""
	case "5":
		label := ask(reader, "Prerelease label", "rc")
		number := 1
		parts := strings.Split(parsed.prerelease, ".")
		if len(parts) == 2 && parts[0] == label {
			if previous, parseErr := strconv.Atoi(parts[1]); parseErr == nil {
				number = previous + 1
			}
		}
		parsed.prerelease = label + "." + ask(reader, "Prerelease number", strconv.Itoa(number))
	case "6":
		return strings.TrimPrefix(ask(reader, "Exact version", current), "v"), nil
	default:
		return "", errors.New("unknown version choice")
	}
	return parsed.String(), nil
}

func chooseTarget(reader *bufio.Reader) (string, error) {
	fmt.Println("\n  1) Windows and Linux")
	fmt.Println("  2) Windows")
	fmt.Println("  3) Linux")
	switch ask(reader, "Target", "1") {
	case "1":
		return "all", nil
	case "2":
		return "windows", nil
	case "3":
		return "linux", nil
	default:
		return "", errors.New("unknown build target")
	}
}

func ask(reader *bufio.Reader, label, fallback string) string {
	fmt.Printf("%s [%s]: ", label, fallback)
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func askYesNo(reader *bufio.Reader, label string, fallback bool) bool {
	hint := "y/N"
	if fallback {
		hint = "Y/n"
	}
	fmt.Printf("%s [%s]: ", label, hint)
	value, _ := reader.ReadString('\n')
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	return value == "y" || value == "yes"
}

func parseVersion(value string) (version, error) {
	match := semverPattern.FindStringSubmatch(value)
	if match == nil {
		return version{}, fmt.Errorf("%q is not a semantic version", value)
	}
	major, _ := strconv.Atoi(match[1])
	minor, _ := strconv.Atoi(match[2])
	patch, _ := strconv.Atoi(match[3])
	return version{major: major, minor: minor, patch: patch, prerelease: match[4]}, nil
}

func (value version) String() string {
	result := fmt.Sprintf("%d.%d.%d", value.major, value.minor, value.patch)
	if value.prerelease != "" {
		result += "-" + value.prerelease
	}
	return result
}

func selectedPlatforms(target string) ([]platform, error) {
	switch target {
	case "all":
		return []platform{{os: "windows", arch: "amd64"}, {os: "linux", arch: "amd64"}}, nil
	case "windows":
		return []platform{{os: "windows", arch: "amd64"}}, nil
	case "linux":
		return []platform{{os: "linux", arch: "amd64"}}, nil
	default:
		return nil, fmt.Errorf("unknown build target %q", target)
	}
}

func build(target platform, buildVersion string) error {
	outputPath := filepath.Join("build", canonicalAssetName(buildVersion, target))
	packagePath := "github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	ldflags := fmt.Sprintf("-s -w -X %s.Version=%s", packagePath, buildVersion)

	fmt.Printf("%sBuilding %s/%s...%s\n", colorBlue, target.os, target.arch, colorReset)
	cmd := exec.Command("go", "build", "-ldflags="+ldflags, "-gcflags=-l=4", "-o", outputPath, "server.go")
	cmd.Env = append(os.Environ(), "GOOS="+target.os, "GOARCH="+target.arch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build %s/%s: %w\n%s", target.os, target.arch, err, output)
	}
	fmt.Printf("%sBuilt %s%s\n", colorGreen, outputPath, colorReset)
	return nil
}

func canonicalAssetName(buildVersion string, target platform) string {
	name := fmt.Sprintf("StationeersServerUI_v%s_%s_%s", buildVersion, target.os, target.arch)
	if target.os == "windows" {
		name += ".exe"
	}
	return name
}

func createLegacyAssets(platforms []platform, buildVersion string) error {
	for _, target := range platforms {
		extension := ".x86_64"
		if target.os == "windows" {
			extension = ".exe"
		}
		from := filepath.Join("build", canonicalAssetName(buildVersion, target))
		to := filepath.Join("build", "StationeersServerControlv"+buildVersion+extension)
		data, err := os.ReadFile(from)
		if err != nil {
			return err
		}
		if err := os.WriteFile(to, data, 0755); err != nil {
			return err
		}
		fmt.Printf("%sCompatibility asset: %s%s\n", colorYellow, to, colorReset)
	}
	return nil
}

func cleanupOldExecutables(buildVersion string) {
	files, err := os.ReadDir("build")
	if err != nil {
		return
	}
	for _, file := range files {
		name := file.Name()
		isBuild := strings.HasPrefix(name, "StationeersServerUI_") || strings.HasPrefix(name, "StationeersServerControl")
		if !isBuild || strings.Contains(name, buildVersion) {
			continue
		}
		path := filepath.Join("build", name)
		if err := os.Remove(path); err != nil {
			fmt.Printf("%sCould not remove %s: %v%s\n", colorRed, path, err, colorReset)
		}
	}
}
