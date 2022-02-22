package main

import (
	_ "yahoofagent/config"
	"yahoofagent/domainservice/crawlerservice"
)

func main() {
	crawlerservice.CrawlerYahooCurrencies()
}
