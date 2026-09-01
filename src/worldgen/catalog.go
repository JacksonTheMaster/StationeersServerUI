package worldgen

import (
	"encoding/xml"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type World struct {
	ID                 string   `json:"id"`
	Label              string   `json:"label"`
	StartConditions    []Option `json:"startConditions"`
	StartLocations     []Option `json:"startLocations"`
	DefaultConditionID string   `json:"defaultConditionId,omitempty"`
}

type Catalog struct {
	Worlds       []World  `json:"worlds"`
	Difficulties []Option `json:"difficulties"`
	Source       string   `json:"source"`
}

type Loader struct {
	mu          sync.Mutex
	fingerprint uint64
	dataRoot    string
	catalog     Catalog
}

var DefaultLoader Loader

type xmlGameData struct {
	Worlds []xmlWorld `xml:"WorldSettings>World"`
}

type xmlWorld struct {
	ID              string              `xml:"Id,attr"`
	Priority        *int                `xml:"Priority,attr"`
	Hidden          bool                `xml:"Hidden,attr"`
	Deprecated      bool                `xml:"Deprecated,attr"`
	StartConditions []xmlStartCondition `xml:"StartCondition"`
	StartLocations  []xmlID             `xml:"StartLocation"`
	RandomLocations []xmlID             `xml:"RandomStartLocation"`
}

type xmlStartCondition struct {
	ID        string `xml:"Id,attr"`
	IsDefault bool   `xml:"IsDefault,attr"`
}

type xmlID struct {
	ID string `xml:"Id,attr"`
}

type xmlDifficultyData struct {
	Difficulties []xmlDifficulty `xml:"DifficultySettings>DifficultySetting"`
}

type xmlDifficulty struct {
	ID      string `xml:"Id,attr"`
	Default bool   `xml:"Default,attr"`
}

func DataRootFromExecutable(executablePath string) string {
	cleanPath := filepath.Clean(executablePath)
	fileName := filepath.Base(cleanPath)
	for _, suffix := range []string{".x86_64", ".exe"} {
		fileName = strings.TrimSuffix(fileName, suffix)
	}
	return filepath.Join(filepath.Dir(cleanPath), fileName+"_Data")
}

func (loader *Loader) Load(executablePath string) (Catalog, error) {
	dataRoot := DataRootFromExecutable(executablePath)
	fingerprint, err := catalogFingerprint(dataRoot)
	if err != nil {
		return FallbackCatalog(), err
	}

	loader.mu.Lock()
	defer loader.mu.Unlock()

	if loader.dataRoot == dataRoot && loader.fingerprint == fingerprint && len(loader.catalog.Worlds) > 0 {
		return cloneCatalog(loader.catalog), nil
	}

	catalog, err := loadCatalog(dataRoot)
	if err != nil {
		return FallbackCatalog(), err
	}
	catalog.Source = "game-files"
	loader.dataRoot = dataRoot
	loader.fingerprint = fingerprint
	loader.catalog = catalog
	return cloneCatalog(catalog), nil
}

func loadCatalog(dataRoot string) (Catalog, error) {
	worldsRoot := filepath.Join(dataRoot, "StreamingAssets", "Worlds")
	var parsedWorlds []xmlWorld
	err := filepath.WalkDir(worldsRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != worldsRoot && strings.HasPrefix(strings.ToLower(entry.Name()), "tutorial") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".xml") {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var data xmlGameData
		if err := xml.Unmarshal(contents, &data); err != nil {
			return fmt.Errorf("parse world data %s: %w", path, err)
		}
		parsedWorlds = append(parsedWorlds, data.Worlds...)
		return nil
	})
	if err != nil {
		return Catalog{}, fmt.Errorf("scan Stationeers worlds: %w", err)
	}

	var catalog Catalog
	for _, parsed := range parsedWorlds {
		// Playable worlds use Priority. This excludes the tutorial/scenario XML files.
		if parsed.ID == "" || parsed.Priority == nil || parsed.Hidden || parsed.Deprecated {
			continue
		}
		world := World{ID: parsed.ID, Label: worldLabel(parsed.ID)}
		for _, condition := range parsed.StartConditions {
			if condition.ID == "" {
				continue
			}
			world.StartConditions = appendUnique(world.StartConditions, option(condition.ID))
			if condition.IsDefault {
				world.DefaultConditionID = condition.ID
			}
		}
		// Round-robin is the useful general/default choice, so present it first.
		for _, location := range parsed.RandomLocations {
			world.StartLocations = appendUnique(world.StartLocations, option(location.ID))
		}
		for _, location := range parsed.StartLocations {
			world.StartLocations = appendUnique(world.StartLocations, option(location.ID))
		}
		catalog.Worlds = append(catalog.Worlds, world)
	}

	sort.SliceStable(catalog.Worlds, func(i, j int) bool {
		left := priorityFor(parsedWorlds, catalog.Worlds[i].ID)
		right := priorityFor(parsedWorlds, catalog.Worlds[j].ID)
		if left == right {
			return catalog.Worlds[i].Label < catalog.Worlds[j].Label
		}
		return left < right
	})

	difficultyPath := filepath.Join(dataRoot, "StreamingAssets", "Data", "difficultySettings.xml")
	difficultyContents, err := os.ReadFile(difficultyPath)
	if err != nil {
		return Catalog{}, fmt.Errorf("read Stationeers difficulties: %w", err)
	}
	var difficultyData xmlDifficultyData
	if err := xml.Unmarshal(difficultyContents, &difficultyData); err != nil {
		return Catalog{}, fmt.Errorf("parse Stationeers difficulties: %w", err)
	}
	for _, difficulty := range difficultyData.Difficulties {
		if difficulty.ID != "" {
			catalog.Difficulties = appendUnique(catalog.Difficulties, option(difficulty.ID))
		}
	}

	if len(catalog.Worlds) == 0 || len(catalog.Difficulties) == 0 {
		return Catalog{}, errors.New("Stationeers world-generation catalog is empty")
	}
	return catalog, nil
}

func catalogFingerprint(dataRoot string) (uint64, error) {
	hasher := fnv.New64a()
	paths := []string{
		filepath.Join(dataRoot, "StreamingAssets", "Worlds"),
		filepath.Join(dataRoot, "StreamingAssets", "Data", "difficultySettings.xml"),
	}

	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			return 0, err
		}
		if !info.IsDir() {
			writeFingerprint(hasher, root, info)
			continue
		}
		if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".xml") {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			writeFingerprint(hasher, path, info)
			return nil
		}); err != nil {
			return 0, err
		}
	}
	return hasher.Sum64(), nil
}

func writeFingerprint(hasher interface{ Write([]byte) (int, error) }, path string, info fs.FileInfo) {
	_, _ = hasher.Write([]byte(path))
	_, _ = hasher.Write([]byte(strconv.FormatInt(info.Size(), 10)))
	_, _ = hasher.Write([]byte(strconv.FormatInt(info.ModTime().UnixNano(), 10)))
}

func priorityFor(worlds []xmlWorld, id string) int {
	for _, world := range worlds {
		if world.ID == id && world.Priority != nil {
			return *world.Priority
		}
	}
	return int(^uint(0) >> 1)
}

func appendUnique(options []Option, candidate Option) []Option {
	if candidate.Value == "" {
		return options
	}
	for _, existing := range options {
		if existing.Value == candidate.Value {
			return options
		}
	}
	return append(options, candidate)
}

func option(value string) Option {
	return Option{Value: value, Label: humanizeIdentifier(value)}
}

func worldLabel(id string) string {
	if label, exists := map[string]string{
		"Mars2":         "Mars",
		"Europa3":       "Europa",
		"Vulcan2":       "Vulcan",
		"MimasHerschel": "Mimas Herschel",
	}[id]; exists {
		return label
	}
	return humanizeIdentifier(id)
}

func humanizeIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if spawnIndex := strings.Index(value, "Spawn"); spawnIndex >= 0 {
		value = value[spawnIndex+len("Spawn"):]
	}
	var result []rune
	var previous rune
	for index, current := range []rune(value) {
		if index > 0 && (unicode.IsUpper(current) && (unicode.IsLower(previous) || unicode.IsDigit(previous)) || unicode.IsDigit(current) && !unicode.IsDigit(previous)) {
			result = append(result, ' ')
		}
		result = append(result, current)
		previous = current
	}
	return string(result)
}

func cloneCatalog(catalog Catalog) Catalog {
	clone := Catalog{Source: catalog.Source}
	clone.Difficulties = append([]Option(nil), catalog.Difficulties...)
	for _, world := range catalog.Worlds {
		world.StartConditions = append([]Option(nil), world.StartConditions...)
		world.StartLocations = append([]Option(nil), world.StartLocations...)
		clone.Worlds = append(clone.Worlds, world)
	}
	return clone
}

func FallbackCatalog() Catalog {
	conditions := func(prefix string) []Option {
		return []Option{
			option(prefix + "Default"), option(prefix + "DefaultCommunity"),
			option(prefix + "Brutal"), option(prefix + "BrutalCommunity"),
		}
	}
	defaultConditions := []Option{option("DefaultStart"), option("DefaultStartCommunity"), option("Brutal"), option("BrutalCommunity")}
	return Catalog{
		Source:       "fallback",
		Difficulties: []Option{option("Creative"), option("Easy"), option("Normal"), option("Stationeer")},
		Worlds: []World{
			{ID: "Lunar", Label: "Lunar", DefaultConditionID: "DefaultStart", StartConditions: defaultConditions, StartLocations: options("LunarSpawnRoundRobin", "LunarSpawnCraterVesper", "LunarSpawnMontesUmbrarum", "LunarSpawnCraterNox", "LunarSpawnMonsArcanus")},
			{ID: "Mars2", Label: "Mars", DefaultConditionID: "DefaultStart", StartConditions: defaultConditions, StartLocations: options("MarsSpawnRoundRobin", "MarsSpawnCanyonOverlook", "MarsSpawnButchersFlat", "MarsSpawnFindersCanyon", "MarsSpawnHellasCrags", "MarsSpawnDonutFlats")},
			{ID: "MimasHerschel", Label: "Mimas Herschel", DefaultConditionID: "MimasDefault", StartConditions: conditions("Mimas"), StartLocations: options("MimasSpawnRoundRobin", "MimasSpawnCentralMesa", "MimasSpawnHarrietCrater", "MimasSpawnCraterField", "MimasSpawnDustBowl")},
			{ID: "Europa3", Label: "Europa", DefaultConditionID: "EuropaDefault", StartConditions: conditions("Europa"), StartLocations: options("EuropaSpawnRoundRobin", "EuropaSpawnIcyBasin", "EuropaSpawnGlacialChannel", "EuropaSpawnBalgatanPass", "EuropaSpawnFrigidHighlands", "EuropaSpawnTyreValley")},
			{ID: "Venus", Label: "Venus", DefaultConditionID: "VenusDefault", StartConditions: options("VenusDefault", "VulcanBrutal", "VenusDefaultCommunity", "VulcanBrutalCommunity"), StartLocations: options("VenusSpawnRoundRobin", "VenusSpawnGaiaValley", "VenusSpawnDaisyValley", "VenusSpawnFaithValley", "VenusSpawnDuskValley")},
			{ID: "Vulcan2", Label: "Vulcan", DefaultConditionID: "VulcanDefault", StartConditions: conditions("Vulcan"), StartLocations: options("VulcanSpawnRoundRobin", "VulcanSpawnVestaValley", "VulcanSpawnEtnasFury", "VulcanSpawnIxionsDemise", "VulcanSpawnTitusReach", "VulcanSpawnSerpentPlateau", "VulcanSpawnKalimarOutlook", "VulcanSpawnTalonHeights", "VulcanSpawnDusterPlateauNorth", "VulcanSpawnCinderPeak", "VulcanSpawnRedJunction", "VulcanSpawnOrensTrack", "VulcanSpawnDogsfoot")},
		},
	}
}

func options(values ...string) []Option {
	result := make([]Option, 0, len(values))
	for _, value := range values {
		result = append(result, option(value))
	}
	return result
}
