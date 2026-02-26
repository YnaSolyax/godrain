package entity

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type Incident struct {
	ID          uint
	Description string
	Source      string
	CreatedAt   time.Time
}

type Defect struct {
	ID          uint
	Description string
	Solution    string
	Vector      pgvector.Vector `db:"vector"`
	CreatedAt   time.Time
}

type LogItem struct {
	ID         uint
	IncidentID uint
	DefectID   uint
	Timestamp  string
	Level      string
	Content    string
	ClusterID  int64
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
