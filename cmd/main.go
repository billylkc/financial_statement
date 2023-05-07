package main

import (
	"fmt"

	"github.com/billylkc/financial_statement/src/src"
)

func main() {
	code := 316
	err := src.GetFinancialRatio(code)
	if err != nil {
		fmt.Println(err)

	}

}
