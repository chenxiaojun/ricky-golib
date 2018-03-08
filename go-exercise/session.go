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
	s.authenticated()
}

// authenticated 检查是否已经登录 (cookies 没有失效)
func (s *Session) authenticated() {
	originURL := makeZhihuLink("/settings/profile")
	fmt.Println(originURL)
}























