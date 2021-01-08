package main

import (
	"flag"

	"github.com/DionTech/scrape/pckg/scrape"
)

func main() {
	flag.Parse()

	domain := flag.Arg(0)

	scrapeDefintion := scrape.ScrapeDefintion{
		Domain: domain}

	if !scrapeDefintion.Validate() {
		return
	}

	scrapeDefintion.Scrape()
}
