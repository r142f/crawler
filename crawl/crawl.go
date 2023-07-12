package crawl

import (
	"crawler/links"
	"log"
	"sync"
)

type item struct {
	link  string
	depth uint
}

type crawler struct {
	queue             chan *item
	visited           map[string]struct{}
	visitedLock       *sync.RWMutex
	wg                *sync.WaitGroup
	parallelismDegree uint
	maxDepth          uint
	rootPath          string
}

func New(parallelismDegree, maxDepth uint, rootPath string) *crawler {
	return &crawler{
		queue:             make(chan *item),
		visited:           make(map[string]struct{}),
		visitedLock:       &sync.RWMutex{},
		wg:                &sync.WaitGroup{},
		parallelismDegree: parallelismDegree,
		maxDepth:          maxDepth,
		rootPath:          rootPath,
	}
}

func (cr *crawler) crawl() {
	for it := range cr.queue {
		link, depth := it.link, it.depth

		cr.visitedLock.Lock()
		if _, ok := cr.visited[link]; ok {
			log.Printf("already crawled %s\n", link)
			cr.visitedLock.Unlock()
			cr.wg.Done()

			continue
		}

		cr.visited[link] = struct{}{}
		cr.visitedLock.Unlock()

		links, err := links.Process(link, cr.rootPath)
		if err != nil {
			log.Printf("processing link %v: %v\n", link, err)
		}

		if depth < cr.maxDepth {
			for _, link := range links {
				cr.visitedLock.RLock()
				if _, ok := cr.visited[link]; !ok {
					cr.wg.Add(1)
					go func(link string) {
						cr.queue <- &item{link, depth + 1}
					}(link)
				}
				cr.visitedLock.RUnlock()
			}
		}

		cr.wg.Done()
	}
}

func (cr *crawler) Run(links []string) {
	for _, link := range links {
		cr.wg.Add(1)
		go func(link string) {
			cr.queue <- &item{link, 0}
		}(link)
	}

	for i := 0; i < int(cr.parallelismDegree); i++ {
		go cr.crawl()
	}

	cr.wg.Wait()
	close(cr.queue)
}
