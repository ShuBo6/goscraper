package goscraper

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
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
	fmt.Printf("\nIcon :\t %v\n", s.Preview.Icon)
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
func TestBase64PNG(t *testing.T) {
	base64String := `data:image/png;base64,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`
	// 将data URI头部的信息去除，只保留base64编码后的数据部分
	dataWithoutHeader := strings.TrimPrefix(base64String, "data:image/png;base64,")

	// 使用base64解码得到原始图片数据
	imageData, err := base64.StdEncoding.DecodeString(dataWithoutHeader)
	if err != nil {
		log.Fatal("Base64 decoding error:", err)
	}

	// 将图片数据写入文件
	err = ioutil.WriteFile("output.png", imageData, 0644)
	if err != nil {
		log.Fatal("Error writing image to file:", err)
	}

	fmt.Println("Image successfully decoded and saved as output.png")

}
