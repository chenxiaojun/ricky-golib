package exercise

import (
	"net/http"
	"net/url"
	"os"
	"encoding/json"
	"path/filepath"
	"fmt"
	"time"
	"io"
	"strings"
	"strconv"
	"io/ioutil"

	"github.com/juju/persistent-cookiejar"
)

type Auth struct {
	Account string `json:"account"`
	Password string `json:"password"`

	loginType string
	loginURL string
}

type Session struct {
	auth *Auth
	client *http.Client
}

type loginResult struct {
	R         int         `json:"r"`
	Msg       string      `json:"msg"`
	ErrorCode int         `json:"errcode"`
	Data      interface{} `json:"data"`
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
}

// Login 登录并保存 cookies
func (s *Session) Login() error {
	if s.authenticated() {
		logger.Success("已经是登录状态, 不需要重复登录")
		return nil
	}

	// _xsrf=&captcha=VRKM&email=rickycxj%40gmail.com&password=test123&remember_me=true
	form := s.buildLoginForm().Encode()
	body := strings.NewReader(form)
	req, err := http.NewRequest("POST", s.auth.loginURL, body)
	if err != nil {
		logger.Error("构造登录请求失败：%s", err.Error())
		return err
	}

	headers := newHTTPHeaders(true)
	headers.Set("Content-Length", strconv.Itoa(len(form)))
	headers.Set("Content-Type", "application/x-www-form-urlencoded")
	headers.Set("Referer", baseZhihuURL)
	req.Header = headers

	logger.Info("登录中，用户名： %s", s.auth.Account)

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error("登录失败: %s", err.Error())
		return err
	}

	logger.Info("content-type: ", resp.Header.Get("Content-Type"))

	if strings.ToLower(resp.Header.Get("Content-Type")) != "application/json" {
		logger.Error("服务器没有返回 json 数据")
		return fmt.Errorf("未知的 Content-Type: %s", resp.Header.Get("Content-Type"))
	}

	defer resp.Body.Close()

	result := loginResult{}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("读取响应内容失败: %s", err.Error())
	}
	logger.Info("登录响应内容： %s", strings.Replace(string(content), "\n", "", -1))

	err = json.Unmarshal(content, &result)
	if err != nil {
		logger.Error("JSON解析失败：%s", err.Error())
		return err
	}

	if result.R == 0 {
		logger.Success("登录成功! ")
		s.client.Jar.(*cookiejar.Jar).Save()
		return nil
	}

	if result.R == 1 {
		logger.Warn("登录失败! 原因：%s", result.Msg)
		return fmt.Errorf("登录失败! 原因： %s", result.Msg)
	}

	logger.Error("登录出现未知错误： %s", string(content))
	return fmt.Errorf("登录失败，未知错误：%s", string(content))
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

func (s *Session) buildLoginForm() url.Values {
	values := s.auth.toForm()
	values.Set("_xsrf", s.searchXSRF())
	values.Set("captcha", s.downloadCaptcha())
	return values
}

func (auth *Auth) toForm() url.Values {
	if auth.isEmail() {
		auth.loginType = "email"
		auth.loginURL = makeZhihuLink("/login/email")
	} else if auth.isPhone() {
		auth.loginType = "phone_num"
		auth.loginURL = makeZhihuLink("/login/phone_num")
	} else {
		panic("无法判断登录类型：" + auth.Account)
	}
	values := url.Values{}
	logger.Info("登录类型： %s, 登录地址：%s", auth.loginType, auth.loginURL)
	values.Set(auth.loginType, auth.Account)
	values.Set("password", auth.Password)
	values.Set("remember_me", "true")
	return values
}

func (auth *Auth) isEmail() bool {
	return isEmail(auth.Account)
}

func (auth *Auth) isPhone() bool {
	return isPhone(auth.Account)
}

func (s *Session) searchXSRF() string {
	resp, err := s.Get(baseZhihuURL)
	if err != nil {
		panic("获取 _xsrf失败：" + err.Error())
	}

	// retrieve from cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_xsrf" {
			logger.Info("_xsrf: ", cookie.Value)
			return cookie.Value
		}
	}
	return ""
}

func (s *Session) downloadCaptcha() string {
	url := makeZhihuLink(fmt.Sprintf("/captcha.gif?r=%d&type=login", 1000*time.Now().Unix()))
	logger.Info("获取验证码： %s", url)
	resp, err := s.Get(url)
	if err != nil {
		panic("获取验证码失败：" + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("获取验证码失败，StatusCode = %d", resp.StatusCode))
	}

	defer resp.Body.Close()

	fileExt := strings.Split(resp.Header.Get("Content-Type"), "/")[1]
	verifyImg := filepath.Join(getCwd(), "verify."+fileExt)
	fd, err := os.OpenFile(verifyImg, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic("打开验证码文件失败：" + err.Error())
	}
	defer fd.Close()

	io.Copy(fd, resp.Body) //保存验证码文件
	openCaptchaFile(verifyImg)
	captcha := readCaptchaInput()

	return captcha
}

func getCwd() string{
	cwd, err := os.Getwd()
	if err != nil {
		panic("获取CWD失败：" + err.Error())
	}
	logger.Info("cwd: ", cwd)
	return cwd
}





















