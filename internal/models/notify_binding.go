package models

import (
	"github.com/engigu/baihu-panel/internal/constant"
)

// NotifyBinding 事件绑定表
type NotifyBinding struct {
	ID        string    `json:"id" gorm:"primaryKey;size:20"`
	Type      string    `json:"type" gorm:"size:20;not null;index"`   // system 或 task
	Event     string    `json:"event" gorm:"size:50;not null;index"`  // 事件类型
	WayID     string    `json:"way_id" gorm:"size:20;not null;index"` // 通知渠道ID
	DataID    string    `json:"data_id" gorm:"size:20;index"`         // 关联ID，系统事件为空，任务事件为任务ID
	Extra     BigText   `json:"extra"`                                // 额外配置（如是否开启日志推送等，对应 BindingExtra 结构）
	CreatedAt LocalTime `json:"created_at"`
	UpdatedAt LocalTime `json:"updated_at"`
}

// CustomVarRule 从日志中提取自定义变量的规则
type CustomVarRule struct {
	Name    string `json:"name"`              // 变量名，在模板中通过 {{name}} 使用
	Regex   string `json:"regex"`             // 正则表达式，捕获组 $1 将作为变量值（若无捕获组则取整体匹配）
	Default string `json:"default,omitempty"` // 匹配失败时的兜底默认值
}

// BindingExtra 存储在 Extra 字段中的 JSON 配置
type BindingExtra struct {
	EnableLog  bool            `json:"enable_log"`
	LogLimit   int             `json:"log_limit"` // 日志字数限制，默认 1000
	CustomVars []CustomVarRule `json:"custom_vars,omitempty"` // 自定义变量提取规则
}

func (NotifyBinding) TableName() string {
	return constant.TablePrefix + "notify_bindings"
}
