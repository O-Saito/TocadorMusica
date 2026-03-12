package domain

import "context"

type YouTubeService interface {
	ParseURL(ctx context.Context, url string) (Track, error)
	Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error)
}

type SearchResult struct {
	Title    string
	URL      string
	Duration string
}
