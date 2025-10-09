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

func (p *ParseLog) Parse() {

	uniqueCluster := make(map[int64]bool, 8)
	parsedLogs := []*ParseLog{}

	logger, err := zap.NewProduction()
	if err != nil {
		return
	}

	defer logger.Sync()

	d, err := drain3.NewDrain(
		drain3.WithDepth(4),
		drain3.WithSimTh(0.1),
		drain3.WithMaxChildren(100),
	)

	if err != nil {
		logger.Error("Newdrain")
		return
	}

	file, err := os.Open("BGL.log")

	if err != nil {
		logger.Error("open file")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for i := 0; i < 1000000 && scanner.Scan(); i++ {

		raw := scanner.Text()
		msg := preprocess(raw)
		tokens := strings.Fields(raw)

		cluster, _, err := d.AddLogMessage(msg)
		if err != nil {
			logger.Error("logMsg")
		}

		if !uniqueCluster[cluster.ClusterId] {
			uniqueCluster[cluster.ClusterId] = true

			parsedLogs = append(parsedLogs, &ParseLog{
				Timestamp: tokens[4],
				Level:     tokens[8],
				Component: strings.Join(tokens[5:8], " "),
				Cluster:   cluster.GetTemplate(),
				ClusterID: cluster.ClusterId,
			})

		}

	}
	jsonData, err := json.MarshalIndent(parsedLogs, "", "  ")
	if err != nil {
		logger.Error("JSON marshal failed", zap.Error(err))
		return
	}
	fmt.Println(string(jsonData))
}
