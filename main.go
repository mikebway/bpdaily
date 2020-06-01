// A command line utility to collate time stamped records read from a CSV file
// into one line per day with the results sorted in ascending date order. The
// timestamp is assumed to be in the first column.
//
// Copyright © 2020 Michael D Broadway <mikebway@mikebway.com>
//
// Licensed under the ISC License (ISC)
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mikebway/bpdaily/dlycsv"
)

// WARNING: THIS IS A VERY CRUDE IMPLEMENTATION. NO OPTIONS. NO FINESSE.

var (
	unitTesting  = false // True if unit testing and NOT to os.Exit from the main function
	executeError error   // The error value obtained by Execute(), captured for unit test purposes
)

// Command line entry point.
func main() {

	// There must be two arguments!
	if len(os.Args) == 3 {

		// Translate the input CSV file into the output CSV file
		// but don't overwrite the output file if it already exists
		executeError = dlycsv.ConvertBloodPressureCSVToDaily(os.Args[1], os.Args[2], false)

	} else {
		executeError = errors.New(`
Usage: 

  bpdaily input-file-path.csv output-file-path
	
`)
	}

	// Display any error that occured
	if executeError != nil {
		fmt.Printf("ERROR - %v\n", executeError.Error())

		// Do not exit if we are unit testing
		if !unitTesting {
			os.Exit(1)
		}
	}
}
