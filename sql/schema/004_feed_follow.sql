-- +goose Up
CREATE TABLE feed_follows (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  feed_id UUID NOT NULL,
  user_id UUID NOT NULL,
  FOREIGN KEY(feed_id) REFERENCES feeds(id),
  FOREIGN KEY(user_id) REFERENCES users(id),
  UNIQUE (feed_id, user_id)
);


-- +goose Down
DROP TABLE feed_follows;
