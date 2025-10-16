package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jaeyo/go-drain3/pkg/drain3"
	"go.uber.org/zap"
)

type ParseLog struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Cluster   string `json:"cluster"`
	ClusterID int64  `json:"cluster_id"`
}

func preprocess(line string) string {
	tokens := strings.Fields(line)

	if len(tokens) > 9 {
		return strings.Join(tokens[9:], " ")
	}
	return line
}

func readLog(fileName string, size int) ([][]string, error) {

	file, err := os.Open(fileName)
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

func parse(line []string, uniqueCluster map[int64]bool) ([]*ParseLog, error) {

	d, err := drain3.NewDrain(
		drain3.WithDepth(4),
		drain3.WithSimTh(0.1),
		drain3.WithMaxChildren(100),
	)

	if err != nil {
		return nil, err
	}

	newLog := []*ParseLog{}

	for _, l := range line {

		msg := preprocess(l)
		tokens := strings.Fields(l)
		cluster, _, err := d.AddLogMessage(msg)

		if err != nil {
			return nil, err
		}

		if !uniqueCluster[cluster.ClusterId] {
			uniqueCluster[cluster.ClusterId] = true

			newLog = append(newLog, &ParseLog{
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

func (p *ParseLog) ParseLog() error {

	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()

	uniqueCluster := make(map[int64]bool, 8)
	parsedLogs := []*ParseLog{}

	files, err := readLog("BGL.log", 250000)

	if err != nil {
		return err
	}

	for _, file := range files {
		log, err := parse(file, uniqueCluster)
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
