package csvlint

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// CSVError returns information about an invalid record in a CSV file
type CSVError struct {
	// Record is the invalid record. This will be nil when we were unable to parse a record.
	Record []string
	// Num is the record number of this record.
	Num int
	// Line is the line number of the error
	Line int
	// Column is the column index of the error, if applicable
	Column *int
	err    error
}

// Error implements the error interface
func (e CSVError) Error() string {
	message := fmt.Sprintf("Record #%d has error: %s on line %d", e.Num, e.err.Error(), e.Line)

	if e.Column != nil {
		return fmt.Sprintf("%s, column %d", message, *e.Column)
	}
	return message
}

func lines_in_record(record []string) int {
	result := 0
	for _, field := range record {
		result += strings.Count(field, "\n")
	}
	// We must also count the newline at the end of the record
	return result + 1
}

// Validate tests whether or not a CSV lints according to RFC 4180.
// The lazyquotes option will attempt to parse lines that aren't quoted properly.
func Validate(reader io.Reader, delimiter rune, lazyquotes bool) ([]CSVError, bool, error) {
	r := csv.NewReader(reader)
	r.TrailingComma = true
	r.FieldsPerRecord = -1
	r.LazyQuotes = lazyquotes
	r.Comma = delimiter

	var header []string
	errors := []CSVError{}
	records := 0
	line_number := 0
	for {
		record, err := r.Read()
		if header != nil {
			records++
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			parsedErr, ok := err.(*csv.ParseError)
			if !ok {
				return errors, true, err
			}
			errors = append(errors, CSVError{
				Record: nil,
				Num:    records,
				Line:   parsedErr.Line,
				Column: &parsedErr.Column,
				err:    parsedErr.Err,
			})
			return errors, true, nil
		}
		if header == nil {
			header = record
		} else if len(record) != len(header) {
			errors = append(errors, CSVError{
				Record: record,
				Num:    records,
				Line:   line_number,
				err:    csv.ErrFieldCount,
			})
		}

		line_number += lines_in_record(record)
	}
	return errors, false, nil
}
