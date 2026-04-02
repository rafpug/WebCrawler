package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

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
	"https://en.wiktionary.org/wiki/Wiktionary:Main_Page",
}

var visited = make(map[string]bool)

var index = make(map[string]map[string]bool)

var maxPages = 1000000

var regex = regexp.MustCompile("[a-z0-9]+")

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
	res := regex.FindAllString(text, -1)
	return filterStopWords(res)
}

func resolveURL(baseStr string, refStr string) (string, error) {
	base, err := url.Parse(baseStr)
	if err != nil {
		return "", err
	}

	ref, err := url.Parse(refStr)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(ref).String(), nil
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
				
				url, err := resolveURL(baseURL, href)
				if err != nil {
					continue
				}
				queue = append(queue, url)
			}
		}
	}

	for doc := n.FirstChild; doc != nil; doc = doc.NextSibling {
		processHTML(baseURL, doc)
	}
}

func fetch(url string) error {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "CrawlerBot")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	processHTML(url, doc)
	visited[url] = true
	return nil
}

func saveIndex(filename string) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	enc := json.NewEncoder(fd)
	enc.SetIndent("", " ")
	return enc.Encode(index)
}

func main() {
	startTime := time.Now()

	logFile, err := os.OpenFile("crawler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		if len(visited) > maxPages {
			fmt.Println("reached max!")
			break
		}

		err := fetch(cur)
		if err != nil {
			fmt.Println("URL: ", cur, " ", err)
		} else {
			logger.Println(cur)
			n := len(visited)
			for n%10 == 0 {
				n /= 10
			}
			if n == 1 {
				fmt.Println("Reached ", len(visited), " pages in ", time.Since(startTime))
			}
		}
	}
	saveIndex("test1")
	fmt.Println("Total Time Elapsed: ", time.Since(startTime))
}