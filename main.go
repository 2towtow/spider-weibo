package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
)

const (
	USERNAME = "root"
	PASSWORD = "xwt011028"
	HOST     = "localhost"
	PORT     = "3306"
	DBNAME   = "spider"
)

type Spiders struct {
	Name string
	Info string
	Time string
}

var db *gorm.DB

func main() {
	err := InitDB()
	if err != nil {
		log.Fatalln("db connect err:", err)
	}
	spider()
	fmt.Println("已完成")
}

func spider() {
	client := http.Client{}

	var key string
	var StartYear, EndYear int

	fmt.Printf("请输入要搜索的关键字:")
	fmt.Scanf("%s\n", &key)
	fmt.Printf("请输入限定年份，如：2020 2022(中间以空格隔开)")
	fmt.Scanf("%d %d", &StartYear, &EndYear)
	fmt.Println("请稍等")

	for i := 1; i <= 50; i++ {
		//1. 发送请求 i为页数
		url := fmt.Sprintf("https://s.weibo.com/weibo?q=%s&page=%d", key, i)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		//添加请求头，伪造成浏览器访问，防止服务器拦截
		HeaderSet(req, "authority", "s.weibo.com")
		HeaderSet(req, "accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		HeaderSet(req, "accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
		HeaderSet(req, "cookie", "SINAGLOBAL=6060167571263.251.1681134030699; SUB=_2A25JMGZfDeRhGeFO7VAZ9ybFyzmIHXVq2woXrDV8PUJbkNAGLUnbkW1NQU9K0He1-FIwHiC6w_EcihOTlEouBN1m; SUBP=0033WrSXqPxfM725Ws9jqgMF55529P9D9W5jhP.pV7rJ2VHoJzib0EZo5NHD95QNehqE1hMR1K5fWs4DqcjDi--fiKyhiK.Ni--Ri-i8i-z7i--fi-zNi-zEi--4iKn7iKL2S02p1hqt; _s_tentry=weibo.com; Apache=8348877140641.762.1681189696368; ULV=1681189696402:2:2:2:8348877140641.762.1681189696368:1681134030758; PC_TOKEN=f3c14077c7; WBStorage=4d96c54e|undefined")
		HeaderSet(req, "sec-ch-ua", `"Microsoft Edge";v="113", "Chromium";v="113", "Not-A.Brand";v="24"`)
		HeaderSet(req, "sec-ch-ua-mobile", "?0")
		HeaderSet(req, "sec-ch-ua-platform", `"Windows"`)
		HeaderSet(req, "sec-fetch-dest", "document")
		HeaderSet(req, "sec-fetch-mode", "navigate")
		HeaderSet(req, "sec-fetch-site", "none")
		HeaderSet(req, "sec-fetch-user", "?1")
		HeaderSet(req, "upgrade-insecure-requests", "1")
		HeaderSet(req, "user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36 Edg/113.0.0.0")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("response err: ", err)
			return
		}
		defer resp.Body.Close()

		//2. 解析页面
		docDetail, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println("detail err: ", err)
			return
		}

		//3. 获取节点信息
		docDetail.Find("#pl_feedlist_index > div:nth-child(1) > div").
			Each(func(i int, s *goquery.Selection) {
				name := s.Find("div.card > div.card-feed > div.content > div.info > div:nth-child(2) > a.name").Text()
				info := s.Find("div.card > div.card-feed > div.content > p").Text()
				time := s.Find("div > div.card > div.card-feed > div.content > p.from > a:nth-child(1)").Text()
				fmt.Sscanf(info, "                    %s​", &info)
				fmt.Sscanf(time, "\n                        %s\n\n", &time)
				var year int
				if len(time) == 17 {
					fmt.Sscanf(time, "%d年", &year)
				} else {
					year = 2023
				}
				if year >= StartYear && year <= EndYear {
					//4.保存信息
					//fmt.Printf("%s1", name)
					//fmt.Printf("%s1\n", info)
					//fmt.Printf("%s1\n", time)
					spiders := new(Spiders)
					spiders.Time = time
					spiders.Name = name
					spiders.Info = info
					db.Create(&spiders)
				} else if year < StartYear {
					return
				}
			})
	}
}

func HeaderSet(req *http.Request, key string, value string) {
	req.Header.Set(key, value)
}

func InitDB() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", USERNAME, PASSWORD, HOST, PORT, DBNAME)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	return
}
