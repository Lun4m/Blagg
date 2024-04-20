package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"blagg/internal/database"

	"github.com/google/uuid"
)

func parseFeed(url string) (*Channel, error) {
	log.Printf("Fetching URL: %v\n", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RSS_feed_bot/3.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error: failed HTTP GET request - %v\n", err.Error())
	}

	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("Status error: %v", resp.Status)
	}

	var rss struct {
		Channel Channel `xml:"channel"`
	}

	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&rss); err != nil {
		return nil, fmt.Errorf("Error: failed decoding XML - %v\n", err.Error())
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

			rssChannel, err := parseFeed(feed.Url)
			if err != nil {
				log.Printf("\nFeed: %s\n%s", feed.Url, err.Error())
				return
			}

			for _, item := range rssChannel.Item {
				if err = self.DB.CreatePost(ctx, database.CreatePostParams{
					ID:          uuid.New(),
					CreatedAt:   time.Now().UTC(),
					UpdatedAt:   time.Now().UTC(),
					Title:       item.Title,
					Url:         item.Link,
					Description: item.Description,
					// TODO: convert this to time.Time so it's sorted correctly when selecting from database
					PublishedAt: item.PubDate,
					FeedID:      feed.ID,
				}); err != nil {
					log.Printf("Error creating post: %v", err.Error())
				}
			}

			if err = self.DB.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
				LastFetchedAt: sql.NullTime{Valid: true, Time: time.Now().UTC()},
				UpdatedAt:     time.Now().UTC(),
				ID:            feed.ID,
			}); err != nil {
				log.Printf("\nFeed: %s\n%s", feed.Url, err.Error())
				return
			}
		}()
	}
	wg.Wait()
}

func (self *apiConfig) fetchFeeds(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			self.fetch(ctx, 3)
		}
	}()
}
