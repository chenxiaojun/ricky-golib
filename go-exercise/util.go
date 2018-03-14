package exercise

import (
	"strings"
	"net/http"
	"regexp"
	"path/filepath"
	"os"
	"os/exec"
	"runtime"
	"fmt"
	
	"github.com/fatih/color"
	"github.com/PuerkitoBio/goquery"
)

const (
	userAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36"
	baseZhihuURL = "https://www.zhihu.com"
)

var (
	reIsPhone = regexp.MustCompile(`(13|14|15|17|18|19)[0-9]{9}`)
	reIsEmail = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	logger = Logger{Enabled: true}
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

func newHTTPHeaders(isXhr bool) http.Header {
	headers := make(http.Header)
	headers.Set("Accept", "*/*")
	headers.Set("Connection", "keep-alive")
	headers.Set("Host", "www.zhihu.com")
	headers.Set("Origin", "http://www.zhihu.com")
	headers.Set("Pragma", "no-cache")
	headers.Set("User-Agent", userAgent)
	if isXhr {
		headers.Set("X-Requested-With", "XMLHttpRequest")
	}
	return headers
}

func isEmail(value string) bool {
	return reIsEmail.MatchString(value)
}

func isPhone(value string) bool {
	return reIsPhone.MatchString(value)
}

func openCaptchaFile(filename string) error {
	logger.Info("调用外部程序渲染验证码......")
	var args []string
	switch runtime.GOOS {
	case "linux":
		args = []string{"xdg-open", filename}
	case "darwin":
		args = []string{"open", filename}
	case "freebsd":
		args = []string{"open", filename}
	case "netbsd":
		args = []string{"open", filename}
	case "windows":
		var (
			cmd = "url.dll, FileProtocolHandler"
			runDll32 = filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
		)
		args = []string{runDll32, cmd, filename}
	default:
		fmt.Printf("无法确定操作系统，请自行打开验证码 %s 文件，并输入验证码。", filename)
	}

	logger.Info("Command: %s", strings.Join(args, " "))

	err := exec.Command(args[0], args[1:]...).Run()
	if err != nil {
		return err
	}

	return nil
}

func readCaptchaInput() string {
	var captcha string
	fmt.Print(color.CyanString("请输入验证码："))
	fmt.Scanf("%s", &captcha)
	return captcha
}

// ZhihuPage 是一个知乎页面，User, Question, Answer, Collection的公共部分
type Page struct {
	// Link 是该页面的链接
	Link string

	// doc 是 HTML document
	doc *goquery.Document

	// fields 是字段缓存，避免重复解析页面
	fields map[string]interface{}
}

// newZhihuPage 是 private 的构造器
func newZhihuPage(link string) *Page {
	return &Page {
		Link: link,
		fields: make(map[string]interface{}),
	}
}

// Doc 用于获取当前问题页面的 HTML document， 惰性求值
func (page *Page) Doc() *goquery.Document {
	if page.doc != nil {
		return page.doc
	}

	err := page.Refresh()
	if err != nil {
		return nil
	}

	return page.doc
}

// Refresh 会重新加载当前页面，获取最新的数据
func (page *Page) Refresh() (err error) {
	page.fields = make(map[string]interface{})
	page.doc, err = newDocumentFromUrl(page.Link)
	return err
}

// newDocumentFromURl 会请求给定的url 并返回一个goquery.Document对象用于解析
func newDocumentFromUrl(url string) (*goquery.Document, error) {
	resp, err := gSession.Get(url)
	if err != nil {
		logger.Error("请求 %s 失败：%s", url, err.Error())
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		logger.Error("解析页面失败：%s", err.Error())
	}
	return doc, err
}

func (page *Page) getStringField(field string) (value string, exists bool) {
	if got, ok := page.fields[field]; ok {
		return got.(string), true
	}
	return "", false
}

func (page *Page) setField(field string, value interface{}) {
	page.fields[field] = value
}

func strip(s string) string {
	return strings.TrimSpace(s)
}





























