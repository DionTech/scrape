package scrape

import (
	"errors"
	"fmt"
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

	// spin up some workers to do the requests
	var wg sync.WaitGroup
	for i := 0; i < scrapeDefintion.Threads; i++ {
		wg.Add(1)

		//this is a really simple rate limiter here at the moment!
		//@TODO make it better!!!
		limiter := time.Tick(scrapeDefintion.RateLimit)

		go func() {
			for req := range requests {
				<-limiter
				responses <- goRequest(req)
			}
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		for req := range requests {
			fmt.Println(req.url)
		}
		wg.Done()
	}()

	// start outputting the response lines; we need a second
	// WaitGroup so we know the outputting has finished
	var owg sync.WaitGroup
	owg.Add(1)

	go scrapeDefintion.respFwd(requests, responses, &owg)

	//Index.Store(scrapeDefintion.Domain, true)
	Index[scrapeDefintion.Domain] = true

	requests <- request{
		method:         "GET",
		url:            scrapeDefintion.Domain,
		followLocation: true}

	// once all of the requests have been sent we can
	// close the requests channel
	//close(requests)

	// wait for all the workers to finish before closing
	// the responses channel
	wg.Wait()
	close(responses)

	owg.Wait()
}

func (scrapeDefintion *ScrapeDefintion) respFwd(requests chan request, responses chan response, owg *sync.WaitGroup) {
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
			if _, ok := Index[additionalURL]; ok == false {
				//Index.Store(additionalURL, true)
				IndexMutex.Lock()
				Index[additionalURL] = true
				IndexMutex.Unlock()

				requests <- request{
					method:         "GET",
					url:            additionalURL,
					followLocation: true}
			}

		}

		//line := fmt.Sprintf("%s %s (%s)\n", path, res.request.URL(), res.status)
	}
	owg.Done()
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
