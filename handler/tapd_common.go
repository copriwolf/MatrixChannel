package handler

import (
	"encoding/json"
	"fmt"
	"matrixChannel/config"
	"matrixChannel/constant"
	. "matrixChannel/proto"
)

//getTapdUserMap 获取用户的 nickName-realName 映射
func getTapdUserMap(serviceConf *config.ServiceConfig, userConf *config.UserConfig) (result map[string]*TapdUser, err error) {
	result = make(map[string]*TapdUser, 0)
	fmt.Println("[getTapdUserMap]获取tapd最新用户映射数据")
	getBody, getErr := GetTapdWithRetry(serviceConf, userConf, makeTapdUserMapUrl(), makeTapdUserMapQuery(serviceConf.TapdCompanyID))
	if getErr != nil {
		err = getErr
		return
	}
	resp, parseErr := parseTapdUserMapResp(getBody)
	if parseErr != nil {
		err = parseErr
		fmt.Printf("[getTapdUserMap]获取失败，错误:%s\n", err.Error())
		return
	}
	if resp == nil || resp.Data == nil {
		return
	}

	for _, item := range resp.Data {
		if item == nil || item.UserWorkspace == nil {
			continue
		}
		result[item.UserWorkspace.User] = item.UserWorkspace
	}

	fmt.Printf("[getTapdUserMap]获取完成，用户数[%d]\n", len(result))
	return
}

// makeTapdUserMapUrl 构造Tapd请求用户姓名数据
func makeTapdUserMapUrl() (result string) {
	return constant.TapdUrlQueryUserMap
}

// makeTapdUserMapQuery 构造Tapd请求参数
func makeTapdUserMapQuery(companyID string) (result map[string]string) {
	result = map[string]string{
		"workspace_id": companyID,
	}
	return
}

// parseTapdUserMapResp 解析Tapd响应报文
func parseTapdUserMapResp(respBody []byte) (result *TapdUserReply, err error) {
	result = &TapdUserReply{}
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

//getTapdIteration 获取迭代数据
func getTapdIteration(serviceConf *config.ServiceConfig, userConf *config.UserConfig, workspaceIDList []string) (result map[string]*TapdIteration, err error) {
	result = make(map[string]*TapdIteration, 0)
	fmt.Println("[getTapdIteration]获取tapd最新迭代数据")

	for _, worksapceID := range workspaceIDList {
		page := int32(1)
		for {
			getBody, getErr := GetTapdWithRetry(serviceConf, userConf, makeTapdIterationUrl(), makeTapdIterationQuery(worksapceID, page))
			if getErr != nil {
				err = getErr
				return
			}
			resp, parseErr := parseTapdIterationResp(getBody)
			if parseErr != nil {
				err = parseErr
				fmt.Printf("[getTapdIteration]获取失败，错误:%s\n", err.Error())
				return
			}
			if resp == nil || resp.Data == nil {
				return
			}

			for _, item := range resp.Data {
				if item == nil || item.Iteration == nil {
					continue
				}
				result[item.Iteration.Id] = item.Iteration
			}
			if (len(resp.Data)) == 0 {
				break
			}
			page++
		}
	}
	fmt.Printf("[getTapdIteration]获取完成，迭代数[%d]\n", len(result))
	return
}

// makeTapdUserMapUrl 构造Tapd请求地址
func makeTapdIterationUrl() (result string) {
	return constant.TapdUrlQueryIteration
}

// makeTapdIterationQuery 构造Tapd请求参数
func makeTapdIterationQuery(workspaceID string, page int32) (result map[string]string) {
	result = map[string]string{
		"workspace_id": workspaceID,
		"page":         fmt.Sprintf("%d", page),
		"limit":        "30", //max 30
	}
	return
}

// parseTapdIterationResp 解析Tapd响应报文
func parseTapdIterationResp(respBody []byte) (result *TapdIterationReply, err error) {
	result = &TapdIterationReply{}
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

// getUserWorkspaceIDList 获取用户涉及的项目ID
func getUserWorkspaceIDList(serviceConf *config.ServiceConfig, userConf *config.UserConfig) (workspaceIDList []string, workspaceMap map[string]string, err error) {
	workspaceIDList = make([]string, 0)
	workspaceMap = make(map[string]string, 0)
	getBody, err := GetTapdWithRetry(serviceConf, userConf, makeTapdUserWorkspaceUrl(), makeTapdUserWorkspaceQuery(serviceConf.TapdCompanyID, userConf.TapdOwner))
	if err != nil {
		return
	}
	resp, err := parseTapdUserWorkspaceResp(getBody)
	if err != nil {
		return
	}
	if len(resp.Data) == 0 {
		return
	}
	for _, item := range resp.Data {
		if item == nil || item.Workspace == nil || item.Workspace.Id == "" {
			continue
		}
		workspaceIDList = append(workspaceIDList, item.Workspace.Id)
		workspaceMap[item.Workspace.Id] = item.Workspace.Name
	}
	fmt.Printf("[getUserJoinWorkspaceList]获取涉及项目数[%d]\n", len(workspaceIDList))
	return
}

// makeTapdUserWorkspaceUrl 构造Tapd请求用户参与的项目数据
func makeTapdUserWorkspaceUrl() (result string) {
	return constant.TapdUrlQueryUserWorkspace
}

// makeTapdUserWorkspaceQuery 构造Tapd请求参数
func makeTapdUserWorkspaceQuery(tapdCompanyID, tapdOwner string) (result map[string]string) {
	result = map[string]string{
		"company_id": tapdCompanyID,
		"nick":       tapdOwner,
	}
	return
}

// parseTapdUserWorkspaceResp 解析Tapd响应报文
func parseTapdUserWorkspaceResp(respBody []byte) (result *TapdUserWorkspaceReply, err error) {
	result = &TapdUserWorkspaceReply{}
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
