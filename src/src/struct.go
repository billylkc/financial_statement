package src

import (
	"fmt"
	"strings"

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

func (rd RowData) ToReadableArray(nrecords int) []string {
	var arr []string
	arr = append(arr, rd.Title)

	if len(rd.Numbers) == 0 { // Fill array up to nrecords
		arr = append(arr, generateEmptyArray(nrecords-1)...)
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
	table.Render() // Send output
	res := sb.String()

	// replace separator to fit org table style
	res = strings.ReplaceAll(res, "-|-", "-+-") // middle separator
	res = strings.ReplaceAll(res, "| ~", "|- ") // row separator

	return res
}
