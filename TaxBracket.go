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

func ReadTaxBracketsConfig(inputFile string) ([]*IncomeTaxBracket, error) {
	fileHandle, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	csvReader := csv.NewReader(fileHandle)
	brackets := []*IncomeTaxBracket{}

	i := 0
	prev_upper := 0.0
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}

			return brackets, err
		}

		// TODO: consider sending entire input row to recod creation function
		if len(row) < 5 { // need at least five fields in a valid record, also check for empty fields
			// return or collect error and move to next record
			return nil, fmt.Errorf("readTaxBrackets(): Minimum number of fields not met in input <%s>\n", row)
		}

		lower, err_low := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)

		top := false // flag indicating topmost tax bracket
		upp := strings.TrimSpace(row[1])
		upper := 0.0
		var err_upp error
		if upp != "" {
			// upper field could be empty if this is the top bracket
			upper, err_upp = strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		} else {
			// set flag indicating uppermost bracket
			top = true
		}

		percent, err_perc := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		lump, err_lump := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
		threshold, err_thr := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)

		if err_low != nil || err_upp != nil || err_perc != nil || err_lump != nil || err_thr != nil {
			fmt.Printf("%v\n", err_thr)
			return nil, fmt.Errorf("readTaxBrackets(): Error reading tax bracket config record: <%s>\n", row)
		}

		if lower >= upper && !top {
			return nil, fmt.Errorf("readTaxBrackets(): Lower limit >= upper limit in input <%s>\n", row)
		}

		if i == 0 {
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
