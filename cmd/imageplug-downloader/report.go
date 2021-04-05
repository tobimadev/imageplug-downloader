package main

import (
	"encoding/json"
	"net/http"
)

type image struct {
	ID  int64  `json:"id"`
	Src string `json:"src"`
}

type product struct {
	ID       int64   `json:"id"`
	Handle   string  `json:"handle"`
	Title    string  `json:"title"`
	Vendor   string  `json:"vendor"`
	ProdType string  `json:"prodType"`
	Images   []image `json:"images"`
}

// type report struct {
// 	Products []product `json:"products"`
// }

func (srv *server) readReport(url string) ([]product, error) {
	// todo: use context
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	report := struct {
		Products []product `json:"products"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, err
	}
	return report.Products, nil
}
