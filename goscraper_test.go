package goscraper

import (
	"fmt"
	"testing"
)

func TestGoScraper(t *testing.T) {
	s, err := Scrape("http://192.168.13.171:9870/", 5)
	debugPrint(s)
	t.Error(err)
	s, err = Scrape("https://shubo6.github.io/", 5)
	debugPrint(s)
	t.Error(err)
	s, err = Scrape("https://demo-company.runnergo.cn/", 5)
	debugPrint(s)
	t.Error(err)
	s, err = Scrape("https://123.57.241.61/", 5)
	debugPrint(s)
	t.Error(err)

}

func debugPrint(s *Document) {
	if s == nil {
		return
	}
	fmt.Printf("\nIcon :\t %s\n", s.Preview.Icon)
	fmt.Printf("Name :\t %s\n", s.Preview.Name)
	fmt.Printf("Title :\t %s\n", s.Preview.Title)
	fmt.Printf("Description :\t %s\n", s.Preview.Description)
	fmt.Printf("Image:\t %v\n", s.Preview.Images)
	fmt.Printf("CssFiles:\t %v\n", s.Preview.CssFiles)
	fmt.Printf("JsFiles:\t %v\n", s.Preview.JsFiles)
	fmt.Printf("Url :\t %s\n\n", s.Preview.Link)
}
