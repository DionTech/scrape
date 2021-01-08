package parsehtml

import (
	"io"
	"strings"

	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

type target struct {
	location string
	r        io.ReadCloser
}

func ExtractSelector(r io.Reader, selector string) ([]string, error) {

	out := []string{}

	sel, err := css.Compile(selector)
	if err != nil {
		return out, err
	}

	node, err := html.Parse(r)
	if err != nil {
		return out, err
	}

	// it's kind of tricky to actually know what to output
	// if the resulting tags contain more than just a text node
	for _, ele := range sel.Select(node) {
		if ele.FirstChild == nil {
			continue
		}
		out = append(out, ele.FirstChild.Data)
	}

	return out, nil
}

func ExtractComments(r io.Reader) []string {

	z := html.NewTokenizer(r)

	out := []string{}
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.CommentToken {
			d := strings.Replace(t.Data, "\n", " ", -1)
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			out = append(out, d)
		}

	}
	return out
}

func ExtractAttribs(r io.Reader, attribs []string) []string {
	z := html.NewTokenizer(r)

	out := []string{}

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		for _, a := range t.Attr {

			if a.Val == "" {
				continue
			}

			for _, attrib := range attribs {
				if attrib == a.Key {
					out = append(out, a.Val)
				}
			}
		}
	}
	return out
}

func ExtractTags(r io.Reader, tags []string) []string {
	z := html.NewTokenizer(r)

	out := []string{}

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.StartTagToken {

			for _, tag := range tags {
				if t.Data == tag {
					if z.Next() == html.TextToken {
						text := strings.TrimSpace(z.Token().Data)
						if text == "" {
							continue
						}
						out = append(out, text)
					}
				}
			}
		}
	}
	return out
}
