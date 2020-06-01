# bpdaily - Organize Omron Blood Pressure History for Charting

`bpdaily` is a command line utility for processing Omron blood pressure reading history
CSV files from one line per reading to one line per day of readings, concatenating
multiple readings from the same day into time sequence order on one line.

Additionally, the output file is sorted in ascending date order rather than having the
most recent reading at the top as rendered by the Omron phone app's export function.

Why do this? The consolidated daily CSV files are much easier to chart with Excel,
Numbers, or Google Sheets: multiple points can be shown for the same day.

## Usage

```bash
bputil <intput-file.csv> <output-file-path.csv>
```

## Possible Enhancements for the Future

There are so many but I am not likely to get around to them because the app does
what I need from it and too few other people will be interested in a command line
tool that only gets them half way to a pretty chart.

* Support other blood pressure machine types other than Omron

* Optionally: Allow time ranges to be defined (e.g. morning, afternoon, evening) and put
readings into the right position according to their time stamp

* Optionally: Discard multiple readings falling in the same time range.

* Optionally: When discarding readings, flag whether to keep the highest or lowest.

* Optionally: Allow the heart rate and notes columns to be excluded.

* Optionally: Make the first column a simple date (no time portion) and exclude the other
timestamps altogether.

* Generalize the `bpdaily/dlycsv` package to support processing of any CSV file
with timestmap columns and multiple records with the same date.

## Unit Testing

Unit test coverage should be kept above 90% by line for all packages if at all
possible. Sadly, that can't be done for the very short main package since the
one `os.Exit(1)` line cannot be covered without terminating the test run.

The unit tests are really more like integration tests in that they will invoke
AWS API calls though successful calls are only achieved through mocking.

You can run all of the unit tests from the command line and receive a coverage
report as follows:

```bash
go test -cover ./...
```

To ensure that all tests are run, and that none are assumed unchanged for the
cache of a previous run, you may add the `-count=1` flag to required that all
tests are run at least and exactly once:

```bash
go test -cover -count=1 ./...
```

For a more detailed reported, broken down by function and with a summary total 
for the complete project, two steps are required:

```bash
go test ./... -coverprofile cover.out
go tool cover -func cover.out
```

The `cover.out` file, and all files with the `.out` extention, are ignored by
git thanks to an entry in the `.gitignore` file.
