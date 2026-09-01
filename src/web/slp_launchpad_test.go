package web

import (
	"reflect"
	"testing"
)

func TestParseWorkshopHandles(t *testing.T) {
	got, err := parseWorkshopHandles([]string{
		"12345, https://steamcommunity.com/sharedfiles/filedetails/?id=67890",
		"12345",
	})
	if err != nil {
		t.Fatalf("parseWorkshopHandles returned an error: %v", err)
	}

	want := []string{"12345", "67890"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseWorkshopHandles() = %v, want %v", got, want)
	}
}

func TestParseWorkshopHandlesRejectsInvalidInput(t *testing.T) {
	for _, inputs := range [][]string{
		nil,
		{"not-a-workshop-item"},
		{"https://steamcommunity.com/sharedfiles/filedetails/"},
		{"0"},
	} {
		if _, err := parseWorkshopHandles(inputs); err == nil {
			t.Fatalf("parseWorkshopHandles(%v) unexpectedly succeeded", inputs)
		}
	}
}
