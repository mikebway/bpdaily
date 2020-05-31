# bpdaily - Organize Omron Blood Pressure History for Charting

`bpdaily` is a command line utility for processing Omron blood pressure reading history
CSV files from one line per reading to one line per day of readings, concatenating
multiplre reads from the same day in time sequence order on one line.

Additionally, the output file is sorted in ascending date order rather than having the
most recent reading at the top as rendered by the Omron phone app's export function.

Why do this? The consolidated daily CSV files are much easier to chart with Excel,
Numbers, or Google Sheets: multiple points can be shown for the same day.

## Usage

```bash
bputil <output-file.csv> <output-file-path.csv>
```

## Unit Testing

Unit test coverage should be kept above 90% by line for all packages.

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
