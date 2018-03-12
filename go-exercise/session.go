package exercise

import (
	"net/http"
	"net/http/cookiejar"
	"os"
	"encoding/json"
	"fmt"
)

type Auth struct {
	Account string
	Password string

	loginType string
	loginURL string
}

type Session struct {
	auth *Auth
	client *http.Client
}

var (
	gSession = NewSession()
)

func Init(cfgFile string) {
	gSession.LoadConfig(cfgFile)
	gSession.Login()
}

// NewSession 创建并返回一个*Session 对象
func NewSession() *Session{
	s := new(Session)
	cookieJar, _ := cookiejar.New(nil)
	s.client = &http.Client{
		Jar: cookieJar,
	}
	return s
}

func (s *Session) LoadConfig(cfg string) {
	file, err := os.Open(cfg)
	if err != nil {
		panic("无法打开配置文件 config.json: " + err.Error())
	}
	defer file.Close()

	auth := new(Auth)
	err = json.NewDecoder(file).Decode(&auth)
	if err != nil {
		panic("解析配置文件出错: " + err.Error())
	}
	s.auth = auth
	fmt.Println(auth)
}

// Login 登录并保存 cookies
func (s *Session) Login() {
	if s.authenticated() {
		logger.Success("已经是登录状态, 不需要重复登录")
		//return nil
	}

	s.buildLoginForm()
}

// authenticated 检查是否已经登录 (cookies 没有失效)
func (s *Session) authenticated() bool {
	originURL := makeZhihuLink("/settings/profile")
	resp, err := s.Get(originURL)
	if err != nil {
		logger.Error("访问 profile 页面出错：%s", err.Error())
		return false
	}

	// 如果没有登录，会跳转到 http://www.zhihu.com/?next=%sFsettings%2Fprofile
	lastURL := resp.Request.URL.String()
	logger.Info("获取 profile 的请求， 跳转到了：%s", lastURL)
	return lastURL == originURL
}

func (s *Session) Get(url string) (*http.Response, error) {
	logger.Info("GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("NewRequest failed with URL: %s", url)
		return nil, err
	}
	req.Header = newHTTPHeaders(false)
	return s.client.Do(req)
}

func (s *Session) buildLoginForm() {
	s.auth.toForm()
}

func (auth *Auth) toForm() {
	fmt.Println("test...", auth.isEmail())
}

func (auth *Auth) isEmail() bool {
	return isEmail(auth.Account)
}























