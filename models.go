package main

type SearchOptions struct {
	query      string
	limit      int
	region     string
	safeSearch int
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
