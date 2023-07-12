package main

import (
	"crawler/crawl"
	"flag"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

var parallelismDegree = flag.Uint("par", 1, "> 0, sets the degree of parallelism. Essentially, it regulates the max number of concurrent requests")
var maxDepth = flag.Uint("depth", 1, ">= 0, sets the depth-limiting parameter")
var rootPath = flag.String("path", "", "sets the path where files will be stored (default cwd)")

func main() {
	flag.Parse()
	
	if *parallelismDegree == 0 {
		log.Fatalln("the degree of parallelism must be positive")
	}

	if absRootPath, err := filepath.Abs(*rootPath); err != nil {
		log.Printf("getting absoulute path %s: %v\n. Using os.Getwd() instead", *rootPath, err)
		dir, err := os.Getwd()
		*rootPath = dir
		if err != nil {
			log.Printf("os.Getwd(): %v\n", err)
			log.Println("serve files using localhost!")
		}
	} else {
		*rootPath = absRootPath
	}

	links := make([]string, 0, len(flag.Args()))
	for _, link := range flag.Args() {
		u, err := url.Parse(link)
		if err != nil {
			log.Printf("error while parsing url %v: %v\n", link, err)
		} else {
			u.Path = "/"
			u.RawQuery = ""
			u.Fragment = ""
			links = append(links, u.String())
		}
	}

	crawler := crawl.New(*parallelismDegree, *maxDepth, *rootPath)
	crawler.Run(links)
}
