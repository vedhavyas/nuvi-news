package parsers

import (
	"net/http"
	"strings"

	"fmt"

	"golang.org/x/net/html"
)

//GetZIPLinksFromURL parses html from the URL and return a list of names of file under <a> tag available in depth 1
func GetZIPLinksFromURL(url string) ([]string, error) {
	return getLinks(url, ".zip")
}

func getLinks(url, fileFormat string) ([]string, error) {
	response, err := http.Get(url)
	if err != nil {
		return []string{}, err
	}

	doc, err := html.Parse(response.Body)
	if err != nil {
		return []string{}, err
	}
	return parseHTML(response.Request.URL.String(), doc, []string{}, fileFormat), nil
}

func parseHTML(baseURL string, content *html.Node, linkFiles []string, fileFormat string) []string {
	if content.Type == html.ElementNode && content.Data == "a" {
		file := content.Attr[0].Val
		// this is <a>
		if strings.HasSuffix(file, fileFormat) {
			linkFiles = append(linkFiles, fmt.Sprintf("%s/%s", baseURL, file))
		}
	}
	for c := content.FirstChild; c != nil; c = c.NextSibling {
		linkFiles = parseHTML(baseURL, c, linkFiles, fileFormat)
	}

	return linkFiles
}
