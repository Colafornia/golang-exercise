package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/tidwall/gjson"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const ssePrefix = "http://www.sse.com.cn"
const nerisPrefix = "https://neris.csrc.gov.cn/falvfagui/rdqsHeader/mainbody?navbarId=1&secFutrsLawId="
const cbircPrefix = "http://www.cbirc.gov.cn/cn/view/pages/ItemDetail.html?itemId=928&generaltype=0&docId="
const (
	ssePageLink   = "http://www.sse.com.cn/home/webupdate/"
	nerisPageLink = "https://neris.csrc.gov.cn/falvfagui/"
	nerisUrl      = "https://neris.csrc.gov.cn/falvfagui/rdqsHeader/informationController"
	sezePageLink  = "http://www.szse.cn/lawrules/rule/new/"
	szseUrl       = "http://www.szse.cn/api/search/content"
	cbircPageLink = "http://www.cbirc.gov.cn/cn/view/pages/ItemList.html?itemPId=923&itemId=928&itemUrl=ItemListRightList.html&itemName=%E8%A7%84%E7%AB%A0%E5%8F%8A%E8%A7%84%E8%8C%83%E6%80%A7%E6%96%87%E4%BB%B6&itemsubPId=926"
	cbircUrl      = "http://www.cbirc.gov.cn/cn/static/data/DocInfo/SelectDocByItemIdAndChild/data_itemId=928,pageIndex=1,pageSize=18.json"
)

type article struct {
	Title string
	Time  string
	Link  string
}

type updatedArticles struct {
	Origin   string
	Name     string
	Time     string
	Articles []article
}

var tpl *template.Template

var yesterday = time.Now().AddDate(0, 0, -1)

func init() {
	time.LoadLocation("Asia/Shanghai")
	t := time.Now()
	fmt.Println("Location:", t.Location(), ":Time:", t)
	tpl = template.Must(template.ParseFiles("./sse-crawler/template.gohtml"))
}

func sendRequest(info []updatedArticles) {
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
		Summary:     "有更新：" + info[0].Articles[0].Title + "...",
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
			}
		})
		info = updatedArticles{
			Origin:   ssePageLink,
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

func requestNerisInfo() updatedArticles {
	resp, err := http.Get(nerisUrl)
	if err != nil {
		fmt.Println("Request Neris API Error")
		fmt.Println(err)
		return updatedArticles{}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var articles = make([]article, 0, 10)
	gjson.Get(string(body), "pageUtil.pageList").ForEach(func(key, value gjson.Result) bool {
		updateDate, _ := time.Parse("20060102", value.Get("secFutrsLawVersion").String())
		isAfterTimeLimit := updateDate.After(yesterday)
		if isAfterTimeLimit {
			link := nerisPrefix + value.Get("secFutrsLawId").String()
			_article := article{
				Title: value.Get("secFutrsLawName").String(),
				Time:  updateDate.Format("2006-01-02"),
				Link:  link,
			}
			articles = append(articles, _article)
			return true
		}
		return false
	})
	fmt.Println(articles)
	info := updatedArticles{
		Origin:   nerisPageLink,
		Name:     "证监会",
		Time:     time.Now().Format("2006-01-02"),
		Articles: articles,
	}
	fmt.Println(info)
	return info
}

func requesSzseInfo() updatedArticles {
	random := rand.Float32()
	s := fmt.Sprintf("%f", random)
	APIURL := szseUrl + "?random=" + s
	fmt.Println(APIURL)
	v := url.Values{}
	v.Set("keyword", "")
	v.Set("range", "title")
	v.Set("currentPage", "1")
	v.Set("pageSize", "20")
	v.Set("scope", "0")
	v.Set("channelCode[]", "szserulesAllRulesBuss")
	resp, err := http.PostForm(APIURL, v)
	if err != nil {
		fmt.Println("Request Neris API Error: ")
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var articles = make([]article, 0, 10)
	gjson.Get(string(body), "data").ForEach(func(key, value gjson.Result) bool {
		msInt, _ := strconv.ParseInt(value.Get("docpubtime").String(), 10, 64)
		updateDate := time.UnixMilli(msInt)
		isAfterTimeLimit := updateDate.After(yesterday)
		if isAfterTimeLimit {
			_article := article{
				Title: value.Get("doctitle").String(),
				Time:  updateDate.Format("2006-01-02"),
				Link:  value.Get("docpuburl").String(),
			}
			articles = append(articles, _article)
			return true
		}
		return false
	})
	info := updatedArticles{
		Origin:   sezePageLink,
		Name:     "深圳证券交易所",
		Time:     time.Now().Format("2006-01-02"),
		Articles: articles,
	}
	fmt.Println(info)
	return info
}

func requestCbircInfo() updatedArticles {
	resp, err := http.Get(cbircUrl)
	if err != nil {
		fmt.Println("Request Cbirc API Error")
		fmt.Println(err)
		return updatedArticles{}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var articles = make([]article, 0, 10)
	gjson.Get(string(body), "data.rows").ForEach(func(key, value gjson.Result) bool {
		fmt.Println(value.Get("publishDate").String())
		updateDate, _ := time.Parse("2006-01-02 15:04:05", value.Get("publishDate").String())
		isAfterTimeLimit := updateDate.After(yesterday)
		if isAfterTimeLimit {
			link := cbircPrefix + value.Get("docId").String()
			_article := article{
				Title: value.Get("docTitle").String(),
				Time:  updateDate.Format("2006-01-02"),
				Link:  link,
			}
			articles = append(articles, _article)
			return true
		}
		return false
	})
	fmt.Println(articles)
	info := updatedArticles{
		Origin:   cbircPageLink,
		Name:     "银保监",
		Time:     time.Now().Format("2006-01-02"),
		Articles: articles,
	}
	fmt.Println(info)
	return info
}

func main() {
	fmt.Println(os.Args)
	var wg sync.WaitGroup
	var info []updatedArticles
	var nerisInfo updatedArticles
	var sseInfo updatedArticles
	var seseInfo updatedArticles
	var cbircInfo updatedArticles
	wg.Add(4)
	go func() {
		nerisInfo = requestNerisInfo()
		if len(nerisInfo.Articles) > 0 {
			info = append(info, nerisInfo)
		}
		wg.Done()
	}()
	go func() {
		seseInfo = requesSzseInfo()
		if len(seseInfo.Articles) > 0 {
			info = append(info, seseInfo)
		}
		wg.Done()
	}()
	go func() {
		sseInfo = crawlrateInfo()
		if len(sseInfo.Articles) > 0 {
			info = append(info, sseInfo)
		}
		wg.Done()
	}()
	go func() {
		cbircInfo = requestCbircInfo()
		if len(cbircInfo.Articles) > 0 {
			info = append(info, cbircInfo)
		}
		wg.Done()
	}()
	wg.Wait()
	fmt.Println(info)
	if len(info) > 0 {
		sendRequest(info)
	} else {
		fmt.Println("No update, No notify")
	}
}
