// Package dlycsv provides functions to collate time stamped records read from a CSV
// file of blood pressure readings, one reading per line, into one line per day with
// potentially many readings concatenated onto that one line.
//
// The input file format is expected to be that exported from Omron blod pressure
// tracking smart phone application.
//
// The output is sorted in ascending date order, ready to be imported into Excel,
// Numbers, or Google Sheets for plotting in a chart.
//
// Copyright © 2020 Michael D Broadway <mikebway@mikebway.com>
//
// Licensed under the ISC License (ISC)
package dlycsv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

// All records that are to be thrown away later will be tagged with a ZZZZ value in their first field
const discardMarker = "ZZZZ"

// ConvertBloodPressureCSVToDaily reads the blood pressure CSV file at the input path, sorts the
// data, then gathers lines that are for the same day into a single line, sending the results to
// a new CSV file at the output path. If the output file alraedy exists, it will only be
// overwritten if the overwrite flag is true.
func ConvertBloodPressureCSVToDaily(inputPath, outputPath string, overwrite bool) error {

	// If we cannot write to the output file for any knowable reason
	// then we should not waste any time processing the input data
	if err := canWeWriteToFile(outputPath, overwrite); err != nil {
		return fmt.Errorf("output file already exists: %w", err)
	}

	// Open the input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("could not open input file: %w", err)
	}
	defer inputFile.Close()

	// Obtain a buffered CSV reader on the input file
	reader := csv.NewReader(bufio.NewReader(inputFile))

	// Handoff to our siblig to do the rest
	return checkForHeaderRecord(reader, outputPath)
}

// canWeWriteToFile determines, the the best of our ability at this point, whether
// we can write to the output file. This may fail for several reasons, returning an error
// explaining why if we cannot.
func canWeWriteToFile(filePath string, overwrite bool) error {

	// Can we stat the file?
	if fileInfo, err := os.Stat(filePath); err == nil {

		// The file exists - is it a directory?
		if fileInfo.Mode().IsDir() {

			// Never mind the overight flag - we can never write to a dorectory
			return fmt.Errorf("cannot overwrite a directory: %s", filePath)
		}

		// If we are not allowed overwrite an existing file - we can't write to this file
		if !overwrite {
			return fmt.Errorf("cannot overwrite existing file (%s): %w", filePath, err)
		}

	} else if !os.IsNotExist(err) {

		// There was an error and the error was not that the file did not exist
		// (Yeah, weird but it can happen in some worlds that the stat operation
		// fails because the file does not exist)
		return fmt.Errorf("cannot access output file (%s): %w", filePath, err)
	}

	// If we got to this point then all is well and the caller is free to go ahead
	// and try to do whtever they like with the target file
	return nil
}

// checkForHeaderRecord checks that the first input record is a valid blood presssure
// column name header record and then hands off to the next step in the flow.
func checkForHeaderRecord(reader *csv.Reader, outputPath string) error {

	// Read the first line of the input CSV file - it should be column titles
	headerRecord, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read blood pressure CSV header record: %w", err)
	}

	// Confirm that the header record contains the expected values for a blood pressure history
	if len(headerRecord) != 5 ||
		headerRecord[0] != "Date Time" ||
		headerRecord[1] != "Systolic" ||
		headerRecord[2] != "Diastolic" ||
		headerRecord[3] != "Pulse" ||
		headerRecord[4] != "Note" {
		return fmt.Errorf("header record of input file does not match blood pressure CSV format")
	}

	// Now that we have confirmed that we have a blood pressure CSV file we can
	// go on to the next phase
	return openOutputFile(reader, outputPath)
}

// openOutputFile opens the output file, truncating any existing content
// then hands off to the next step in the flow.
func openOutputFile(reader *csv.Reader, outputPath string) error {

	// Open the output file, recreating/emptying it if it already exists
	outputFile, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	// We can safely close the file on exit since the CSV writer used further down
	// the stack flushes output
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)

	// Have our deeper sibling do the remainder of the reading and writing
	return sortInput(reader, writer)
}

// sortInput loads the rest of the input file, sorts those records into ascending order,
// then hands off to the next step in the flow.
func sortInput(reader *csv.Reader, writer *csv.Writer) error {

	// Load the input CSV data (excluding the already processed inputHeader)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read body of input file: %w", err)
	}

	// Convert the date time value in each record into a sortable format
	convertBPDateTimes(&records)

	// Sort the records into descending order
	sort.Slice(records, func(i, j int) bool { return records[i][0] < records[j][0] })

	// Combine records for the same date into single records
	maxReadingsInOneDay := combineRecordsForSameDay(&records)

	// Write a header record, repeating the column names to match the most readings for a single day
	header := buildHeaderRecord(maxReadingsInOneDay)
	err = writer.Write(header)
	if err != nil {
		return fmt.Errorf("failed to write header to output file: %w", err)
	}

	// Eliminate all the records records marked for discard
	discardMarkedRecords(&records)

	// Make sure we flush the writer when we are done
	defer writer.Flush()

	// Write the body of the data
	err = writer.WriteAll(records)
	if err != nil {
		return fmt.Errorf("failed to write blood pressure data to output file: %w", err)
	}

	// Glorious - we are completely finished
	return nil
}

// combineRecordsForSameDay merges consecutive records for the same day onto the end of
// the first record for that date, marking the following records as discardable.
//
// Returns the maxium number of readings accumulated into a single day.
func combineRecordsForSameDay(records *[][]string) int {

	// Assume that we will accumulate the second record into the first
	accumulateIndex := 0

	// When we find the first date string, this is where we will track
	// the date we are accumulating for
	var accumulateDate string

	// How many readinsg have we accumulated in the current target so far
	readingsAccumulatedSoFar := 1

	// We also track the maxium number of readings found for the same day
	// so that we can put out a header record with column names to match
	maxReadingsInOneDay := 1

	// Loop through all of the records
	for index, record := range *records {

		// If we have reached a discardable record, we can stop looping.
		// Every record beyond this one will also be discardable
		if record[0] == discardMarker {
			break
		}

		// Collect the date portion of the first field
		recordDate := record[0][0:10]

		// Special case - the first record has to be stepped over as there
		// is no prior record to accumulate into. We will look to accumulate
		// into this one.
		if index == 0 {
			accumulateDate = recordDate
			continue
		}

		// The normal case, does the date of this record match the one we are
		// accumulating into?
		if recordDate == accumulateDate {

			// We have a match - append this record's fields to the accumulate record
			(*records)[accumulateIndex] = append((*records)[accumulateIndex], record...)

			// Keep track of the number of accumulations into the target record so far
			readingsAccumulatedSoFar++

			// Mark the current record for discard
			record[0] = discardMarker

			// Move on to the next record
			continue
		}

		// The current record does not match the date of the previous accumulation record
		// First lets see if we have a new high point for the number of readings made on a single day
		if readingsAccumulatedSoFar > maxReadingsInOneDay {
			maxReadingsInOneDay = readingsAccumulatedSoFar
		}

		// Tag the new guy as the accumulator from now on
		accumulateIndex = index
		accumulateDate = recordDate
		readingsAccumulatedSoFar = 1
	}

	// Return the maxium number of accumulations into a single day
	return maxReadingsInOneDay
}

// convertBPDateTimes converts the date-time values in the first field of each of
// the given blood pressure records to a sortable YYYY-MM-DD hh:mm:ss form.
//
// If any record is found not to contain a date value, its first field will be set to "ZZZZ"
// so that it can later be sorted to the end of the set and easily removed
func convertBPDateTimes(records *[][]string) {

	// Loop through all of the records
	for index, record := range *records {

		// Check we have a non-zero length record!
		if len(record) != 0 {

			// Convert the date time string in the first field to a time value
			datetime, err := time.Parse("Jan 02 2006 15:04:05", record[0])

			// If the first field was a valid date time, put it back in YYYY-MM-DD hh:mm:ss form
			if err == nil {

				// Convert the time value to our desired form and stuff it back in the record
				(*records)[index][0] = datetime.Format("2006-01-02 15:04:05")

			} else {

				// Darn - this reord is duff
				(*records)[index][0] = discardMarker
			}

		} else {

			// NOTE: This is a safety net, just in case. In practice the
			// encoding/csv package discards blank lines so we should never
			// get here.

			// We have an empty record, add the discard marker
			(*records)[index] = append((*records)[index], discardMarker)
		}
	}
}

// discardMarkedRecords eliminates all records records marked for discard.
func discardMarkedRecords(records *[][]string) {

	// Sort the records into descending order
	sort.Slice(*records, func(i, j int) bool { return (*records)[i][0] < (*records)[j][0] })

	// Start at the bottom and work back up to find the first legitimate record
	index := len(*records) - 1
	for ; index > 0; index-- {
		if (*records)[index][0] != discardMarker {
			break
		}
	}

	// Index is now the last good record, we discard the rest
	*records = (*records)[:index+1]
}

// buildHeaderRecord assembles one or more sets of blood pressure CSV file column headers
// into a string array record.
func buildHeaderRecord(maxReadingsInOneDay int) []string {

	// Build our header record here
	var header []string

	// Loop for the max reading count adding numbered header sections
	for i := 0; i < maxReadingsInOneDay; {

		// We want to start our numbering at 1 so increment the loop index inside the loop
		i++

		// Add a numbered set of column headings
		addHeadingSet(&header, i)
	}

	// And we have our finished header
	return header
}

// addHeadingSet appends one set of column names to the header record
func addHeadingSet(header *[]string, setNumber int) {

	// Convert the set number to text
	setNumberText := strconv.Itoa(setNumber)

	// Build an array of numbered heading names
	headingSet := []string{
		"Date Time " + setNumberText,
		"Systolic " + setNumberText,
		"Diastolic " + setNumberText,
		"Pulse " + setNumberText,
		"Note " + setNumberText,
	}

	// Add the set to the header
	*header = append(*header, headingSet...)
}
