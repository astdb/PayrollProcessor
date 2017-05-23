package main

import (
	"TaxBracket"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// if inputfiles aren't provided on command line, show usage message and abort
	if len(os.Args) <= 2 {
		fmt.Println("Usage: > go run PayrollProcessor.go <inputfile> <taxconfigfile>")
		return
	}

	inFile := os.Args[1]			// 
	taxConfigFile := os.Args[2]

	taxBrackets, err := TaxBracket.ReadTaxBracketsConfig(taxConfigFile)
	if err != nil {
		// error reading tax bracket config - quit
		fmt.Printf("Error reading tax brackets config: %v\n", err)
		return
	}

	for _, v := range taxBrackets {
		v.Print()
	}

	payrollRecords, err := readPayrollRecords(inFile)
	if err != nil {
		fmt.Printf("Error reading payroll record input: %v\n", err)
	}

	// for k, v := range payrollRecords {
	// 	fmt.Printf("%2d. ", k)

	// 	if v != nil {
	// 		v.Print()
	// 	} else {
	// 		fmt.Printf("Nil record object\n")
	// 	}
	// }

	err = writeOutputFile(inFile, payrollRecords, taxBrackets)
	if err != nil {
		fmt.Printf("Error writing payroll record output: %v", err)
	}
}

func readPayrollRecords(inputFile string) ([]*PayrollRecord, error) {
	fileHandle, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	csvReader := csv.NewReader(fileHandle)
	records := []*PayrollRecord{}

	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}

			// TODO: returning at the first erroneous row? can we skip over til EOF?
			return records, err
		}

		// TODO: consider sending entire input row to recod creation function
		if len(row) < 5 { // need at least five fields in a valid record, also check for empty fields
			// return or collect error and move to next record
			// return records, fmt.Errorf("Minimum number of fields not met in input <%s>\n", row)
			continue // read next row
		}

		newPayrollRecord, err := createPayrollRecord(row[0], row[1], row[2], row[3], row[4])
		if err != nil {
			// throw error
			// fmt.Errorf("Error creating payroll record with input %s\n", row)
			// return records, fmt.Errorf("Minimum number of fields not met in input <%s>\n", row)
			continue // read next row
		}

		records = append(records, newPayrollRecord)
	}
}

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

func (rec *PayrollRecord) fullName() string {
	return rec.FirstName + " " + rec.LastName
}

func (rec *PayrollRecord) payPeriod() string {
	return rec.PaymentDate
}

func (rec *PayrollRecord) grossIncome() float64 {
	return round(rec.AnnualSalary / 12)
}

func (rec *PayrollRecord) incomeTax(taxBrackets []*TaxBracket.IncomeTaxBracket) (float64, error) {
	// iterate through tax brackets and find the right tax percentage for this salary
	perc := 0.0
	lump := 0.0
	abv := 0.0
	set := false
	// fmt.Println("Entering income tax bracket loop")
	for _, brac := range taxBrackets {
		fmt.Printf("Curret bracket: [Lower: %f, Upper: %f]\n", brac.Lower, brac.Upper)
		if brac.Upper == 0 {
			// top tax bracket
			if rec.AnnualSalary >= brac.Lower {
				perc = brac.Percent
				lump = brac.Lump
				abv = brac.Above
				set = true
			}

		} else {
			if rec.AnnualSalary >= brac.Lower && rec.AnnualSalary <= brac.Upper {
				perc = brac.Percent
				lump = brac.Lump
				abv = brac.Above
				set = true
			}
		}
	}

	if !set {
		// no fitting bracket found for this salary
		return -1.0, fmt.Errorf("No fitting tax bracket was found for salary amount %f", rec.AnnualSalary)
	}

	percentageTax := ((rec.AnnualSalary - abv) * perc) / 100
	return round((percentageTax + lump) / 12), nil
}

func (rec *PayrollRecord) netIncome(taxBrackets []*TaxBracket.IncomeTaxBracket) (float64, error) {
	gross := rec.grossIncome()
	tax, err := rec.incomeTax(taxBrackets)

	if err != nil {
		return -1.0, err
	}

	if tax > gross {
		return -1.0, fmt.Errorf("Taxed amount (%f) larger than gross income (%f)", tax, gross)
	}

	return round(gross - tax), nil
}

func (rec *PayrollRecord) superAmount() (float64, error) {
	if rec.SuperRate < 0 || rec.SuperRate > 50 {
		return -1.0, fmt.Errorf("Invalid super rate (%f)", rec.SuperRate)
	}

	super := (rec.grossIncome() * rec.SuperRate) / 100
	return round(super), nil
}

func (rec *PayrollRecord) Print() {
	if rec.Valid {
		fmt.Printf("%s, %s, %.2f, %.2f%%, %s, %f\n", rec.FirstName, rec.LastName, rec.AnnualSalary, rec.SuperRate, rec.PaymentDate, rec.grossIncome())
	} else {
		fmt.Printf("Invalid record %s\n", rec.ErrorStr)
	}
}

func createPayrollRecord(firstname_str string, lastname_str string, annualsalary_str string, superrate_str string, paymentdate_str string) (*PayrollRecord, error) {
	var newRecord PayrollRecord

	// prepare input
	FirstName := strings.TrimSpace(firstname_str)
	LastName := strings.TrimSpace(lastname_str)
	AnnualSalary, err_sal := strconv.ParseFloat(annualsalary_str, 64)
	SuperRate := strings.TrimSpace(superrate_str)
	PaymentDate := strings.TrimSpace(paymentdate_str)

	// validate [TODO: may improve by returning specific error, consider ignoring erroneous record and reading ahead]
	// extract numeric super percentage
	rexp, _ := regexp.Compile("^[0-9]+")
	SuperRate_f, err_sr := strconv.ParseFloat(rexp.FindString(SuperRate), 64)

	if FirstName == "" || LastName == "" || (err_sal != nil) || (err_sr != nil) || AnnualSalary <= 0 || SuperRate_f < 0 || SuperRate_f > 50 {
		newRecord.Valid = false
		newRecord.ErrorStr = fmt.Sprintf("Invalid input record: [%s] [%s] [%s] [%s] [%s]", firstname_str, lastname_str, annualsalary_str, superrate_str, paymentdate_str)
		return &newRecord, fmt.Errorf("Invalid data in payroll input record\n")
	}

	newRecord.FirstName = FirstName
	newRecord.LastName = LastName
	newRecord.AnnualSalary = AnnualSalary
	newRecord.SuperRate = SuperRate_f
	newRecord.PaymentDate = PaymentDate
	newRecord.Valid = true

	return &newRecord, nil
}

func round(n float64) float64 {
	n_floor := math.Floor(n)
	n_dec := n - n_floor

	if n_dec >= .50 {
		// round up
		return n_floor + 1.0
	}

	// round down
	return n_floor
}

func writeOutputFile(inFileName string, records []*PayrollRecord, taxBrackets []*TaxBracket.IncomeTaxBracket) error {
	if strings.TrimSpace(inFileName) == "" {
		return fmt.Errorf("Invalid input filename")
	}

	outFileName := strings.Split(strings.TrimSpace(inFileName), ".")[0] + "-out.csv"
	f, err := os.Create(outFileName)

	if err != nil {
		return fmt.Errorf("Error creating outputfile <%s>: %v", outFileName, err)
	}

	defer f.Close()

	for _, rec := range records {
		if rec.Valid {
			// print
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

			_, err = f.WriteString(name + ", " + payp + ", " + strconv.Itoa(int(gross)) + ", " + strconv.Itoa(int(tax)) + ", " + strconv.Itoa(int(net)) + ", " + strconv.Itoa(int(super)) + "\n")
			if err != nil {
				return fmt.Errorf("Error writing CSV output: %v", err)
			}
		}
	}

	return nil
}
