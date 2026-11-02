# ddgosearch

A lightweight DuckDuckGo search tool and API written in Go.

Search DuckDuckGo from your terminal or run a simple HTTP API server with small binary.

## Features

* Search DuckDuckGo from the command line
* JSON output support
* Lightweight HTTP API server
* JSONL mode through stdio
* Minimal dependencies
* Single static binary

## Installation

### Go Install

```bash
go install -trimpath -ldflags "-s -w" github.com/kr4phy/ddgosearch@latest
```

Make sure your Go bin directory is in your PATH.

### Build from Source

```bash
git clone https://github.com/kr4phy/ddgosearch.git
cd ddgosearch
go build
```

## Usage

### CLI

```bash
ddgosearch golang
```

```bash
ddgosearch -json "open source licenses"
```

### API Server

```bash
ddgosearch -api-server -port 8080
```

Search endpoint:

```text
GET /api/v1/search?q=golang
```

### JSONL mode

```
ddgosearch -jsonl-stdio
```
Use json with `query`, `limit` field and optionally with safesearch option.
Add json field `quit` with value `true` to exit

## Philosophy

This project prioritizes:

* Simplicity
* Minimal dependencies
* Small binary size
* Easy deployment
