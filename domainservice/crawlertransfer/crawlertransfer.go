package crawlertransfer

import (
	"fmt"
	"regexp"
	"strconv"
	"tczbgo/config"
	"tczbgo/logger"

	"github.com/gocolly/colly"

	. "forexpackage/model"
	. "yahoofagent/model"
)

var (
	appSettings      AppSettings
	getDatumReg      = regexp.MustCompile(`.+\(|=\S+`)
	replaceComma     = regexp.MustCompile(",")
	getErrorDatumReg = regexp.MustCompile(`.+_`)
	getErrorNameReg  = regexp.MustCompile(`_.+`)
	site             string
)

func init() {
	config.GetAppSettings(&appSettings)
	site = appSettings.Site
}

func GetCurrenciesData(doc *colly.HTMLElement, requestTime int64) (ForexData, bool) {
	var forexData ForexData
	var askString string
	var bidString string
	var errDatas []bool = []bool{false, false}
	var errData bool = false
	// var updateTime string
	name := doc.ChildText("font[color=blue]")
	forexData.Name = "USD"
	forexData.Datum = getDatumReg.ReplaceAllString(name, "")
	forexData.Site = site
	forexData.RequestTime = requestTime
	doc.ForEach("tbody > tr[align=center]", func(i int, e *colly.HTMLElement) {
		if i == 1 {
			e.ForEach("td", func(j int, e *colly.HTMLElement) {
				if j == 2 {
					bidString = replaceComma.ReplaceAllString(e.Text, "")
				} else if j == 3 {
					askString = replaceComma.ReplaceAllString(e.Text, "")
				}
			})
		}
	})

	forexData.Ask, errDatas[0] = parseFloat(askString)
	forexData.Bid, errDatas[1] = parseFloat(bidString)
	for _, err := range errDatas {
		if err {
			errData = true
		}
	}
	return forexData, errData
}

func GetErrorData(page string, requestTime int64) ForexData {
	var forexData ForexData
	forexData.Datum = getErrorDatumReg.ReplaceAllString(page, "")
	forexData.Name = getErrorNameReg.ReplaceAllString(page, "")
	forexData.Site = site
	forexData.RequestTime = requestTime
	forexData.Ask = -1
	forexData.Bid = -1
	return forexData
}

func parseFloat(dataString string) (float64, bool) {
	var dataFloat float64
	dataString = replaceComma.ReplaceAllString(dataString, "")
	dataFloat, err := strconv.ParseFloat(dataString, 64)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("parsefloat fail, Error: %v", err.Error()))
		return -1.0, true
	}
	return dataFloat, false
}
