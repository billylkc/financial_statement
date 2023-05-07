package financial_statement

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/billylkc/financial_statement/src"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

func Blah() error {

	code := 11
	err := src.GetFinancialRatio(code)
	if err != nil {
		return err
	}
	return nil
}

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

func dev() error {

	return nil
}

func getFinancialRatio(code int) error {

	// https://github.com/jedib0t/go-pretty

	/*
	   - Financial Ratio http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_ratio.php?code=316
	   - Income Statement http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_pl.php?code=316
	   - Financial Position http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_bs.php?code=316
	   - Cash Flow Statement http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_cashflow.php?code=316&quarter=latest
	*/

	// Financial position
	url := fmt.Sprintf("https://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_ratio.php?code=%d", code)
	r, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		log.Fatalf("failed to fetch data: %d %s", r.StatusCode, r.Status)
	}
	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var rowArrs [][]string
	doc.Find(".figureTable").Each(func(tbID int, tablehtml *goquery.Selection) {
		if tbID <= 10 {
			tablehtml.Find("tr").Each(func(trID int, rowhtml *goquery.Selection) {
				var row []string
				rowhtml.Find("td").Each(func(tdID int, tdhtml *goquery.Selection) {
					t := strings.TrimSpace(tdhtml.Text())

					// handle Year
					if (tbID == 0) && (t == "") {
						t = ""
					}
					t = strings.ReplaceAll(t, "  ", " ") // replace double spacing, with single
					row = append(row, t)
				})

				if len(row) > 0 { // Rows with data only
					rowArrs = append(rowArrs, row)
				}
			})
		}
	})

	form, m, err := getFormData()
	if err != nil {
		return err
	}

	// Handle table header (year)
	var header []string
	header = rowArrs[0]
	rowArrs = rowArrs[1:]
	_ = header

	// Fill crawled web data back to form
	var missing []string // For missing keys
	for _, row := range rowArrs {
		if len(row) == 0 {
			panic("something went wrong, row length is zero")
		}

		key := strings.TrimSpace(row[0])
		val := row[1:]
		if idx, ok := m[key]; ok {
			form[idx].Numbers = parseNumbers(val)
		} else {
			// warning

			missing = append(missing, key)

		}
	}
	fmt.Printf("key not found. Please check. %s.\n", strings.Join(missing, ", "))

	// Remove any unmatched metrics for readability
	var nonEmptyForm FormData
	for _, row := range form {
		// only append those with values
		if len(row.Numbers) > 0 {
			nonEmptyForm = append(nonEmptyForm, row)
		}
	}

	// Create form data with tr row
	var (
		newForm   FormData // Form with tr rows
		prevGroup string   // prev group
	)
	for _, row := range nonEmptyForm {
		currentGroup := row.Group

		if currentGroup == prevGroup {
			newForm = append(newForm, row)

		} else {
			// Add group and separator, if group changed
			tr := RowData{Title: currentGroup}
			sep := RowData{Title: "~"} // special character for easy replace for org format

			if prevGroup != "" {
				newForm = append(newForm, sep)
			}
			prevGroup = currentGroup
			newForm = append(newForm, tr)
			newForm = append(newForm, sep)
			newForm = append(newForm, row)
		}

	}
	res := newForm.Render(header)
	fmt.Println(res)
	return nil
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
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
	n := Number{}
	s = strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), ",", "")

	// Handle null, na values
	if s == "N/A" {
		return n
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
		n.Float = f
	}
	return n
}