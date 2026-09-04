package backupmgr

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	worldMetaFilename   = "world_meta.xml"
	worldFilename       = "world.xml"
	filetimeEpochOffset = 116444736000000000 // difference between 1601 and 1970 in 100-ns units

	// Metadata is normally less than 1 KiB. Keeping a generous hard limit prevents
	// malformed archives from turning the cheap summary path into an expensive one.
	maxWorldMetaSize = 1 << 20
	// Large, long-running worlds can legitimately have very large world.xml files.
	// The limit is only a safety boundary against accidental or malicious zip bombs.
	maxWorldSize = 2 << 30
)

// SaveSummary contains information that can be read cheaply from world_meta.xml.
type SaveSummary struct {
	ID                 string    `json:"id"`
	GameVersion        string    `json:"gameVersion"`
	SavedAt            time.Time `json:"savedAt"`
	DaysPlayed         int64     `json:"daysPlayed"`
	WorldName          string    `json:"worldName"`
	WorldFileName      string    `json:"worldFileName"`
	Rooms              int64     `json:"rooms"`
	PipeNetworks       int64     `json:"pipeNetworks"`
	CableNetworks      int64     `json:"cableNetworks"`
	Things             int64     `json:"things"`
	Atmospheres        int64     `json:"atmospheres"`
	ArchiveSize        int64     `json:"archiveSize"`
	WorldXMLSize       uint64    `json:"worldXmlSize"`
	WorldXMLCompressed uint64    `json:"worldXmlCompressedSize"`
}

// SaveAnalysis contains values that require streaming through world.xml.
type SaveAnalysis struct {
	SaveSummary
	Players            int64 `json:"players"`
	PlayersAlive       int64 `json:"playersAlive"`
	PlayersUnconscious int64 `json:"playersUnconscious"`
	Furnaces           int64 `json:"furnaces"`
	DestroyedFurnaces  int64 `json:"destroyedFurnaces"`
}

type worldMetaData struct {
	XMLName               xml.Name `xml:"WorldMetaData"`
	ID                    string   `xml:"Id,attr"`
	GameVersion           string   `xml:"GameVersion"`
	DateTime              int64    `xml:"DateTime"`
	DaysPast              int64    `xml:"DaysPast"`
	WorldName             string   `xml:"WorldName"`
	WorldFileName         string   `xml:"WorldFileName"`
	NumberOfRooms         int64    `xml:"NumberOfRooms"`
	NumberOfPipeNetworks  int64    `xml:"NumberOfPipeNetworks"`
	NumberOfCableNetworks int64    `xml:"NumberOfCableNetworks"`
	NumberOfThings        int64    `xml:"NumberOfThings"`
	NumberOfAtmospheres   int64    `xml:"NumberOfAtmospheres"`
}

// ReadSaveSummary reads only the small world_meta.xml member of a Stationeers save.
func ReadSaveSummary(path string) (SaveSummary, error) {
	archive, stat, err := openSaveArchive(path)
	if err != nil {
		return SaveSummary{}, err
	}
	defer archive.Close()

	metaFile := findZipMember(archive.File, worldMetaFilename)
	if metaFile == nil {
		return SaveSummary{}, fmt.Errorf("save archive is missing %s", worldMetaFilename)
	}
	if metaFile.UncompressedSize64 > maxWorldMetaSize {
		return SaveSummary{}, fmt.Errorf("%s exceeds the %d-byte safety limit", worldMetaFilename, maxWorldMetaSize)
	}

	metaReader, err := metaFile.Open()
	if err != nil {
		return SaveSummary{}, fmt.Errorf("open %s: %w", worldMetaFilename, err)
	}
	defer metaReader.Close()

	var meta worldMetaData
	limited := io.LimitReader(metaReader, maxWorldMetaSize+1)
	if err := xml.NewDecoder(&saveXMLReader{reader: limited}).Decode(&meta); err != nil {
		return SaveSummary{}, fmt.Errorf("decode %s: %w", worldMetaFilename, err)
	}

	worldFile := findZipMember(archive.File, worldFilename)
	if worldFile == nil {
		return SaveSummary{}, fmt.Errorf("save archive is missing %s", worldFilename)
	}

	return summaryFromMetadata(meta, stat.Size(), worldFile), nil
}

// AnalyzeSave streams world.xml once. It does not extract the archive to disk or
// retain the XML document in memory, and it stops promptly when ctx is cancelled.
func AnalyzeSave(ctx context.Context, path string) (SaveAnalysis, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return SaveAnalysis{}, err
	}

	archive, stat, err := openSaveArchive(path)
	if err != nil {
		return SaveAnalysis{}, err
	}
	defer archive.Close()

	meta, err := readMetadataFromArchive(archive)
	if err != nil {
		return SaveAnalysis{}, err
	}
	worldFile := findZipMember(archive.File, worldFilename)
	if worldFile == nil {
		return SaveAnalysis{}, fmt.Errorf("save archive is missing %s", worldFilename)
	}
	if worldFile.UncompressedSize64 > maxWorldSize {
		return SaveAnalysis{}, fmt.Errorf("%s exceeds the %d-byte safety limit", worldFilename, maxWorldSize)
	}

	worldReader, err := worldFile.Open()
	if err != nil {
		return SaveAnalysis{}, fmt.Errorf("open %s: %w", worldFilename, err)
	}
	defer worldReader.Close()

	result := SaveAnalysis{SaveSummary: summaryFromMetadata(meta, stat.Size(), worldFile)}
	limited := io.LimitReader(worldReader, maxWorldSize+1)
	if err := scanWorld(ctx, limited, &result); err != nil {
		return SaveAnalysis{}, fmt.Errorf("scan %s: %w", worldFilename, err)
	}
	return result, nil
}

type scannedThing struct {
	isHuman            bool
	isFurnace          bool
	isAlive            bool
	isUnconscious      bool
	hasSpawnedWreckage bool
}

type capturedValue uint8

const (
	captureNone capturedValue = iota
	capturePrefab
	captureState
	captureWreckage
)

func scanWorld(ctx context.Context, source io.Reader, analysis *SaveAnalysis) error {
	reader := bufio.NewReaderSize(&contextReader{ctx: ctx, reader: source}, 64*1024)
	var thing scannedThing
	inThing := false
	depth := 0
	capture := captureNone

	for {
		text, err := reader.ReadSlice('<')
		if err != nil && err != bufio.ErrBufferFull && err != io.EOF {
			return err
		}
		if len(text) > 0 && text[len(text)-1] == '<' {
			text = text[:len(text)-1]
		}
		if capture != captureNone {
			value := bytes.TrimSpace(text)
			switch capture {
			case capturePrefab:
				thing.isFurnace = bytes.HasPrefix(value, []byte("Structure")) && bytes.Contains(value, []byte("Furnace"))
			case captureState:
				thing.isAlive = bytes.EqualFold(value, []byte("Alive"))
				thing.isUnconscious = bytes.EqualFold(value, []byte("Unconscious"))
			case captureWreckage:
				thing.hasSpawnedWreckage = bytes.EqualFold(value, []byte("true"))
			}
		}
		if err == bufio.ErrBufferFull {
			if capture != captureNone {
				return fmt.Errorf("value for tracked element exceeds scanner buffer")
			}
			continue
		}
		if err == io.EOF {
			if inThing {
				return io.ErrUnexpectedEOF
			}
			return nil
		}

		tag, err := reader.ReadSlice('>')
		if err != nil {
			if err == io.EOF {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		tag = bytes.TrimSpace(tag[:len(tag)-1])
		if len(tag) == 0 || tag[0] == '?' || tag[0] == '!' {
			continue
		}

		closing := tag[0] == '/'
		selfClosing := tag[len(tag)-1] == '/'
		name := tagName(tag, closing)

		if !inThing {
			if !closing && bytes.Equal(name, []byte("ThingSaveData")) {
				inThing = true
				depth = 0
				capture = captureNone
				thing = scannedThing{isHuman: bytes.Contains(tag, []byte(`type="HumanSaveData"`))}
			}
			continue
		}

		if closing {
			if depth == 0 && bytes.Equal(name, []byte("ThingSaveData")) {
				analysis.addThing(thing)
				inThing = false
				capture = captureNone
				continue
			}
			if depth <= 0 {
				return fmt.Errorf("unexpected closing element %q", name)
			}
			depth--
			if depth == 0 {
				capture = captureNone
			}
			continue
		}

		if depth == 0 {
			switch {
			case bytes.Equal(name, []byte("PrefabName")):
				capture = capturePrefab
			case thing.isHuman && bytes.Equal(name, []byte("State")):
				capture = captureState
			case bytes.Equal(name, []byte("HasSpawnedWreckage")):
				capture = captureWreckage
			default:
				capture = captureNone
			}
		}
		if !selfClosing {
			depth++
		}
	}
}

func tagName(tag []byte, closing bool) []byte {
	if closing {
		tag = bytes.TrimSpace(tag[1:])
	}
	if index := bytes.IndexAny(tag, " \t\r\n/"); index >= 0 {
		return tag[:index]
	}
	return tag
}

func (analysis *SaveAnalysis) addThing(thing scannedThing) {
	if thing.isHuman {
		analysis.Players++
		switch {
		case thing.isAlive:
			analysis.PlayersAlive++
		case thing.isUnconscious:
			analysis.PlayersUnconscious++
		}
	}

	if thing.isFurnace {
		analysis.Furnaces++
		if thing.hasSpawnedWreckage {
			analysis.DestroyedFurnaces++
		}
	}
}

func openSaveArchive(path string) (*zip.ReadCloser, os.FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("stat save archive: %w", err)
	}
	if !stat.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("save archive is not a regular file")
	}

	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open save archive: %w", err)
	}
	return archive, stat, nil
}

func readMetadataFromArchive(archive *zip.ReadCloser) (worldMetaData, error) {
	metaFile := findZipMember(archive.File, worldMetaFilename)
	if metaFile == nil {
		return worldMetaData{}, fmt.Errorf("save archive is missing %s", worldMetaFilename)
	}
	if metaFile.UncompressedSize64 > maxWorldMetaSize {
		return worldMetaData{}, fmt.Errorf("%s exceeds the %d-byte safety limit", worldMetaFilename, maxWorldMetaSize)
	}

	reader, err := metaFile.Open()
	if err != nil {
		return worldMetaData{}, fmt.Errorf("open %s: %w", worldMetaFilename, err)
	}
	defer reader.Close()

	var meta worldMetaData
	if err := xml.NewDecoder(&saveXMLReader{reader: io.LimitReader(reader, maxWorldMetaSize+1)}).Decode(&meta); err != nil {
		return worldMetaData{}, fmt.Errorf("decode %s: %w", worldMetaFilename, err)
	}
	return meta, nil
}

func findZipMember(files []*zip.File, name string) *zip.File {
	for _, file := range files {
		if file.Name == name {
			return file
		}
	}
	return nil
}

func summaryFromMetadata(meta worldMetaData, archiveSize int64, worldFile *zip.File) SaveSummary {
	return SaveSummary{
		ID:                 meta.ID,
		GameVersion:        meta.GameVersion,
		SavedAt:            filetimeToTime(meta.DateTime),
		DaysPlayed:         meta.DaysPast,
		WorldName:          meta.WorldName,
		WorldFileName:      meta.WorldFileName,
		Rooms:              meta.NumberOfRooms,
		PipeNetworks:       meta.NumberOfPipeNetworks,
		CableNetworks:      meta.NumberOfCableNetworks,
		Things:             meta.NumberOfThings,
		Atmospheres:        meta.NumberOfAtmospheres,
		ArchiveSize:        archiveSize,
		WorldXMLSize:       worldFile.UncompressedSize64,
		WorldXMLCompressed: worldFile.CompressedSize64,
	}
}

func filetimeToTime(value int64) time.Time {
	ticks := value - filetimeEpochOffset
	seconds := ticks / 10_000_000
	nanoseconds := (ticks % 10_000_000) * 100
	return time.Unix(seconds, nanoseconds)
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (reader *contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-reader.ctx.Done():
		return 0, reader.ctx.Err()
	default:
		return reader.reader.Read(buffer)
	}
}
