package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var stopWords = map[string]bool {
	"a": true,
	"the": true,
	"is": true,
	"and": true,
	"of": true,
	"to": true,
	"in": true,
}

var queue = []string {
	"https://example.com",
}

var visited = make(map[string]bool)

var index = make(map[string]map[string]bool)



func filterStopWords(words []string) []string {
	var res []string
	for _, w := range words {
		if !stopWords[w] {
			res = append(res, w)
		}
	}
	return res
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	regex := regexp.MustCompile("[a-z0-9]+")
	res := regex.FindAllString(text, -1)
	return filterStopWords(res)
}

func resolveURL(baseStr string, refStr string) string {
	base, err := url.Parse(baseStr)
	if err != nil {
		panic(err)
	}

	ref, err := url.Parse(refStr)
	if err != nil {
		panic(err)
	}

	return base.ResolveReference(ref).String()
}

func processHTML(baseURL string, n *html.Node) {
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}
	
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			for _, word := range tokenize(text) {
				if index[word] == nil {
					index[word] = make(map[string]bool)
				}
				index[word][baseURL] = true
			}
		}
	}

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				href := strings.TrimSpace(attr.Val)

				if href == "" ||
					strings.HasPrefix(href, "#") ||
					strings.HasPrefix(href, "javscript:") ||
					strings.HasPrefix(href, "mailto:") {
						continue
				}
				
				url := resolveURL(baseURL, href)
				queue = append(queue, url)
			}
		}
	}

	for doc := n.FirstChild; doc != nil; doc = doc.NextSibling {
		processHTML(baseURL, doc)
	}
}

func fetch(url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}


}

func saveIndex(filename string, idx InvertedIndex) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	enc := json.NewEncoder(fd)
	enc.SetIndent("", " ")
	return enc.Encode(idx)
}

func sayHello(name string) {
	fmt.Println("Hello from another function! ", name)
}

func main() {
	fmt.Println("Hello, World!")
	sayHello("Alice")
}