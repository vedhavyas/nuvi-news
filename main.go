package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"strings"

	"github.com/keimoon/gore"
	"github.com/vedhavyas/nuvi-news/util"
)

var redisPool gore.Pool

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

	err := redisPool.Dial("localhost:6379")
	if err != nil {
		log.Fatal(err)
	}

	fileName := "1472752689118.zip"
	util.Unzip(fileName, strings.Replace(fileName, ".", "_", -1))
	processXMLs("1472752689118_zip")

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

func processXMLs(dir string) error {
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return uploadToRedis(path)
	})

	if err != nil {
		return err
	}

	return nil
}

func uploadToRedis(filePath string) error {
	dataFile, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}

	defer dataFile.Close()

	dataBytes, err := ioutil.ReadAll(dataFile)

	data := string(dataBytes)

	fmt.Println(data)
	return nil
}

func filterList(conn gore.Conn, key string, fullList []string) ([]string, error) {
	reply, err := gore.NewCommand("lrange", key, "0 -1").Run(conn)
	if err != nil {
		return err
	}

	finishedList := []string{}
	reply.Slice(&finishedList)

	for _, finishedFile := range finishedList {
		for i, queueFile := range fullList {
			if queueFile == finishedFile {
				fullList = append(fullList[:i], fullList[i+1:]...)
				continue
			}
		}

	}

	return fullList
}
