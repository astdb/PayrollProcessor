// PayrollProcessor reads in a set of payroll records from an input CSV file and writes processed data per requirements onto an output CSV file.
// The code uses internal struct PayrollRecord to model input records and implemets a number of associated methods to that type in order to
// calculate output parameters. Tax bracket information is stored in a configuration file, TAX_CONFIG.csv. The tax bracket information is managed
// by a separate custom package, TaxBracket. It reads the tax bracket data and makes it available within the program as required.

package main

// import required external pakages
import (
	"PayrollRecord" // custom package providing functionality to manage payroll input records
	"TaxBracket"    // custom package provides functionality to read external tax bracket connfiguration
	"fmt"
	"os"
)

// ------------ main method ----------------
func main() {
	// if inputfiles aren't provided on command line, show usage message and abort
	if len(os.Args) <= 2 {
		fmt.Println("Usage: > go run PayrollProcessor.go <inputfile> <taxconfigfile>")
		return
	}

	inFile := os.Args[1]        // employee details input file
	taxConfigFile := os.Args[2] // tax bracket cnfiguration file

	// read tax bracket configs and handle any errors
	taxBrackets, err := TaxBracket.ReadTaxBracketsConfig(taxConfigFile)
	if err != nil {
		// error reading tax bracket config - abort with error
		fmt.Printf("Error reading tax brackets config: %v\n", err)
		return
	}

	// read employee payroll information (from CSV  format file)
	payrollRecords, err := PayrollRecord.ReadPayrollRecords(inFile)
	if err != nil {
		fmt.Printf("Error reading payroll record input: %v\n", err)
	}

	// once data is read in, pass them into along with input filename and tax bracket information to write output file (CSV)
	err = PayrollRecord.WriteOutputFile(inFile, payrollRecords, taxBrackets)
	if err != nil {
		fmt.Printf("Error writing payroll record output: %v", err)
	}
}
