package scrape

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DionTech/stdoutformat"
)

type scrapeIndex map[string]bool

/** ScrapeDefintion
holds all options about doing scraping
*/
type ScrapeDefintion struct {
	Domain         string
	OutputDir      string
	Threads        int
	TemplatePath   string
	RateLimit      time.Duration
	FollowInternal bool
	PathBinding    string
	Requests       chan request
	Responses      chan response
}

var Index scrapeIndex
var OutgoingLinks scrapeIndex

var IndexMutex = sync.RWMutex{}

var RequestWaitGroup sync.WaitGroup
var ResponseWaitGroup sync.WaitGroup

func (scrapeDefintion *ScrapeDefintion) Init() {
	scrapeDefintion.Requests = make(chan request, scrapeDefintion.Threads)
	scrapeDefintion.Responses = make(chan response)
}

func (scrapeDefintion *ScrapeDefintion) Save() {
	indexFileName := "index"
	indexFile, err := os.Create(indexFileName)

	if err != nil {
		stdoutformat.Fatalf("cannot create file: %s", err)
	}
	writer := bufio.NewWriter(indexFile)

	for url := range Index {
		writer.WriteString(url + "\n")
	}

	writer.Flush()

	outgoingLinksFileName := "outgoing-links"
	file, err := os.Create(outgoingLinksFileName)

	if err != nil {
		stdoutformat.Fatalf("cannot create file: %s", err)
	}

	w := bufio.NewWriter(file)

	for url := range OutgoingLinks {
		fmt.Println(url)
		w.WriteString(url + "\n")
	}

	w.Flush()
}

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

	if scrapeDefintion.TemplatePath != "" {
		if _, err := os.Stat(scrapeDefintion.TemplatePath); os.IsNotExist(err) {
			stdoutformat.Error(errors.New("Template Path must be valid"))
			return false
		}
	}

	return true
}

func (scrapeDefintion *ScrapeDefintion) Scrape() {
	scrapeDefintion.mkOutPutDir()

	Index = make(scrapeIndex, 0)
	OutgoingLinks = make(scrapeIndex, 0)

	for i := 0; i < scrapeDefintion.Threads; i++ {
		go func() {
			//waiting x ms after each request
			limiter := time.Tick(scrapeDefintion.RateLimit)
			for req := range scrapeDefintion.Requests {
				scrapeDefintion.Responses <- goRequest(req)
				<-limiter
			}
		}()
	}

	go scrapeDefintion.respFwd()

	ResponseWaitGroup.Add(1)
	go scrapeDefintion.pushRequest(scrapeDefintion.Domain)

	if scrapeDefintion.TemplatePath != "" {
		scrapeDefintion.pushTemplate()
	}

	RequestWaitGroup.Wait()
	ResponseWaitGroup.Wait()

}

func (scrapeDefintion *ScrapeDefintion) pushRequest(domain string) {
	IndexMutex.Lock()
	Index[domain] = true
	IndexMutex.Unlock()
	RequestWaitGroup.Add(1)

	scrapeDefintion.Requests <- request{
		method:         "GET",
		url:            domain,
		followLocation: true}
}

func (scrapeDefintion *ScrapeDefintion) respFwd() {
	for res := range scrapeDefintion.Responses {
		if res.err != nil {
			stdoutformat.Printf("request failed: %s\n", res.err)
			continue
		}
		_, err := res.save(scrapeDefintion.OutputDir)
		if err != nil {
			stdoutformat.Printf("failed to save file: %s\n", err)
		}

		//when not follow internal, we can quit here
		if scrapeDefintion.FollowInternal == false {
			ResponseWaitGroup.Done()
			continue
		}

		additionalURLs, outgoingLinks := res.scan()

		go func(outgoingLinks []string) {
			for _, outgoingLink := range outgoingLinks {
				if _, ok := OutgoingLinks[outgoingLink]; ok == false {
					OutgoingLinks[outgoingLink] = true
				}
			}
		}(outgoingLinks)

		for _, additionalURL := range additionalURLs {

			IndexMutex.RLock()
			if _, ok := Index[additionalURL]; ok == false {
				if scrapeDefintion.PathBinding != "" {
					if strings.Contains(additionalURL, scrapeDefintion.PathBinding) {
						ResponseWaitGroup.Add(1)
						go scrapeDefintion.pushRequest(additionalURL)
					}
				} else {
					ResponseWaitGroup.Add(1)
					go scrapeDefintion.pushRequest(additionalURL)
				}

			}
			IndexMutex.RUnlock()

		}

		ResponseWaitGroup.Done()
	}
}

func (scrapeDefintion *ScrapeDefintion) pushTemplate() {
	file, _ := os.Open(filepath.FromSlash(scrapeDefintion.TemplatePath))

	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		additionalURL := scrapeDefintion.Domain + scanner.Text()
		ResponseWaitGroup.Add(1)
		go scrapeDefintion.pushRequest(additionalURL)
	}
}

func (scrapeDefintion *ScrapeDefintion) mkOutPutDir() {
	//check if to have to make a dir
	dir := filepath.FromSlash(scrapeDefintion.OutputDir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)

		if err != nil {
			stdoutformat.Fatalf("cannot create output dir %s", dir)
		}
	}
}
