package chrome_hepler

import (
	"fmt"
	"github.com/ShuBo6/goscraper"
	"log"
	"net/url"
	"testing"
)

func TestMurHash(t *testing.T) {

	// create chrome instance
	//ctx, cancel := chromedp.NewContext(
	//	context.Background(),
	//	// chromedp.WithDebugf(log.Printf),
	//)
	//defer cancel()
	//
	//// create a timeout
	//ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	//defer cancel()
	//
	//// navigate to a page, wait for an element, click
	//var example string
	//err := chromedp.Run(ctx,
	//	chromedp.Navigate(`http://172.16.9.99/images/bg.png`),
	//	chromedp.
	//)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("Go's time.After example:\n%s", example)
	//fingerprint.Judge_Favicohash()
}

func TestScrape(t *testing.T) {
	//s, err := goscraper.Scrape("http://192.168.13.171:9870/dfshealth.html#tab-overview", 5)
	s, err := goscraper.Scrape("https://demo-company.runnergo.cn/", 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("body:\n", s.Body)
	fmt.Printf("Icon :\t %s\n", s.Preview.Icon)
	fmt.Printf("Name :\t %s\n", s.Preview.Name)
	fmt.Printf("Title :\t %s\n", s.Preview.Title)
	fmt.Printf("Description :\t %s\n", s.Preview.Description)
	fmt.Printf("Image:\t %v\n", s.Preview.Images)
	fmt.Printf("CssFiles:\t %v\n", s.Preview.CssFiles)
	fmt.Printf("JsFiles:\t %v\n", s.Preview.JsFiles)
	fmt.Printf("Url :\t %s\n", s.Preview.Link)
}
func TestUrlPath(t *testing.T) {
	urlParse, err := url.Parse(`http://172.16.9.99`)
	if err != nil {
		log.Panic(err)
	}
	urlParse = urlParse.JoinPath(`./sxxxx.js`)

	fmt.Println(urlParse.String())
}
