package main

import (
	q "github.com/cedricmar/go-quandl/quandl"
)

func main() {

	apiKey := ""
	format := ""
	symbol := "WIKI/AAPL"

	api := q.NewAPI(apiKey, format)

	api.GetMeta(symbol)

	/*
		sym := api.GetSymbol(symbol, map[string]string{
			"collapse":   "annual",
			"start_date": "2015-12-31",
			"end_date":   "2018-12-31",
		})

		for _, d := range sym.Data {
			for i, val := range d {
				fmt.Printf("%s: %v\n", sym.ColumnNames[i], val)
			}
			fmt.Println("- - - - - - - - -")
		}
	*/

}
