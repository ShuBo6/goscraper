package goscraper

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"
)

func TestGoScraperChromeDP(t *testing.T) {
	s, err := ScrapeChromeDP("https://172.16.9.22:8999", 5)
	debugPrint(s)
	t.Error(err)
	t.Log("-------------")
	s, err = Scrape("https://172.16.9.22:8999", 5)
	debugPrint(s)
	t.Error(err)
}

func TestGoScraper(t *testing.T) {
	s, err := Scrape("http://192.168.13.171:9870/", 5)
	debugPrint(s)
	t.Error(err)
	//s, err = Scrape("https://shubo6.github.io/", 5)
	//debugPrint(s)
	//t.Error(err)
	//s, err = Scrape("https://demo-company.runnergo.cn/", 5)
	//debugPrint(s)
	//t.Error(err)
	//s, err = Scrape("https://123.57.241.61/", 5)
	//debugPrint(s)
	//t.Error(err)
	s, err = Scrape("http://172.16.9.213:9090/", 5)
	debugPrint(s)
	t.Error(err)

}

func TestDebugString(t *testing.T) {
	s, _ := Scrape("http://192.168.13.171:9870/", 5)
	fmt.Println(ToDebugString(s))
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
	fmt.Printf("Response:\t %v\n", s.Response)
	fmt.Printf("Url :\t %s\n\n", s.Preview.Link)
	fmt.Printf("Headers :\t %s\n\n", s.Headers)
}
func ToDebugString(target interface{}) string {
	targetValue := reflect.ValueOf(target)
	targetType := reflect.Indirect(targetValue).Type()

	var str, value, name string
	for i := 0; i < targetType.NumField(); i++ {
		name = targetType.Field(i).Name
		for j, v := range name {
			if j == 0 {
				if unicode.IsLower(v) {
					name = ""
				}
			} else {
				break
			}
		}
		if name != "" {
			value = strings.ReplaceAll(value, "\n", " ")
			str += fmt.Sprintf("\t%s: [%v]\n", name, value)
		}
	}
	return fmt.Sprintf("{\n%s}", str)
}
