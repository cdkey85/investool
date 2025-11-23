// 留言评论模型

package models

import (
	"time"
)

// Comment 留言评论结构体
type Comment struct {
	ID        int64     `json:"id"`          // 主键ID
	StockCode string    `json:"stock_code"`  // 股票代码（可选）
	StockName string    `json:"stock_name"`  // 股票名称（可选）
	Content   string    `json:"content"`     // 留言内容
	CreatedAt time.Time `json:"created_at"`  // 创建时间
	UpdatedAt time.Time `json:"updated_at"`  // 更新时间
}

// TableName 自定义表名
func (Comment) TableName() string {
	return "comments"
}