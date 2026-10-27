package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const sample = `
FameAndFortuneDB = {
["char"] = {
["Ostbur - Doomhowl"] = {
["seq"] = 41,
},
},
["global"] = {
["installId"] = "abc123def0",
%s
},
}
`

func file(global string) string { return sprintf(sample, global) }

func TestReadsTheOneSavedCode(t *testing.T) {
	got := readSavedCode(file(`["latestCode"] = "FF1:0a1b2c3d:abcDEF0123()xyz",
["latestCodeAt"] = 1793120400,
["latestCodeVersion"] = "1.2.0",`))
	want := savedCode{Codes: []string{"FF1:0a1b2c3d:abcDEF0123()xyz"}, At: 1793120400, Version: "1.2.0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestReadsParts(t *testing.T) {
	got := readSavedCode(file(`["latestParts"] = {
"FF1:00000001:aaaa",
"FF1:00000002:bbbb",
},`))
	if !reflect.DeepEqual(got.Codes, []string{"FF1:00000001:aaaa", "FF1:00000002:bbbb"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestSaysWhyThereIsNoCode(t *testing.T) {
	for text, reason := range map[string]string{
		"WeakAurasSaved = {}":         "not-ours",
		file(""):                      "no-code",
		file(`["saveCode"] = false,`): "turned-off",
		file(`["lastSoftReserve"] = "FF1:99999999:z",`): "no-code",
	} {
		if got := readSavedCode(text); got.Reason != reason {
			t.Errorf("%q: got %q, want %q", text, got.Reason, reason)
		}
	}
}

func TestFindsEverySavedDataFile(t *testing.T) {
	root := t.TempDir()
	forever := filepath.Join(root, "World of Warcraft", "_forever_")
	for _, acct := range []string{"WOW1", "12345678#1"} {
		dir := filepath.Join(forever, "WTF", "Account", acct, "SavedVariables")
		must(t, os.MkdirAll(dir, 0o755))
		must(t, os.WriteFile(filepath.Join(dir, savedFile), []byte(file("")), 0o644))
	}
	// Character folders and other games' files are left alone.
	must(t, os.MkdirAll(filepath.Join(forever, "WTF", "Account", "WOW1", "Doomhowl", "Ostbur", "SavedVariables"), 0o755))
	for _, start := range []string{filepath.Join(root, "World of Warcraft"), forever} {
		got := savedFiles(start)
		if len(got) != 2 || got[0].Account != "12345678#1" || got[1].Account != "WOW1" {
			t.Fatalf("from %s: %+v", start, got)
		}
	}
	if forever != foreverDir(filepath.Join(root, "World of Warcraft")) {
		t.Fatal("the World of Warcraft folder leads to _forever_")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
