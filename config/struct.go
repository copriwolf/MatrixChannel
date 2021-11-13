package config

import "time"

type Config struct {
	Service *ServiceConfig         `yaml:"Service"`
	User    map[string]*UserConfig `yaml:"User"`
}

type ServiceConfig struct {
	// 刷新间隔
	RefreshInterval time.Duration `yaml:"refreshInterval"`
	// 请求失败的重试次数
	HttpRequestFailSleepTime time.Duration `yaml:"httpRequestFailSleepTime"`
	// 请求失败的睡眠间隔
	HttpRequestAttempts int `yaml:"httpRequestAttempts"`
	// Tapd API 用户ID
	TapdApiUser string `yaml:"tapdApiUser"`
	// Tapd API 密钥
	TapdApiPassword string `yaml:"tapdApiPassword"`
	// Tapd 公司ID
	TapdCompanyID string `yaml:"tapdCompanyID"`

	// Notion 机器人 ID（多人共享机器人才需要）
	NotionBotClientID string `yaml:"notionBotClientID"`
	// Notion 机器人 密钥（多人共享机器人才需要）
	NotionBotClientSecret string `yaml:"notionBotClientSecret"`
	// Notion 机器人 回调地址（多人共享机器人才需要）
	NotionBotRedirectUri string `yaml:"notionBotRedirectUri"`
	// 回调成功推送企业微信机器人地址（多人共享机器人才需要）
	WxBotNotifyUri string `yaml:"wxBotNotifyUri"`
	// Https 服务器证书路径
	TlsCertFilePath string `yaml:"tlsCertFilePath"`
	// Https 服务器密钥路径
	TlsKeyFilePath string `yaml:"tlsKeyFilePath"`
}

type UserConfig struct {
	// 开启同步的数据 ["task"|"story"]
	Enable []string `yaml:"enable"`
	// Tapd 处理人
	TapdOwner string `yaml:"tapdOwner"`
	// Notion 机器人密钥
	NotionBotSecret string `yaml:"notionBotSecret"`
	// NotionDatabaseTask
	NotionDbTaskID string `yaml:"notionDbTaskID"`
	// NotionDatabaseStory
	NotionDbStoryID string `yaml:"notionDbStoryID"`
	// 每次都强制更新
	ForceUpdate bool `yaml:"forceUpdate"`
}
