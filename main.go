package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	//url := "http://bitly.com/nuvi-plz"
	//
	//fileList, err := parsers.GetZIPLinksFromURL(url)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//for _, file := range fileList {
	//	fmt.Println(file)
	//}

	//fileURL := "http://feed.omgili.com/5Rh5AMTrc4Pv/mainstream/posts/1472752689118.zip"
	//splitData := strings.Split(fileURL, "/")
	//fileName := splitData[len(splitData)-1]
	//if err := downloadAndProcessFile(fileURL, fileName); err != nil {
	//	log.Fatal(err)
	//}

	//fileName := "1472752689118.zip"
	//util.Unzip(fileName, strings.Replace(fileName, ".", "_", -1))
}

func downloadAndProcessFile(url, fileName string) error {
	fileOut, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer fileOut.Close()

	fmt.Println("Downnloading file ", fileName)
	res, err := http.Get(url)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}

	_, err = io.Copy(fileOut, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func processXMLs(dir string) {

	dir, err := os.Open(dir)
	if err != nil {
		return err
	}

}
