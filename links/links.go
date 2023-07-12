package links

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/net/html"
)

func getIdxOfAttributeWithLink(n *html.Node, u *url.URL) int {
	if _, ok := TAGS_WITH_SRC_OR_HREF_ATTRIBUTE[n.Data]; n.Type == html.ElementNode && ok {
		for i, attribute := range n.Attr {
			if attribute.Key != "href" && attribute.Key != "src" {
				continue
			}

			link, err := u.Parse(n.Attr[i].Val)
			if err != nil || link.Hostname() != u.Hostname() {
				continue
			}

			return i
		}
	}

	return -1
}

func extractLinkFromNode(n *html.Node, u *url.URL) (string, error) {
	if n.Data != "a" && n.Data != "area" && n.Data != "base" {
		return "", nil
	}

	i := getIdxOfAttributeWithLink(n, u)
	if i == -1 {
		return "", nil
	}

	link, err := u.Parse(n.Attr[i].Val)
	if err != nil {
		return "", fmt.Errorf("parsing url %s: %v", n.Attr[i].Val, err)
	}
	link.RawQuery = ""
	link.Fragment = ""

	return link.String(), nil
}

func localizeLinkInNode(n *html.Node, u *url.URL, rootPath string) error {
	i := getIdxOfAttributeWithLink(n, u)
	if i == -1 {
		return nil
	}

	link, err := u.Parse(n.Attr[i].Val)
	if err != nil {
		return fmt.Errorf("parsing url %s: %v", n.Attr[i].Val, err)
	}

	if n.Data == "a" || n.Data == "area" || n.Data == "base" {
		path := filepath.Join(rootPath, link.Hostname(), link.Path)
		if filepath.Ext(path) != ".html" {
			link = link.JoinPath("index.html")
		}
		link.Path = fmt.Sprintf("%s/%s%s", rootPath, link.Hostname(), link.Path)
		link.Scheme = ""
		link.User = nil
		link.Host = ""
	}

	n.Attr[i].Val = link.String()

	return nil
}

func processLinksFromPage(doc *html.Node, resp *http.Response, rootPath string) (links []string) {
	queue := list.New()
	pushChildNodes(queue, doc)

	for queue.Len() > 0 {
		node := queue.Remove(queue.Front()).(*html.Node)

		extractedlink, err := extractLinkFromNode(node, resp.Request.URL)
		if err != nil {
			log.Printf("extractLinkFromNode(%v, %v): %v\n", node, resp.Request.URL, err)
		}
		if extractedlink != "" {
			links = append(links, extractedlink)
		}
		err = localizeLinkInNode(node, resp.Request.URL, rootPath)
		if err != nil {
			log.Printf("localizeLinkInNode(%v, %v, %v): %v\n", node, resp.Request.URL, rootPath, err)
		}

		pushChildNodes(queue, node)
	}

	return links
}

func safePage(doc *html.Node, resp *http.Response, rootPath string) error {
	u := resp.Request.URL

	path := filepath.Join(rootPath, u.Hostname(), u.Path)
	if filepath.Ext(path) != ".html" {
		path = filepath.Join(path, "index.html")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return fmt.Errorf("creating parent directories of %s: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file with path %s: %v", path, err)
	}
	defer file.Close()

	if err = html.Render(file, doc); err != nil {
		return fmt.Errorf("rendering html to %s: %v", path, err)
	}

	log.Printf("saved %s\n", path)

	return nil
}

func Process(link, rootPath string) (links []string, err error) {
	log.Printf("started to process %s\n", link)

	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting %s: %s", link, resp.Status)
	}

	if !checkHeader(resp, "Content-Type", "text/html") {
		log.Printf("not html page: %s\n", link)
		return nil, nil
	}

	if u, err := url.Parse(link); err != nil {
		return nil, fmt.Errorf("parsing url %s: %v", link, err)
	} else if resp.Request.URL.Hostname() != u.Hostname() {
		log.Printf("got redirected to another domain from %s to %s\n", link, resp.Request.URL.String())
		return nil, nil
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing %s as HTML: %v", link, err)
	}

	links = processLinksFromPage(doc, resp, rootPath)

	err = safePage(doc, resp, rootPath)
	if err != nil {
		return links, fmt.Errorf("saving %s: %v", link, err)
	}

	return links, nil
}
