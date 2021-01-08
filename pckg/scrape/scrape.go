package scrape

import (
	"errors"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/DionTech/stdoutformat"
)

type scrapeIndex map[string]bool

type ScrapeDefintion struct {
	Domain    string
	OutputDir string
	Threads   int
	RateLimit time.Duration
}

var Index scrapeIndex

var IndexMutex = sync.RWMutex{}

var RequestWaitGroup sync.WaitGroup
var ResponseWaitGroup sync.WaitGroup

func (scrapeDefintion *ScrapeDefintion) Validate() bool {
	if scrapeDefintion.Threads == 0 {
		scrapeDefintion.Threads = 1
	}

	if scrapeDefintion.RateLimit == 0 {
		scrapeDefintion.RateLimit = time.Duration(100) * time.Millisecond
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

	Index = make(scrapeIndex, 0)

	// the request and response channels for
	// the worker pool
	requests := make(chan request, scrapeDefintion.Threads)
	responses := make(chan response)

	for i := 0; i < scrapeDefintion.Threads; i++ {
		go func() {
			//this is a really simple rate limiter here at the moment!
			//@TODO make it better!!!
			limiter := time.Tick(scrapeDefintion.RateLimit)
			for req := range requests {
				responses <- goRequest(req)
				<-limiter
			}
		}()
	}

	go scrapeDefintion.respFwd(requests, responses)

	ResponseWaitGroup.Add(1)
	go scrapeDefintion.pushRequest(scrapeDefintion.Domain, requests)

	RequestWaitGroup.Wait()
	ResponseWaitGroup.Wait()

}

func (scrapeDefintion *ScrapeDefintion) pushRequest(domain string, requests chan request) {
	IndexMutex.Lock()
	Index[domain] = true
	IndexMutex.Unlock()
	RequestWaitGroup.Add(1)

	requests <- request{
		method:         "GET",
		url:            domain,
		followLocation: true}
}

func (scrapeDefintion *ScrapeDefintion) respFwd(requests chan request, responses chan response) {
	for res := range responses {
		if res.err != nil {
			stdoutformat.Printf("request failed: %s\n", res.err)
			continue
		}
		_, err := res.save(scrapeDefintion.OutputDir)
		if err != nil {
			stdoutformat.Printf("failed to save file: %s\n", err)
		}

		additionalURLs := res.scan()

		for _, additionalURL := range additionalURLs {

			IndexMutex.RLock()
			if _, ok := Index[additionalURL]; ok == false {
				ResponseWaitGroup.Add(1)
				go scrapeDefintion.pushRequest(additionalURL, requests)
			}
			IndexMutex.RUnlock()

		}
		//line := fmt.Sprintf("%s %s (%s)\n", path, res.request.URL(), res.status)
		ResponseWaitGroup.Done()
	}
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
