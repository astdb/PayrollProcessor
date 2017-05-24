// Go tests are placed in files with the pattern *_test.go. Test Methods have the signature func <TestMethod>(t *testing.T) The tests are run with "go test" command.

package main

import (
	"TaxBracket"
	"strings"
	"testing"
)

// tests for readPayrollRecords(inputFile string) ([]*PayrollRecord, error)
func TestReadPayrollRecords(t *testing.T) {
	// set of valid input files should produce a nil error
	files := []string{"test_input_01.csv"}

	for _, fn := range files {
		_, err := readPayrollRecords(fn)

		if err != nil {
			t.Errorf("FAILED: error reading payroll input records file %v", err)
		}
	}
}

// tests for PayrollRecord.fullName()
func TestFullName(t *testing.T) {
	// create some test payroll record objects and expected fullnames
	tests := map[*PayrollRecord]string{}

	var prr *PayrollRecord
	prr, _ = createPayrollRecord(strings.Split("David,Rudd,60050,9%,01 March – 31 March", ","))
	tests[prr] = "David Rudd"
	prr, _ = createPayrollRecord(strings.Split("Ryan,Chen,120000,10%,01 March – 31 March", ","))
	tests[prr] = "Ryan Chen"
	prr, _ = createPayrollRecord(strings.Split("Marcus,Aurelius,850000,25%,01 May – 31 May", ","))
	tests[prr] = "Marcus Aurelius"
	prr, _ = createPayrollRecord(strings.Split("Georgy,Zukhov,850000,25%,01 May – 31 May", ","))
	tests[prr] = "Georgy Zukhov"
	prr, _ = createPayrollRecord(strings.Split("John, Citizen,850000,25%,01 May – 31 May", ","))
	tests[prr] = "John Citizen"

	// test
	for prr, expected := range tests {
		got := prr.fullName()
		if prr.Valid {
			if got != expected {
				t.Errorf("FAILED: PayrollRecord.fullName() = %v: expected <%s>", got, expected)
			}
		} else {
			t.Errorf("FAILED: invalid payroll record")
		}
	}

}

// tests for PayrollRecord.payPeriod()
func TestPayPeriod(t *testing.T) {
	// create some test payroll record objects and expected pay periods
	tests := map[*PayrollRecord]string{}

	var prr *PayrollRecord
	prr, _ = createPayrollRecord(strings.Split("David,Rudd,60050,9%,01 March – 31 March", ","))
	tests[prr] = "01 March – 31 March"
	prr, _ = createPayrollRecord(strings.Split("Ryan,Chen,120000,10%,01 Apr – 30 Apr", ","))
	tests[prr] = "01 Apr – 30 Apr"
	prr, _ = createPayrollRecord(strings.Split("Marcus,Aurelius,850000,25%,01 May – 31 May", ","))
	tests[prr] = "01 May – 31 May"
	prr, _ = createPayrollRecord(strings.Split("Georgy,Zukhov,850000,25%,01 Dec – 31 Dec", ","))
	tests[prr] = "01 Dec – 31 Dec"
	prr, _ = createPayrollRecord(strings.Split("John, Citizen,850000,25%,01 Jun – 30 Jun", ","))
	tests[prr] = "01 Jun – 30 Jun"

	// test
	for prr, expected := range tests {
		got := prr.payPeriod()
		if prr.Valid {
			if got != expected {
				t.Errorf("FAILED: PayrollRecord.payPeriod() = %v: expected <%s>", got, expected)
			}
		} else {
			t.Errorf("FAILED: invalid payroll record")
		}
	}
}

// tests for PayrollRecord.grossIncome()
func TestGrossIncome(t *testing.T) {
	tests := map[*PayrollRecord]float64{}

	var prr *PayrollRecord
	prr, _ = createPayrollRecord(strings.Split("David,Rudd,60050,9%,01 March – 31 March", ","))
	tests[prr] = 5004.0
	prr, _ = createPayrollRecord(strings.Split("Ryan,Chen,120000,10%,01 Apr – 30 Apr", ","))
	tests[prr] = 10000.0
	prr, _ = createPayrollRecord(strings.Split("Marcus,Aurelius,850000,25%,01 May – 31 May", ","))
	tests[prr] = 70833.0
	prr, _ = createPayrollRecord(strings.Split("Georgy,Zukhov,32185,25%,01 Dec – 31 Dec", ","))
	tests[prr] = 2682.0
	prr, _ = createPayrollRecord(strings.Split("John, Citizen,895642,25%,01 Jun – 30 Jun", ","))
	tests[prr] = 74637.0

	// test
	for prr, expected := range tests {
		got := prr.grossIncome()

		if prr.Valid {
			if got != expected {
				t.Errorf("FAILED: PayrollRecord.grossIncome() = %v: expected <%.2f>", got, expected)
			}
		} else {
			t.Errorf("FAILED: invalid payroll record")
		}
	}
}

// tests for PayrollRecord.incomeTax()
func TestIncomeTax(t *testing.T) {
	tests := map[*PayrollRecord]float64{}

	var prr *PayrollRecord
	prr, _ = createPayrollRecord(strings.Split("David,Rudd,60050,9%,01 March – 31 March", ","))
	tests[prr] = 922.0
	// prr, _ = createPayrollRecord(strings.Split("David,Rudd,0,9%,01 March – 31 March", ","))
	// tests[prr] = 0.0
	prr, _ = createPayrollRecord(strings.Split("Ryan,Chen,120000,10%,01 Apr – 30 Apr", ","))
	tests[prr] = 2696.0
	prr, _ = createPayrollRecord(strings.Split("Marcus,Aurelius,850000,25%,01 May – 31 May", ","))
	tests[prr] = 29671.0
	prr, _ = createPayrollRecord(strings.Split("Georgy,Zukhov,32185,25%,01 Dec – 31 Dec", ","))
	tests[prr] = 221.0
	prr, _ = createPayrollRecord(strings.Split("John, Citizen,895642,25%,01 Jun – 30 Jun", ","))
	tests[prr] = 31382.0

	// test
	taxConfigFile := "TAX_CONFIG.csv"
	taxBrackets, err := TaxBracket.ReadTaxBracketsConfig(taxConfigFile)

	if err != nil {
		t.Errorf("FAILED: error loading tax brackets config: %v", err)
	}

	for prr, expected := range tests {
		got, err := prr.incomeTax(taxBrackets)

		if err == nil {
			if prr.Valid {
				if got != expected {
					t.Errorf("FAILED: PayrollRecord.incomeTax() = %v: expected <%.2f>", got, expected)
				}
			} else {
				t.Errorf("FAILED: invalid payroll record")
			}
		} else {
			t.Errorf("FAILED: error calculating income tax: %v", err)
		}
	}
}

// tests for PayrollRecord.netIncome()
func TestNetIncome(t *testing.T) {
	tests := map[*PayrollRecord]float64{}

	var prr *PayrollRecord
	prr, _ = createPayrollRecord(strings.Split("David,Rudd,60050,9%,01 March – 31 March", ","))
	tests[prr] = 4082.0
	// prr, _ = createPayrollRecord(strings.Split("David,Rudd,0,9%,01 March – 31 March", ","))
	// tests[prr] = 0.0
	prr, _ = createPayrollRecord(strings.Split("Ryan,Chen,120000,10%,01 Apr – 30 Apr", ","))
	tests[prr] = 7304.0
	prr, _ = createPayrollRecord(strings.Split("Marcus,Aurelius,850000,25%,01 May – 31 May", ","))
	tests[prr] = 41162.0
	prr, _ = createPayrollRecord(strings.Split("Georgy,Zukhov,32185,25%,01 Dec – 31 Dec", ","))
	tests[prr] = 2461.0
	prr, _ = createPayrollRecord(strings.Split("John, Citizen,895642,25%,01 Jun – 30 Jun", ","))
	tests[prr] = 43255.0

	// test
	taxConfigFile := "TAX_CONFIG.csv"
	taxBrackets, err := TaxBracket.ReadTaxBracketsConfig(taxConfigFile)

	if err != nil {
		t.Errorf("FAILED: error loading tax brackets config: %v", err)
	}

	for prr, expected := range tests {
		got, err := prr.netIncome(taxBrackets)

		if err == nil {
			if prr.Valid {
				if got != expected {
					t.Errorf("FAILED: PayrollRecord.netIncome() = %v: expected <%.2f>", got, expected)
				}
			} else {
				t.Errorf("FAILED: invalid payroll record")
			}
		} else {
			t.Errorf("FAILED: error calculating net income: %v", err)
		}
	}
}

// tests for PayrollRecord.SuperAmount()
func TestSuperAmount(t *testing.T) {
	tests := map[*PayrollRecord]float64{}

	var prr *PayrollRecord
	prr, _ = createPayrollRecord(strings.Split("David,Rudd,60050,9%,01 March – 31 March", ","))
	tests[prr] = 450.0
	// prr, _ = createPayrollRecord(strings.Split("David,Rudd,0,9%,01 March – 31 March", ","))
	// tests[prr] = 0.0
	prr, _ = createPayrollRecord(strings.Split("Ryan,Chen,120000,10%,01 Apr – 30 Apr", ","))
	tests[prr] = 1000.0
	prr, _ = createPayrollRecord(strings.Split("Marcus,Aurelius,850000,25%,01 May – 31 May", ","))
	tests[prr] = 17708.0
	prr, _ = createPayrollRecord(strings.Split("Georgy,Zukhov,32185,25%,01 Dec – 31 Dec", ","))
	tests[prr] = 671.0
	prr, _ = createPayrollRecord(strings.Split("John, Citizen,895642,25%,01 Jun – 30 Jun", ","))
	tests[prr] = 18659.0

	// test
	for prr, expected := range tests {
		got, err := prr.superAmount()

		if err == nil {
			if prr.Valid {
				if got != expected {
					t.Errorf("FAILED: PayrollRecord.SuperAmount() = %v: expected <%.2f>", got, expected)
				}
			} else {
				t.Errorf("FAILED: invalid payroll record")
			}
		} else {
			t.Errorf("FAILED: error calculating super: %v", err)
		}
	}
}

// tests for PayrollRecord.Print()
func TestPrint_PayrollRecord(t *testing.T) {

}

// tests for PayrollRecord.CreatePayrollRecord()
func TestCreatePayrollRecord(t *testing.T) {
	
}

// tests for round() routine
func TestRound(t *testing.T) {
	var tests = []struct {
		input float64
		want  float64
	}{
		{0.0, 0.0},
		{0.3, 0.0},
		{0.5, 1.0},
		{0.6, 1.0},
		{10.3, 10.0},
		{10.5, 11.0},
		{10.8, 11.0},
	}

	for _, test := range tests {
		if got := round(test.input); got != test.want {
			t.Errorf("FAILED: round(%f) = %v", test, got)
		}
	}
}

// test writeOutputFile(inFileName string, records []*PayrollRecord, taxBrackets []*TaxBracket.IncomeTaxBracket) error
func TestWriteOutputFile(t *testing.T) {

}

// -------------------- Tests for TaxBracket package --------------------------

// test TaxBracket.Print()
func TestWritePrint_TaxBracket(t *testing.T) {

}

// test TaxBracket.Print()
func TestReadTaxBracketsConfig(t *testing.T) {

}
