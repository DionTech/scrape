package main

import (
	"flag"

	"github.com/DionTech/scrape/pckg/scrape"
)

func main() {
	var help bool
	flag.BoolVar(&help, "help", false, "show help")

	var outputDir string
	flag.StringVar(&outputDir, "output-directory", "./scrape", "where to store scrape")
	flag.StringVar(&outputDir, "o", "./scrape", "where to store scrape")

	//options: RateLimit, Threads

	flag.Parse()

	domain := flag.Arg(0)

	if domain == "" || help {
		flag.PrintDefaults()
		return
	}

	scrapeDefintion := scrape.ScrapeDefintion{
		Domain:    domain,
		OutputDir: outputDir}

	if !scrapeDefintion.Validate() {
		return
	}

	scrapeDefintion.Scrape()
}
