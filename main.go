package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/keimoon/gore"
	"github.com/vedhavyas/nuvi-news/parsers"
	"github.com/vedhavyas/nuvi-news/util"
)

var redisPool = &gore.Pool{
	InitialConn: 10,
	MaximumConn: 20,
}

const NEWS_LIST = "NEWS_XML"
const PROC_ZIPS = "PROCESSED_ZIPS"

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	url := flag.String("URL", "http://bitly.com/nuvi-plz", "")
	redisAddr := flag.String("redis-url", "localhost:6379", "redis-url")
	flag.Parse()

	filesMap, err := parsers.GetZIPLinksFromURL(*url)
	if err != nil {
		log.Fatal(err)
	}

	err = redisPool.Dial(*redisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer redisPool.Close()

	conn, err := redisPool.Acquire()
	if err != nil {
		log.Fatal(err)
	}

	if conn == nil {
		log.Fatal("Connection to Redis Failed")
	}

	filesMap, err = util.FilterMap(conn, PROC_ZIPS, filesMap)
	redisPool.Release(conn)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for name, url := range filesMap {
		wg.Add(1)
		go func(name, url string) {
			if err := downloadAndProcessFile(name, url); err != nil {
				log.Println(err)
			}
			wg.Done()
		}(name, url)
	}

	wg.Wait()
	log.Println("All done")
}

func downloadAndProcessFile(fileName, url string) error {
	log.Println("Downnloading zip", fileName)
	filePath := fmt.Sprintf("%s/%s", "/tmp", fileName)
	fileOut, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer fileOut.Close()

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

	dir := strings.Replace(filePath, ".", "_", -1)
	if err := util.Unzip(filePath, dir); err != nil {
		return err
	}

	if err := processXMLs(dir, fileName); err != nil {
		return err
	}

	return nil
}

func processXMLs(dir, name string) error {
	var fileMap = make(map[string]string, 0)
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileMap[f.Name()] = path
			return nil
		}
		return nil
	})

	if err != nil {
		return err
	}

	conn, err := redisPool.Acquire()
	if err != nil {
		return err
	}

	if conn == nil {
		return errors.New("Connection to Redis Failed")
	}

	defer redisPool.Release(conn)

	fileMap, err = util.FilterMap(conn, name, fileMap)
	if err != nil {
		return err
	}

	for _, file := range fileMap {
		err := uploadToRedis(conn, dir, file)
		if err != nil {
			return err
		}
	}

	_, err = gore.NewCommand("RPUSH", PROC_ZIPS, name).Run(conn)

	return err
}

func uploadToRedis(conn *gore.Conn, dir, filePath string) error {
	dataFile, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer dataFile.Close()

	dataBytes, err := ioutil.ReadAll(dataFile)
	_, err = gore.NewCommand("RPUSH", NEWS_LIST, string(dataBytes)).Run(conn)
	if err != nil {
		return err
	}

	splitData := strings.Split(filePath, "/")
	fileName := splitData[len(splitData)-1]
	_, err = gore.NewCommand("RPUSH", dir, fileName).Run(conn)
	return err
}
