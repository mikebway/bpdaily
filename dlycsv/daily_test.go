package dlycsv

// Unit tests for the dlycsv package.
//
// Copyright © 2020 Michael D Broadway <mikebway@mikebway.com>
//
// Licensed under the ISC License (ISC)

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// A structure used to contain all of the file paths used in a single test
type TestFilePaths struct {
	InputPath    string // The input CSV file path
	OutputPath   string // The path at which to write the output file
	ExpectedPath string // The path to a file that contains exactly the data that we expected the output file to contain
}

// TestHappyPath processes a messy but still legal blood pressure CSV file.
// Messy in that it includes invalid lines and blank lines
func TestHappyPath(t *testing.T) {

	// The file paths that we work with in this test (and the owverite tests)
	filePaths := buildHappyFilePaths()

	// Make sure the output file does not exist
	err := removeFile(filePaths.OutputPath)
	require.Nil(t, err, "could not delete output file: %v", err)

	// Fake the arguments for processing the happy path input file
	os.Args[1] = filePaths.InputPath
	os.Args[2] = filePaths.OutputPath

	// Run the target function
	err = ConvertBloodPressureCSVToDaily(filePaths.InputPath, filePaths.OutputPath, false)
	require.Nil(t, err, "ConvertBloodPressureCSVToDaily returned an error: %v", err)

	// Confirm that the output obtained matches that expected
	err = outputIsAsExpected(filePaths)
	require.Nil(t, err, "output content did not match expected: %v", err)
}

// TestNoOverwrite confirms that an existing output file will not be overwritten if
// we did not ask for it to be.
func TestNoOverwrite(t *testing.T) {

	// Fetch the happy file paths - we will try write to the output of the hapy path test
	filePaths := buildHappyFilePaths()

	// Run the happy path test to make sure that the output file exists
	TestHappyPath(t)

	// Knowing the output file exists - confirm that a second run would fail for that reason
	err := ConvertBloodPressureCSVToDaily(filePaths.InputPath, filePaths.OutputPath, false)
	require.NotNil(t, err, "should have failed because output file already exists")
	require.Contains(t, err.Error(), "output file already exists")
}

// TestOverwrite confirms that an existing output file be overwritten if we ask for it to be.
func TestOverwrite(t *testing.T) {

	// We work with two sets of file paths in this test
	happyFilePaths := buildHappyFilePaths()
	overwriteFilePaths := buildTestFilePaths("../testdata/overwrite")

	// Run the happy path test to make sure that the output file exists
	TestHappyPath(t)

	// Blend the happy output path (a file we know exists) with our overwrite file
	// paths to get a set that will have different output than the happy path test
	// but written to the same file
	overwriteFilePaths.OutputPath = happyFilePaths.OutputPath

	// Knowing the output file exists - confirm that a second run will overwrite it when asked to
	err := ConvertBloodPressureCSVToDaily(overwriteFilePaths.InputPath, overwriteFilePaths.OutputPath, true)
	require.Nil(t, err, "ConvertBloodPressureCSVToDaily returned an error: %v", err)

	// Confirm that the overwritten output obtained matches that expected
	err = outputIsAsExpected(overwriteFilePaths)
	require.Nil(t, err, "output content did not match expected: %v", err)
}

// TestNoOverwriteDir confirms that we get an error if we try to overwite a directory
func TestNoOverwriteDir(t *testing.T) {

	// We work with two sets of file paths in this test (making the target a directory)
	filePaths := buildHappyFilePaths()
	filePaths.OutputPath = "../testdata"

	// Knowing the output file is a directory - confirm that we get an error if we try to write to it
	err := ConvertBloodPressureCSVToDaily(filePaths.InputPath, filePaths.OutputPath, true)
	require.NotNil(t, err, "expected error because output file is a directory")
	require.Contains(t, err.Error(), "cannot overwrite a directory")
}

// TestMissingInput confirms that the appropriate error is returned if we ask to read from
// an input CSV file that does not exist.
func TestMissingInput(t *testing.T) {

	// We work with two sets of file paths in this test
	filePaths := buildTestFilePaths("../no-such/thing")

	// It does not matter that we are willing to overwrite the output file if there is no input file
	err := ConvertBloodPressureCSVToDaily(filePaths.InputPath, filePaths.OutputPath, true)
	require.NotNil(t, err, "expected error because input file did not exist")
	require.Contains(t, err.Error(), "could not open input file")
}

// TestEmptyInput confirms that the appropriate error is returned if we ask to read from
// an empty input CSV file.
func TestEmptyInput(t *testing.T) {

	// We work with two sets of file paths in this test
	filePaths := buildTestFilePaths("../testdata/empty")

	// You cannot convert an empty input file
	err := ConvertBloodPressureCSVToDaily(filePaths.InputPath, filePaths.OutputPath, true)
	require.NotNil(t, err, "expected error because input file is empty")
	require.Contains(t, err.Error(), "failed to read blood pressure CSV header record")
}

// TestBadHeader confirms that the appropriate error is returned if we ask to read from
// an input CSV file with the wrong column names in its header.
func TestBadHeader(t *testing.T) {

	// We work with two sets of file paths in this test
	filePaths := buildTestFilePaths("../testdata/badheader")

	// You cannot convert an input file with the wrong column names
	err := ConvertBloodPressureCSVToDaily(filePaths.InputPath, filePaths.OutputPath, true)
	require.NotNil(t, err, "expected error because input file has a bad header")
	require.Contains(t, err.Error(), "header record of input file does not match blood pressure CSV format")
}

// TestBadBody confirms that the appropriate error is returned if we ask to read from
// an input CSV file with a corrupt data body (missing or invalid fields).
func TestBadBody(t *testing.T) {

	// We work with two sets of file paths in this test
	filePaths := buildTestFilePaths("../testdata/badbody")

	// You cannot convert an input file with missing or too many data fields in some rows
	err := ConvertBloodPressureCSVToDaily(filePaths.InputPath, filePaths.OutputPath, true)
	require.NotNil(t, err, "expected error because input file has a bad data set")
	require.Contains(t, err.Error(), "failed to read body of input file")
}

// TestConversionOfEmptyRecords exercises the low level convertBPDateTimes(..)
// function to confirm that it would correctly handle empty records if the
// encoding/csv package ever changed its practice and failed to strip them
// when the file is read.
func TestConversionOfEmptyRecords(t *testing.T) {

	// Pass in two empty records and confirm that they get a discard marker
	// field added to them
	var records = make([][]string, 2, 2)
	convertBPDateTimes(&records)

	// We should have two entries now with teh discard marker in their first (and only) field
	require.Equal(t, len(records), 2, "there should still be only two records")
	require.Equal(t, len(records[0]), 1, "first record should have one field")
	require.Equal(t, records[0][0], discardMarker, "first record should have a discard marker")
	require.Equal(t, len(records[1]), 1, "second record should have one field")
	require.Equal(t, records[1][0], discardMarker, "second record should have a discard marker")
}

// buildHappyFilePaths constructs the file paths of the input, output, and expected
// comparison file for the happy path and overwrite tests.
func buildHappyFilePaths() *TestFilePaths {
	return buildTestFilePaths("../testdata/happypath")
}

// buildTestFilePaths constructs the file paths of the input, output, and expected
// comparison file by appending to given root path.
func buildTestFilePaths(rootPath string) *TestFilePaths {
	return &TestFilePaths{
		InputPath:    rootPath + ".in.csv",
		OutputPath:   rootPath + ".out.csv",
		ExpectedPath: rootPath + ".expected.csv",
	}
}

// removeFile deletes any existing file at the given path
func removeFile(filePath string) error {

	// Only bother if the file exists!
	_, err := os.Stat(filePath)
	if err == nil {
		return os.Remove(filePath)
	}

	// Only get here if the file does not exist or we could not stat it
	if !os.IsNotExist(err) {
		return err // Some weired error doing the stat operation
	}

	// All is good - the file does not exist to delete
	return nil
}

// outputIsAsExpected checks whether the output file and expected out file
// have the same content. Returns nil if the files are the same, and error
// if not.
func outputIsAsExpected(filePaths *TestFilePaths) error {

	// Open both files
	outputFile, outputFileErr := os.Open(filePaths.OutputPath)
	if outputFileErr != nil {
		return outputFileErr
	}
	defer outputFile.Close()
	expectedFile, expectedFileErr := os.Open(filePaths.ExpectedPath)
	if expectedFileErr != nil {
		return expectedFileErr
	}
	defer expectedFile.Close()

	// Establish line scanners for both files
	outputScanner := bufio.NewScanner(outputFile)
	outputScanner.Split(bufio.ScanLines)
	expectedScanner := bufio.NewScanner(expectedFile)
	expectedScanner.Split(bufio.ScanLines)

	// Load the line content of both files
	var outputlines []string
	for outputScanner.Scan() {
		outputlines = append(outputlines, outputScanner.Text())
	}
	var expectedlines []string
	for expectedScanner.Scan() {
		expectedlines = append(expectedlines, expectedScanner.Text())
	}

	// Check that both files have the same number of lines
	outputlineCount := len(outputlines)
	expectedlineCount := len(expectedlines)
	if outputlineCount != expectedlineCount {
		return fmt.Errorf("%s has %d lines, %s has %d lines",
			filePaths.OutputPath, outputlineCount, filePaths.ExpectedPath, expectedlineCount)
	}

	// Compare each line, quiting on a mismatch
	for i := 0; i < expectedlineCount; i++ {
		if outputlines[i] != expectedlines[i] {
			return fmt.Errorf("expected\n\t%s \nbut found\n\t%s", expectedlines[i], outputlines[i])
		}
	}

	// The files match!
	return nil
}
