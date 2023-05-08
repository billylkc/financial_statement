package src

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// - Financial Ratio http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_ratio.php?code=316
// - Income Statement http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_pl.php?code=316
// - Financial Position http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_bs.php?code=316
// - Cash Flow Statement http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_cashflow.php?code=316&quarter=latest
func GetFinancialData(page string, code int) (string, error) {

	var table string
	url, jsonData := getPageData(page, code)

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

	form, m, err := getFormData(jsonData)
	if err != nil {
		return table, err
	}

	// Handle table header (year)
	var header []string
	header = rowArrs[0]
	rowArrs = rowArrs[1:]
	_ = header

	// Fill crawled web data back to form
	// TODO: handle others and subtotal rows
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
	table = newForm.Render(header)

	return table, err
}

func getPageData(page string, code int) (string, string) {

	var (
		url      string
		jsonData string
	)

	switch page {
	case "financial_position":
		url = fmt.Sprintf("http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_bs.php?code=%d", code)
		jsonData = `
[
  {"Group":"Assets", "Title": "Cash & Short-Term Funds"},
  {"Group":"Assets", "Title": "Placings with Banks"},
  {"Group":"Assets", "Title": "Gov't Cert. of Indebtedness"},
  {"Group":"Assets", "Title": "Advances to Customers"},
  {"Group":"Assets", "Title": "Financial Assets at FVTPL"},
  {"Group":"Assets", "Title": "Financial Investments"},
  {"Group":"Assets", "Title": "Derivative Financial Assets"},
  {"Group":"Assets", "Title": "Interests in Asso. & JCEs"},
  {"Group":"Assets", "Title": "Intangible Assets"},
  {"Group":"Assets", "Title": "Investment Properties"},
  {"Group":"Assets", "Title": "Property, plant, equip. & others"},
  {"Group":"Assets", "Title": "Leasehold Land"},
  {"Group":"Assets", "Title": "Other Assets"},

  {"Group":"Current Assets", "Title": "Inventories"},
  {"Group":"Current Assets", "Title": "Cash & Bank Balances"},
  {"Group":"Current Assets", "Title": "Other Current Assets"},
  {"Group":"Current Assets", "Title": "Assets Held for Sale"},

  {"Group":"Non-current Assets", "Title": "Investment Properties"},
  {"Group":"Non-current Assets", "Title": "Property, plant, equip. & others"},
  {"Group":"Non-current Assets", "Title": "Leasehold Land"},
  {"Group":"Non-current Assets", "Title": "Intangible Assets"},
  {"Group":"Non-current Assets", "Title": "Interests in Asso. & JCEs"},
  {"Group":"Non-current Assets", "Title": "Other Non-current Assets"},



  {"Group":"Liabilities", "Title": "Currency Notes in Circulation"},
  {"Group":"Liabilities", "Title": "Bank Deposits"},
  {"Group":"Liabilities", "Title": "Customers Deposits"},
  {"Group":"Liabilities", "Title": "CD & Other Debt Securities Issued"},
  {"Group":"Liabilities", "Title": "Financial Liabilities at FVTPL"},
  {"Group":"Liabilities", "Title": "Derivative Financial Liabilities"},
  {"Group":"Liabilities", "Title": "Subordinated Liabilities"},
  {"Group":"Liabilities", "Title": "Other Liabilities"},
  {"Group":"Liabilities", "Title": ""},



  {"Group":"Current Liabilities", "Title": "Other Current Liabilities"},
  {"Group":"Current Liabilities", "Title": "Liab asso w/ Assets Held for Sale"},
  {"Group":"Current Liabilities", "Title": "Net Current Assets"},
  {"Group":"Current Liabilities", "Title": "Total Assets Less Current Liabilities"},

  {"Group":"Capital and Reserves", "Title": "Share Capital"},
  {"Group":"Capital and Reserves", "Title": "Reserves"},
  {"Group":"Capital and Reserves", "Title": "Others"},
  {"Group":"Capital and Reserves", "Title": "Shareholders' Funds"},
  {"Group":"Capital and Reserves", "Title": "Non-controlling Interests"},
  {"Group":"Capital and Reserves", "Title": "Others"},
  {"Group":"Capital and Reserves", "Title": "Others"},

  {"Group":"Commitments and Contingent Liabilities", "Title": "Commitments"},
  {"Group":"Commitments and Contingent Liabilities", "Title": "Commitments and Contingent Liabilities"}

]
`

	case "income_statement":
		url = fmt.Sprintf("http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_pl.php?code=%d", code)
		jsonData = `
[
  {"Group":"Income Statement", "Title": "Interest Income"},
  {"Group":"Income Statement", "Title": "Interest Expense"},
  {"Group":"Income Statement", "Title": "Net Interest Income"},
  {"Group":"Income Statement", "Title": "Other Operating Income"},
  {"Group":"Income Statement", "Title": "Total Operating Income"},
  {"Group":"Income Statement", "Title": "Net Insurance Claims Incurred & Movement in Policyholders' Liabilities"},
  {"Group":"Income Statement", "Title": "Net Operating Income"},
  {"Group":"Income Statement", "Title": "Operating Expenses"},
  {"Group":"Income Statement", "Title": "Impairment Losses on Loans & Advances"},
  {"Group":"Income Statement", "Title": "Other Impairment Losses"},
  {"Group":"Income Statement", "Title": "Operating Profit"},

  {"Group":"Income Statement", "Title": "Turnover"},
  {"Group":"Income Statement", "Title": "Cost of Sales"},
  {"Group":"Income Statement", "Title": "Gross Profit"},
  {"Group":"Income Statement", "Title": "Change in FV & Impairment on Inv. Prop."},
  {"Group":"Income Statement", "Title": "Change in FV & Impairment on Others"},
  {"Group":"Income Statement", "Title": "Profit / (Loss) on Disposal"},
  {"Group":"Income Statement", "Title": "Other Non-operating Items"},
  {"Group":"Income Statement", "Title": "Share of Results of Asso. & JCEs"},
  {"Group":"Income Statement", "Title": "Profit / (Loss) before Taxation"},
  {"Group":"Income Statement", "Title": "Taxation"},
  {"Group":"Income Statement", "Title": "Profit / (Loss) from Discontinued Operations"},
  {"Group":"Income Statement", "Title": "Non-controlling Interests"},
  {"Group":"Income Statement", "Title": "Others"},
  {"Group":"Income Statement", "Title": "Profit / (Loss) Attributable to Shareholders"},
  {"Group":"Income Statement", "Title": "Net Finance Costs / (Income)"},
  {"Group":"Income Statement", "Title": "Depreciation & Amortisation"},
  {"Group":"Income Statement", "Title": "Directors' Emoluments"},

  {"Group":"Market Valuation Indicators", "Title": "EPS (cts)"},
  {"Group":"Market Valuation Indicators", "Title": "DPS (cts)"},
  {"Group":"Market Valuation Indicators", "Title": "Dividend Payout Ratio (%)"},
  {"Group":"Market Valuation Indicators", "Title": "Cash flow per share ($)"},
  {"Group":"Market Valuation Indicators", "Title": "NBV per share ($)"}
]
`

	case "financial_ratio":
		url = fmt.Sprintf("https://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_ratio.php?code=%d", code)
		jsonData = `
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

	default:
		fmt.Printf("Else")
	}
	return url, jsonData
}

func getFormData(jsonData string) (FormData, map[string]int, error) {
	// {"Group":"", "Title":""},

	var (
		fd FormData
		m  map[string]int
	)
	// Parse json to form data
	err := json.Unmarshal([]byte(jsonData), &fd)
	if err != nil {
		return fd, m, fmt.Errorf("Can not unmarshal json data. %w", err)
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
