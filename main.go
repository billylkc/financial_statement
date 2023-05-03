package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape() error {
	// Request the HTML page.
	res, err := http.Get("http://www.etnet.com.hk/www/eng/stocks/realtime/quote_ci_cashflow.php?code=316&quarter=latest")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return err
	}

	// Load the HTML document
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(content))
	fmt.Println("-----------------")

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err

	}

	// Find the review items
	// doc.Find(".left-content article .post-title")

	// doc.Find(".DivFigureContent table").Each(func(i int, s *goquery.Selection) {
	// 	// For each item found, get the title
	// 	title := s.Find("a").Text()
	// 	fmt.Printf("Review %d: %s\n", i, title)
	// })
	fmt.Println(PrettyPrint(doc))
	return nil
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func main() {
	ExampleScrape()
}
