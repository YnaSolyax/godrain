package logparser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	storage "github.com/YnaSolyax/godrain/storage/db"
	"github.com/YnaSolyax/godrain/storage/entity"
	"github.com/jaeyo/go-drain3/pkg/drain3"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type Log struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Cluster   string `json:"cluster"`
	ClusterID int64  `json:"cluster_id"`
}

type fields struct {
	Timestamp int
	lvl       int
	content   int
}

func newFileds(Timestamp int, lvl int, content int) *fields {
	return &fields{
		Timestamp: Timestamp,
		lvl:       lvl,
		content:   content,
	}
}

func GetLogFields(format string) *fields {
	field := newFileds(-1, -1, -1)
	findformat := strings.Fields(format)

	for i, ff := range findformat {
		clean := strings.Trim(ff, "[]")
		switch clean {
		case "Timestamp":
			field.Timestamp = i
		case "Level":
			field.lvl = i
		case "Content":
			field.content = i
		}
	}

	return field
}

func processBatch(d *drain3.Drain, db *storage.DBStorage, incidentID uint, lines []string, field *fields) error {

	for _, l := range lines {
		if len(strings.TrimSpace(l)) == 0 {
			continue
		}

		found := strings.Fields(l)
		time, lvl, cont := "", "", ""

		if field.Timestamp != -1 && len(found) > field.Timestamp {
			time = found[field.Timestamp]
		}
		if field.lvl != -1 && len(found) > field.lvl {
			lvl = found[field.lvl]
		}
		if field.content != -1 && len(found) > field.content {
			cont = strings.Join(found[field.content:], " ")
		} else {
			cont = l
		}

		cluster, _, err := d.AddLogMessage(cont)
		if err != nil {
			return err
		}
		template := cluster.GetTemplate()

		vec := pgvector.NewVector(make([]float32, 384))

		defectID, foundInDB, err := db.FindDefectByVector(vec, 0.8, &zap.Logger{})
		if err != nil {
			fmt.Printf("Ошибка поиска дефекта: %v\n", err)
			continue
		}

		if !foundInDB {
			defectID, err = db.CreateDefect(template, vec, &zap.Logger{})
			if err != nil {
				fmt.Printf("Ошибка создания дефекта: %v\n", err)
				continue
			}
			fmt.Printf(" [DB] New defect detected: %s\n", template)
		}

		err = db.SaveLogItem(entity.LogItem{
			IncidentID: incidentID,
			DefectID:   defectID,
			Timestamp:  time,
			Level:      lvl,
			Content:    template,
			ClusterID:  cluster.ClusterId,
		})
		if err != nil {
			fmt.Printf("Ошибка сохранения лога в БД: %v\n", err)
		}
	}
	return nil
}

func (p *Log) ParseLog(db *storage.DBStorage, filename string, format string) error {

	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()

	var incidentID uint = 1

	d, err := drain3.NewDrain(
		drain3.WithDepth(4),
		drain3.WithSimTh(0.4),
		drain3.WithMaxChildren(100),
	)
	if err != nil {
		return err
	}

	indexes := GetLogFields(format)

	logDir := "logs"
	fullPath := filepath.Join(logDir, filename)
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	batchSize := 1000
	batch := make([]string, 0, batchSize)

	fmt.Printf("Начинаю анализ файла: %s\n", fullPath)

	for scanner.Scan() {
		line := scanner.Text()
		batch = append(batch, line)

		if len(batch) >= batchSize {
			processBatch(d, db, incidentID, batch, indexes)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		processBatch(d, db, incidentID, batch, indexes)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Ошибка чтения файла", zap.Error(err))
		return err
	}

	return nil
}
