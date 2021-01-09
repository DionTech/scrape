package scrape

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

var UserAgent = "Mozilla/5.0 (compatible; scrape/1.0; +github.com/DionTech/scrape)"

var transport = &http.Transport{
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	DisableKeepAlives: true,
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: time.Second,
		DualStack: true,
	}).DialContext,
}

var httpClient = &http.Client{
	Transport: transport,
}

func goRequest(r request) response {
	fmt.Println(r.url)
	httpClient.Timeout = r.timeout

	if !r.followLocation {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	var req *http.Request
	var err error
	req, err = http.NewRequest(r.method, r.url, nil)

	if err != nil {
		RequestWaitGroup.Done()
		return response{request: r, err: err}
	}
	req.Close = true

	if !r.HasHeader("User-Agent") {
		r.headers = append(r.headers, fmt.Sprintf("User-Agent: %s", UserAgent))
	}

	for _, h := range r.headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			continue
		}

		req.Header.Set(parts[0], parts[1])
	}

	resp, err := httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		RequestWaitGroup.Done()
		return response{request: r, err: err}
	}
	body, _ := ioutil.ReadAll(resp.Body)

	// extract the response headers
	hs := make([]string, 0)
	for k, vs := range resp.Header {
		for _, v := range vs {
			hs = append(hs, fmt.Sprintf("%s: %s", k, v))
		}
	}

	RequestWaitGroup.Done()

	return response{
		request:    r,
		status:     resp.Status,
		statusCode: resp.StatusCode,
		headers:    hs,
		body:       body,
	}
}
