package crawlerservice

import (
	"fmt"
	. "forexpackage/model"
	"math/rand"
	"os"
	"tczbgo/config"
	"tczbgo/logger"
	"tczbgo/system/zbtime"
	"time"

	"yahoofagent/domainservice/crawlertransfer"
	"yahoofagent/infrastructure/crawlerprovider"

	. "yahoofagent/model"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

var (
	appSettings AppSettings
	pageList    = []string{}
	maxWork     int
)

func init() {
	config.GetAppSettings(&appSettings)
	maxWork = appSettings.MaxWork
}

func CrawlerYahooCurrencies() {
	now := time.Now().Format("15:04:05")
	logger.Log(logger.LogLevel.Debug, fmt.Sprintf("agent start: %v", now))
	go setTimeCloseExe()
	go heartBeat()

	t := time.NewTicker(1 * time.Minute)
	for {
		<-t.C
		if len(pageList) < maxWork {
			go getPage()
		}
	}
}

func setTimeCloseExe() {
	t := 6 * time.Hour
	min := -30
	max := 30
	rangeTime := rangeRadom(min, max)
	t = t + time.Duration(rangeTime)*time.Minute
	time.AfterFunc(t, closeExe)
}

func rangeRadom(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func closeExe() {
	now := time.Now().Format("15:04:05")
	logger.Log(logger.LogLevel.Debug, fmt.Sprintf("agent close: %v", now))
	closeAllPage()
	time.Sleep(5 * time.Second)
	os.Exit(0)
}

func closeAllPage() {
	for index := range pageList {
		go sendStop(pageList[index])
	}
}

func heartBeat() {
	defer panicRecover("service/heartBeat")
	t := time.NewTicker(1 * time.Minute)
	for {
		<-t.C
		if len(pageList) > 0 {
			crawlerprovider.SendHeartBeat(pageList)
		}
		crawlerprovider.SendMachineHeartBeat(pageList)
	}
}

func checkIPnotZB() bool {
	// ip := crawlerProvider.GetIP()

	// return ip != "220.135.64.15"
	return true
}

func getPage() {
	defer panicRecover("service/getPage")
	var page Page

	notZB := checkIPnotZB()
	if notZB {
		page = crawlerprovider.GetPage()
		if (page != Page{}) {
			fmt.Println(page.PageName)
			pageList = append(pageList, page.PageName)
			crawler(page)
			sendStop(page.PageName)
		}
	} else {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("check IP is ZB"))
		time.Sleep(10 * time.Minute)
	}
}

func crawler(page Page) {
	errorCount := 0
	sixHourTimeout := false
	var forexDataList []ForexData

	c := colly.NewCollector(colly.Async(true)) // 在colly中使用 Collector 這類物件 來做事情

	c.AllowURLRevisit = true //允許對同一 URL 進行多次下載

	extensions.RandomUserAgent(c) //隨機生成 user-agent

	c.DisableCookies() //關閉 cookie 處理

	c.OnError(func(r *colly.Response, err error) {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("colly onError, Error: %v", err.Error()))
		errorCount++
		if errorCount >= 3 {
			errorData := crawlertransfer.GetErrorData(page.PageName, zbtime.UnixTimeNow(zbtime.Duration.Millisecond))
			forexDataList = append(forexDataList, errorData)
		}
	})

	c.OnHTML("table[cellpadding='3']", func(e *colly.HTMLElement) { // 對 html 尋找 div.table-row
		if e.Index == 1 {
			forexData, errData := crawlertransfer.GetCurrenciesData(e, zbtime.UnixTimeNow(zbtime.Duration.Millisecond))
			if !errData {
				errorCount = 0
				forexDataList = append(forexDataList, forexData)
			} else {
				errorCount++
				if errorCount >= 3 {
					errorData := crawlertransfer.GetErrorData(page.PageName, zbtime.UnixTimeNow(zbtime.Duration.Millisecond))
					forexDataList = append(forexDataList, errorData)
				}
			}
		}
		if len(forexDataList) > 0 {
			crawlerprovider.SendKafka(forexDataList)
			fmt.Printf("%v, %v/%v\n", time.Now().Format("15:04:05"), forexDataList[0].Datum, forexDataList[0].Name)
		}
		forexDataList = []ForexData{}
	})

	c.OnRequest(func(r *colly.Request) { // Set Header
	})

	time.AfterFunc(6*time.Hour, func() { sixHourTimeout = true })

	for !sixHourTimeout {
		c.Visit(page.Url)
		c.Wait()

		time.Sleep(1 * time.Minute)
		sixHourTimeout = !checkIPnotZB()
	}
}

func sendStop(pageName string) {
	crawlerprovider.SendStop(pageName)
	for index, page := range pageList {
		if page == pageName {
			pageList = sliceRemove(pageList, index)
			break
		}
	}
}

func sliceRemove(slice []string, index int) []string {
	next := index + 1
	return append(slice[:index], slice[next:]...)
}

func panicRecover(funcName string) {
	if r := recover(); r != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("%v panicRecover, Error: %v", funcName, r))
	}
}
