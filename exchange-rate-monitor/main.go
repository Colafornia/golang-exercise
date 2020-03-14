package main

import (
	"bytes"
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"os"
	"html/template"
	"strconv"
	"io/ioutil"

	"github.com/gocolly/colly"
)

const (
	rateHost = "http://fx.cmbchina.com/hq/"
)

type rateInfo struct {
	Name         string
	Time         string
	CurrentVal   float64
	Rate         float64
}

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseFiles("./exchange-rate-monitor/template.gohtml"))
}

func sendRequest(info rateInfo) {
	type RequestBody struct {
		ContentType  int      `json:"contentType"`
		Content      string   `json:"content"`
		AppToken     string   `json:"appToken"`
		TopicIds     []string `json:"topicIds"`
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
	    Content: html,
		AppToken: os.Getenv("APP_TOKEN"),
		TopicIds: []string{"103"},
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

func crawlrateInfo() rateInfo {
	c := colly.NewCollector()
	var info rateInfo

	c.OnHTML("body", func(e *colly.HTMLElement) {
		tr := e.DOM.Find("tr:contains(欧元)")
		list := strings.Fields(tr.First().Text())
		fmt.Printf("%+v\n", list)
	    currentVal, _ := strconv.ParseFloat(list[3], 64)
		info = rateInfo{
			Name:         list[0],
			Time:         list[7],
			CurrentVal:   currentVal,
			Rate:         currentVal/100,
		}
		fmt.Println(info)
	})

	err := c.Visit(rateHost)

	if err != nil {
		fmt.Println("Visit RateHost Error")
	}
	return info
}

func main() {
	fmt.Println(os.Args)
	info := crawlrateInfo()
	if os.Args[1] == "push" || info.CurrentVal < 7.66 {
		sendRequest(info)
	} else {
		fmt.Println("Larger than 7.65, not notify")
	}
}
