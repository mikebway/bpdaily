package main

// Unit tests for the slogs S3 read functions
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

	// The file paths that we work with in this test
	filePaths := buildTestFilePaths("./testdata/happypath")

	// Make sure the output file does not exist
	err := removeFile(filePaths.OutputPath)
	require.Nil(t, err, "could not delete output file: %v", err)

	// Fake the arguments for processing the happy path input file
	os.Args[1] = filePaths.InputPath
	os.Args[2] = filePaths.OutputPath

	// Run the program
	main()

	// Confirm that the output obtained matches that expected
	err = outputIsAsExpected(filePaths)
	require.Nil(t, err, "output content did not match expected: %v", err)
}

// buildTestFilePaths constucts the file paths of the input, output, and expected
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
