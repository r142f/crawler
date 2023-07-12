package links

import (
	"container/list"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

var TAGS_WITH_SRC_OR_HREF_ATTRIBUTE = map[string]struct{}{
	"audio":  {},
	"embed":  {},
	"iframe": {},
	"img":    {},
	"input":  {},
	"script": {},
	"source": {},
	"track":  {},
	"video":  {},
	"a":      {},
	"link":   {},
	"area":   {},
	"base":   {},
}

func checkHeader(resp *http.Response, key, value string) bool {
	for _, v := range resp.Header[key] {
		if strings.Contains(v, value) {
			return true
		}
	}

	return false
}

func pushChildNodes(queue *list.List, node *html.Node) {
	for node := node.FirstChild; node != nil; node = node.NextSibling {
		queue.PushBack(node)
	}
}
