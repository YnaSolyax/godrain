package entity

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type Defect struct {
	ID          uint `gorm:"primaryKey"`
	Description string
	Solution    string
	Vector      pgvector.Vector `gorm:"type:vector(384)"`
	CreatedAt   time.Time
}

type LogItem struct {
	ID        uint  `gorm:"primaryKey"`
	DefectID  *uint `gorm:"index"`
	Content   string
	CreatedAt time.Time
}
