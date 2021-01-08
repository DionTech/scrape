package scrape

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/DionTech/stdoutformat"
)

type ScrapeDefintion struct {
	Domain    string
	OutputDir string
	Threads   int
	RateLimit time.Duration
	Index     map[string]bool
}

func (scrapeDefintion *ScrapeDefintion) Validate() bool {
	if scrapeDefintion.Threads == 0 {
		scrapeDefintion.Threads = 1
	}

	if scrapeDefintion.RateLimit == 0 {
		scrapeDefintion.RateLimit = time.Duration(500) * time.Millisecond
	}

	if _, err := url.ParseRequestURI(scrapeDefintion.Domain); err != nil {
		stdoutformat.Error(err)
		return false
	}

	if scrapeDefintion.OutputDir == "" {
		stdoutformat.Error(errors.New("OutputDir must be set"))
		return false
	}

	return true
}

func (scrapeDefintion *ScrapeDefintion) Scrape() {
	scrapeDefintion.mkOutPutDir()

	fmt.Println(scrapeDefintion)
}

func (scrapeDefintion *ScrapeDefintion) mkOutPutDir() {
	//check if to have to make a dir
	if _, err := os.Stat(scrapeDefintion.OutputDir); os.IsNotExist(err) {
		err := os.Mkdir(scrapeDefintion.OutputDir, 0755)

		if err != nil {
			stdoutformat.Fatalf("cannot create output dir %s", scrapeDefintion.OutputDir)
		}
	}
}
