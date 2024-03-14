# goscraper

基于 ： https://github.com/badoux/goscraper 改造

增加了
`CssFiles`和`JsFiles`

用于在特定场景下收集web信息。

---

```go
package main

import (
	"fmt"
	"github.com/ShuBo6/goscraper"
	"log"
)

func main() {
	s, err := goscraper.Scrape("https://shubo6.github.io/", 5)
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

	//fmt.Printf("Body :\n %s\n", s.Body)

}

```

```shell
Icon :   https://shubo6.github.io/images/logo.svg
Name :   shubo的博客
Title :  shubo的博客
Description :    
Image:   [https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/1.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/2.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/3.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/4.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/5.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/7.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/6.png https://shubo6.github.io/2023/08/08/homelab%E5%AE%9E%E7%8E%B0%E5%85%8D%E7%AB%AF%E5%8F%A3https%E8%AE%BF%E9%97%AE/cloudflare.png https://shubo6.github.io/2023/08/08/homelab%E5%AE%9E%E7%8E%B0%E5%85%8D%E7%AB%AF%E5%8F%A3https%E8%AE%BF%E9%97%AE/nginx.png https://shubo6.github.io/2023/05/08/%E4%BD%BF%E7%94%A8kubeadm-%E5%9C%A8%E8%99%9A%E6%8B%9F%E6%9C%BA%E4%B8%AD%E5%88%9B%E5%BB%BA%E4%BC%AA%E5%A4%9A%E8%8A%82%E7%82%B9%E9%9B%86%E7%BE%A4/1.jpg]
CssFiles:        [https://shubo6.github.io/css/main.css https://shubo6.github.io/fonts.googleapis.com/css%3Ffamily=Roboto%20Mono:300,300italic,400,400italic,700,700italic&display=swap&subset=latin,latin-ext https://shubo6.github.io/lib/font-awesome/css/all.min.css https://shubo6.github.io/cdn.jsdelivr.net/gh/fancyapps/fancybox@3/dist/jquery.fancybox.min.css]
JsFiles:         [https://shubo6.github.io/lib/anime.min.js https://shubo6.github.io/cdn.jsdelivr.net/npm/jquery@3/dist/jquery.min.js https://shubo6.github.io/cdn.jsdelivr.net/gh/fancyapps/fancybox@3/dist/jquery.fancybox.min.js https://shubo6.github.io/lib/velocity/velocity.min.js https://shubo6.github.io/lib/velocity/velocity.ui.min.js https://shubo6.github.io/js/utils.js https://shubo6.github.io/js/motion.js https://shubo6.github.io/js/schemes/pisces.js https://shubo6.github.io/js/next-boot.js https://shubo6.github.io/js/cursor/fireworks.js]
Url :    https://shubo6.github.io/index.html
Headers :        map[Accept-Ranges:[bytes] Access-Control-Allow-Origin:[*] Age:[0] Cache-Control:[max-age=600] Content-Type:[text/html; charset=utf-8] Date:[Thu, 14 Mar 2024 01:41:05 GMT] Etag:[W/"656d30e7-162d8"] Expires:[Thu, 14 Mar 2024 01:37:04 GMT] Last-Modified:[Mon, 04 Dec 2023 01:52:39 GMT] Permissions-Policy:[interest-cohort=()] Server:[GitHub.com] Strict-Transport-Security:[max-age=31556952] Vary:[Accept-Encoding] Via:[1.1 varnish] X-Cache:[HIT] X-Cache-Hits:[1] X-Fastly-Request-Id:[4f6e33b7ff57256b9c8cc0257d3eec0278be6c98] X-Github-Request-Id:[85AA:396E9:1D9D32:1EF37A:65F25267] X-Proxy-Cache:[MISS] X-Served-By:[cache-nrt-rjtf7700046-NRT] X-Timer:[S1710380465.366303,VS0,VE246]]


```
