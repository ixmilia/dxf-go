package dxf

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type odaHelper struct {
	OdaPath    string
	TempPath   string
	InputPath  string
	OutputPath string
}

func newOdaHelper() (helper *odaHelper, err error) {
	odaPath, err := getOdaPath()
	if err != nil {
		return
	}

	tempPath, err := ioutil.TempDir("", "OdaIntegrationTest")
	if err != nil {
		return
	}

	inputPath := filepath.Join(tempPath, "input")
	outputPath := filepath.Join(tempPath, "output")

	err = os.MkdirAll(inputPath, os.ModePerm)
	if err != nil {
		return
	}

	err = os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return
	}

	helper = &odaHelper{
		OdaPath:    odaPath,
		TempPath:   tempPath,
		InputPath:  inputPath,
		OutputPath: outputPath,
	}

	return
}

func odaVersionString(t *testing.T, v AcadVersion) string {
	switch v {
	case R9:
		return "ACAD9"
	case R10:
		return "ACAD10"
	case R12:
		return "ACAD12"
	case R13:
		return "ACAD13"
	case R14:
		return "ACAD14"
	case R2000:
		return "ACAD2000"
	case R2004:
		return "ACAD2004"
	case R2007:
		return "ACAD2007"
	case R2010:
		return "ACAD2010"
	case R2013:
		return "ACAD2013"
	case R2018:
		return "ACAD2018"
	default:
		t.Errorf("Unsupported drawing version: %s", v)
		return "UNSUPPORTED"
	}
}

func (o *odaHelper) convertDrawing(t *testing.T, d *Drawing, v AcadVersion) (result Drawing) {
	err := d.SaveFile(filepath.Join(o.InputPath, "drawing.dxf"))
	if err != nil {
		t.Fatalf("Unable to save drawing to temporarly location %v", o.InputPath)
	}

	args := []string{
		o.InputPath,
		o.OutputPath,
		odaVersionString(t, v),
		"DXF", // output file type
		"0",   // recurse
		"1",   // audit
	}
	cmdErr := exec.Command(o.OdaPath, args...).Run()

	errors := ""
	errorFiles, _ := filepath.Glob(o.OutputPath)
	for _, errorFile := range errorFiles {
		bytes, _ := ioutil.ReadFile(errorFile)
		errors += string(bytes)
	}

	if len(errors) > 0 {
		t.Fatalf("Error converting files:\n%v", errors)
	}

	if cmdErr != nil {
		t.Fatal("Unable to convert drawing file")
	}

	result, err = ReadFile(filepath.Join(o.OutputPath, "drawing.dxf"))
	if err != nil {
		t.Fatalf("Unable to re-read converted drawing: %v", err)
	}

	return
}

func (o *odaHelper) cleanup(t *testing.T) {
	if t.Failed() {
		t.Logf("Temporary file location: %v", o.TempPath)
	} else {
		os.RemoveAll(o.TempPath)
	}
}

func getOdaPath() (odaPath string, err error) {
	// Find ODA converter.  Final path looks something like:
	//   C:\Program Files\ODA\ODAFileConverter_title 20.12.0\ODAFileConverter.exe
	matches, err := filepath.Glob("C:/Program Files/ODA/ODAFileConverter*/ODAFileConverter.exe")
	if err != nil {
		// don't really care about the error, just quit
		return
	}

	if len(matches) == 0 {
		// nothing found, don't care
		err = errors.New("non-fatal, no ODA converter found")
		return
	}

	odaPath = matches[0]
	return
}
