package goscraper

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

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
	Icon        UriFile
	Name        string
	Title       string
	Description string
	Images      []UriFile
	JsFiles     []UriFile
	jsFileMap   map[string]*struct{} // 去重用
	CssFiles    []UriFile
	cssFileMap  map[string]*struct{} // 去重用
	Link        string
}

type UriFile struct {
	Path   string
	Data   *UrlSchemaFile
	Schema string // http , https, data
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
	var httpclient = &http.Client{Timeout: time.Second * 5} // 五秒超时
	if scraper.Url.Scheme == "https" {
		httpclient = &http.Client{Timeout: time.Second * 5, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
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
	doc.Preview.Images = []UriFile{}
	// saves previews' link in case that <link rel="canonical"> is found after <meta property="og:url">
	link := doc.Preview.Link
	// set default value to site name if <meta property="og:site_name"> not found
	doc.Preview.Name = scraper.Url.Host
	// set default icon to web root if <link rel="icon" href="/favicon.ico"> not found
	doc.Preview.Icon = UriFile{
		Path:   fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, "/favicon.ico"),
		Schema: scraper.Url.Scheme,
	}
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
				doc.Preview.Icon = *scraper.convertFullUrl(l.Href)
			}
			// 处理css
			if strings.Contains(l.Rel, "stylesheet") && doc.Preview.cssFileMap[l.Href] == nil {
				doc.Preview.cssFileMap[l.Href] = &emptyStruct
				//url.JoinPath(scraper.getUrl(), l.Href)
				cssFiles := scraper.convertFullUrl(l.Href)
				if cssFiles != nil {
					doc.Preview.CssFiles = append(doc.Preview.CssFiles, *cssFiles)
				}
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
				jsFiles := scraper.convertFullUrl(s.Src)
				if jsFiles != nil {
					doc.Preview.JsFiles = append(doc.Preview.JsFiles, *jsFiles)
				}
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
				img := scraper.convertFullUrl(content)
				if img != nil {
					doc.Preview.Images = append(doc.Preview.Images, *img)
				}

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
					img := scraper.convertFullUrl(strings.TrimSpace(attr.Val))
					if img != nil {
						doc.Preview.Images = append(doc.Preview.Images, *img)
					}
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
			httpEquivRefreshFullUrl := scraper.convertFullUrl(httpEquivRefreshUrl)
			if httpEquivRefreshFullUrl == nil {
				return nil
			}
			u, err := url.Parse(httpEquivRefreshFullUrl.Path)
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

func (scraper *Scraper) convertFullUrl(u string) *UriFile {

	urlParse, err := url.Parse(u)
	if err != nil {
		log.Println("[convertFullUrl]", err)
		return nil
	}
	if urlParse.Scheme == "data" {
		urlSchemaFile, decodeErr := urlSchemaFileDecode([]byte(u))
		if decodeErr != nil {
			log.Println("[convertFullUrl]", decodeErr)
			return nil
		}
		return &UriFile{
			Data:   urlSchemaFile,
			Schema: urlParse.Scheme,
		}
	}
	if !urlParse.IsAbs() {
		hostUrl, _ := url.Parse(fmt.Sprintf("%s://%s", scraper.Url.Scheme, scraper.Url.Host))
		return &UriFile{
			Path:   hostUrl.JoinPath(u).String(),
			Schema: urlParse.Scheme,
		}
	}

	return &UriFile{
		Path:   u,
		Schema: urlParse.Scheme,
	}
}

type UrlSchemaFile struct {
	dataType      string
	fileExt       string
	rawData       []byte
	urlSchemaData []byte
}

/*
fileBase64Decode

data:,文本数据
data:text/plain,文本数据
data:text/html,HTML代码
data:text/html;base64,base64编码的HTML代码
data:text/css,CSS代码
data:text/css;base64,base64编码的CSS代码
data:text/JavaScript,Javascript代码
data:text/javascript;base64,base64编码的Javascript代码
data:image/gif;base64,base64编码的gif图片数据
data:image/png;base64,base64编码的png图片数据
data:image/jpeg;base64,base64编码的jpeg图片数据
data:image/x-icon;base64,base64编码的icon图片数据

示例数据：data:image/png;base64,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
*/
func urlSchemaFileDecode(srcData []byte) (*UrlSchemaFile, error) {
	if !bytes.HasPrefix(srcData, []byte("data:")) {
		return nil, fmt.Errorf("数据格式错误: %s", string(srcData))
	}
	dataSlice := bytes.Split(srcData, []byte(`,`))
	if len(dataSlice) != 2 {
		return nil, fmt.Errorf("数据格式错误: %s", string(srcData))
	}

	var (
		ans = &UrlSchemaFile{
			urlSchemaData: srcData,
		}
		isBase64 bool
		err      error
	)

	for _, item := range bytes.Split(dataSlice[0], []byte(";")) {
		if string(item) == "base64" {
			isBase64 = true
		} else {
			tp := bytes.TrimPrefix(item, []byte("data:"))
			tpslice := bytes.Split(tp, []byte("/"))
			if len(tpslice) != 2 {
				continue
			}
			ans.dataType = string(tpslice[0])
			ans.fileExt = string(tpslice[1])
			if string(tpslice[1]) == "x-icon" {
				ans.fileExt = "icon"
			}
		}
	}

	if isBase64 {
		ans.rawData, err = base64.StdEncoding.DecodeString(string(dataSlice[1]))
		if err != nil {
			return nil, fmt.Errorf("base64解码失败, rawData:%s , err:%v", string(dataSlice[1]), err)
		}
	} else {
		ans.rawData = dataSlice[1]
	}
	return ans, nil

}
