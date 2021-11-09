package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"matrixChannel/config"
	matrix_channel_pb "matrixChannel/proto"
	"net/http"
	"time"
)

func GetTapdWithRetry(serviceConf *config.ServiceConfig, userConf *config.UserConfig, url string, queryParams map[string]string) (result []byte, err error) {
	errString := ""
	for idx := 0; idx < serviceConf.HttpRequestAttempts; idx++ {
		result, err = getTapd(serviceConf, userConf, url, queryParams)
		if err == nil {
			return
		}
		if idx != 0 {
			errString += "|" + err.Error()
		} else {
			errString = err.Error()
		}
		time.Sleep(serviceConf.HttpRequestFailSleepTime * time.Duration(2*idx+1))
	}
	err = fmt.Errorf("SendRetry err: %s", errString)
	return
}

func getTapd(serviceConf *config.ServiceConfig, userConf *config.UserConfig, url string, queryParams map[string]string) (result []byte, err error) {
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf("构建 tapd 请求失败, err[%d]", err)
		return
	}
	request.SetBasicAuth(serviceConf.TapdApiUser, serviceConf.TapdApiPassword)
	query := request.URL.Query()
	for key, item := range queryParams {
		query.Add(key, item)
	}
	request.URL.RawQuery = query.Encode()
	resp, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("请求异常，err[%s]", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	//fmt.Println(resp.StatusCode)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("返回code[%d] 不为 200", resp.StatusCode)
		return
	}
	fmt.Printf(".")
	result = body
	return
}

func GetNotionWithRetry(serviceConf *config.ServiceConfig, userConf *config.UserConfig, method, url string, payload io.Reader) (result []byte, err error) {
	errString := ""
	for idx := 0; idx < serviceConf.HttpRequestAttempts; idx++ {
		result, err = getNotion(userConf.NotionBotSecret, method, url, payload)
		if err == nil {
			return
		}
		if idx != 0 {
			errString += "|" + err.Error()
		} else {
			errString = err.Error()
		}
		time.Sleep(serviceConf.HttpRequestFailSleepTime * time.Duration(2*idx+1))
	}
	err = fmt.Errorf("SendRetry err: %s", errString)
	return
}

func getNotion(notionBotSecret, method, url string, payload io.Reader) (result []byte, err error) {
	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		err = fmt.Errorf("构建 notion 请求失败, err[%d]", err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", notionBotSecret))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Notion-Version", "2021-08-16")

	res, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("请求异常, err[%s]", err.Error())
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("读取响应内容异常, err[%s]", err.Error())
		return
	}
	if res.StatusCode != 200 {
		errRes := &matrix_channel_pb.NotionRequestErrReply{}
		_ = json.Unmarshal(body, errRes)
		err = fmt.Errorf("[%d]%s", res.StatusCode, errRes)
		return
	}

	fmt.Printf(".")
	result = body
	return
}
