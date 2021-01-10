package scrape

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/DionTech/scrape/pckg/parsehtml"
)

// a response is a wrapper around an HTTP response;
// it contains the request value for context.
type response struct {
	request request

	status     string
	statusCode int
	headers    []string
	body       []byte
	err        error
}

// String returns a string representation of the request and response
func (r response) String() string {
	b := &bytes.Buffer{}

	b.WriteString(r.request.url)
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("> %s %s HTTP/1.1\n", r.request.method, r.request.url))

	// request headers
	for _, h := range r.request.headers {
		b.WriteString(fmt.Sprintf("> %s\n", h))
	}
	b.WriteString("\n")

	// status line
	b.WriteString(fmt.Sprintf("< HTTP/1.1 %s\n", r.status))

	// response headers
	for _, h := range r.headers {
		b.WriteString(fmt.Sprintf("< %s\n", h))
	}
	b.WriteString("\n")

	// body
	b.Write(r.body)

	return b.String()
}

func (r response) StringNoHeaders() string {
	b := &bytes.Buffer{}

	b.Write(r.body)

	return b.String()
}

// save write a request and response output to disk
func (r response) save(pathPrefix string) (string, error) {

	//DK: changed, cause we are intersted in headers!;
	//we can think about making this optionable
	//content := []byte(r.StringNoHeaders())
	content := []byte(r.String())

	checksum := sha1.Sum(content)
	parts := []string{pathPrefix}

	parts = append(parts, r.request.url)
	parts = append(parts, fmt.Sprintf("%x", checksum))

	p := filepath.Join(parts...)

	if _, err := os.Stat(filepath.Dir(p)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(p), 0750)
		if err != nil {
			return p, err
		}
	}

	err := ioutil.WriteFile(p, content, 0640)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (r response) scan() (additionalURLs []string, outgoingLinks []string) {
	additionalURLs = make([]string, 0)
	outgoingLinks = make([]string, 0)

	requestURL, err := url.Parse(r.request.url)
	if err != nil {
		log.Fatal(err)
		return nil, additionalURLs
	}

	for _, additionalURL := range parsehtml.ExtractAttribs(bytes.NewReader(r.body),
		[]string{"src", "href", "action"}) {
		parsed, err := url.Parse(additionalURL)
		if err != nil {
			//fmt.Println(err)
			continue
		}

		if parsed.Scheme == "" {
			parsed.Scheme = requestURL.Scheme
		}

		if parsed.Host == "" {
			parsed.Host = requestURL.Host
		}

		parsedHostname := parsed.Hostname()
		requestHostname := requestURL.Hostname()

		parsedHostname = strings.Replace(parsedHostname, "www.", "", 1)
		requestHostname = strings.Replace(requestHostname, "www.", "", 1)

		if parsedHostname != requestHostname {
			if _, err := url.ParseRequestURI(parsed.String()); err == nil {
				outgoingLinks = append(outgoingLinks, parsed.String())
			}
			continue
		}

		additionalURLs = append(additionalURLs, parsed.String())
	}

	return additionalURLs, outgoingLinks
}
