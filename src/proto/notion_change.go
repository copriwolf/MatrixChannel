package matrix_channel_pb

import (
	"matrixChannel/constant"
	"matrixChannel/util"
)

func (x *NotionPage) ToTaskChange(changeType string) (result *NotionTaskChange) {
	result = &NotionTaskChange{
		ChangeType:       changeType,
		TapdID:           GetRichText(x.Properties[constant.PropertyNameTapdTaskID]),
		NotionID:         x.Id,
		LastModifiedTime: GetStartDate(x.Properties[constant.PropertyNameTapdLastModifiedTime]),
		NotionData:       x,
	}
	return
}

func (x *NotionPage) ToStoryChange(changeType string) (result *NotionStoryChange) {
	result = &NotionStoryChange{
		ChangeType:       changeType,
		StoryID:          GetRichText(x.Properties[constant.PropertyNameTapdStoryID]),
		NotionID:         x.Id,
		LastModifiedTime: GetStartDate(x.Properties[constant.PropertyNameTapdLastModifiedTime]),
		NotionData:       x,
	}
	return
}

func GetTitle(property *NotionProperty) (result string) {
	result = ""
	if util.IsZeroOfUnderlyingType(property) {
		return
	}
	result = property.Title[0].PlainText
	return
}

func GetRichText(property *NotionProperty) (result string) {
	result = ""
	if util.IsZeroOfUnderlyingType(property) {
		return
	}
	result = property.RichText[0].PlainText
	return
}

func GetStartDate(property *NotionProperty) (result string) {
	result = ""
	if util.IsZeroOfUnderlyingType(property) {
		return
	}
	result = property.Date.Start
	return
}

func (x *TapdTask) ToNotionPageProperty(userMap map[string]*TapdUser, iterationMap map[string]*TapdIteration, userWorkspaceMap map[string]string) (result map[string]*NotionPageProperty) {
	result = make(map[string]*NotionPageProperty, 0)
	result[constant.PropertyNamePriority] = PropertyValue(x.Priority).InitZero().FormatPriority().ToSelect()
	result[constant.PropertyNameCreator] = PropertyValue(x.Creator).InitZero().FormatUserName(userMap, false, "").ToSelect()
	result[constant.PropertyNameTitle] = PropertyValue(x.Name).InitZero().ToTitle()
	result[constant.PropertyNameTapdName] = PropertyValue(x.Name).InitZero().ToRichText()
	result[constant.PropertyNameTapdTaskID] = PropertyValue(x.Id).InitZero().ToRichText()
	result[constant.PropertyNameTapdStoryID] = PropertyValue(x.StoryId).InitZero().ToRichText()
	result[constant.PropertyNameTapdLastModifiedTime] = PropertyValue(x.Modified).InitZero().ToDate()
	result[constant.PropertyNameTapdCompleted] = PropertyValue(x.Completed).InitZero().ToDate()
	// updateTime
	// Tags
	result[constant.PropertyNameTapdIterationID] = PropertyValue(x.IterationId).InitZero().ToRichText()
	result[constant.PropertyNameTapdIterationName] = PropertyValue(x.IterationId).InitZero().FormatIteration(iterationMap).ToSelect()
	result[constant.PropertyNameTapdEffort] = PropertyValue(x.Effort).InitZero().ToNumber()
	result[constant.PropertyNameTapdStatus] = PropertyValue(x.Status).InitZero().FormatStatus().ToSelect()
	result[constant.PropertyNameTapdCreated] = PropertyValue(x.Created).InitZero().ToDate()
	result[constant.PropertyNameTapdWorkspaceID] = PropertyValue(x.WorkspaceId).InitZero().ToRichText()
	result[constant.PropertyNameTapdWorkspaceName] = PropertyValue(x.WorkspaceId).InitZero().FormatWorkspaceName(userWorkspaceMap).ToSelect()
	result[constant.PropertyNameTapdDescription] = PropertyValue(x.Description).InitZero().ToRichText()
	result[constant.PropertyNameTapdName] = PropertyValue(x.Name).InitZero().ToRichText()
	result[constant.PropertyNameTapdOwner] = PropertyValue(x.Owner).InitZero().FormatUserName(userMap, true, ";").ToMultiSelect(";")

	// 额外生成
	result[constant.PropertyNameTaskUrl] = GenTaskUrl(x.WorkspaceId, x.Id).InitZero().ToUrl()
	result[constant.PropertyNameStoryUrl] = GenStoryUrl(x.WorkspaceId, x.StoryId).InitZero().ToUrl()

	return
}

func (x *TapdStory) ToNotionPageProperty(statusMap map[string]StoryMapStatus, userMap map[string]*TapdUser, iterationMap map[string]*TapdIteration, workspaceMap map[string]string) (result map[string]*NotionPageProperty) {
	result = make(map[string]*NotionPageProperty, 0)
	result[constant.PropertyNamePriority] = PropertyValue(x.Priority).InitZero().FormatPriority().ToSelect()
	result[constant.PropertyNameCreator] = PropertyValue(x.Creator).InitZero().FormatUserName(userMap, false, "").ToSelect()
	result[constant.PropertyNameTitle] = PropertyValue(x.Name).InitZero().ToTitle()
	result[constant.PropertyNameTapdName] = PropertyValue(x.Name).InitZero().ToRichText()
	result[constant.PropertyNameTapdStoryID] = PropertyValue(x.Id).InitZero().ToRichText()
	result[constant.PropertyNameTapdLastModifiedTime] = PropertyValue(x.Modified).InitZero().ToDate()
	result[constant.PropertyNameTapdCompleted] = PropertyValue(x.Completed).InitZero().ToDate()
	// updateTime
	// Tags
	result[constant.PropertyNameTapdIterationID] = PropertyValue(x.IterationId).InitZero().ToRichText()
	result[constant.PropertyNameTapdIterationName] = PropertyValue(x.IterationId).InitZero().FormatIteration(iterationMap).ToSelect()
	result[constant.PropertyNameTapdEffort] = PropertyValue(x.Effort).InitZero().ToNumber()
	result[constant.PropertyNameTapdStatus] = PropertyValue(x.Status).InitZero().FormatStoryStatus(statusMap[x.WorkspaceId]).ToSelect()
	result[constant.PropertyNameTapdCreated] = PropertyValue(x.Created).InitZero().ToDate()
	result[constant.PropertyNameTapdWorkspaceID] = PropertyValue(x.WorkspaceId).InitZero().ToRichText()
	result[constant.PropertyNameTapdWorkspaceName] = PropertyValue(x.WorkspaceId).InitZero().FormatWorkspaceName(workspaceMap).ToSelect()
	result[constant.PropertyNameTapdDescription] = PropertyValue(x.Description).InitZero().ToRichText()
	result[constant.PropertyNameTapdName] = PropertyValue(x.Name).InitZero().ToRichText()
	result[constant.PropertyNameTapdOwner] = PropertyValue(x.Owner).InitZero().FormatUserName(userMap, true, ";").ToMultiSelect(";")

	// 额外生成
	result[constant.PropertyNameStoryUrl] = GenStoryUrl(x.WorkspaceId, x.Id).InitZero().ToUrl()

	return
}
