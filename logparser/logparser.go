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
	Component string `json:"component"`
	Cluster   string `json:"cluster"`
	ClusterID int64  `json:"cluster_id"`
}

func GetLogFields(filename string, format string) {
	logDir := "logs"
	line := ""
	haveFormat := map[string]bool{
		"Timestamp": true,
		"Level":     true,
		"Content":   true,
	}
	indexes := make([]int, 0, 5)

	fullPath := filepath.Join(logDir, filename)

	file, err := os.Open(fullPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	if scanner.Scan() {
		line = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return
	}
	findformat := strings.Fields(format)
	fmt.Println(findformat)
	for i, ff := range findformat {
		cutFF := strings.Trim(ff, "[]")

		if haveFormat[cutFF] {
			indexes = append(indexes, i)
		}
	}

	lenFormat := len(findformat)
	foundIndexes := strings.SplitN(line, " ", lenFormat)
	fmt.Println(foundIndexes)

	for _, ind := range indexes {
		fmt.Println(foundIndexes[ind])
	}
	fmt.Println(line)

}

func readLog(fileName string, size int) ([][]string, error) {

	logDir := "logs"

	fullPath := filepath.Join(logDir, fileName)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	batch := make([]string, 0, size)
	allBatches := make([][]string, 0, 20)
	currentBatch := 0

	for scanner.Scan() {
		line := scanner.Text()
		batch = append(batch, line)

		if len(batch) == size {
			currentBatch++
			allBatches = append(allBatches, batch)
			batch = batch[:0]
		}

	}

	if len(batch) > 0 {
		fmt.Printf("Batch %d: %d lines read\n", currentBatch, len(batch))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return allBatches, nil

}

func parse(d *drain3.Drain, line []string, uniqueCluster map[int64]bool, msg string) ([]*Log, error) {

	newLog := []*Log{}

	for _, l := range line {

		tokens := strings.Fields(l)
		cluster, _, err := d.AddLogMessage(msg)

		if err != nil {
			return nil, err
		}

		if !uniqueCluster[cluster.ClusterId] {
			uniqueCluster[cluster.ClusterId] = true

			newLog = append(newLog, &Log{
				Timestamp: tokens[4],
				Level:     tokens[8],
				Component: strings.Join(tokens[5:8], " "),
				Cluster:   cluster.GetTemplate(),
				ClusterID: cluster.ClusterId,
			})
		}
	}
	return newLog, nil

}

func (p *Log) ParseLog(filename string, msg string) error {

	d, err := drain3.NewDrain(
		drain3.WithDepth(4),
		drain3.WithSimTh(0.1),
		drain3.WithMaxChildren(100),
	)

	if err != nil {
		return err
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()

	uniqueCluster := make(map[int64]bool, 8)
	parsedLogs := []*Log{}

	files, err := readLog(filename, 250000)

	if err != nil {
		return err
	}

	for _, file := range files {
		log, err := parse(d, file, uniqueCluster, msg)
		if err != nil {
			return err
		}
		jsonData, err := json.MarshalIndent(log, "", "  ")
		if err != nil {
			logger.Error("JSON marshal failed", zap.Error(err))
			return err
		}
		fmt.Println(string(jsonData))
	}

	info, _ := os.Stat("BGL.log")
	size := info.Size()
	fmt.Printf("Размер файла %s: %d байт\n", "BGL.log", size)

	jsonData, err := json.MarshalIndent(parsedLogs, "", "  ")
	if err != nil {
		logger.Error("JSON marshal failed", zap.Error(err))
		return err
	}
	fmt.Println(string(jsonData))

	return nil
}
