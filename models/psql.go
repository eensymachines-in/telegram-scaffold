package models

import "github.com/lib/pq"

type TelegGrp struct {
	ChatID     int64          `gorm:"chat_id;type:bigint;primary_key"`
	Title      string         `gorm:"title;type:text"`
	Categories pq.StringArray `gorm:"categories;type:text[]"`
}

// TableName : table name override for TelegramGrp
func (tg TelegGrp) TableName() string {
	return "groups"
}
