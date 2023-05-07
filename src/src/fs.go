package src

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// - Financial Ratio http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_ratio.php?code=316
// - Income Statement http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_pl.php?code=316
// - Financial Position http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_bs.php?code=316
// - Cash Flow Statement http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_cashflow.php?code=316&quarter=latest
func GetFinancialRatio(code int) error {

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

func getFormData() (FormData, map[string]int, error) {
	// {"Group":"", "Title":""},
	jsonData := `
[
  {"Group":"Profitability Analysis", "Title":"ROAA (%)"},
  {"Group":"Profitability Analysis", "Title":"ROAE (%)"},
  {"Group":"Profitability Analysis", "Title":"Return on Capital Employed (%)"},

  {"Group":"Fundamental Indicators", "Title":"Net Interest Margin (%)"},
  {"Group":"Fundamental Indicators", "Title":"Net Interest Spread (%)"},
  {"Group":"Fundamental Indicators", "Title":"Capital Adequacy Ratio (%)"},
  {"Group":"Fundamental Indicators", "Title":"Tier 1 Capital Ratio (%)"},
  {"Group":"Fundamental Indicators", "Title":"Core Capital Ratio (%)"},
  {"Group":"Fundamental Indicators", "Title":"Liquidity Ratio (%)"},

  {"Group":"Margin Analysis",        "Title":"Gross Profit Margin (%)"},
  {"Group":"Margin Analysis",        "Title":"EBITDA Margin (%)"},
  {"Group":"Margin Analysis",        "Title":"Pre-tax Profit Margin (%)"},
  {"Group":"Margin Analysis",        "Title":"Net Profit Margin (%)"},

  {"Group":"Liquidity Analysis",     "Title":"Current Ratio (X)"},
  {"Group":"Liquidity Analysis",     "Title":"Quick Ratio (X)"},

  {"Group":"Operating Performance Analysis", "Title":"Interest Expense / Interest Income (%)"},
  {"Group":"Operating Performance Analysis", "Title":"Other Operating Income / Operating Income (%)"},
  {"Group":"Operating Performance Analysis", "Title":"Operating Expense / Operating Income (%)"},

  {"Group":"Loan & Depostis Analysis", "Title":"Liquidity / Customer Deposits (%)"},
  {"Group":"Loan & Depostis Analysis", "Title":"Provision for Bad Debt / Customer Loans (%)"},
  {"Group":"Loan & Depostis Analysis", "Title":"Loans / Deposits (%)"},
  {"Group":"Loan & Depostis Analysis", "Title":"Customer Loans / Total Assets (%)"},

  {"Group":"Leverage Analysis",      "Title":"Total Liabilities / Total Assets(%)"},
  {"Group":"Leverage Analysis",      "Title":"Total Liabilities / Shareholders' Funds (%)"},
  {"Group":"Leverage Analysis",      "Title":"Non-current Liabilities / Shareholders' Funds (%)"},
  {"Group":"Leverage Analysis",      "Title":"Interest Coverage Ratio (X)"},

  {"Group":"Efficiency Analysis",    "Title":"Inventory Turnover on Sales (Day)"}
]
`

	var (
		fd FormData
		m  map[string]int
	)
	// Parse json to form data
	err := json.Unmarshal([]byte(jsonData), &fd)
	if err != nil {
		return fd, m, err
	}

	// Assign index for ordering and look up
	for i := 0; i < len(fd); i++ {
		fd[i].Index = i
	}

	// build mapping
	m = make(map[string]int)
	for _, v := range fd {
		m[v.Title] = v.Index
	}
	return fd, m, nil

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
