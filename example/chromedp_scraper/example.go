package main

import (
	"fmt"
	"github.com/ShuBo6/goscraper"
	"golang.org/x/net/context"
	"log"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()

	s, err := goscraper.ChromeDPScrape("https://shubo6.github.io/", 5, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Icon :\t %s\n", s.Preview.Icon)
	fmt.Printf("Name :\t %s\n", s.Preview.Name)
	fmt.Printf("Title :\t %s\n", s.Preview.Title)
	fmt.Printf("Description :\t %s\n", s.Preview.Description)
	fmt.Printf("Image:\t %v\n", s.Preview.Images)
	fmt.Printf("CssFiles:\t %v\n", s.Preview.CssFiles)
	fmt.Printf("JsFiles:\t %v\n", s.Preview.JsFiles)
	fmt.Printf("Url :\t %s\n", s.Preview.Link)

	fmt.Printf("Headers :\t %v\n", s.Headers)
	fmt.Printf("Status :\t %v\n", s.Response.Status)
	fmt.Printf("Proto :\t %v\n", s.Response.Proto)

}
