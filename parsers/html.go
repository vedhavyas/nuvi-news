package parsers

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

//GetZIPLinksFromURL parses html from the URL and return a list of names of file under <a> tag available in depth 1
func GetZIPLinksFromURL(url string) (map[string]string, error) {
	return getLinks(url, ".zip")
}

func getLinks(url, fileFormat string) (map[string]string, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(response.Body)
	if err != nil {
		return nil, err
	}
	return parseHTML(response.Request.URL.String(), doc, make(map[string]string), fileFormat), nil
}

func parseHTML(baseURL string, content *html.Node, linksMap map[string]string, fileFormat string) map[string]string {
	if content.Type == html.ElementNode && content.Data == "a" {
		file := content.Attr[0].Val
		// this is <a>
		if strings.HasSuffix(file, fileFormat) {
			linksMap[file] = fmt.Sprintf("%s/%s", baseURL, file)
		}
	}
	for c := content.FirstChild; c != nil; c = c.NextSibling {
		linksMap = parseHTML(baseURL, c, linksMap, fileFormat)
	}

	return linksMap
}
