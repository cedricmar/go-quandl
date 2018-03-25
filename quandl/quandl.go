// Quandl library to communicate with the API
// Full API documentation can be found here https://docs.quandl.com
// Created with the help of the PHP library https://github.com/DannyBen/php-quandl/blob/master/Quandl.php

package quandl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type API struct {
	apiKey       string
	format       string
	cacheHandler string
	wasCached    bool
	//force_curl    = false;
	//no_ssl_verify = false; // disable ssl verification for curl
	timeout int
	lastURL string
	err     error
}

type JSONData struct {
	Dataset Dataset `json:"dataset"`
}

type Dataset struct {
	ID                  int             `json:"id"`
	DatasetCode         string          `json:"dataset_code"`
	DatabaseCode        string          `json:"database_code"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	RefreshedAt         string          `json:"refreshed_at"`
	NewestAvailable     string          `json:"newest_available_date"`
	OldestAvailableDate string          `json:"oldest_available_date"`
	ColumnNames         []string        `json:"column_names"`
	Frequency           string          `json:"frequency"` // none|daily|weekly|monthly|quarterly|annual
	Type                string          `json:"type"`
	Premium             bool            `json:"premium"`
	Limit               string          `json:"limit"`
	Transform           string          `json:"transform"`
	ColumnIndex         string          `json:"column_index"`
	StartDate           string          `json:"start_date"`
	EndDate             string          `json:"end_date"`
	Data                [][]interface{} `json:"data"`
	Collapse            string          `json:"collapse"`
	Order               string          `json:"order"`
	DatabaseID          int             `json:"database_id"`
}

var urlTemplates = map[string]string{
	"direct": "https://www.quandl.com/api/v3/%s.%s?%s",
	"symbol": "https://www.quandl.com/api/v3/datasets/%s.%s?%s",
	"search": "https://www.quandl.com/api/v3/datasets.%s?%s",
	"list":   "https://www.quandl.com/api/v3/datasets.%s?%s",
	"meta":   "https://www.quandl.com/api/v3/datasets/%s/metadata.%s",
	"dbs":    "https://www.quandl.com/api/v3/databases.%s?%s",
	"bulk":   "https://www.quandl.com/api/v3/databases/%s/data?%s",
}

// NewAPI inits the api.
func NewAPI(k string, format string) *API {
	if format == "" {
		format = "object"
	}
	return &API{apiKey: k, format: format}
}

// SetTimeout specify a timeout in seconds,
// the default timeout is 5 seconds.
func (api *API) SetTimeout(t int) {
	api.timeout = t
}

// Get calls the API.
func (api *API) Get(path string, params map[string]string) Dataset {
	url := api.getURL(
		"direct",
		path,
		api.getFormat(true),
		api.arrangeParams(params),
	)
	return api.getData(url)
}

// GetSymbol returns a Dataset for the given symbol.
func (api *API) GetSymbol(sym string, params map[string]string) Dataset {
	url := api.getURL(
		"symbol",
		sym,
		api.getFormat(true),
		api.arrangeParams(params),
	)
	return api.getData(url)
}

/*
// GetBulk downloads an entire database to a ZIP file.
func (api *API) GetBulk(database string, filename string, complete bool) {
	params := map[string]string{}
	if complete {
		params["download_type"] = "complete"
	} else {
		params["download_type"] = "partial"
	}
	url := api.getURL("bulk", database, api.arrangeParams(params))
	return api.downloadToFile(url, filename)
}
*/

// GetMeta returns metadata for a given symbol.
func (api *API) GetMeta(symbol string) Dataset {
	url := api.getURL(
		"meta",
		symbol,
		api.getFormat(false),
	)
	return api.getData(url)
}

// GetDatabases returns the list of databases.
// Quandl limits it to 100 per page at most.
func (api *API) GetDatabases(page int, perPage int) Dataset {
	if page == 0 {
		page = 1
	}
	if perPage == 0 || perPage > 100 {
		perPage = 100
	}
	params := map[string]string{
		"per_page": strconv.Itoa(perPage),
		"page":     strconv.Itoa(page),
	}
	url := api.getURL(
		"dbs",
		api.getFormat(false),
		api.arrangeParams(params),
	)
	return api.getData(url)
}

// GetSearch returns results for a search query.
// CSV output is not supported so it will fall back to object mode.
func (api *API) GetSearch(query string, page int, perPage int) Dataset {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 300
	}
	params := map[string]string{
		"per_page": strconv.Itoa(perPage),
		"page":     strconv.Itoa(page),
		"query":    query,
	}
	url := api.getURL(
		"search",
		api.getFormat(true),
		api.arrangeParams(params),
	)
	return api.getData(url)
}

/*
// GetList returns the list of symbols
// func (api *API) GetList(source string, page int, perPage int) Dataset {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 300
	}
	// etc...
}
*/

func (api *API) getURL(k string, args ...interface{}) string {
	url := urlTemplates[k]

	fmt.Println(fmt.Sprintf(url, args...))

	api.lastURL = strings.Trim(fmt.Sprintf(url, args...), "?&")
	return api.lastURL
}

func (api *API) getFormat(omitCsv bool) string {
	if (api.format == "csv" && omitCsv) || api.format == "object" {
		return "json"
	}
	return api.format
}

func (api *API) getData(url string) Dataset {
	res := api.executeDownload(url)
	/*if api.format == "object" {
		return res
	}*/

	d := JSONData{}
	err := json.Unmarshal(res, &d)
	if err != nil {
		log.Fatal(err)
	}

	return d.Dataset
}

func (api *API) executeDownload(url string) []byte {
	var data []byte
	//if api.cacheHandler == "" {
	data, err := api.download(url)
	if err != nil {
		// return error resp
	}
	//}
	/* else {
	    data = api.attemptGetFromCache(url)
	}*/

	return data
}

func (api *API) arrangeParams(params map[string]string) string {
	if api.apiKey != "" {
		params["auth_token"] = api.apiKey
	}
	if len(params) == 0 {
		return ""
	}

	/*
	   trims := []string{"trim_start", "trim_end"}
	   for _, v := range trims {
	       if params[v] != "" {
	           params[v] = convertToQuandlDate(params[v])
	       }
	   }
	*/

	return httpBuildQuery(params)
}

func (api *API) download(url string) ([]byte, error) {
	ts := 5
	if api.timeout != 0 {
		ts = api.timeout
	}
	timeout := time.Duration(ts) * time.Second

	client := &http.Client{
		Timeout: timeout,
	}
	res, err := client.Get(url)
	if err != nil {
		// @TODO - implement
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err, _ := ioutil.ReadAll(res.Body)
		log.Fatalf("%d %s\n", res.StatusCode, string(err))
	}
	return ioutil.ReadAll(res.Body)
}

func httpBuildQuery(params map[string]string) string {
	v := url.Values{}
	for k, p := range params {
		v.Add(k, p)
	}
	return v.Encode()
}
