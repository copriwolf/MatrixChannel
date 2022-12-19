package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"matrixChannel/config"
	"matrixChannel/constant"
	. "matrixChannel/proto"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type NotionCallback struct {
	ServerConf *config.ServiceConfig
}

func HttpSvr(conf *config.ServiceConfig) {
	if conf.NotionBotRedirectUri == "" {
		log.Println("Public bot not configured, http svr sleep.")
		return
	}

	start := time.Now()
	port := "8443"
	srv := http.Server{
		Addr:    ":" + port,
		Handler: http.DefaultServeMux,
	}

	notionCallback := &NotionCallback{conf}
	http.HandleFunc("/oauth", notionCallback.oauthHandler)
	http.HandleFunc("/callback", notionCallback.callbackHandler)

	//gracefully shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		data := <-sigint
		log.Printf("received signal: " + data.String())
		log.Printf("start to shutdown...")

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("HTTP server Shutdown: %v", err)
		}
	}()

	err := srv.ListenAndServeTLS(notionCallback.ServerConf.TlsCertFilePath, notionCallback.ServerConf.TlsKeyFilePath)
	//if err != nil {
	//	log.Fatal("ListenAndServe: ", err)
	//}
	log.Printf("listen on %s", port)
	log.Printf("time elapse: %d ms", time.Now().Sub(start).Milliseconds())

	//serve
	//err = srv.Serve(ln)

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	} else {
		log.Printf("successfully shutdown: %v", err)
	}
}

// oauthHandler 用户 OAuth 校验页
func (n *NotionCallback) oauthHandler(w http.ResponseWriter, r *http.Request) {
	redirectUri := url.Values{}
	redirectUri.Add("owner", "user")
	redirectUri.Add("client_id", n.ServerConf.NotionBotClientID)
	redirectUri.Add("response_type", "code")
	//redirectUri.Add("redirect_url", n.ServerConf.NotionBotRedirectUri)
	//redirectUri.Add("state", "copriwolf")
	fmt.Fprintf(w, fmt.Sprintf("<a href=\"https://api.notion.com/v1/oauth/authorize?%s\">Add to Notion</a>", redirectUri.Encode()))
}

// callbackHandler Notion 回传页
func (n *NotionCallback) callbackHandler(w http.ResponseWriter, r *http.Request) {
	res, err := n.parseNotionCallback(r)
	if err != nil {
		_, _ = fmt.Fprintf(w, fmt.Sprintf("%+v", err))
		return
	}
	err = n.notifyWxBot(res)
	if err != nil {
		_, _ = fmt.Fprintf(w, fmt.Sprintf("%+v", err))
		return
	}
	_, _ = fmt.Fprintf(w, fmt.Sprintf("%s Hello! It's Done. \n originalData: \n %+v", res.Owner.User.Name, res))
}

// notifyWxBot 推送数据到企业微信
func (n NotionCallback) notifyWxBot(callback *NotionExchangeTokenReply) (err error) {
	postStr, _ := json.Marshal(callback)
	wxBotUrl := n.ServerConf.WxBotNotifyUri
	if wxBotUrl == "" {
		return
	}
	wxBot := &WxNotify{
		Msgtype:  "markdown",
		Markdown: &WxNotifyTypeMarkdown{Content: fmt.Sprintf("# MatrixChannel Oauth \n## ![](%s).%s \n > %s", callback.WorkspaceIcon, callback.Owner.User.Name, string(postStr))},
	}
	wxBotPayload, _ := json.Marshal(wxBot)

	fmt.Println(wxBotUrl, string(wxBotPayload))

	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(http.MethodPost, wxBotUrl, bytes.NewReader(wxBotPayload))
	if err != nil {
		err = fmt.Errorf("构建企业微信推送请求失败, err[%d]", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	return
}

// parseNotionCallback 解析 Notion 在 OAuth 授权后返回的用户数据
func (n *NotionCallback) parseNotionCallback(r *http.Request) (result *NotionExchangeTokenReply, err error) {
	getParams := r.URL.Query()
	code, state := getParams.Get("code"), getParams.Get("state")
	if code == "" {
		err = fmt.Errorf("获取回传数据错误，关键数据 code 为空 state[%s]", state)
		return
	}
	result, err = n.exchangingAccessToken(code)
	if err != nil {
		return
	}
	return
}

// exchangingAccessToken 使用 GrantCode 换取用户 Db 的 AccessToken
func (n *NotionCallback) exchangingAccessToken(code string) (result *NotionExchangeTokenReply, err error) {
	result = &NotionExchangeTokenReply{}
	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(http.MethodPost, constant.NotionUrlExchangeAccessToken, n.makeExchangeQuery(code, n.ServerConf.NotionBotRedirectUri))
	if err != nil {
		err = fmt.Errorf("构建 notion 请求失败, err[%d]", err)
		return
	}
	req.SetBasicAuth(n.ServerConf.NotionBotClientID, n.ServerConf.NotionBotClientSecret)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Notion-Version", "2021-08-16")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	if res.StatusCode != 200 {
		errRes := &NotionExchangeErrReply{}
		_ = json.Unmarshal(body, errRes)
		err = fmt.Errorf("[%d]%s", res.StatusCode, errRes)
		return
	}

	fmt.Printf(".")
	fmt.Println(string(body))
	err = json.Unmarshal(body, result)
	if err != nil {
		err = fmt.Errorf("解析返回失败：%s", err)
		return
	}
	return
}

// makeExchangeQuery 构造 exchange 请求
func (n *NotionCallback) makeExchangeQuery(code, redirectUri string) (result io.Reader) {
	req := &NotionExchangeTokenRequest{
		GrantType: "authorization_code",
		Code:      code,
		//RedirectUri: redirectUri,
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}

func (n *NotionCallback) queryPageDetail(accessToken, method, url string, payload io.Reader) (result []byte, err error) {
	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		err = fmt.Errorf("构建 notion 请求失败, err[%d]", err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Notion-Version", "2021-08-16")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	if res.StatusCode != 200 {
		errRes := &NotionRequestErrReply{}
		_ = json.Unmarshal(body, errRes)
		err = fmt.Errorf("[%d]%s", res.StatusCode, errRes)
		return
	}

	fmt.Printf(".")
	result = body
	return
}

// todo
func (n *NotionCallback) getPageDetail(callback *NotionExchangeTokenReply) (result string, err error) {
	queryRes, err := n.queryPageDetail(callback.AccessToken, http.MethodPost, n.makePageQueryUrl(callback.WorkspaceId), n.makePageQueryPayload())
	if err != nil {
		return
	}
	pageDetail := &NotionDataBaseQueryReply{}
	err = json.Unmarshal(queryRes, pageDetail)
	if err != nil {
		return
	}
	return
}

// makePageQueryUrl
func (n *NotionCallback) makePageQueryUrl(workspaceId string) (result string) {
	return fmt.Sprintf(constant.NotionUrlQueryDb, workspaceId)
}

// makePageQueryPayload
func (n *NotionCallback) makePageQueryPayload() (result *bytes.Reader) {
	req := &NotionDataBaseQueryRequest{
		Sorts: []*NotionDataBaseQuerySort{{
			Property:  "createTime",
			Direction: "ascending",
		}},
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}
