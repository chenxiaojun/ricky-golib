package exercise

import "strings"

const (
	baseZhihuURL = "https://www.zhihu.com"
)

func makeZhihuLink(path string) string {
	return urlJoin(baseZhihuURL, path)
}

func urlJoin(base, path string) string {
	if strings.HasSuffix(base, "/") {
		base = strings.TrimRight(base, "/")
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimLeft(path, "/")
	}
	return base + "/" + path
}
