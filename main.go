package main

import "fmt"

type DefaultParser struct {
}
type SeoData struct {
	URL             string
	MetaDescription string
	StatusCode      int
	Title           string
	H1              string
}
type parser interface{}

func ScrapeURLs([]string)[]SeoData {

}
func ScrapeSitemap(url string) []SeoData {
	urls := ExtractSitemapURLs(url)
	data := ScrapeURLs(urls)
	return data
}
func makeRequest(url string){

}
func ExtractSitemapURLs(startURL string) []string {
	workList := make(chan []string)
	toCrawl := []string{}
	var n int
	n++
	go func(){
		workList <- []string{startURL}
	}()
	for ;n>0;n--{
	list:= <-workList
for _,link:=range list{
	go func (link string)  {
		response,err:=makeRequest(link)
		if(err!=nil){
			log.Printf("Error fetching %s: %v",link,err)
		}
		urls,_:=extractURLs(response)
		if(err!=nil){
			log.Printf("Error fetching %s: %v",link,err)
		}
	}(link)
}}
return toCrawl
}
func main() {
	p := DefaultParser{}
	results := ScrapeSitemap("")
	for _, result := range results {
		fmt.Println(result)
	}
}
