package goscraper

import (
	"bytes"
	"context"
	chrome_hepler "github.com/ShuBo6/goscraper/chromedp_helper"
	"log"
	"net/url"
	"strings"
	"time"
)

type ScraperChromeDP struct {
	*Scraper
	*chrome_hepler.ChromeHelper
}

func ScrapeChromeDP(uri string, maxRedirect int) (*Document, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	return (&ScraperChromeDP{
		Scraper: &Scraper{
			Url: u, MaxRedirect: maxRedirect,
		},
		ChromeHelper: chrome_hepler.NewChrome(),
	}).Scrape()
}
func (scraper *ScraperChromeDP) Scrape() (*Document, error) {

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
func (scraper *ScraperChromeDP) getDocument() (*Document, error) {

	scraper.MaxRedirect -= 1
	if strings.Contains(scraper.Url.String(), "#!") {
		scraper.toFragmentUrl()
	}
	if strings.Contains(scraper.Url.String(), EscapedFragment) {
		scraper.EscapedFragmentUrl = scraper.Url
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	scraper.ChromeHelper.Run(ctx, scraper.Url)

	if scraper.ChromeHelper.ResponseUrl != scraper.getUrl() {
		scraper.EscapedFragmentUrl = nil
		u, err := url.Parse(scraper.ChromeHelper.ResponseUrl)
		if err != nil {
			log.Println("url.Parse(scraper.ChromeHelper.ResponseUrl) failed", err)
		}
		scraper.Url = u
	}
	//b, err := convertUTF8(resp.Body, resp.Header.Get("content-type"))
	//if err != nil {
	//	return nil, err
	//}
	b := bytes.NewBuffer([]byte(scraper.ChromeHelper.ResponseBody))
	scraper.ChromeHelper.Response.Header = scraper.ChromeHelper.Headers
	doc := &Document{
		Headers:  scraper.ChromeHelper.Headers,
		Body:     b.Bytes(),
		Response: scraper.ChromeHelper.Response,
		Preview: DocumentPreview{
			jsFileMap:  make(map[string]*struct{}),
			cssFileMap: make(map[string]*struct{}),
			Link:       scraper.Url.String(),
		}}

	return doc, nil
}
