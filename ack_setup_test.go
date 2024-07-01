package ptxt

// This file contains a fake test ensuring that test assets are available,
// sets up a few important variables and provides some helper methods.

import "os"
import "io/fs"
import "fmt"
import "embed"
import "sync"
import "strings"
import "testing"

import "image"
import "image/png"

import "github.com/tinne26/ggfnt"

//go:embed test/fonts/*
var testFS embed.FS

var testFontsDir string = "test/fonts"
var testFont *ggfnt.Font
var assetsLoadMutex sync.Mutex
var testAssetsLoaded bool

func TestAssetAvailability(t *testing.T) {
	if len(testWarnings) > 0 {
		t.Fatalf("missing test assets\n%s", testWarnings)
	}
}

var testWarnings string
func ensureTestAssetsLoaded() {
	// assets load access control
	assetsLoadMutex.Lock()
	defer assetsLoadMutex.Unlock()
	if testAssetsLoaded { return }
	testAssetsLoaded = true

	// try to load one font from the embedded folder
	var fonts []*ggfnt.Font
	err := fs.WalkDir(testFS, testFontsDir,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil { return err }
			if entry.IsDir() {
				if path == testFontsDir { return nil }
				return fs.SkipDir
			}

			if !strings.HasSuffix(path, ".ggfnt") {
				return nil
			}

			file, err := testFS.Open(path)
			if err != nil { return err }
			font, err := ggfnt.Parse(file)
			if err != nil {
				_ = file.Close()
				return err
			}
			fonts = append(fonts, font)
			return file.Close()
		})
	if err != nil {
		fmt.Printf("TESTS INIT: %s", err.Error())
		os.Exit(1)
	}

	// set fonts and/or warnings for missing fonts
	if len(fonts) == 0 {
		testWarnings = "WARNING: Expected at least one .ggfnt font in " + testFontsDir + "/ (found 0)\n" +
		               "WARNING: Most tests will be skipped\n"
	} else {
		testFont = fonts[0]
	}
}

// --- helpers ---

func exportAsPNG(filename string, img image.Image) {
	file, err := os.Create(filename)
	if err != nil { panic(err) }
	err = png.Encode(file, img)
	if err != nil { panic(err) }
	err = file.Close()
	if err != nil { panic(err) }
}

func equalSlices(a, b []byte) bool {
	if len(a) != len(b) { return false }
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] { return false }
	}
	return true
}
