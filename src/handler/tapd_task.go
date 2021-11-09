package handler

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

type TaskHandler struct {
	ServiceConf *config.ServiceConfig
	UserConf    *config.UserConfig
}

type TaskTapdMap struct {
	UserMap             map[string]*TapdUser
	UserWorkspaceIdList []string
	UserWorkspaceMap    map[string]string
	Iteration           map[string]*TapdIteration
}

func (u *TaskHandler) GetName() string {
	return constant.HandlerTask
}

func (u *TaskHandler) New(service *config.ServiceConfig, user *config.UserConfig) (result *TaskHandler) {
	result = &TaskHandler{
		ServiceConf: service,
		UserConf:    user,
	}
	return
}

func (u *TaskHandler) Do() (err error) {
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Printf("TaskHandler begin. Owner[%s] Page[%s]\n", u.UserConf.TapdOwner, u.UserConf.NotionDbTaskID)
	// initNotionData
	notionData, err := u.initNotion()
	if err != nil {
		return
	}
	// queryTapdMap
	tapdMap, err := u.getTapdMap()
	if err != nil {
		return
	}
	// queryTapdData
	tapdData, err := u.getFromTapd(tapdMap.UserWorkspaceIdList)
	if err != nil {
		return
	}
	// compare => updateOrCreate
	updateData, createData, err := u.compare(notionData, tapdData)
	err = u.do(tapdMap, updateData, createData)
	return
}

// getTapdMap 获取通用的映射数据
func (u TaskHandler) getTapdMap() (result *TaskTapdMap, err error) {
	result = &TaskTapdMap{}
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
	result = &TaskTapdMap{
		UserMap:             tapdUserMap,
		UserWorkspaceIdList: tapdUserWorkspaceIdList,
		UserWorkspaceMap:    tapdUserWorkspaceMap,
		Iteration:           tapdIteration,
	}
	return
}

// compare 比对区别(notion&&tapd)
func (u TaskHandler) compare(notion map[string]*NotionTaskChange, tapd map[string]*TapdTask) (updateData, createData map[string]*NotionTaskChange, err error) {
	updateData, createData = make(map[string]*NotionTaskChange, 0), make(map[string]*NotionTaskChange, 0)
	if len(tapd) == 0 {
		return
	}
	for tapdTaskId, tapdItem := range tapd {
		// update
		if notionItem, ok := notion[tapdTaskId]; ok {
			tapdTime, _ := time.ParseInLocation(constant.TimeFormatDate, tapdItem.Modified, time.Local)
			notionTime, _ := time.ParseInLocation(time.RFC3339, notionItem.NotionData.Properties[constant.PropertyNameTapdLastModifiedTime].Date.Start, time.Local)
			// 相差 60s 认为修改一样
			if tapdTime.Sub(notionTime).Seconds() < 60 && !u.UserConf.ForceUpdate {
				continue
			}
			updateData[tapdTaskId] = notion[tapdTaskId]
			updateData[tapdTaskId].ChangeType = "UPDATE"
			updateData[tapdTaskId].TapdData = tapdItem
			continue
		}
		// create
		createData[tapdTaskId] = &NotionTaskChange{
			ChangeType: "CREATE",
			TapdID:     tapdTaskId,
			TapdData:   tapdItem,
		}
		continue
	}
	return
}

func (u TaskHandler) do(tapdMap *TaskTapdMap, updateData, createData map[string]*NotionTaskChange) (err error) {
	if err = u.doUpdate(tapdMap, updateData); err != nil {
		return
	}
	if err = u.doCreate(tapdMap, createData); err != nil {
		return
	}
	fmt.Printf("[do]完成数据更新[%d]条，完成数据传入[%d]条\n", len(updateData), len(createData))
	return
}

func (u TaskHandler) doUpdate(tapdMap *TaskTapdMap, updateData map[string]*NotionTaskChange) (err error) {
	fmt.Printf("[doUpdate]修改数据，一共需要处理数据[%d]\n", len(updateData))
	cnt := 0
	total := len(updateData)
	if total < 1 {
		return
	}
	for _, item := range updateData {
		_, poetErr := GetNotionWithRetry(
			u.ServiceConf, u.UserConf,
			http.MethodPatch, u.makeNotionUpdatePage(item.NotionID), u.makeNotionUpdatePagePayload(tapdMap, item),
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

func (u TaskHandler) doCreate(tapdMap *TaskTapdMap, createData map[string]*NotionTaskChange) (err error) {
	fmt.Printf("[doCreate]新增数据，一共需要处理数据[%d]\n", len(createData))
	cnt, total := 0, len(createData)
	if total < 1 {
		return
	}
	for _, item := range createData {
		_, poetErr := GetNotionWithRetry(
			u.ServiceConf, u.UserConf, http.MethodPost,
			u.makeNotionCreatePage(), u.makeNotionCreatePagePayload(u.getNotionDbId(u.UserConf), tapdMap, item))
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

// initNotion 初始化notion数据，从notion获取page并生成映射 map[nickName] = pageID
func (u *TaskHandler) initNotion() (result map[string]*NotionTaskChange, err error) {
	fmt.Println("[initNotion]获取notion已有数据")
	// queryNotionData
	notionDataMap := make(map[string]*NotionTaskChange, 0)
	// parseIntoMap
	cursor := ""
	for {
		// queryData
		notionRes, gerErr := GetNotionWithRetry(
			u.ServiceConf, u.UserConf,
			http.MethodPost, u.makeNotionQueryDbUrl(u.UserConf), u.makeNotionQueryDbPayload(cursor),
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
		//fmt.Println(string(notionRes))
		// appendData
		notionDataMap = u.appendNotionMap(notionDataMap, data)
		// judge continue
		cursor = data.NextCursor
		//fmt.Printf("[initNotin]nextCursor[%s]\n", cursor)
		fmt.Printf("[initNotin]从page[%s]解析[%d]条数据\n", cursor, len(data.Results))
		if cursor == "" {
			break
		}
	}
	result = notionDataMap
	fmt.Printf("[initNotion]获取完成，已有数据[%d]\n", len(result))
	return
}

// appendNotionMap 追加数据 map[taskID]=*NotionTaskChange
func (u *TaskHandler) appendNotionMap(notionData map[string]*NotionTaskChange, queryData *NotionDataBaseQueryReply) (result map[string]*NotionTaskChange) {
	result = notionData
	if queryData == nil || len(queryData.Results) < 1 {
		return
	}

	for _, item := range queryData.Results {
		if len(item.Properties[constant.PropertyNameTitle].Title) < 1 {
			continue
		}
		if GetTitle(item.Properties[constant.PropertyNameTitle]) != "" {
			if item.Properties[constant.PropertyNameTapdTaskID] == nil ||
				item.Properties[constant.PropertyNameTapdTaskID].RichText == nil ||
				len(item.Properties[constant.PropertyNameTapdTaskID].RichText) == 0 {
				continue
			}
			taskID := item.Properties[constant.PropertyNameTapdTaskID].RichText[0].PlainText
			notionData[taskID] = item.ToTaskChange("INIT")
		}
	}
	return
}

// makeNotionQueryDbUrl 构建notion查询page的url
func (u *TaskHandler) makeNotionQueryDbUrl(userConf *config.UserConfig) (result string) {
	return fmt.Sprintf(constant.NotionUrlQueryDb, userConf.NotionDbTaskID)
}

// makeNotionQueryDbPayload 构建notion查询page的jsonPayload
func (u *TaskHandler) makeNotionQueryDbPayload(cursor string) (result *bytes.Reader) {
	req := &NotionDataBaseQueryRequest{
		StartCursor: cursor,
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}

// makeNotionUpdatePage
func (u *TaskHandler) makeNotionUpdatePage(pageID string) (result string) {
	return constant.GetNotionUrlUpdatePage(pageID)
}

// makeNotionCreatePagePayload
func (u *TaskHandler) makeNotionUpdatePagePayload(tapdMap *TaskTapdMap, input *NotionTaskChange) (result *bytes.Reader) {
	req := &NotionPageUpdateRequest{
		Properties: input.TapdData.ToNotionPageProperty(tapdMap.UserMap, tapdMap.Iteration, tapdMap.UserWorkspaceMap),
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}

// makeNotionCreatePage
func (u *TaskHandler) makeNotionCreatePage() (result string) {
	return constant.NotionUrlCreatePage
}

// makeNotionCreatePagePayload
func (u *TaskHandler) makeNotionCreatePagePayload(notionDbId string, tapdMap *TaskTapdMap, input *NotionTaskChange) (result *bytes.Reader) {
	req := &NotionPageCreateRequest{
		Parent:     &NotionPageParent{DatabaseId: notionDbId},
		Properties: input.TapdData.ToNotionPageProperty(tapdMap.UserMap, tapdMap.Iteration, tapdMap.UserWorkspaceMap),
	}
	reqStr, _ := json.Marshal(req)
	result = bytes.NewReader(reqStr)
	return
}

// getFromTapd 从Tapd拉取最新数据 map[tapdID]tapdData
func (u *TaskHandler) getFromTapd(userWorkspaceIDList []string) (result map[string]*TapdTask, err error) {
	result = make(map[string]*TapdTask, 0)
	fmt.Println("[getFromTapd]获取tapd最新数据")
	if len(userWorkspaceIDList) == 0 {
		return
	}

	for _, workspaceID := range userWorkspaceIDList {
		page := int32(1)
		for {
			// getWorkS
			getBody, getErr := GetTapdWithRetry(
				u.ServiceConf, u.UserConf,
				u.makeTapdTaskUrl(), u.makeTapdTaskQuery(workspaceID, u.UserConf.TapdOwner, page),
			)
			if getErr != nil {
				err = getErr
				return
			}
			resp, parseErr := u.parseTapdTaskResp(getBody)
			if parseErr != nil {
				err = parseErr
				return
			}
			if (len(resp.Data)) == 0 {
				break
			}
			cnt := len(result)
			result = u.appendTapdMap(result, resp)
			fmt.Printf("[getFromTapd]项目[%s]页数[%d]获取任务数[%d],总数[%d]\n", workspaceID, page, len(result)-cnt, len(result))
			page++
		}
	}
	fmt.Printf("[getFromTapd]获取完成，涉及项目数[%d] 总任务数[%d]\n", len(userWorkspaceIDList), len(result))
	return
}

func (u TaskHandler) appendTapdMap(tapdData map[string]*TapdTask, queryData *TapdTaskReply) (result map[string]*TapdTask) {
	result = tapdData
	fmt.Printf("[getFromTapd][appendTapdTaskMap]待处理数据[%d]\n", len(queryData.Data))
	for _, item := range queryData.Data {
		result[item.Task.Id] = item.Task
	}
	return
}

// makeTapdTaskUrl 构造Tapd请求用户任务数据
func (u *TaskHandler) makeTapdTaskUrl() (result string) {
	return constant.TapdUrlQueryTask
}

// makeTapdTaskQuery 构造Tapd请求参数
func (u *TaskHandler) makeTapdTaskQuery(workspaceID, owner string, page int32) (result map[string]string) {
	result = map[string]string{
		"workspace_id": workspaceID,
		"owner":        owner,
		"page":         fmt.Sprintf("%d", page),
		"limit":        "30", //max 30
	}
	return
}

// parseTapdTaskResp 解析Tapd响应报文
func (u *TaskHandler) parseTapdTaskResp(respBody []byte) (result *TapdTaskReply, err error) {
	result = &TapdTaskReply{}
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

func (u TaskHandler) getNotionDbId(userConfig *config.UserConfig) string {
	return userConfig.NotionDbTaskID
}
