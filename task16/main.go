package main

import (
	"flag"
	"log"
	"net/url"
	"time"
)

func main() {
	startURL := flag.String("url", "", "start URL, for example: https://example.com")
	depth := flag.Int("depth", 1, "recursion depth for links")
	outDir := flag.String("out", "mirror", "output directory")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTP timeout")

	flag.Parse()

	if *startURL == "" {
		log.Fatal("please provide -url")
	}

	parsed, err := url.Parse(*startURL)
	if err != nil {
		log.Fatalf("invalid URL: %v", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		log.Fatal("only http and https URLs are supported")
	}

	d := NewDownloader(*outDir, *timeout)

	if err := d.Mirror(*startURL, *depth); err != nil {
		log.Fatalf("mirror failed: %v", err)
	}

	log.Println("done")
}
