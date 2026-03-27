package entity

import "time"

type Permission struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"type:text;not null;unique"`
	Description string    `gorm:"type:text;not null"`
	Category    string    `gorm:"type:text;not null;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (Permission) TableName() string { return "permissions" }
