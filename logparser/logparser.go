package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaeyo/go-drain3/pkg/drain3"
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

func processBatch(d *drain3.Drain, lines []string, uniqueCluster map[int64]bool, field *fields) {

	processedLogs := make([]*Log, 0, len(lines))

	for _, l := range lines {
		if len(strings.TrimSpace(l)) == 0 {
			continue
		}

		found := strings.Fields(l)

		time, lvl, cont := "", "", ""

		if field.Timestamp != -1 {
			time = found[field.Timestamp]
		}

		if field.lvl != -1 {
			lvl = found[field.lvl]
		}

		if field.content != -1 {
			cont = strings.Join(found[field.content:], " ")
		} else {
			cont = l
		}

		cluster, _, err := d.AddLogMessage(cont)
		if err != nil {
			continue
		}

		if !uniqueCluster[cluster.ClusterId] {
			uniqueCluster[cluster.ClusterId] = true

			newLog := &Log{
				Timestamp: time,
				Level:     lvl,
				Cluster:   cluster.GetTemplate(),
				ClusterID: cluster.ClusterId,
			}
			processedLogs = append(processedLogs, newLog)
		}
	}

	if len(processedLogs) > 0 {
		jsonData, err := json.MarshalIndent(processedLogs, "", "  ")
		if err == nil {
			fmt.Println(string(jsonData))
		}
	}
}

func (p *Log) ParseLog(filename string, format string) error {

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
	uniqueCluster := make(map[int64]bool)

	fmt.Printf("Начинаю анализ файла: %s\n", fullPath)

	for scanner.Scan() {
		line := scanner.Text()
		batch = append(batch, line)

		if len(batch) >= batchSize {
			processBatch(d, batch, uniqueCluster, indexes)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		processBatch(d, batch, uniqueCluster, indexes)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Ошибка чтения файла", zap.Error(err))
		return err
	}

	return nil
}
