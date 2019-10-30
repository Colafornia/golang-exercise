package main

import (
	"errors"
	"fmt"
	"os"
	"bytes"
	"html/template"

	"github.com/gocolly/colly"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/gomail.v2"
)

const (
	fundHost = "http://fund.eastmoney.com/100032.html"
)

type fundInfo struct{
	Name string
	Time string
	CurrentVal string
	Rate string
	MinInHistory string
}

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseFiles("./fund-valuation-monitor/email.gohtml"))
}

func min(values []string) (min string, e error) {
    if len(values) == 0 {
        return "0", errors.New("Cannot detect a minimum value in an empty slice")
    }

    min = values[0]
    for _, v := range values {
            if (v < min) {
                min = v
            }
    }

    return min, nil
}

func sendEmail(info fundInfo)  {
	var tplBytes bytes.Buffer
	tpl.Execute(&tplBytes, info)
	emailHtml := tplBytes.String()
	fmt.Println(emailHtml)
	if emailHtml == "" {
		return
	}
	emailName := os.Getenv("EMAIL_NAME")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	m := gomail.NewMessage()
	m.SetHeader("From", emailName)
	m.SetHeader("To", emailName)
	m.SetHeader("Subject", "基金涨跌监控-今日估值已低于近期历史最低值")
	m.SetBody("text/html", emailHtml)
	d := gomail.NewDialer("smtp.163.com", 465, emailName, emailPassword)
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func crawlFundInfo() fundInfo {
	c := colly.NewCollector()
	var info fundInfo

	c.OnHTML("#body", func(e *colly.HTMLElement) {
		currentVal := e.DOM.Find("#gz_gsz").First().Text()
		rate := e.DOM.Find("#gz_gszzl").First().Text()
		time := e.DOM.Find("#gz_gztime").First().Text()
		name := e.DOM.Find(".fundDetail-tit").First().Text()

		var history []string
		historyNodes := e.DOM.Find("#Li1 table tbody tr")
		historyNodes.Each(func(_ int, selection *goquery.Selection) {
			value := selection.Find("td").Eq(1).Text()
			if value != "" {
				history = append(history, value)
			}
		})

		minVal, _ := min(history)

		info = fundInfo{
			Name: name,
			Time: time,
			CurrentVal: currentVal,
			Rate: rate,
			MinInHistory: minVal,
		}
		fmt.Println(info)
	})

	err := c.Visit(fundHost)

	if err != nil {
		fmt.Println("Visit FundHost Error")
	}
	return info
}

func main() {
	info := crawlFundInfo()

	if info.CurrentVal < info.MinInHistory {
		sendEmail(info)
	} else {
		fmt.Println("Larger than MinInHistory, not notify")
	}
}