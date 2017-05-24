// PayrollProcessor reads in a set of payroll records from an input CSV file and writes processed data per requirements onto an output CSV file.
// The code uses internal struct PayrollRecord to model input records and implemets a number of associated methods to that type in order to
// calculate output parameters. Tax bracket information is stored in a configuration file, TAX_CONFIG.csv. The tax bracket information is managed
// by a separate custom package, TaxBracket. It reads the tax bracket data and makes it available within the program as required. 

package main

// import required external pakages
import (
	"TaxBracket" // custom packageL provides functionality to read external tax bracket connfiguration
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	payrollRecords, err := readPayrollRecords(inFile)
	if err != nil {
		fmt.Printf("Error reading payroll record input: %v\n", err)
	}

	// once data is read in, pass them into along with input filename and tax bracket information to write output file (CSV)
	err = writeOutputFile(inFile, payrollRecords, taxBrackets)
	if err != nil {
		fmt.Printf("Error writing payroll record output: %v", err)
	}
}

// ------------ support functions and structures ----------------

// readPayrollRecords reads in a set of employe payroll info from a specified file and returns that data along with any error encountered
func readPayrollRecords(inputFile string) ([]*PayrollRecord, error) {
	// open input file
	fileHandle, err := os.Open(inputFile)
	if err != nil {
		return nil, err // if an error encountered opening file, return error to caller with nil data (go methods can return multiple values)
	}
	defer fileHandle.Close() // defer file closure so file will auto-close at function return

	// prepare new CSV reader and slice of PayrollRecord objects to read in data
	csvReader := csv.NewReader(fileHandle)
	records := []*PayrollRecord{}

	// per row in input CSV file
	for {
		row, err := csvReader.Read() // read row
		if err != nil {
			if err == io.EOF {
				err = nil // if EOF, set error to nil: in that case we'll return nil error and the read set of records
			}

			// return set of records read, along with any error encountered
			return records, err
		}

		// send read-in row to create new payroll input record object
		newPayrollRecord, err := createPayrollRecord(row)
		if err != nil {
			// output error and move onto next record
			fmt.Printf("Error creating payroll input record object %v\n", err)
			continue
		}

		records = append(records, newPayrollRecord)
	}
}

// struct representing a payroll record from  the input file: a struct is a composite data type which can have associated methods
type PayrollRecord struct {
	FirstName    string
	LastName     string
	AnnualSalary float64
	SuperRate    float64
	PaymentDate  string
	Valid        bool   //	indicates if the record object is valid
	ErrorStr     string // if Valid == false, contains the input data from the input file leading to invalid object
	TaxBrackets  []*TaxBracket.IncomeTaxBracket
}

// -------- methods associated with the PayrollRecord struct -----------

// get full name for this payroll record
func (rec *PayrollRecord) fullName() string {
	return rec.FirstName + " " + rec.LastName
}

// get pay period for this payroll record
func (rec *PayrollRecord) payPeriod() string {
	return rec.PaymentDate
}

// get gross income for this payroll record
func (rec *PayrollRecord) grossIncome() float64 {
	return round(rec.AnnualSalary / 12)
}

// calculate monthly income tax for this payroll record (takes the income tax bracket data provided by the TaxBracket package)
func (rec *PayrollRecord) incomeTax(taxBrackets []*TaxBracket.IncomeTaxBracket) (float64, error) {
	// iterate through tax brackets and find the right tax percentage, limit above which percentage tax is payable and any lump sum payable for this salary amount
	perc := 0.0
	lump := 0.0
	abv := 0.0
	set := false // flag denoting that a fitting tax bracket was found

	// for each tax bracket configured
	for _, brac := range taxBrackets {
		if brac.Upper == 0 {
			// upper limit would be automatically set to zero for topmost tax bracket - check if salary is bigger than top bracket lower
			if rec.AnnualSalary >= brac.Lower {
				perc = brac.Percent
				lump = brac.Lump
				abv = brac.Above
				set = true
			}

		} else {
			// if not the topmost tax bracket, check if the salary falls between the lower and upper limist of this bracket
			if rec.AnnualSalary >= brac.Lower && rec.AnnualSalary <= brac.Upper {
				perc = brac.Percent
				lump = brac.Lump
				abv = brac.Above
				set = true
			}
		}
	}

	if !set {
		// no fitting bracket found for this salary - return error
		return -1.0, fmt.Errorf("No fitting tax bracket was found for salary amount %f", rec.AnnualSalary)
	}

	// percentage annual tax payable is the set percentage of percentage taxable portion
	percentageTax := ((rec.AnnualSalary - abv) * perc) / 100

	// add any applicable lump payment to annual percentage tax and divide by 12 to get monthly payable tax - use round() to round to given specification
	return round((percentageTax + lump) / 12), nil
}

// calculate net income value for this salary (and return any error)
func (rec *PayrollRecord) netIncome(taxBrackets []*TaxBracket.IncomeTaxBracket) (float64, error) {
	gross := rec.grossIncome()             // gross monthly income
	tax, err := rec.incomeTax(taxBrackets) // monthly tax payable

	if err != nil {
		return -1.0, err // return error if any encountered calculating income tax
	}

	// sanity check to ensure tax payable isn't larger than gross income
	if tax > gross {
		return -1.0, fmt.Errorf("Taxed amount (%f) larger than gross income (%f)", tax, gross)
	}

	// return net income value
	return round(gross - tax), nil
}

// calculate superannuation
func (rec *PayrollRecord) superAmount() (float64, error) {
	if rec.SuperRate < 0 || rec.SuperRate > 50 {
		return -1.0, fmt.Errorf("Invalid super rate (%f)", rec.SuperRate)
	}

	super := (rec.grossIncome() * rec.SuperRate) / 100
	return round(super), nil
}

// print payroll input record, mostly for debug puposes
func (rec *PayrollRecord) Print() {
	if rec.Valid {
		fmt.Printf("%s, %s, %.2f, %.2f%%, %s, %f\n", rec.FirstName, rec.LastName, rec.AnnualSalary, rec.SuperRate, rec.PaymentDate, rec.grossIncome())
	} else {
		fmt.Printf("Invalid record %s\n", rec.ErrorStr)
	}
}

// createPayrollRecord takes a string slice (input row from input records file), creates a payroll record struct instance, and returns a reference to it
func createPayrollRecord(inputRow []string) (*PayrollRecord, error) {
	// ensure input CSV row has minimum required number of fields (first, last, annual, super, startdate)
	if len(inputRow) < 5 {
		return nil, fmt.Errorf("Input row must have atleast five fields %v", inputRow)
	}

	// declare new empty record struct instance
	var newRecord PayrollRecord

	// prepare input
	FirstName := strings.TrimSpace(inputRow[0])
	LastName := strings.TrimSpace(inputRow[1])
	AnnualSalary, err_sal := strconv.ParseFloat(inputRow[2], 64)
	SuperRate := strings.TrimSpace(inputRow[3])
	PaymentDate := strings.TrimSpace(inputRow[4])

	// extract numeric super percentage value e.g. 50 from "50%"
	rexp, _ := regexp.Compile("^[0-9]+")                                      // use regular expression to match numeric portion
	SuperRate_f, err_sr := strconv.ParseFloat(rexp.FindString(SuperRate), 64) // convert it to numeric value

	// sanity check values
	// if an error is encountered and the program is unable to create a valid record, a struct instance with
	// Valid attribute set to false will be returned with an error string rather than aborting.
	if FirstName == "" || LastName == "" || (err_sal != nil) || (err_sr != nil) || AnnualSalary <= 0 || SuperRate_f < 0 || SuperRate_f > 50 {
		newRecord.Valid = false
		newRecord.ErrorStr = fmt.Sprintf("Invalid input record: [%s] [%s] [%s] [%s] [%s]", inputRow[0], inputRow[1], inputRow[2], inputRow[3], inputRow[4])
		return &newRecord, fmt.Errorf("Invalid data in payroll input record\n")
	}

	// assign values to struct instance's fields
	newRecord.FirstName = FirstName
	newRecord.LastName = LastName
	newRecord.AnnualSalary = AnnualSalary
	newRecord.SuperRate = SuperRate_f
	newRecord.PaymentDate = PaymentDate
	newRecord.Valid = true

	// return reference to struct and nil error
	return &newRecord, nil
}

// ------- standalone utility functions ---------------

// round() round values to whole dollar (if  >= .50 round up, else round down)
func round(n float64) float64 {
	n_floor := math.Floor(n) // extract whole value
	n_dec := n - n_floor     // extract decimal portion

	// round up if >= .50
	if n_dec >= .50 {
		// round up
		return n_floor + 1.0
	}

	// else round down
	return n_floor
}

// writeOutputFile() takes the input filename, slice of read-in payroll structs and tax bracket config and writes the required output file (CSV)
func writeOutputFile(inFileName string, records []*PayrollRecord, taxBrackets []*TaxBracket.IncomeTaxBracket) error {
	// sanity check input filename
	if strings.TrimSpace(inFileName) == "" {
		return fmt.Errorf("Invalid input filename")
	}

	// create output filename e.g. input.csv -> input-out.csv
	outFileName := strings.Split(strings.TrimSpace(inFileName), ".")[0] + "-out.csv"
	f, err := os.Create(outFileName)

	if err != nil {
		return fmt.Errorf("Error creating outputfile <%s>: %v", outFileName, err) // return error if encountered creating file
	}

	defer f.Close() // defer file closure to function exit

	// for each payroll input record
	for _, rec := range records {
		if rec.Valid { // if record is valid
			// get record data
			name := rec.fullName()
			payp := rec.payPeriod()
			gross := rec.grossIncome()
			tax, err := rec.incomeTax(taxBrackets)
			if err != nil {
				return fmt.Errorf("Error getting income tax: %v", err)
			}

			net, err := rec.netIncome(taxBrackets)
			if err != nil {
				return fmt.Errorf("Error getting net income: %v", err)
			}

			super, err := rec.superAmount()
			if err != nil {
				return fmt.Errorf("Error getting net income: %v", err)
			}

			// write output
			_, err = f.WriteString(fmt.Sprintf("%s, %s, %.0f,%.0f, %.0f, %.0f\n", name, payp, gross, tax, net, super))
			if err != nil {
				return fmt.Errorf("Error writing CSV output: %v", err)
			}
		} else {
			// invalid record
			_, err := f.WriteString("Invalid payroll record: no output.\n")
			if err != nil {
				return fmt.Errorf("Error writing CSV output: %v", err)
			}
		}
	}

	return nil
}
