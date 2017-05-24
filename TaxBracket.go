// Package TaxBracket provides structures and methods to read and store a set of income tax brackets in memory from disk file.
package TaxBracket

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Type struct representing an income tax bracket
type IncomeTaxBracket struct {
	Lower   float64 // lower salary limit of tax bracket
	Upper   float64 // upper salary limit of tax bracket
	Percent float64 // percentage tax to be levied above a certain threshold
	Lump    float64 // lump sum to be paid for this bracket, if any
	Above   float64 // threshold above which percentage tax has to be paid
}

// utility method to print out a tax bracket - mostly used for debug purposes
func (itb *IncomeTaxBracket) Print() {
	fmt.Printf("[Lower: $%.2f, Upper: $%.2f, Percent: %.2f%%, Lump Sum: $%.2f, Above: $%.2f]\n", itb.Lower, itb.Upper, itb.Percent, itb.Lump, itb.Above)
}

// ReadTaxBracketsConfig takes in a config file name and reads in a set of tax bracket configurations
// The config file is expected to be a comma-separatd values file
// Each row would define an income tax bracket, row format as below:
// lower_income_limit, upper_income_limit, percentage_payable_over_threshold, lumpsum, threshold
func ReadTaxBracketsConfig(inputFile string) ([]*IncomeTaxBracket, error) {
	// open specified config file and handle any error encountered
	fileHandle, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close() // defrer file closure to function exit

	csvReader := csv.NewReader(fileHandle) // initialize CSV reader
	brackets := []*IncomeTaxBracket{}      // initialize empty slice ofIncomeTaxBracket struct references to store read-in brackets

	i := 0            // counter to keep track of number or rows read
	prev_upper := 0.0 // placecholder to keep track of the previously-read bracket's upper income limit

	// for each row in tx brackets config file
	for {
		row, err := csvReader.Read() // read row
		if err != nil {
			if err == io.EOF {
				err = nil // if end of file reached, set error to nil
			}

			return brackets, err // return read in tax bracket data with any encountered error (error -> nil for EOF)
		}

		if len(row) < 5 { // need at least five fields in a valid record, also check for empty fields
			// return or collect error and move to next record
			return nil, fmt.Errorf("readTaxBrackets(): Minimum number of fields not met in input <%s>\n", row)
		}

		// read in first field of row: lower limit of income for this tax bracket
		lower, err_low := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)

		top := false                     // flag indicating topmost tax bracket
		upp := strings.TrimSpace(row[1]) // tentatively read in first field of row: lower limimt of income for this tax bracket
		upper := 0.0                     // placeholder for actual upper income limit
		var err_upp error                // placeholder for any read error
		if upp != "" {
			// if value defined for second field, read it in: upper taxable income limit for this bracket
			upper, err_upp = strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		} else {
			// upper field could be empty if this is the topmost bracket (e.g. 180,000 and up)
			// set flag indicating uppermost bracket
			top = true
		}

		percent, err_perc := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)  // read in tax percentage value: third field of row
		lump, err_lump := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)     // read in lump sum value: fourth field of row
		threshold, err_thr := strconv.ParseFloat(strings.TrimSpace(row[4]), 64) // read in tax percentage value: fifth field of row

		// ensure all values were read in without error
		if err_low != nil || err_upp != nil || err_perc != nil || err_lump != nil || err_thr != nil {
			return nil, fmt.Errorf("readTaxBrackets(): Error reading tax bracket config record: <%s>\n", row)
		}

		// sanity check - upper limit cannot be smaller than lower limit
		if lower >= upper && !top {
			return nil, fmt.Errorf("readTaxBrackets(): Lower limit >= upper limit in input <%s>\n", row)
		}

		// if this is the first row
		if i == 0 {
			// first bracket lower limit must be zero
			if lower != 0 {
				return nil, fmt.Errorf("readTaxBrackets(): First bracket lower limit != 0 in input <%s>\n", row)
			}
		} else {
			if lower <= prev_upper {
				return nil, fmt.Errorf("readTaxBrackets(): Current bracket's lower limit <= previous upper limit in input <%s>\n", row)
			}
		}

		prev_upper = upper
		newTaxBracket := &IncomeTaxBracket{lower, upper, percent, lump, threshold}
		brackets = append(brackets, newTaxBracket)
		i++

		if top {
			break
		}
	}

	if i == 0 {
		return nil, fmt.Errorf("readTaxBrackets(): No valid tax brackets found\n")
	}

	return brackets, nil
}
