package safer

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

const (
	companySnapshotURL = "https://safer.fmcsa.dot.gov/query.asp"
	searchURL          = "https://safer.fmcsa.dot.gov/keywordx.asp"
	paramUSDOT         = "USDOT"
	paramMCMX          = "MC_MX"
)

var headers = http.Header{
	"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	"Accept-Encoding":           {"gzip, deflate, br"},
	"Accept-Language":           {"en-US,en;q=0.9"},
	"Cache-Control":             {"max-age=0"},
	"Connection":                {"keep-alive"},
	"Host":                      {"safer.fmcsa.dot.gov"},
	"Upgrade-Insecure-Requests": {"1"},
	"User-Agent":                {"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Mobile Safari/537.36"},
}

type scraper struct {
	baseCollector      *colly.Collector
	companySnapshotURL string
	searchURL          string
}

func (s *scraper) scrapeCompanySnapshot(queryParam, queryString string) (*CompanySnapshot, error) {
	// build output snapshot and scraping collector
	var (
		snapshot  = new(CompanySnapshot)
		collector = s.baseCollector.Clone()
	)

	// checks to see if the returned page is a not found error, this is only called when the xpath is matched
	var notFound bool
	collector.OnXML(snapshotNotFoundXpath, func(element *colly.XMLElement) {
		notFound = true
	})

	// add handler to extract the latest update date
	collector.OnXML(latestUpdateDateXpath, func(element *colly.XMLElement) {
		snapshot.LatestUpdateDate = parseDate(element.Text)
	})

	// add handler to extract values from tables
	var tableIdx int
	collector.OnXML(tableXpath, func(element *colly.XMLElement) {
		if mapFunc, ok := snapshotTableXpathMapping[tableIdx]; ok {
			mapFunc(element, snapshot)
		}
		tableIdx++
	})

	// build POST data
	data := url.Values{
		"searchType":   {"ANY"},
		"query_type":   {"queryCarrierSnapshot"},
		"query_param":  {queryParam},
		"query_string": {queryString},
	}.Encode()

	// send POST and start collector job to parse values
	if err := collector.Request(http.MethodPost, s.companySnapshotURL, strings.NewReader(data), nil, headers); err != nil {
		return nil, err
	}

	if notFound {
		return nil, ErrCompanyNotFound
	}

	return snapshot, nil
}

func (s *scraper) scrapeCompanyNameSearch(queryString string) ([]CompanyResult, error) {
	collector := s.baseCollector.Clone()

	// add handler to parse output into the result array
	var output []CompanyResult
	collector.OnXML(companyResultXpath, func(element *colly.XMLElement) {
		output = append(output, companyResultStructFromXpath(element))
	})

	// build POST data
	searchString := "*" + strings.ToUpper(queryString) + "*" // e.g. `*SEARCH TERM*`
	data := url.Values{"searchstring": {searchString}, "SEARCHTYPE": {""}}.Encode()

	// send POST and start collector job to parse values
	if err := collector.Request(http.MethodPost, s.searchURL, strings.NewReader(data), nil, headers); err != nil {
		return nil, err
	}
	return output, nil
}
