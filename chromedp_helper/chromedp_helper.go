package chrome_hepler

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/songzhibin97/gkit/distributed/retry"
	"log"
	"strings"
	"time"
)

type ChromeHelper struct {
	timeout      time.Duration
	options      []chromedp.ExecAllocatorOption
	Headers      string
	Server       string
	ResponseBody string
	Title        string
	HtmlBody     string
	Banner       string
	Status       string
	StatusCode   int64
	Proto        string
	retryCount   int
}

func NewChrome() *ChromeHelper {
	chrome := &ChromeHelper{}
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

func (chrome *ChromeHelper) Run(ctx context.Context, protocol string, ip string, port string) {
retry:
	url := fmt.Sprintf("%s://%s:%s", protocol, ip, port)
	log.Println("开始web请求: ", url)
	defer func() {
		log.Println("完成web请求: ", url)
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
	chrome.Headers = ""
	chrome.Banner = ""
	timeoutCtx, timeoutCtxCancelFunc := context.WithTimeout(ctx, chrome.timeout)
	allocatorCtx, allocatorCancelFunc := chromedp.NewExecAllocator(timeoutCtx, chrome.options...)
	chromeCtx, chromeCtxCancelFunc := chromedp.NewContext(allocatorCtx)
	defer func() {
		chromeCtxCancelFunc()
		allocatorCancelFunc()
		timeoutCtxCancelFunc()
	}()
	urls := []string{url, url + "/"}
	if port == "80" || (port == "443" && protocol == "https") {
		urls = append(urls, fmt.Sprintf("%s://%s", protocol, ip), fmt.Sprintf("%s://%s/", protocol, ip))
	}
	firstResp := true
	chromedp.ListenTarget(chromeCtx, func(ev interface{}) {
		switch ev.(type) {
		case *network.EventResponseReceived:
			resp := ev.(*network.EventResponseReceived).Response
			if !firstResp {
				shouldReturn := true
				for _, u := range urls {
					if u == resp.URL {
						shouldReturn = false
					}
				}
				if shouldReturn {
					return
				}
			}
			firstResp = false
			chrome.Banner += fmt.Sprintf("%s %d %s\n", resp.Protocol, resp.Status, resp.StatusText)
			chrome.Status = fmt.Sprintf("%d %s\n", resp.Status, resp.StatusText)
			chrome.StatusCode = resp.Status
			chrome.Proto = resp.Protocol
			for k, v := range resp.Headers {
				if k == "Server" {
					chrome.Server = fmt.Sprintf("%s", v)
				}
				chrome.Headers += fmt.Sprintf("%s: %s;", k, v)
				chrome.Banner += fmt.Sprintf("%s: %s\n", k, v)
			}
		}
	})
	var nodes []*cdp.Node
	var html string
	err := chromedp.Run(chromeCtx,
		chromedp.Navigate(url),
	)
	if err != nil {
		log.Println(url + " " + err.Error())
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
		log.Println(chrome.Title)
		if len(chrome.Title) > 0 {
			if preDataLen >= len(data) { // TODO 这里没看懂为什么要这样才退出
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
