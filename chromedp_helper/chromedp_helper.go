package chrome_hepler

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/songzhibin97/gkit/distributed/retry"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ChromeHelper struct {
	timeout      time.Duration
	options      []chromedp.ExecAllocatorOption
	Headers      http.Header
	Server       string
	ResponseBody string
	Response     *http.Response
	ResponseUrl  string
	Title        string
	HtmlBody     string
	Banner       string
	retryCount   int
}

func NewChrome() *ChromeHelper {
	chrome := &ChromeHelper{
		Headers:  make(http.Header),
		Response: &http.Response{},
	}
	chrome.options = []chromedp.ExecAllocatorOption{
		// 龙芯设备添加 ExecPath
		// chromedp.ExecPath("/usr/bin/lbrowser"),
		chromedp.IgnoreCertErrors,
		chromedp.DisableGPU,
		chromedp.Flag("blink-settings", "imageEnable=false"),
		chromedp.Flag("enable-automation", true),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.66 Safari/537.36`),
	}
	chrome.options = append(chromedp.DefaultExecAllocatorOptions[:], chrome.options...)
	chrome.timeout = 10 * time.Second
	chrome.retryCount = 2
	return chrome
}

var emptyStruct = struct{}{}

func (chrome *ChromeHelper) Run(ctx context.Context, u *url.URL) {
retry:
	urlStr := u.String()
	protocol := u.Scheme
	port := 0
	addr := u.Host
	hostSlice := strings.Split(u.Host, ":")
	if len(hostSlice) == 2 {
		addr = hostSlice[0]
		port, _ = strconv.Atoi(hostSlice[1])
	}
	if port == 0 { // 处理默认情况的端口号
		if u.Scheme == "http" {
			port = 80
		} else {
			port = 443
		}
	}

	//log.Println("开始web请求: ", urlStr)
	defer func() {
		//log.Println("完成web请求: ", urlStr)
		var err any
		err = recover()
		if err != nil {
			log.Println(err)
		}
	}()
	chrome.HtmlBody = ""
	chrome.Title = ""
	chrome.ResponseBody = ""
	chrome.Server = ""
	chrome.Headers = make(http.Header)
	chrome.Banner = ""
	timeoutCtx, timeoutCtxCancelFunc := context.WithTimeout(ctx, chrome.timeout)
	allocatorCtx, allocatorCancelFunc := chromedp.NewExecAllocator(timeoutCtx, chrome.options...)
	chromeCtx, chromeCtxCancelFunc := chromedp.NewContext(allocatorCtx)
	defer func() {
		chromeCtxCancelFunc()
		allocatorCancelFunc()
		timeoutCtxCancelFunc()
	}()

	urlWightList := map[string]struct{}{
		urlStr:       emptyStruct,
		urlStr + "/": emptyStruct,
	}

	if port == 80 || (port == 443 && protocol == "https") {
		urlWightList[fmt.Sprintf("%s://%s", protocol, addr)] = emptyStruct
		urlWightList[fmt.Sprintf("%s://%s/", protocol, addr)] = emptyStruct
	}
	firstResp := true
	chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
		switch ev.(type) {
		case *network.EventResponseReceived:
			resp := ev.(*network.EventResponseReceived).Response
			if !firstResp {
				shouldReturn := true
				_, ok := urlWightList[resp.URL]
				shouldReturn = !ok
				if shouldReturn {
					return
				}
			}
			firstResp = false
			chrome.ResponseUrl = resp.URL
			chrome.Response.Proto = resp.Protocol
			chrome.Response.StatusCode = int(resp.Status)
			chrome.Response.Status = fmt.Sprintf("%d %s", resp.Status, resp.StatusText) // 转换成go http包的格式
			chrome.Banner += fmt.Sprintf("%s %d %s\n", resp.Protocol, resp.Status, resp.StatusText)
			for k, v := range resp.Headers {
				if k == "Server" {
					chrome.Server = fmt.Sprintf("%s", v)
				}
				chrome.Headers[k] = append(chrome.Headers[k], fmt.Sprintf("%v", v))
				chrome.Banner += fmt.Sprintf("%s: %s\n", k, v)
			}
		}
	})
	var nodes []*cdp.Node
	var html string
	err := chromedp.Run(chromeCtx,
		chromedp.Navigate(urlStr),
	)
	if err != nil {
		log.Println(urlStr + " " + err.Error())
		return
	}
	preDataLen := 0
	retryCnt := 10
	for i := 0; i < retryCnt; i++ {
		var data []byte
		err = chromedp.Run(chromeCtx,
			chromedp.Sleep(time.Second*2),
			chromedp.Nodes(`document.querySelector("html")`, &nodes, chromedp.ByJSPath),
			chromedp.OuterHTML(`document.querySelector("html")`, &html, chromedp.ByJSPath),
			chromedp.Title(&chrome.Title),
			chromedp.OuterHTML(`document.querySelector("Body")`, &chrome.HtmlBody, chromedp.ByJSPath),
			chromedp.FullScreenshot(&data, 80),
		)
		if err != nil {
			t := time.Duration(retry.FibonacciNext(i))
			log.Println(fmt.Sprintf("err:%v,将会在%d秒后重试", err, t))
			time.Sleep(t * time.Second)
			continue
		}
		//log.Println(chrome.Title)
		if len(chrome.Title) > 0 {
			if preDataLen >= len(data) { // TODO 这里没看懂为什么要这样才退出，这样最少也要两次正确的请求才会退出
				break
			}
			preDataLen = len(data)
		}
	}
	if len(chrome.Title) == 0 && chrome.retryCount >= 0 {
		chrome.retryCount--
		goto retry
	}
	if nodes != nil && len(nodes) > 0 {
		for _, htmlNode := range nodes {
			if strings.ToLower(htmlNode.NodeName) == "html" {
				document := htmlNode.Parent
				if document != nil && len(document.Children) > 0 {
					for _, child := range document.Children {
						nodeName := strings.ToLower(child.NodeName)
						if child.NodeType == cdp.NodeTypeComment {
							chrome.ResponseBody += "<!--" + child.NodeValue + "-->\n"
						} else if child.NodeValue != "" {
							chrome.ResponseBody += "<" + nodeName + ">" + child.NodeValue + "</" + nodeName + ">\n"
						}
					}
				}
				break
			}
		}
	}
	chrome.ResponseBody += html
	chrome.Banner += fmt.Sprintf("\n%s", chrome.ResponseBody)
}
