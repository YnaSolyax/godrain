package logparser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	storage "github.com/YnaSolyax/godrain/storage/db"
	"github.com/YnaSolyax/godrain/storage/entity"
	"github.com/jaeyo/go-drain3/pkg/drain3"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

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

func processBatch(d *drain3.Drain, db *storage.DBStorage, incidentID uint, lines []string, field *fields, logger *zap.Logger) error {

	for _, l := range lines {
		if len(strings.TrimSpace(l)) == 0 {
			logger.Info("log empty")
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
			logger.Error("drain AddLog error", zap.Any("details", err))
			return err
		}
		template := cluster.GetTemplate()

		vec := pgvector.NewVector(make([]float32, 384))

		defectID, foundInDB, err := db.FindDefectByVector(vec, 0.8, logger)
		if err != nil {
			logger.Error("defect not found")
			continue
		}

		if !foundInDB {
			defectID, err = db.CreateDefect(template, vec, logger)
			if err != nil {
				logger.Error("can't create deffect")
				continue
			}
			logger.Info("Success add defect")
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
			logger.Error("can't add log to db")
			return err
		}
	}
	return nil
}

func ParseLog(db *storage.DBStorage, filename string, format string) error {

	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()

	d, err := drain3.NewDrain(
		drain3.WithDepth(4),
		drain3.WithSimTh(0.4),
		drain3.WithMaxChildren(100),
	)
	if err != nil {
		return err
	}

	indexes := GetLogFields(format)

	logDir := "logs" //hardcode
	fullPath := filepath.Join(logDir, filename)
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	incidentID, err := db.CreateIncident(filename)
	if err != nil {
		logger.Error("can't create incident")
		return err
	}
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	batchSize := 1000
	batch := make([]string, 0, batchSize)

	for scanner.Scan() {
		line := scanner.Text()
		batch = append(batch, line)

		if len(batch) >= batchSize {
			processBatch(d, db, incidentID, batch, indexes, logger)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		processBatch(d, db, incidentID, batch, indexes, logger)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("can't read file", zap.Error(err))
		return err
	}

	return nil
}
