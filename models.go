package main

type SearchOptions struct {
	query      string
	limit      int
	region     string
	safeSearch int
}

type JSONLRequest struct {
	Query      string `json:"query"`
	Limit      int    `json:"limit"`
	Region     string `json:"region"`
	SafeSearch *int   `json:"safeSearch"`
	Quit     bool   `json:"quit"`
}

type SearchResult struct {
	Index       int
	Title       string
	URL         string
	Description string
}

type MinimalSearchResult struct {
	Index int
	Title string
	URL   string
}
