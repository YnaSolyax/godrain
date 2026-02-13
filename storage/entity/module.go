package entity

import (
	"gorm.io/gorm"
)

type Incident struct {
	gorm.Model
	Description string
	Source      string
	Logs        []LogItem `gorm:"foreignKey:IncidentID"`
}

type Defect struct {
	gorm.Model
	Description string
	Solution    string
	Logs        []LogItem `gorm:"foreignKey:DefectID"`
}

type LogItem struct {
	gorm.Model
	IncidentID uint
	DefectID   *uint
	Timestamp  string
	Level      string
	Content    string
	Vector     []byte
}
