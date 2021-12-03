// This file contains a bunch of unit tests, but the tests don't assert anything important.  They exist primarily to
// give short, self-contained, individually runnable examples of how to use this library.  Many examples will save
// their output to the `saved_examples/` directory.

package dxf

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestReadDrawing(t *testing.T) {
	drawing, err := ReadFile(path.Join(getThisDirectory(), "sample_drawing.dxf"))
	if err != nil {
		t.Errorf("Error reading file: %s", err)
	}

	for _, entity := range drawing.Entities {
		switch ent := entity.(type) {
		case *Line:
			fmt.Printf("Line from %s to %s\n", ent.P1.String(), ent.P2.String())
		}
	}
}

func TestWriteSimpleLineR12(t *testing.T) {
	drawing := NewDrawing()
	drawing.Header.Version = R12
	line := NewLine()
	line.P1 = Point{X: 0.0, Y: 0.0, Z: 0.0}
	line.P2 = Point{X: 1.0, Y: 1.0, Z: 0.0}
	drawing.Entities = append(drawing.Entities, line)
	err := drawing.SaveFile(path.Join(getSampleOutputDirectory(), "simple_line_r12.dxf"))
	if err != nil {
		t.Errorf("Error saving file: %v", err)
	}
}

func getThisDirectory() string {
	_, filename, _, _ := runtime.Caller(0)
	thisDir := path.Dir(filename)
	return thisDir
}

func getSampleOutputDirectory() string {
	thisDir := getThisDirectory()
	examplesDir := path.Join(thisDir, "saved_examples")
	os.Mkdir(examplesDir, 0777)
	return examplesDir
}
