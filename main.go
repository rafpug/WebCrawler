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

type InvertedIndex map[string][]string

var stopWords = map[string]bool {
	"a": true,
	"the": true,
	"is": true,
	"and": true,
	"of": true,
	"to": true,
	"in": true,
}

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

func processHTML(n *html.Node) {
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}
	
	if n.Type == html.TextNode {

	}

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				href := strings.TrimSpace(attr.Val)
			}
		}
	}

	for doc := n.FirstChild; doc != nil; doc = doc.NextSibling {
		processHTML(doc)
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

	doc, err := html.Parse(resp)
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