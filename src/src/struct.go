package src

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/billylkc/financial_statement/src/utils"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

// Number is a struct to parse numeric numbers as int or float
type Number struct {
	Int   int64
	Float float64
}

// Row data is a self defined mapping to capture different metrics in groups
// e.g. Profitability Analysis -> [ROAA, ROAE, etc..]
type RowData struct {
	Index   int      `json:"Index"`  // Auto generated according to sequence
	Group   string   `json:"Group"`  // Profitability Analysis
	Title   string   `json:"Title"`  // ROAA (%)
	Numbers []Number `json:"Number"` // Metrics for different years in array
}

// Final tabular data scraped
type FormData []RowData

// ToReadableArray converts the Number struct to readable int or float
func (rd RowData) ToReadableArray(nrecords int) []string {
	var arr []string
	arr = append(arr, rd.Title)

	if len(rd.Numbers) == 0 { // Fill array up to nrecords
		arr = append(arr, utils.GenerateEmptyArray(nrecords-1)...)
	}

	for _, num := range rd.Numbers {
		var a string
		if num.Float == 0.0 {
			if num.Int == 0 {
				a = ""
			} else {
				a = humanize.Comma(num.Int)
			}
		} else {
			a = fmt.Sprintf("%.3f", num.Float)
		}
		arr = append(arr, a)
	}
	return arr
}

// Render renders the data and return printable org table
func (fd *FormData) Render(header []string) string {
	sb := &strings.Builder{}
	table := tablewriter.NewWriter(sb)

	// Parse Form data to list of strings
	var data [][]string
	nrecords := len(header)
	for _, rd := range *fd {
		data = append(data, rd.ToReadableArray(nrecords))
	}

	// Generate table
	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetCenterSeparator("|")
	table.SetCaption(true, "blah")

	for _, v := range data {
		table.Append(v)
	}
	var (
		res    []string
		result string
	)
	table.Render() // Send output
	tableStr := sb.String()

	// replace separator to fit org table style

	for _, line := range strings.Split(tableStr, "\n") {
		if strings.HasPrefix(line, "| ~") {
			line = strings.ReplaceAll(line, "~", "-")
			line = strings.ReplaceAll(line, " ", "-")
		}

		res = append(res, line)
	}
	result = strings.Join(res, "\n")
	result = strings.ReplaceAll(result, "-|-", "-+-") // middle separator
	return result
}

func parseNumbers(ss []string) []Number {
	var (
		number []Number
		num    Number
	)
	for _, s := range ss {
		num = parseNumber(s)
		number = append(number, num)
	}
	return number
}

func parseNumber(s string) Number {
	var isPercentage bool // Check for percentage sign

	n := Number{}
	s = strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), ",", "")

	// Handle null, na values
	if s == "N/A" {
		return n
	}

	// Handle percentage
	if strings.Contains(s, "%") {
		s = strings.ReplaceAll(s, "%", "")
		isPercentage = true
	}

	// Handle negative values
	if strings.Contains(s, "(") {
		s = strings.ReplaceAll(s, "(", "")
		s = strings.ReplaceAll(s, ")", "")
		s = "-" + s // add negative sign
	}

	if i, err := strconv.Atoi(s); err == nil {
		n.Int = int64(i)
	} else if f, err := strconv.ParseFloat(s, 64); err == nil {
		if isPercentage {
			f = f / 100
		}
		n.Float = f
	}
	return n
}
