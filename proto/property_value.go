package matrix_channel_pb

import (
	"fmt"
	"matrixChannel/constant"
	"strconv"
	"strings"
	"time"
)

type PropertyValue string
type StoryMapStatus map[string]string

// 生成任务 Url
func GenTaskUrl(workspaceID, taskID string) (result PropertyValue) {
	result = PropertyValue(fmt.Sprintf(constant.TapdUrlViewTask, workspaceID, taskID))
	return
}

// 生成需求 Url
func GenStoryUrl(workspaceID, storyID string) (result PropertyValue) {
	result = PropertyValue(fmt.Sprintf(constant.TapdUrlViewStory, workspaceID, storyID))
	return
}

// 格式化优先级
func (v PropertyValue) FormatPriority() (result PropertyValue) {
	result = v
	if value, ok := constant.TapdPriorityMap[string(v)]; ok {
		result = PropertyValue(value)
	}
	return
}

// 格式化状态
func (v PropertyValue) FormatStatus() (result PropertyValue) {
	result = v
	if value, ok := constant.TapdStatusMap[string(v)]; ok {
		result = PropertyValue(value)
	}
	return
}

// 格式化【需求】的自定义状态
func (v PropertyValue) FormatStoryStatus(statusMap StoryMapStatus) (result PropertyValue) {
	result = v
	if value, ok := statusMap[string(v)]; ok {
		result = PropertyValue(value)
	}
	return
}

// 格式化【项目】名称
func (v PropertyValue) FormatWorkspaceName(workspaceMap map[string]string) (result PropertyValue) {
	result = v
	if value, ok := workspaceMap[string(v)]; ok {
		result = PropertyValue(value)
	}
	return
}

// 格式化姓名包含真实中文名
func (v PropertyValue) FormatUserName(userMap map[string]*TapdUser, isMulti bool, multiSep string) (result PropertyValue) {
	result = v
	formatFunc := func(input string) string {
		if value, ok := userMap[input]; ok {
			return fmt.Sprintf("%s(%s)", value.User, value.Name)
		}
		return input
	}
	if isMulti {
		dataSet := strings.Split(string(v), multiSep)
		for idx, item := range dataSet {
			if item != "" {
				dataSet[idx] = formatFunc(item)
			}
		}
		result = PropertyValue(strings.Join(dataSet, multiSep))
	} else {
		result = PropertyValue(formatFunc(string(v)))
	}
	return
}

// 格式化【需求】的自定义状态
func (v PropertyValue) FormatIteration(iterationMap map[string]*TapdIteration) (result PropertyValue) {
	result = "空"
	if value, ok := iterationMap[string(v)]; ok {
		startDateList := strings.Split(value.Startdate, "-")
		endDateList := strings.Split(value.Enddate, "-")
		result = PropertyValue(fmt.Sprintf("%s(%s-%s)", value.Name, strings.Join(startDateList[1:], ""), strings.Join(endDateList[1:], "")))
	}
	return
}

func (v PropertyValue) InitZero() (result PropertyValue) {
	content := " "
	if string(v) != "" {
		content = string(v)
	}
	result = PropertyValue(content)
	return
}

func (v PropertyValue) ToTitle() (result *NotionPageProperty) {
	return &NotionPageProperty{
		Title: []*NotionTypeTitle{{
			Text: &NotionTypeTextInternal{
				Content: string(v),
			},
		}},
	}
}

func (v PropertyValue) ToRichText() (result *NotionPageProperty) {
	content := string(v)
	if len(content) > 2000 {
		content = content[:2000]
	}
	return &NotionPageProperty{
		RichText: []*NotionTypeRichText{{
			Text: &NotionTypeTextInternal{
				Content: content,
			}},
		},
	}
}

func (v PropertyValue) ToSelect() (result *NotionPageProperty) {
	return &NotionPageProperty{
		Select: &NotionTypeSelect{
			Name: string(v),
		},
	}
}

func (v PropertyValue) ToDate() (result *NotionPageProperty) {
	date, _ := time.ParseInLocation(constant.TimeFormatDate, string(v), time.Local)

	return &NotionPageProperty{
		Date: &NotionTypeDate{
			Start: date.Format(time.RFC3339),
		},
	}
}

func (v PropertyValue) ToNumber() (result *NotionPageProperty) {
	num, err := strconv.Atoi(string(v))
	if err != nil || num == 0 {
		num = -1
	}
	return &NotionPageProperty{
		Number: int64(num),
	}
}

func (v PropertyValue) ToMultiSelect(sep string) (result *NotionPageProperty) {

	selectList := strings.Split(string(v), sep)
	multi := make([]*NotionTypeSelect, 0)
	for _, vv := range selectList {
		if vv == "" {
			continue
		}
		multi = append(multi, &NotionTypeSelect{
			Name: vv,
		})
	}

	return &NotionPageProperty{
		MultiSelect: multi,
	}
}

func (v PropertyValue) ToUrl() (result *NotionPageProperty) {
	return &NotionPageProperty{
		Url: string(v),
	}
}
