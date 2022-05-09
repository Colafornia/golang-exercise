package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const ssePrefix = "http://www.sse.com.cn/"
const (
	ssePageLink = "http://www.sse.com.cn/home/webupdate/"
)

type article struct {
	Title string
	Time  string
	Link  string
}

type updatedArticles struct {
	Name     string
	Time     string
	Articles []article
}

var tpl *template.Template

func init() {
	time.LoadLocation("Asia/Shanghai")
	t := time.Now()
	fmt.Println("Location:", t.Location(), ":Time:", t)
	tpl = template.Must(template.ParseFiles("./sse-crawler/template.gohtml"))
}

func sendRequest(info updatedArticles) {
	type RequestBody struct {
		Summary     string   `json:"summary"`
		ContentType int      `json:"contentType"`
		Content     string   `json:"content"`
		AppToken    string   `json:"appToken"`
		TopicIds    []string `json:"topicIds"`
	}
	var tplBytes bytes.Buffer
	tpl.Execute(&tplBytes, info)
	html := tplBytes.String()
	fmt.Println(html)
	if html == "" {
		return
	}
	contentBody := RequestBody{
		ContentType: 2,
		Content:     html,
		Summary:     "有更新：" + info.Articles[0].Title + "...",
		AppToken:    os.Getenv("APP_TOKEN"),
		TopicIds:    []string{"5747"},
		// 测试频道    5749
		// 正式频道    5747
	}
	content, _ := json.Marshal(contentBody)
	resp, err := http.Post("http://wxpusher.zjiecode.com/api/send/message", "application/json", bytes.NewBuffer(content))
	if err != nil {
		fmt.Println("err:", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func crawlrateInfo() updatedArticles {
	c := colly.NewCollector()
	var info updatedArticles
	yesterday := time.Now().Add(-36 * time.Hour)
	fmt.Println(yesterday)
	c.OnHTML("body", func(e *colly.HTMLElement) {
		var articles = make([]article, 0, 10)
		e.DOM.Find("#sse_list_1 dd").Each(func(i int, selection *goquery.Selection) {
			selection.Find("span").Text()
			link, ok := selection.Find("a").Attr("href")
			var _article article
			updateDate, err := time.Parse("2006-01-02", selection.Find("span").Text())
			if err != nil {
				return
			}
			isAfterTimeLimit := updateDate.After(yesterday)
			if ok && isAfterTimeLimit {
				_article = article{
					Title: selection.Find("a").Text(),
					Time:  selection.Find("span").Text(),
					Link:  ssePrefix + link,
				}
				articles = append(articles, _article)
				fmt.Println(selection.Find("span").Text())
				fmt.Println(selection.Find("a").Text())
				fmt.Println(ssePrefix + link)
			}
		})
		info = updatedArticles{
			Name:     "上海证券交易所",
			Time:     time.Now().Format("2006-01-02"),
			Articles: articles,
		}
		fmt.Println(info)
	})

	err := c.Visit(ssePageLink)

	if err != nil {
		fmt.Println("Visit SSEPageLink Error")
	}
	return info
}

func main() {
	fmt.Println(os.Args)
	info := crawlrateInfo()
	fmt.Println(info)
	if len(info.Articles) > 0 {
		sendRequest(info)
	} else {
		fmt.Println("No update, No notify")
	}
}
