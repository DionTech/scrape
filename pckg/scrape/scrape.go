package scrape

import (
	"fmt"
	"net/url"

	"github.com/DionTech/stdoutformat"
)

type ScrapeDefintion struct {
	Domain string
	Index  map[string]bool
}

func (scrapeDefintion *ScrapeDefintion) Validate() bool {
	if _, err := url.ParseRequestURI(scrapeDefintion.Domain); err != nil {
		stdoutformat.Error(err)
		return false
	}

	return true
}

func (scrapeDefintion *ScrapeDefintion) Scrape() {
	fmt.Println(scrapeDefintion)
}
