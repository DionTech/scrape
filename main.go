package main

import (
	"flag"
	"time"

	"github.com/DionTech/scrape/pckg/scrape"
)

func main() {
	var help bool
	flag.BoolVar(&help, "help", false, "show help")

	var outputDir string
	flag.StringVar(&outputDir, "output-directory", "./scrape", "where to store scrape")
	flag.StringVar(&outputDir, "o", "./scrape", "short for output-directory")

	//options: RateLimit, Threads
	var threads int
	flag.IntVar(&threads, "threads", 1, "army of threads to parse")
	flag.IntVar(&threads, "t", 1, "short for threads")

	var rateLimit int
	flag.IntVar(&rateLimit, "rate-limit", 250, "rate limit in ms between two requests in a thread")
	flag.IntVar(&rateLimit, "r", 250, "short for rate-limit")

	flag.Parse()

	domain := flag.Arg(0)

	if domain == "" || help {
		flag.PrintDefaults()
		return
	}

	scrapeDefintion := scrape.ScrapeDefintion{
		Domain:    domain,
		OutputDir: outputDir,
		Threads:   threads,
		RateLimit: time.Duration(rateLimit) * time.Millisecond}

	if !scrapeDefintion.Validate() {
		return
	}

	scrapeDefintion.Scrape()
}
