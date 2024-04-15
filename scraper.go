package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"log"
	"net/http"
	"sync"
	"time"

	"blagg/internal/database"
)

func parseFeed(url string) (*Channel, error) {
	log.Printf("Fetching URL: %v", url)
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

func (self *apiConfig) fetch(ctx context.Context, limit int32) {
	var wg sync.WaitGroup

	feeds, err := self.DB.GetNextFeedsToFetch(ctx, limit)
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, feed := range feeds {
		wg.Add(1)

		go func() {
			defer wg.Done()

			rss, err := parseFeed(feed.Url)
			if err != nil {
				log.Println(err.Error())
				return
			}
			log.Printf("Feed %s - printing titles", feed.Url)
			for _, item := range rss.Item {
				log.Println(item.Title)
			}

			log.Printf("Feed %s - marking as fetched", feed.Url)
			err = self.DB.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
				LastFetchedAt: sql.NullTime{Valid: true, Time: time.Now().UTC()},
				UpdatedAt:     time.Now().UTC(),
				ID:            feed.ID,
			})
			if err != nil {
				log.Println(err.Error())
				return
			}
		}()
	}
	wg.Wait()
}

func (self *apiConfig) feedFetchWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			self.fetch(ctx, 10)
		}
	}()
}
