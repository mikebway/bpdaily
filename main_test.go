package main

// Unit tests for the slogs S3 read functions
//
// Copyright © 2020 Michael D Broadway <mikebway@mikebway.com>
//
// Licensed under the ISC License (ISC)

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// beforeEach should be run at the start of each test to ensure that the main
// package has been initialized for unit testing.
func beforeEach() {
	unitTesting = true
	executeError = nil
}

// TestTooFewParameters checks that the program will object if less than two parameters
// are provided.
func TestTooFewParameters(t *testing.T) {

	// Make sure the main() function does not exit altogether
	beforeEach()

	// Replace the argument list with one of our own with only a single file parameter (and a program name)
	os.Args = []string{
		"TestTooFewParameters",      // Our fake program name
		"./ThereIsNo/InputFile.csv", // An input file that does not exist
	}

	// Run the program
	main()

	// There should be an error reporting an invalid parameter count
	require.NotNil(t, executeError, "should have failed for too few parameters")
	require.Contains(t, executeError.Error(), "Please provide two arguments")
}

// TestTooManyParameters checks that the program will object if more than two parameters
// are provided.
func TestTooManyParameters(t *testing.T) {

	// Make sure the main() function does not exit altogether
	beforeEach()

	// Replace the argument list with one of our own with three parameters (and a program name)
	os.Args = []string{
		"TestTooManyParameters",      // Our fake program name
		"./ThereIsNo/InputFile.csv",  // An input file that does not exist
		"./ThereIsNo/OutputFile.csv", // An output file that does not exist
		"OneTooManyParameters",       // An unwanted extra parameter
	}

	// Run the program
	main()

	// There should be an error reporting an invalid parameter count
	require.NotNil(t, executeError, "should have failed for too many parameters")
	require.Contains(t, executeError.Error(), "Please provide two arguments")
}

// TestMissingInputFile checks that the program will object if the input file
// does not exist.
func TestMissingInputFile(t *testing.T) {

	// Make sure the main() function does not exit altogether
	beforeEach()

	// Replace the argument list with one of our own with only a single file parameter
	os.Args = []string{
		"TestMissingInputFile",       // Our fake program name
		"./ThereIsNo/InputFile.csv",  // An input file that does not exist
		"./ThereIsNo/OutputFile.csv", // An output file that does not exist
	}

	// Run the program
	main()

	// There should be an error reporting an invalid parameter count
	require.NotNil(t, executeError, "should have failed for input file not found")
	require.Contains(t, executeError.Error(), "could not open input file")
}
