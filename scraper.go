package main

import (
	"encoding/xml"
	"net/http"
)

func parseFeed(url string) (*Channel, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var rss struct {
		Channel Channel `xml:"channel"`
	}

	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&rss); err != nil {
		return nil, err
	}
	return &rss.Channel, nil
}
