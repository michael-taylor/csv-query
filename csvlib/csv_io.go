package csvlib

import (
	"bufio"
	"errors"
	"io"
	"iter"
	"os"
	"strings"
)

// CSVParams contains configuration parameters for parsing a CSV-formatted file.
type CSVParams struct {
	Delimiter string
	Header    bool
	SkipLines int
}

// A CSVFile represents a CSV formatted file and provides operations for reading its lines.
type CSVFile struct {
	fileName   string
	file       *os.File
	reader     io.Reader
	fields     []string
	parameters CSVParams
}

// DefaultCSVParams constructs a CSVParams with safe default values.
func DefaultCSVParams() CSVParams {
	return CSVParams{
		Delimiter: ",",
		Header:    false,
		SkipLines: 0,
	}
}

// OpenCSVFile opens a CSV file with the given fileName.
// The parameters parameter contains the configuration options used when parsing the file.
// An error will be returned if the file doesn't exist or can't be opened.
func OpenCSVFile(fileName string, parameters CSVParams) (CSVFile, error) {
	if _, err := os.Stat(fileName); err != nil {
		return CSVFile{}, err
	}
	file, err := os.Open(fileName)
	if err != nil {
		return CSVFile{}, err
	}

	csv := CSVFile{fileName: fileName, file: file, reader: file, parameters: parameters}

	// Read field titles if header is present
	if parameters.Header {
		scanner := bufio.NewScanner(file)
		for range parameters.SkipLines {
			if !scanner.Scan() {
				return CSVFile{}, scanner.Err()
			}
		}
		if !scanner.Scan() {
			return CSVFile{}, errors.New("could not read header line")
		}
		csv.fields = strings.Split(scanner.Text(), parameters.Delimiter)
	}

	return csv, nil
}

// LoadCSVData reads CSV-formatted data from a string.
// The parameters parameter contains the configuration options used when parsing the file.
// An error will be returned if the file doesn't exist or can't be opened.
func LoadCSVData(data string, parameters CSVParams) (CSVFile, error) {
	reader := strings.NewReader(data)
	csv := CSVFile{file: nil, reader: reader, parameters: parameters}

	// Read field titles if header is present
	if parameters.Header {
		scanner := bufio.NewScanner(reader)
		for range parameters.SkipLines {
			if !scanner.Scan() {
				return CSVFile{}, scanner.Err()
			}
		}
		if !scanner.Scan() {
			return CSVFile{}, errors.New("could not read header line")
		}
		csv.fields = strings.Split(scanner.Text(), parameters.Delimiter)
	}

	return csv, nil
}

// Close closes the underlying file handle used by the CSVFile.
// It should be called with a `defer` after the OpenCSV function.
func (csv CSVFile) Close() {
	err := csv.file.Close()
	if err != nil {
		return
	}
}

// Headers returns an array of the field names if they exist. An empty array will be returned if they don't.
func (csv CSVFile) Headers() []string {
	return csv.fields
}

// Records returns an iterator that returns an array of string-formatted values for each record (line) in the CSV file.
func (csv CSVFile) Records() iter.Seq2[[]string, error] {
	skipLines := csv.parameters.SkipLines
	if csv.parameters.Header {
		skipLines += 1
	}

	scanner := bufio.NewScanner(csv.reader)
	for range skipLines {
		scanner.Scan()
	}

	numFields := len(csv.fields)

	return func(yield func([]string, error) bool) {
		for scanner.Scan() {
			fields := strings.Split(scanner.Text(), csv.parameters.Delimiter)
			if numFields == 0 {
				numFields = len(fields)
			} else if numFields != len(fields) {
				if !yield([]string{}, errors.New("incorrect number of fields")) {
					return
				}
			} else {
				if !yield(fields, nil) {
					return
				}
			}
		}
	}
}
