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

}


```

```shell
Icon :   https://shubo6.github.io/images/logo.svg
Name :   shubo的博客
Title :  shubo的博客
Description :    
Image:   [https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/1.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/2.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/3.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/4.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/5.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/7.png https://shubo6.github.io/2023/12/04/DELL-PowerEdge-R730-iDRAC-CheatSheet/6.png https://shubo6.github.io/2023/08/08/homelab实现免端口https访问/cloudflare.png https://shubo6.github.io/2023/08/08/homelab实现免端口https访问/nginx.pno6.github.io/2023/05/08/使用kubeadm-在虚拟机中创建伪多节点集群/1.jpg]
CssFiles:        [https://shubo6.github.io/css/main.css https://shubo6.github.io/css https://shubo6.github.io/lib/font-awesome/css/all.min.css https://shubo6.github.io/gh/fancyapps/fancybox@3/dist/jquery.fancybox.min.css]
JsFiles:         [https://shubo6.github.io/lib/anime.min.js https://shubo6.github.io/npm/jquery@3/dist/jquery.min.js https://shubo6.github.io/gh/fancyapps/fancybox@3/dist/jquery.fancybox.min.js https://shubo6.github.io/lib/velocity/velocity.min.js https://shubo6.github.io/lib/velocity/velocity.ui.min.js https://shubo6.github.io/js/utils.js https://shubo6.github.io/js/motion.js https://shubo6.github.io/js/schemes/pisces.js https://shubo6.github.io/js/next-boot.js https://shubo6.github.io/js/cursor/fireworks.js]
Url :    https://shubo6.github.io/index.html

```
