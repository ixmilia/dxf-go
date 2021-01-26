package dxf

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func roundTripDrawing(t *testing.T, d *Drawing) (result Drawing) {
	s := d.String()
	result = parse(t, s)
	return
}

func assert(t *testing.T, condition bool, message string) {
	if !condition {
		t.Error(message)
	}
}

func assertContains(t *testing.T, expected, actual string) {
	if !strings.Contains(actual, expected) {
		t.Errorf("Unable to find '%s' in '%s'", expected, actual)
	}
}

func assertNotContains(t *testing.T, notExpected, actual string) {
	if strings.Contains(actual, notExpected) {
		t.Errorf("Unexpectedly found '%s' in '%s'", notExpected, actual)
	}
}

func assertEqCodePairs(t *testing.T, expected, actual []CodePair) {
	isCorrect := len(actual) == len(expected)
	actualText := ""
	for _, p := range actual {
		actualText += p.String() + "\n"
	}
	expectedText := ""
	for _, p := range expected {
		expectedText += p.String() + "\n"
	}

	for i := 0; i < len(actual); i++ {
		isCorrect = isCorrect && reflect.DeepEqual(actual[i], expected[i])
	}

	if !isCorrect {
		t.Errorf("Expected code pairs [\n%s] but found [\n%s]", expectedText, actualText)
	}
}

func assertContainsCodePairs(t *testing.T, expected, actual []CodePair) {
	for i := 0; i < len(actual)-len(expected)+1; i++ {
		match := true
		for j := 0; j < len(expected) && match; j++ {
			match = match && reflect.DeepEqual(actual[i+j], expected[j])
		}

		if match {
			return
		}
	}

	expectedText := "\n"
	for _, p := range expected {
		expectedText += p.String() + "\n"
	}

	actualText := "\n"
	for _, p := range actual {
		actualText += p.String() + "\n"
	}

	t.Errorf("Unable to find %s in %s", expectedText, actualText)
}

func assertNotContainsCodePairs(t *testing.T, notExpected, actual []CodePair) {
	for i := 0; i < len(actual)-len(notExpected)+1; i++ {
		match := false
		for j := 0; j < len(notExpected) && !match; j++ {
			match = match || reflect.DeepEqual(actual[i+j], notExpected[j])
		}

		if match {
			notExpectedText := "\n"
			for _, p := range notExpected {
				notExpectedText += p.String() + "\n"
			}

			actualText := "\n"
			for _, p := range actual {
				actualText += p.String() + "\n"
			}

			t.Errorf("Unexpectedly found %s in %s", notExpectedText, actualText)
		}
	}
}

func assertEqShort(t *testing.T, expected, actual int16) {
	assert(t, expected == actual, fmt.Sprintf(expectedActualString("d"), expected, actual))
}

func assertEqInt(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected: %d\nActual: %d", expected, actual)
	}
}

func assertEqFloat64(t *testing.T, expected, actual float64) {
	assert(t, expected == actual, fmt.Sprintf(expectedActualString("f"), expected, actual))
}

func assertEqPoint(t *testing.T, expected, actual Point) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected.String(), actual.String())
	}
}

func assertEqVector(t *testing.T, expected, actual Vector) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected.String(), actual.String())
	}
}

func assertEqString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}

func assertEqUInt64(t *testing.T, expected, actual uint64) {
	assert(t, expected == actual, fmt.Sprintf(expectedActualString("X"), expected, actual))
}

func assertEqByteArray(t *testing.T, expected, actual []byte) {
	assert(t, len(expected) == len(actual), fmt.Sprintf(expectedActualString("d"), len(expected), len(actual)))
	for i := range expected {
		assert(t, expected[i] == actual[i], fmt.Sprintf("Difference at offset %d: "+expectedActualString("x"), i, expected[i], actual[i]))
	}
}

func join(vals ...string) string {
	return strings.Join(vals, "\r\n")
}

func parse(t *testing.T, content string) Drawing {
	drawing, err := ParseDrawing(strings.TrimSpace(content))
	if err != nil {
		t.Error(err)
	}

	return drawing
}

func parseFromCodePairs(t *testing.T, codePairs ...CodePair) Drawing {
	drawing, err := ParseDrawingFromCodePairs(codePairs...)
	if err != nil {
		t.Error(err)
	}

	return drawing
}

func expectedActualString(placeholder string) string {
	return fmt.Sprintf("Expected: %%%s\nActual %%%s", placeholder, placeholder)
}

func drawingCodePairs(t *testing.T, drawing Drawing) (codePairs []CodePair) {
	codePairs, err := drawing.CodePairs()
	if err != nil {
		t.Error(err)
	}

	return
}

func drawingCodePairsFromEntity(t *testing.T, entity Entity, version AcadVersion) (codePairs []CodePair) {
	drawing := *NewDrawing()
	drawing.Header.Version = version
	drawing.Entities = append(drawing.Entities, entity)
	return drawingCodePairs(t, drawing)
}
