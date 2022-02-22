package crawlerprovider

import (
	"encoding/json"
	"fmt"
	. "forexpackage/model"
	"tczbgo/config"
	"tczbgo/kafka"
	"tczbgo/logger"
	"tczbgo/system/zbhttp"
	"tczbgo/system/zbos"
	"tczbgo/system/zbtime"

	. "yahoofagent/model"
)

var (
	appSettings            AppSettings
	machineHeartBeatApiUrl string
	sgetPageApiUrl         string
	heartbeatApiUrl        string
	sendStopApiUrl         string
	provider               string
	topic                  string
	heartBeat              string
	version                string
)

func init() {
	provider, _ = zbos.HostName()
	version = config.Version
	topic = kafka.Topic
	config.GetAppSettings(&appSettings)
	sgetPageApiUrl = fmt.Sprintf("%v%v", appSettings.GetPageApiUrl, provider)
	heartbeatApiUrl = fmt.Sprintf("%v%v", appSettings.HeartBeatApiUrl, provider)
	sendStopApiUrl = fmt.Sprintf("%v%v", appSettings.SendStopApiUrl, provider)
	machineHeartBeatApiUrl = fmt.Sprintf(appSettings.MachineHeartBeatApiUrl, provider)
}

func GetIP() string {
	var ip string
	ip, err := zbhttp.GetIP()

	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("getIP fail, Error: %v", err.Error()))
		ip = "220.135.64.15"
	}

	return ip
}

func GetPage() Page {
	var page Page
	_, response, err := zbhttp.Get(sgetPageApiUrl)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("getpage get fail, Error: %v", err.Error()))
		return page
	} else {
		err = json.Unmarshal(response, &page)
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("getpage unmarshal fail, Error: %v", err.Error()))
		}
	}
	return page
}

func SendKafka(forexDataList []ForexData) {
	sendTime := zbtime.UnixTimeNow(zbtime.Duration.Millisecond)
	for i := range forexDataList {
		forexDataList[i].SendTime = sendTime
	}
	forexDataByte, err := json.Marshal(forexDataList)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("sendkafka marshal fail, Error: %v", err.Error()))
	} else {
		err = kafka.SendMessage(topic, forexDataByte)
		if err != nil {
			logger.Log(logger.LogLevel.Error, fmt.Sprintf("sendkafka fail, Error: %v", err.Error()))
		}
	}
}

func SendStop(pageName string) {
	api := fmt.Sprintf("%v/%v", sendStopApiUrl, pageName)

	_, _, err := zbhttp.NewHttp(zbhttp.Method.Patch, api, nil, nil)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("sendstop patch fail, Error: %v", err.Error()))
	}
}

func SendHeartBeat(pageList []string) {
	pageData, err := json.Marshal(pageList)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("heartbeat marshal fail, Error: %v", err.Error()))
		return
	}

	header := map[string][]string{"Content-Type": {"application/json"}}
	_, _, err = zbhttp.NewHttp(zbhttp.Method.Patch, heartbeatApiUrl, pageData, header)

	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("heartbeat fail, Error: %v", err.Error()))
	}
}

func SendMachineHeartBeat(pageList []string) {
	heartBeat = fmt.Sprintf("W:%v,V:%v,H:", len(pageList), version)
	statusData, err := json.Marshal(heartBeat)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("machineheartbeat marshal fail, Error: %v", err.Error()))
		return
	}
	_, _, err = zbhttp.Post(fmt.Sprintf("%v?status=%v", machineHeartBeatApiUrl, heartBeat), statusData)
	if err != nil {
		logger.Log(logger.LogLevel.Error, fmt.Sprintf("machineheartbeat post fail, Error: %v", err.Error()))
	}
}
