package constant

const (
	// Api 根路径
	TapdUrlApiRoot = "https://api.tapd.cn/"
	// 获取公司下或者项目下成员
	TapdUrlQueryUser = TapdUrlApiRoot + "workspaces/users"
	// 获取任务列表(有分页，limit30max)
	TapdUrlQueryTask = TapdUrlApiRoot + "tasks"
	// 获取需求列表(有分页，limit30max)
	TapdUrlQueryStory = TapdUrlApiRoot + "stories"
	// 获取用户参与的项目列表（无分页）
	TapdUrlQueryUserWorkspace = TapdUrlApiRoot + "workspaces/user_participant_projects"
	// 获取项目自定义状态映射（无分页）
	TapdUrlQueryStoryStatusMap = TapdUrlApiRoot + "workflows/status_map"
	// 获取公司用户数据映射（无分页）
	TapdUrlQueryUserMap = TapdUrlApiRoot + "workspaces/users"
	// 获取获取迭代接口by项目ID（有分页，limit30max）
	TapdUrlQueryIteration = TapdUrlApiRoot + "iterations"

	// 浏览跟路径
	TapdUrlViewRoot = "https://www.tapd.cn/"
	// 浏览任务
	TapdUrlViewTask = TapdUrlViewRoot + "%s/prong/tasks/view/%s" // workspaceID / taskID
	// 浏览需求
	TapdUrlViewStory = TapdUrlViewRoot + "%s/prong/stories/view/%s" // workspaceID / storyID
)

// 优先级字段
const (
	TapdPriorityHigh       = "4"
	TapdPriorityMiddle     = "3"
	TapdPriorityLow        = "2"
	TapdPriorityNiceToHave = "1"
)

var (
	TapdPriorityMap = map[string]string{
		TapdPriorityHigh:       "High",
		TapdPriorityMiddle:     "Middle",
		TapdPriorityLow:        "Low",
		TapdPriorityNiceToHave: "NiceToHave",
	}
)

// 状态字段
const (
	// 任务·通用
	TapdStatusOpen        = "open"
	TapdStatusProgressing = "progressing"
	TapdStatusDone        = "done"
	// 需求·枚举\

	TapdStoryStatusPlanning   = "planning"
	TapdStoryStatusDeveloping = "developing"
	TapdStoryStatusResolved   = "resolved"
	TapdStoryStatusRejected   = "rejected"
	TapdStoryStatus2          = "status_2"
	TapdStoryStatus3          = "status_3"
	TapdStoryStatus4          = "status_4"
)

var (
	TapdStatusMap = map[string]string{
		TapdStatusOpen:        "未开始",
		TapdStatusProgressing: "进行中",
		TapdStatusDone:        "已完成",

		TapdStoryStatusPlanning:   "待规划",
		TapdStoryStatusDeveloping: "开发中",
		TapdStoryStatusResolved:   "已实现",
		TapdStoryStatusRejected:   "已拒绝",
		TapdStoryStatus2:          "已规划",
		TapdStoryStatus3:          "开发完成",
		TapdStoryStatus4:          "测试中",
	}
)
