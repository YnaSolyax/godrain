package entity

import (
	"gorm.io/gorm"
)

type Incident struct {
	gorm.Model
	Description string
	Source      string
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

/*
table1 -incident
ID       desc source
...      ..  BGL.log

table2 - defectss
ID_logs        desc solution
logs1 ... 100  ...   "упал сервер"
при том хрнаить лог как json

неизвестная длина для сравнения
пример в базе 100, в поступившем 1000

vector - отдельная сущность как суммарный контент всех логов
на уровне дефекта

переделать алгоритм вектора и использовать библиотеки bgVector
*/
