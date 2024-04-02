package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

type DefaultParser struct {
}
type SeoData struct {
	URL             string
	MetaDescription string
	StatusCode      int
	Title           string
	H1              string
}

func isSitemap(urls []string) ([]string, []string) {
	sitemapFiles := []string{}
	pages := []string{}
	for _, page := range urls {
		foundSitemap := strings.Contains(page, "xml")
		if foundSitemap == true {
			fmt.Println("Found Sitemap", page)
			sitemapFiles = append(sitemapFiles, page)
		} else {
			pages = append(pages, page)
		}
	}
	return sitemapFiles, pages
}
func (d DefaultParser) getSeoData(resp *http.Response) (SeoData, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return SeoData{}, err

	}
	result := SeoData{}
	result.URL = resp.Request.URL.String()
	result.StatusCode = resp.StatusCode
	result.Title = doc.Find("title").First().Text()
	result.MetaDescription, _ = doc.Find("meta[name=description]").Attr("content")
	result.H1 = doc.Find("h1").First().Text()
	return result, nil

}

type Parser interface {
	getSeoData(*http.Response) (SeoData, error)
}

func randomUserAgent() string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}
func ScrapeURLs(urls []string, parser Parser, concurrency int) []SeoData {
	tokens := make(chan struct{}, concurrency)
	var n int
	n++
	worklist := make(chan []string)
	results := []SeoData{}
	go func() { worklist <- urls }()
	for ; n > 0; n-- {
		list := <-worklist
		for _, url := range list {
			if url != "" {
				n++
				go func(url string, token chan struct{}) {
					log.Printf("Requesting URL: %s", url)
					res, err := scrapePage(url, tokens, parser)
					if err != nil {
						log.Printf("Encountered error, URL: %s", url)
					} else {
						results = append(results, res)
					}
					worklist <- []string{}
				}(url, tokens)
			}
		}
	}
	return results
}
func ScrapeSitemap(url string, parser Parser, concurrency int) []SeoData {
	urls := ExtractSitemapURLs(url)
	log.Print(urls)
	data := ScrapeURLs(urls, parser, concurrency)
	return data
}
func makeRequest(url string) (*http.Response, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", randomUserAgent())
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func extractURLs(response *http.Response) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}
	results := []string{}
	sel := doc.Find("loc")
	for i := range sel.Nodes {
		loc := sel.Eq(i)
		result := loc.Text()
		results = append(results, result)
	}
	return results, nil

}
func ExtractSitemapURLs(startURL string) []string {
	workList := make(chan []string)
	toCrawl := []string{}
	var n int
	n++
	go func() {
		workList <- []string{startURL}
	}()
	for ; n > 0; n-- {
		list := <-workList
		for _, link := range list {
			n++
			go func(link string) {
				response, err := makeRequest(link)
				if err != nil {
					log.Printf("Error fetching %s: %v", link, err)
				}
				urls, _ := extractURLs(response)
				if err != nil {
					log.Printf("Error fetching %s: %v", link, err)
				}
				sitemapFiles, pages := isSitemap(urls)
				if sitemapFiles != nil {
					workList <- sitemapFiles
				}
				for _, page := range pages {
					toCrawl = append(toCrawl, page)
				}

			}(link)
		}
	}
	return toCrawl
}
func scrapePage(url string, token chan struct{}, parser Parser) (SeoData, error) {
	res, err := crawlPage(url, token)
	if err != nil {
		return SeoData{}, err
	}
	data, err := parser.getSeoData(res)
	if err != nil {
		return SeoData{}, err
	}
	return data, nil

}
func crawlPage(url string, tokens chan struct{}) (*http.Response, error) {
	tokens <- struct{}{}
	res, err := makeRequest(url)
	<-tokens
	if err != nil {
		return nil, err
	}
	return res, nil

}
func main() {
	log.Print("Starting the scraper")
	p := DefaultParser{}
	results := ScrapeSitemap("https://www.quicksprout.com/sitemap.xml", p, 10)
	for _, result := range results {
		fmt.Println(result)
	}
}
