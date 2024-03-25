package goscraper

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var (
	EscapedFragment string = "_escaped_fragment_="
	fragmentRegexp         = regexp.MustCompile("#!(.*)")
	emptyStruct            = struct{}{}
)

type Scraper struct {
	Url                *url.URL
	EscapedFragmentUrl *url.URL
	MaxRedirect        int
}

type Document struct {
	Body     []byte
	Response *http.Response
	Headers  map[string][]string
	buff     bytes.Buffer
	Preview  DocumentPreview
}

type DocumentPreview struct {
	Icon        string
	Name        string
	Title       string
	Description string
	Images      []string
	JsFiles     []string
	jsFileMap   map[string]*struct{} // 去重用
	CssFiles    []string
	cssFileMap  map[string]*struct{} // 去重用
	Link        string
}

func Scrape(uri string, maxRedirect int) (*Document, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	return (&Scraper{Url: u, MaxRedirect: maxRedirect}).Scrape()
}

func (scraper *Scraper) Scrape() (*Document, error) {
	doc, err := scraper.getDocument()
	if err != nil {
		return nil, err
	}
	err = scraper.parseDocument(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (scraper *Scraper) getUrl() string {
	if scraper.EscapedFragmentUrl != nil {
		return scraper.EscapedFragmentUrl.String()
	}
	return scraper.Url.String()
}

func (scraper *Scraper) toFragmentUrl() error {
	unescapedurl, err := url.QueryUnescape(scraper.Url.String())
	if err != nil {
		return err
	}
	matches := fragmentRegexp.FindStringSubmatch(unescapedurl)
	if len(matches) > 1 {
		escapedFragment := EscapedFragment
		for _, r := range matches[1] {
			b := byte(r)
			if avoidByte(b) {
				continue
			}
			if escapeByte(b) {
				escapedFragment += url.QueryEscape(string(r))
			} else {
				escapedFragment += string(r)
			}
		}

		p := "?"
		if len(scraper.Url.Query()) > 0 {
			p = "&"
		}
		fragmentUrl, err := url.Parse(strings.Replace(unescapedurl, matches[0], p+escapedFragment, 1))
		if err != nil {
			return err
		}
		scraper.EscapedFragmentUrl = fragmentUrl
	} else {
		p := "?"
		if len(scraper.Url.Query()) > 0 {
			p = "&"
		}
		fragmentUrl, err := url.Parse(unescapedurl + p + EscapedFragment)
		if err != nil {
			return err
		}
		scraper.EscapedFragmentUrl = fragmentUrl
	}
	return nil
}

func (scraper *Scraper) getDocument() (*Document, error) {
	scraper.MaxRedirect -= 1
	if strings.Contains(scraper.Url.String(), "#!") {
		scraper.toFragmentUrl()
	}
	if strings.Contains(scraper.Url.String(), EscapedFragment) {
		scraper.EscapedFragmentUrl = scraper.Url
	}

	req, err := http.NewRequest("GET", scraper.getUrl(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "GoScraper")
	var httpclient = http.DefaultClient
	if scraper.Url.Scheme == "https" {
		httpclient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	}

	resp, err := httpclient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if resp.Request.URL.String() != scraper.getUrl() {
		scraper.EscapedFragmentUrl = nil
		scraper.Url = resp.Request.URL
	}
	b, err := convertUTF8(resp.Body, resp.Header.Get("content-type"))
	if err != nil {
		return nil, err
	}
	doc := &Document{
		buff:     b,
		Headers:  resp.Header,
		Body:     b.Bytes(),
		Response: resp,
		Preview: DocumentPreview{
			jsFileMap:  make(map[string]*struct{}),
			cssFileMap: make(map[string]*struct{}),
			Link:       scraper.Url.String(),
		}}

	return doc, nil
}

func convertUTF8(content io.Reader, contentType string) (bytes.Buffer, error) {
	buff := bytes.Buffer{}
	content, err := charset.NewReader(content, contentType)
	if err != nil {
		return buff, err
	}
	_, err = io.Copy(&buff, content)
	if err != nil {
		return buff, err
	}
	return buff, nil
}

type linkHtmlNode struct {
	Rel  string
	Href string
}
type scriptHtmlNode struct {
	Type string
	Src  string
}

func (scraper *Scraper) parseDocument(doc *Document) error {
	t := html.NewTokenizer(&doc.buff)
	var ogImage bool
	var headPassed bool
	var hasFragment bool
	var hasCanonical bool
	var hasHttpEquivRefresh bool
	var httpEquivRefreshUrl string

	var canonicalUrl *url.URL
	doc.Preview.Images = []string{}
	// saves previews' link in case that <link rel="canonical"> is found after <meta property="og:url">
	link := doc.Preview.Link
	// set default value to site name if <meta property="og:site_name"> not found
	doc.Preview.Name = scraper.Url.Host
	// set default icon to web root if <link rel="icon" href="/favicon.ico"> not found
	doc.Preview.Icon = fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, "/favicon.ico")
	for {
		tokenType := t.Next()
		if tokenType == html.ErrorToken {
			return nil
		}
		if tokenType != html.SelfClosingTagToken && tokenType != html.StartTagToken && tokenType != html.EndTagToken {
			continue
		}
		token := t.Token()

		switch token.Data {
		case "head":
			if tokenType == html.EndTagToken {
				headPassed = true
			}
		case "body":
			headPassed = true

		case "link":
			// key 是 rel的值，value是href的
			var l = linkHtmlNode{}
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) != "rel" && cleanStr(attr.Key) != "href" { // 其他的不处理
					continue
				}
				if cleanStr(attr.Key) == "href" {
					l.Href = strings.TrimSpace(attr.Val)
				}
				if cleanStr(attr.Key) == "rel" {
					l.Rel = strings.TrimSpace(attr.Val)
				}
			}

			if len(l.Rel) == 0 || len(l.Href) == 0 { // 无效的值直接break此case
				break
			}
			// 处理canonical
			if l.Rel == "canonical" && link != l.Href {
				hasCanonical = true
				var err error
				canonicalUrl, err = url.Parse(l.Href)
				if err != nil {
					return err
				}
			}
			// 处理icon
			if strings.Contains(l.Rel, "icon") {
				doc.Preview.Icon = scraper.convertFullUrl(l.Href)
			}
			// 处理css
			if strings.Contains(l.Rel, "stylesheet") && doc.Preview.cssFileMap[l.Href] == nil {
				doc.Preview.cssFileMap[l.Href] = &emptyStruct
				//url.JoinPath(scraper.getUrl(), l.Href)
				doc.Preview.CssFiles = append(doc.Preview.CssFiles, scraper.convertFullUrl(l.Href))
			}
		case "script": // 仅匹配html中引入的js文件
			var s scriptHtmlNode
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) != "type" && cleanStr(attr.Key) != "src" { // 其他的不处理
					continue
				}
				if cleanStr(attr.Key) == "type" {
					s.Type = strings.TrimSpace(attr.Val)
				}
				if cleanStr(attr.Key) == "src" {
					s.Src = strings.TrimSpace(attr.Val)
				}
			}

			if len(s.Src) == 0 { // 无效的值直接break此case
				break
			}
			if (strings.HasSuffix(s.Type, "javascript") || strings.HasSuffix(s.Src, "js")) && doc.Preview.jsFileMap[s.Src] == nil {
				doc.Preview.jsFileMap[s.Src] = &emptyStruct
				doc.Preview.JsFiles = append(doc.Preview.JsFiles, scraper.convertFullUrl(s.Src))
			}

		case "meta":
			if len(token.Attr) != 2 {
				break
			}
			if metaFragment(token) && scraper.EscapedFragmentUrl == nil {
				hasFragment = true
			}
			var property string
			var content string
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) == "property" || cleanStr(attr.Key) == "name" {
					property = attr.Val
				}
				if cleanStr(attr.Key) == "http-equiv" && strings.ToLower(cleanStr(attr.Val)) == "refresh" {
					hasHttpEquivRefresh = true
				}
				if cleanStr(attr.Key) == "content" {
					content = attr.Val
				}

			}

			if hasHttpEquivRefresh && strings.Contains(strings.ToLower(cleanStr(content)), "url=") {
				for _, s := range strings.Split(strings.ToLower(cleanStr(content)), ";") {
					if strings.Contains(s, "url=") {
						httpEquivRefreshUrl = strings.TrimPrefix(s, "url=")
					}
				}
			}

			switch cleanStr(property) {
			case "og:site_name":
				doc.Preview.Name = content
			case "og:title":
				doc.Preview.Title = content
			case "og:description":
				doc.Preview.Description = content
			case "description":
				if len(doc.Preview.Description) == 0 {
					doc.Preview.Description = content
				}
			case "og:url":
				doc.Preview.Link = content
			case "og:image":
				ogImage = true
				//ogImgUrl, err := url.Parse(content)
				//if err != nil {
				//	return err
				//}
				//if !ogImgUrl.IsAbs() {
				//	ogImgUrl, err = url.Parse(fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, ogImgUrl.Path))
				//	if err != nil {
				//		return err
				//	}
				//}

				doc.Preview.Images = append(doc.Preview.Images, scraper.convertFullUrl(content))

			}

		case "title":
			if tokenType == html.StartTagToken {
				t.Next()
				token = t.Token()
				if len(doc.Preview.Title) == 0 {
					doc.Preview.Title = token.Data
				}
			}

		case "img":
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) == "src" {
					doc.Preview.Images = append(doc.Preview.Images, scraper.convertFullUrl(strings.TrimSpace(attr.Val)))
					//imgUrl, err := url.Parse(attr.Val)
					//if err != nil {
					//	return err
					//}
					//if !imgUrl.IsAbs() {
					//	doc.Preview.Images = append(doc.Preview.Images, fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, imgUrl.Path))
					//} else {
					//	doc.Preview.Images = append(doc.Preview.Images, attr.Val)
					//}

				}
			}
		}
		/*
		   处理这种情况:
		       <meta http-equiv="REFRESH" content="0;url=dfshealth.html" />
		*/
		if hasHttpEquivRefresh && headPassed && scraper.MaxRedirect > 0 {
			u, err := url.Parse(scraper.convertFullUrl(httpEquivRefreshUrl))
			if err != nil {
				return err
			}
			scraper.Url = u
			scraper.EscapedFragmentUrl = nil
			fdoc, err := scraper.getDocument()
			if err != nil {
				return err
			}
			*doc = *fdoc
			return scraper.parseDocument(doc)
		}
		if hasCanonical && headPassed && scraper.MaxRedirect > 0 {
			if !canonicalUrl.IsAbs() {
				//absCanonical, err :=      url.Parse(fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, canonicalUrl.Path))
				//if err != nil {
				//	return err
				//}
				canonicalUrl = scraper.Url.JoinPath(canonicalUrl.Path)
			}
			scraper.Url = canonicalUrl
			scraper.EscapedFragmentUrl = nil
			fdoc, err := scraper.getDocument()
			if err != nil {
				return err
			}
			*doc = *fdoc
			return scraper.parseDocument(doc)
		}

		if hasFragment && headPassed && scraper.MaxRedirect > 0 {
			scraper.toFragmentUrl()
			fdoc, err := scraper.getDocument()
			if err != nil {
				return err
			}
			*doc = *fdoc
			return scraper.parseDocument(doc)
		}

		if len(doc.Preview.Title) > 0 && len(doc.Preview.Description) > 0 && ogImage && headPassed {
			return nil
		}

	}

	return nil
}

func avoidByte(b byte) bool {
	i := int(b)
	if i == 127 || (i >= 0 && i <= 31) {
		return true
	}
	return false
}

func escapeByte(b byte) bool {
	i := int(b)
	if i == 32 || i == 35 || i == 37 || i == 38 || i == 43 || (i >= 127 && i <= 255) {
		return true
	}
	return false
}

func metaFragment(token html.Token) bool {
	var name string
	var content string

	for _, attr := range token.Attr {
		if cleanStr(attr.Key) == "name" {
			name = attr.Val
		}
		if cleanStr(attr.Key) == "content" {
			content = attr.Val
		}
	}
	if name == "fragment" && content == "!" {
		return true
	}
	return false
}

func cleanStr(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}

func (scraper *Scraper) convertFullUrl(u string) string {
	urlParse, err := url.Parse(u)
	if err != nil {
		log.Println("[convertFullUrl]", err)
		return ""
	}
	if !urlParse.IsAbs() {
		hostUrl, _ := url.Parse(fmt.Sprintf("%s://%s", scraper.Url.Scheme, scraper.Url.Host))
		return hostUrl.JoinPath(u).String()
	}
	return u
}
