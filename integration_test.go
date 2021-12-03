package dxf

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRoundTripSingleLineR12(t *testing.T) {
	drawing := *NewDrawing()
	line := NewLine()
	line.P1 = Point{X: 0.0, Y: 0.0, Z: 0.0}
	line.P2 = Point{X: 1.0, Y: 1.0, Z: 0.0}
	drawing.Entities = append(drawing.Entities, line)

	roundTrippedDrawing := roundTripAutoCad(&drawing, R12, t)
	assertEqInt(t, 1, len(roundTrippedDrawing.Entities))
	roundTrippedLine := roundTrippedDrawing.Entities[0].(*Line)
	assertEqPoint(t, line.P1, roundTrippedLine.P1)
	assertEqPoint(t, line.P2, roundTrippedLine.P2)
}

func roundTripAutoCad(d *Drawing, version AcadVersion, t *testing.T) Drawing {
	acadPath := pathToAcad()
	if acadPath == "" {
		// acad.exe not found, return identity
		return *d
	}

	tempPath := getTempPath()
	err := os.MkdirAll(tempPath, 0777)
	if err != nil {
		t.Errorf("Could not create temp directory: %s", err)
	}

	d.Header.Version = version
	err = d.SaveFile(path.Join(tempPath, "input.dxf"))
	if err != nil {
		t.Errorf("Could not save file: %s", err)
	}

	var script strings.Builder
	script.WriteString("FILEDIA 0\n")
	script.WriteString(fmt.Sprintf("DXFIN \"%s\"\n", path.Join(tempPath, "input.dxf")))
	script.WriteString(fmt.Sprintf("DXFOUT \"%s\" V %s 16\n", path.Join(tempPath, "output.dxf"), versionToAcadVersion(version)))
	script.WriteString("FILEDIA 1\n")
	script.WriteString("QUIT Y\n")
	f, err := os.Create(path.Join(tempPath, "script.scr"))
	if err != nil {
		t.Errorf("Could not create script file: %s", err)
	}
	_, err = f.WriteString(script.String())
	if err != nil {
		t.Errorf("Could not write script file: %s", err)
	}
	err = f.Close()
	if err != nil {
		t.Errorf("Could not close script file: %s", err)
	}

	acadCommand := exec.Command(acadPath, "/b", path.Join(tempPath, "script.scr"))
	err = acadCommand.Start()
	if err != nil {
		t.Errorf("Error starting AutoCAD: %v", err)
	}

	err = acadCommand.Wait()
	if err != nil {
		t.Errorf("Error running acad.exe: %v", err)
	}

	roundTrippedDrawing, err := ReadFile(path.Join(tempPath, "output.dxf"))
	if err != nil {
		t.Errorf("Error reading output file: %v", err)
	}

	return roundTrippedDrawing
}

func pathToAcad() string {
	matches, err := filepath.Glob("C:/Program Files/Autodesk/AutoCAD */acad.exe")
	if matches == nil || err != nil {
		return ""
	}

	return matches[0]
}

func versionToAcadVersion(version AcadVersion) string {
	switch version {
	case R12:
		return "R12"
	case R2000:
		return "2000"
	case R2004:
		return "2004"
	case R2007:
		return "2007"
	case R2010:
		return "2010"
	case R2013:
		return "2013"
	case R2018:
		return "2018"
	}

	panic(fmt.Sprintf("Unsupported version %s", version))
}

func getTempPath() string {
	_, filename, _, _ := runtime.Caller(0)
	thisDir := path.Dir(filename)
	nanos := time.Now().UnixNano()
	tempPath := path.Join(thisDir, "integration_tests", fmt.Sprint(nanos))
	return tempPath
}
