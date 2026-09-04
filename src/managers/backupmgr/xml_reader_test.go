package backupmgr

import (
	"context"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestSaveXMLReaderHandlesChunkBoundaries(t *testing.T) {
	for _, text := range []string{
		"<WorldData><Text>value&#xB;second&#11;third\x0b</Text></WorldData>",
		strings.Repeat(" ", 32760) + "<WorldData>hello&#xB;world</WorldData>",
	} {
		for _, reader := range []io.Reader{strings.NewReader(text), iotest.OneByteReader(strings.NewReader(text))} {
			if err := validateSaveXML(context.Background(), reader, "WorldData"); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, text := range []string{"<WorldData>&unknown;</WorldData>", "<WorldData>&#xB", "<WorldData>&#xFFFFFFFF;</WorldData>"} {
		if err := validateSaveXML(context.Background(), iotest.OneByteReader(strings.NewReader(text)), "WorldData"); err == nil {
			t.Fatalf("accepted malformed XML: %s", text)
		}
	}
}
