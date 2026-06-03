package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	query := r.URL.Query().Get("q")
	limitParam := r.URL.Query().Get("limit")
	region := r.URL.Query().Get("region")
	safeSearchParam := r.URL.Query().Get("safeSearch")

	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		limit = 10
	}

	safeSearch, err := strconv.Atoi(safeSearchParam)
	if err != nil {
		safeSearch = -1
	}

	results, err := ScrapeDuckDuckGo(SearchOptions{
		query:      query,
		limit:      limit,
		region:     region,
		safeSearch: safeSearch,
	})
	if err != nil {
		http.Error(w, "Error scraping DuckDuckGo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func createAPIServer(port int) (*http.Server, error) {
	fmt.Printf("Starting server on port %d...\n", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search", searchHandler)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	return server, nil
}

func main() {

	var (
		limit         int
		minimalOutput bool
		region        string
		safeSearch    int
		jsonOutput    bool
		showVersion   bool
		apiServerMode bool
		port          int
	)

	flag.BoolVar(&apiServerMode, "api-server", false, "Run as an API server instead of a CLI tool")
	flag.IntVar(&port, "port", 8080, "Port to run the api server on")
	flag.IntVar(&port, "p", 8080, "Alias for -port")
	flag.IntVar(&limit, "limit", 10, "Limit the number of results")
	flag.IntVar(&limit, "l", 10, "Alias for --limit")
	flag.BoolVar(&minimalOutput, "minimal-output", false, "Show only title and URL (omit description)")
	flag.BoolVar(&minimalOutput, "m", false, "Alias for --minimal-output")
	flag.StringVar(&region, "region", "wt-wt", "Set search region (for example: wt-wt, us-en, kr-kr)")
	flag.StringVar(&region, "kl", "wt-wt", "Alias for --region")
	flag.IntVar(&safeSearch, "safe-search", -1, "Set safe search: 1=on, -1=moderate, -2=off")
	flag.IntVar(&safeSearch, "kp", -1, "Alias for --safe-search")
	flag.BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	flag.BoolVar(&jsonOutput, "j", false, "Alias for --json")
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")
	flag.BoolVar(&showVersion, "v", false, "Alias for --version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ddgosearch: Get search results from DuckDuckGo in anywhere\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [options] <query>\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintf(os.Stderr, "  %s github\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -limit 5 golang cli\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -m -region us-en -safe-search 1 github actions\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -json open source licenses\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Notes:")
		fmt.Fprintln(os.Stderr, "  - Put the query after options.")
		fmt.Fprintln(os.Stderr, "  - All remaining arguments are joined as a single query string.")
		fmt.Fprintln(os.Stderr, "  - Flag -port will be ignored when not running in API server mode.")
	}
	flag.Parse()

	if showVersion {
		version := "unknown"
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
		fmt.Println("ddgosearch version", version)
		return
	}

	if apiServerMode {
		server, err := createAPIServer(port)
		sigChan := make(chan os.Signal, 1)

		if err != nil {
			log.Fatal(err)
		}
		go server.ListenAndServe()
		fmt.Printf("API server is running on %s\n", server.Addr)

		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nShutting down server...")
		server.Close()
		fmt.Println("Server stopped.")

		return
	}

	if limit < 1 {
		log.Fatal("limit must be a positive integer")
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	query := strings.TrimSpace(strings.Join(args, " "))

	fullResults, err := ScrapeDuckDuckGo(SearchOptions{
		query:      query,
		limit:      limit,
		region:     region,
		safeSearch: safeSearch,
	})

	if err != nil {
		log.Fatal(err)
	}

	if len(fullResults) == 0 {
		fmt.Println("No results found.")
		return
	}

	var minimalResults []MinimalSearchResult

	if minimalOutput {
		minimalResults = make([]MinimalSearchResult, len(fullResults))
		for i, r := range fullResults {
			minimalResults[i] = MinimalSearchResult{
				Index: r.Index,
				Title: r.Title,
				URL:   r.URL,
			}
		}
	}

	if jsonOutput {
		if minimalOutput {
			jsonData, err := json.MarshalIndent(minimalResults, "", "  ")
			if err != nil {
				log.Fatal("Error encoding results to JSON:", err)
			}

			fmt.Println(string(jsonData))
			return
		}

		jsonData, err := json.MarshalIndent(fullResults, "", "  ")
		if err != nil {
			log.Fatal("Error encoding results to JSON:", err)
		}

		fmt.Println(string(jsonData))
		return
	} else {
		if minimalOutput {
			for _, result := range minimalResults {
				fmt.Printf("%d.\t%s\n", result.Index, result.Title)
				fmt.Printf("\tURL: %s\n", result.URL)
				fmt.Println()
			}
			return
		}

		for _, result := range fullResults {
			fmt.Printf("%d.\t%s\n", result.Index, result.Title)
			fmt.Printf("\tURL: %s\n", result.URL)
			if result.Description != "" {
				fmt.Printf("\tDescription: %s\n", result.Description)
			}
			fmt.Println()
		}
		return
	}
}
