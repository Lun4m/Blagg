package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JSONFeed struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Url           string    `json:"url"`
	UserID        uuid.UUID `json:"user_id"`
	LastFetchedAt NullTime  `json:"last_fetched_at"`
}

type NullTime sql.NullTime

func (self *NullTime) MarshalJSON() ([]byte, error) {
	if !self.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(self.Time)
}

func (self *Feed) Json() JSONFeed {
	return JSONFeed{
		ID:            self.ID,
		CreatedAt:     self.CreatedAt,
		UpdatedAt:     self.UpdatedAt,
		Name:          self.Name,
		Url:           self.Url,
		UserID:        self.UserID,
		LastFetchedAt: NullTime(self.LastFetchedAt),
	}
}
