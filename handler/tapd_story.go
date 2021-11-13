package handler

import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"matrixChannel/config"
	"matrixChannel/constant"
	. "matrixChannel/proto"
	"net/http"
	"time"
)

type StoryHandler struct {
	ServiceConf *config.ServiceConfig
	UserConf    *config.UserConfig
}

type StoryTapdMap struct {
	UserMap             map[string]*TapdUser
	UserWorkspaceIdList []string
	UserWorkspaceMap    map[string]string
	Iteration           map[string]*TapdIteration
	StoryStatusMap      map[string]StoryMapStatus
}

func (u *StoryHandler) GetName() string {
	return constant.HandlerStory
}

func (u *StoryHandler) New(service *config.ServiceConfig, user *config.UserConfig) (result *StoryHandler) {
	result = &StoryHandler{
		ServiceConf: service,
		UserConf:    user,
	}
	return
}

func (u *StoryHandler) Do() (err error) {
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Printf("StoryHandler begin. Owner[%s] Page[%s]\n", u.UserConf.TapdOwner, u.UserConf.NotionDbStoryID)
	// initNotionData
	notionData, err := u.initNotion()
	if err != nil {
		return
	}
	// queryTapdCommonMap
	tapdMap, err := u.getTapdMap()
	if err != nil {
		return
	}
	// queryTapdData
	tapdData, tapdStatusMap, err := u.getFromTapd(tapdMap.UserWorkspaceIdList)
	if err != nil {
		return
	}
	tapdMap.StoryStatusMap = tapdStatusMap
	// compare => updateOrCreate
	updateData, createData, err := u.compare(notionData, tapdData)
	err = u.do(tapdMap, updateData, createData)
	return
}

// getTapdMap 获取通用的映射数据
func (u StoryHandler) getTapdMap() (result *StoryTapdMap, err error) {
	result = &StoryTapdMap{}
	tapdUserMap, err := getTapdUserMap(u.ServiceConf, u.UserConf)
	if err != nil {
		return
	}
	tapdUserWorkspaceIdList, tapdUserWorkspaceMap, err := getUserWorkspaceIDList(u.ServiceConf, u.UserConf)
	if err != nil {
		return
	}
	tapdIteration, err := getTapdIteration(u.ServiceConf, u.UserConf, tapdUserWorkspaceIdList)
	if err != nil {
		return
	}
	result = &StoryTapdMap{
		UserMap:             tapdUserMap,
		UserWorkspaceIdList: tapdUserWorkspaceIdList,
		UserWorkspaceMap:    tapdUserWorkspaceMap,
		Iteration:           tapdIteration,
	}
	return
}

// compare 比对区别(notion&&tapd)
func (u StoryHandler) compare(notion map[string]*NotionStoryChange, tapd map[string]*TapdStory) (updateData, createData map[string]*NotionStoryChange, err error) {
	updateData, createData = make(map[string]*NotionStoryChange, 0), make(map[string]*NotionStoryChange, 0)
	if len(tapd) == 0 {
		return
	}
	for tapdStoryId, tapdItem := range tapd {
		// update
		if notionItem, ok := notion[tapdStoryId]; ok {
			tapdTime, _ := time.ParseInLocation(constant.TimeFormatDate, tapdItem.Modified, time.Local)
			notionTime, _ := time.ParseInLocation(time.RFC3339, notionItem.LastModifiedTime, time.Local)
			// 相差 60s 认为修改一样
			if tapdTime.Sub(notionTime).Seconds() < 60 && !u.UserConf.ForceUpdate {
				continue
			}
			updateData[tapdStoryId] = notion[tapdStoryId]
			updateData[tapdStoryId].ChangeType = "UPDATE"
			updateData[tapdStoryId].TapdData = tapdItem
			continue
		}
		// create
		createData[tapdStoryId] = &NotionStoryChange{
			ChangeType: "CREATE",
			StoryID:    tapdStoryId,
			TapdData:   tapdItem,
		}
		continue
	}
	return
}

func (u StoryHandler) do(tapdMap *StoryTapdMap, updateData, createData map[string]*NotionStoryChange) (err error) {
	if err = u.doUpdate(tapdMap, updateData); err != nil {
		return
	}
	if err = u.doCreate(tapdMap, createData); err != nil {
		return
	}
	fmt.Printf("[do]完成数据更新[%d]条，完成数据传入[%d]条\n", len(updateData), len(createData))
	return
}

func (u StoryHandler) doUpdate(tapdMap *StoryTapdMap, updateData map[string]*NotionStoryChange) (err error) {
	fmt.Printf("[doUpdate]修改数据，一共需要处理数据[%d]\n", len(updateData))
	cnt := 0
	total := len(updateData)
	if total < 1 {
		return
	}
	for _, item := range updateData {
		_, poetErr := GetNotionWithRetry(
			u.ServiceConf, u.UserConf,
			http.MethodPatch, u.makeNotionUpdatePage(item.NotionID),
			u.makeNotionUpdatePagePayload(tapdMap, item),
		)
		if poetErr != nil {
			err = poetErr
			fmt.Printf("[doUpdate][%d/%d][❌]%s\n", cnt, total, err)
			return
		}
		cnt++
		fmt.Printf("[doUpdate][%d/%d][✅]%s\n", cnt, total, item.TapdData.Name)
	}

	return
}

func (u StoryHandler) doCreate(tapdMap *StoryTapdMap, createData map[string]*NotionStoryChange) (err error) {
	fmt.Printf("[doCreate]新增数据，一共需要处理数据[%d]\n", len(createData))
	cnt, total := 0, len(createData)
	if total < 1 {
		return
	}
	for _, item := range createData {
		_, poetErr := GetNotionWithRetry(
			u.ServiceConf, u.UserConf, http.MethodPost,
			u.makeNotionCreatePage(),
			u.makeNotionCreatePagePayload(u.UserConf.NotionDbStoryID, tapdMap, item),
		)
		cnt++
		if poetErr != nil {
			err = poetErr
			fmt.Printf("[doCreate][%d/%d][❌][%s]%s\n", cnt, total, item.TapdData.Name, err)
			return
		}
		fmt.Printf("[doCreate][%d/%d][✅]%s\n", cnt, total, item.TapdData.Name)
	}

	return
}

// initNotion 初始化notion数据，从notion获取page并生成映射 map[tapdStoryID] = pageID
func (u *StoryHandler) initNotion() (result map[string]*NotionStoryChange, err error) {
	fmt.Println("[initNotion]获取notion已有数据")
	// queryNotionData
	notionDataMap := make(map[string]*NotionStoryChange, 0)
	// parseIntoMap
	var cursor string
	for {
		// queryData
		notionRes, gerErr := GetNotionWithRetry(
			u.ServiceConf, u.UserConf, http.MethodPost,
			u.makeNotionQueryDbUrl(u.UserConf.NotionDbStoryID), u.makeNotionQueryDbPayload(cursor),
		)
		data := &NotionDataBaseQueryReply{}
		if gerErr != nil {
			err = fmt.Errorf("[initNotin]请求失败，err: %s", gerErr)
			return
		}
		jsonErr := json.Unmarshal(notionRes, data)
		if jsonErr != nil {
			err = fmt.Errorf("[initNotin]解析json数据失败，err: %s", jsonErr)
			return
		}
		// output log
		//fmt.Println(notionRes)
		// appendData
		notionDataMap = u.appendNotionMap(notionDataMap, data)
		// judge continue
		cursor = data.NextCursor
		fmt.Printf("[initNotin]从[%s]解析[%d]条解析数据\n", cursor, len(data.Results))
		if cursor == "" {
			break
		}
	}
	result = notionDataMap
	fmt.Printf("[initNotion]获取完成，已有数据[%d]\n", len(result))
	return
}

// appendNotionMap 追加数据 map[storyID]=*NotionStoryChange
func (u *StoryHandler) appendNotionMap(notionData map[string]*NotionStoryChange, queryData *NotionDataBaseQueryReply) (result map[string]*NotionStoryChange) {
	result = notionData
	if queryData == nil || len(queryData.Results) < 1 {
		return
	}

	for _, item := range queryData.Results {
		if len(item.Properties[constant.PropertyNameTitle].Title) < 1 {
			continue
		}
		if GetTitle(item.Properties[constant.PropertyNameTitle]) != "" {
			storyID := item.Properties[constant.PropertyNameTapdStoryID].RichText[0].PlainText
			notionData[storyID] = item.ToStoryChange("INIT")
		}
	}
	return
}

// makeNotionQueryDbUrl 构建notion查询page的url
func (u *StoryHandler) makeNotionQueryDbUrl(notionDbStoryID string) (result string) {
	return fmt.Sprintf(constant.NotionUrlQueryDb, notionDbStoryID)
}

// makeNotionQueryDbPayload 构建notion查询page的jsonPayload
func (u *StoryHandler) makeNotionQueryDbPayload(cursor string) (result *bytes.Reader) {
	req := &NotionDataBaseQueryRequest{
		StartCursor: cursor,
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}

// makeNotionUpdatePage
func (u *StoryHandler) makeNotionUpdatePage(pageID string) (result string) {
	return constant.GetNotionUrlUpdatePage(pageID)
}

// makeNotionCreatePagePayload
func (u *StoryHandler) makeNotionUpdatePagePayload(tapdMap *StoryTapdMap, input *NotionStoryChange) (result *bytes.Reader) {
	req := &NotionPageUpdateRequest{
		Properties: input.TapdData.ToNotionPageProperty(tapdMap.StoryStatusMap, tapdMap.UserMap, tapdMap.Iteration, tapdMap.UserWorkspaceMap),
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}

// makeNotionCreatePage
func (u *StoryHandler) makeNotionCreatePage() (result string) {
	return constant.NotionUrlCreatePage
}

// makeNotionCreatePagePayload
func (u *StoryHandler) makeNotionCreatePagePayload(notionDbId string, tapdMap *StoryTapdMap, input *NotionStoryChange) (result *bytes.Reader) {
	req := &NotionPageCreateRequest{
		Parent:     &NotionPageParent{DatabaseId: notionDbId},
		Properties: input.TapdData.ToNotionPageProperty(tapdMap.StoryStatusMap, tapdMap.UserMap, tapdMap.Iteration, tapdMap.UserWorkspaceMap),
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}

// getFromTapd 从Tapd拉取最新数据 map[storyID]tapdData
func (u *StoryHandler) getFromTapd(userWorkspaceIDList []string) (result map[string]*TapdStory, statusMap map[string]StoryMapStatus, err error) {
	result = make(map[string]*TapdStory, 0)
	statusMap = make(map[string]StoryMapStatus, 0)
	fmt.Println("[getFromTapd]获取tapd最新数据")

	if len(userWorkspaceIDList) == 0 {
		return
	}
	// 获取迭代数据
	for _, workspaceID := range userWorkspaceIDList {
		page := int32(1)
		for {
			// getStatusMap
			mapResp, mapErr := u.getStatusMap(u.ServiceConf, u.UserConf, statusMap, workspaceID)
			if mapErr != nil {
				err = mapErr
				return
			}
			statusMap = mapResp

			// getStory
			getBody, getErr := GetTapdWithRetry(
				u.ServiceConf, u.UserConf,
				u.makeTapdStoryUrl(), u.makeTapdStoryQuery(workspaceID, u.UserConf.TapdOwner, page),
			)
			if getErr != nil {
				err = getErr
				return
			}
			resp, parseErr := u.parseTapdStoryResp(getBody)
			if parseErr != nil {
				err = parseErr
				return
			}
			if (len(resp.Data)) == 0 {
				break
			}
			cnt := len(result)
			result = u.appendTapdTaskMap(result, resp)
			fmt.Printf("[getFromTapd]项目[%s]页数[%d]获取任务数[%d],总数[%d]\n", workspaceID, page, len(result)-cnt, len(result))
			page++
		}
	}
	fmt.Printf("[getFromTapd]获取完成，涉及项目数[%d] 总任务数[%d]\n", len(userWorkspaceIDList), len(result))
	return
}

// getStatusMap 获取需求-自定义状态的映射
func (u StoryHandler) getStatusMap(serviceConf *config.ServiceConfig, userConf *config.UserConfig, input map[string]StoryMapStatus, workspaceID string) (result map[string]StoryMapStatus, err error) {
	result = input
	getBody, getErr := GetTapdWithRetry(serviceConf, userConf, u.makeTapdStoryStatusMapUrl(), u.makeTapdStoryStatusMapQuery(workspaceID))
	if getErr != nil {
		err = getErr
		return
	}
	resp, parseErr := u.parseTapdStatusMapResp(getBody)
	if parseErr != nil {
		err = parseErr
		return
	}
	if resp == nil || resp.Data == nil {
		return
	}
	result[workspaceID] = resp.Data
	return
}

func (u StoryHandler) appendTapdTaskMap(tapdData map[string]*TapdStory, queryData *TapdStoryReply) (result map[string]*TapdStory) {
	result = tapdData
	fmt.Printf("[getFromTapd][appendTapdTaskMap]待处理数据[%d]\n", len(queryData.Data))
	for _, item := range queryData.Data {
		result[item.Story.Id] = item.Story
	}
	return
}

// makeTapdStoryUrl 构造Tapd请求用户任务数据
func (u *StoryHandler) makeTapdStoryUrl() (result string) {
	return constant.TapdUrlQueryStory
}

// makeTapdStoryStatusMapUrl 构造Tapd请求需求自定义状态的映射
func (u *StoryHandler) makeTapdStoryStatusMapUrl() (result string) {
	return constant.TapdUrlQueryStoryStatusMap
}

// makeTapdStoryQuery 构造Tapd请求参数
func (u *StoryHandler) makeTapdStoryQuery(workspaceID, owner string, page int32) (result map[string]string) {
	result = map[string]string{
		"workspace_id": workspaceID,
		"owner":        owner,
		"page":         fmt.Sprintf("%d", page),
		"limit":        "30", //max 30
	}
	return
}

// makeTapdStoryStatusMapQuery 构造Tapd请求参数
func (u *StoryHandler) makeTapdStoryStatusMapQuery(workspaceID string) (result map[string]string) {
	result = map[string]string{
		"workspace_id": workspaceID,
		"system":       "story",
	}
	return
}

// parseTapdStoryResp 解析Tapd响应报文
func (u *StoryHandler) parseTapdStoryResp(respBody []byte) (result *TapdStoryReply, err error) {
	result = &TapdStoryReply{}
	err = json.Unmarshal(respBody, result)
	if err != nil {
		err = fmt.Errorf("解析请求响应失败, err[%s]", err.Error())
		return
	}

	if result.Status != 1 {
		err = fmt.Errorf("请求响应失败, status[%d] info[%s]", result.Status, result.Info)
		return
	}
	return
}

// parseTapdStatusMapResp 解析Tapd响应报文
func (u *StoryHandler) parseTapdStatusMapResp(respBody []byte) (result *TapdStoryStatusMapReply, err error) {
	result = &TapdStoryStatusMapReply{}
	err = json.Unmarshal(respBody, result)
	if err != nil {
		// 应付骚气 tapd 空返回是个数组的问题
		emptyRes := &TapdStoryStatusMapEmptyReply{}
		if emptyErr := json.Unmarshal(respBody, emptyRes); emptyErr == nil {
			err = nil
			return
		}
		err = fmt.Errorf("解析请求响应失败, err[%s]", err.Error())
		return
	}

	if result.Status != 1 {
		err = fmt.Errorf("请求响应失败, status[%d] info[%s]", result.Status, result.Info)
		return
	}
	return
}
