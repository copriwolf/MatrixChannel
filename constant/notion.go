package constant

import "fmt"

const (
	// 根路径
	NotionUrlBaseUrl = "https://api.notion.com/v1/"
)

var (
	// OAuth 获取 AccessToken
	NotionUrlExchangeAccessToken = fmt.Sprintf("%soauth/token", NotionUrlBaseUrl)
	// 查询 Db 数据
	NotionUrlQueryDb = fmt.Sprintf("%sdatabases/%s", NotionUrlBaseUrl, "%s/query")
	// 创建 Page
	NotionUrlCreatePage = fmt.Sprintf("%spages/", NotionUrlBaseUrl)
)

func GetNotionUrlUpdatePage(pageID string) string {
	return fmt.Sprintf("%spages/%s", NotionUrlBaseUrl, pageID)
}

const (
	PropertyNamePriority             = "优先级"
	PropertyNameCreator              = "创建人"
	PropertyNameTitle                = "Title"
	PropertyNameTapdStoryID          = "需求ID"
	PropertyNameTapdIterationID      = "迭代ID"
	PropertyNameTapdIterationName    = "迭代"
	PropertyNameTapdLastModifiedTime = "最后修改时间"
	PropertyNameTapdTaskID           = "TaskID"
	PropertyNameTapdEffort           = "工时"
	PropertyNameTapdStatus           = "任务状态"
	PropertyNameTapdCreated          = "创建时间"
	PropertyNameTapdCompleted        = "完成时间"
	PropertyNameTapdWorkspaceID      = "项目ID"
	PropertyNameTapdWorkspaceName    = "项目"
	PropertyNameTapdDescription      = "描述"
	PropertyNameTapdName             = "标题"
	PropertyNameTapdOwner            = "处理人"
	PropertyNameTaskUrl              = "TaskUrl"
	PropertyNameStoryUrl             = "StoryUrl"
)

const (
	PropertyTypeSelect         = "select"
	PropertyTypeTitle          = "title"
	PropertyTypeDate           = "date"
	PropertyTypeRichText       = "rich_text"
	PropertyTypeLastEditedTime = "last_edited_time"
	PropertyTypeCreatedTime    = "created_time"
	PropertyTypeMultiSelect    = "multi_select"
	PropertyTypeNumber         = "number"
)
