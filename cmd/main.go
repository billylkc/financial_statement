package main

import (
	"fmt"

	"github.com/billylkc/financial_statement/src/src"
)

func main() {

	page := "financial_ratio" // financial_ration, income_statement
	page = "income_statement"
	page = "financial_position"

	code := 11
	table, err := src.GetFinancialData(page, code)
	if err != nil {
		fmt.Println(err)

	}

	fmt.Println(table)

}
