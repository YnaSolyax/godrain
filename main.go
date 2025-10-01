package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jaeyo/go-drain3/pkg/drain3"
	"go.uber.org/zap"
)

func preprocess(line string) string {
	tokens := strings.Fields(line)

	if len(tokens) > 5 {
		return strings.Join(tokens[5:], " ")
	}
	return line
}

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d, err := drain3.NewDrain(
		drain3.WithDepth(4),
		drain3.WithSimTh(0.7),
		drain3.WithMaxChildren(100),
	)

	if err != nil {
		logger.Error("Newdrain")
		return
	}

	file, err := os.Open("BGL.log")

	if err != nil {
		logger.Error("Newdrain")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for i := 0; i < 20 && scanner.Scan(); i++ {
		raw := scanner.Text()
		msg := preprocess(raw)

		cluster, clusterType, err := d.AddLogMessage(msg)
		if err != nil {
			logger.Error("logMsg")
		}

		fmt.Printf(
			"Line %d\n", i+1,
		)
		fmt.Printf("  Cluster ID:    %d\n", cluster.ClusterId)
		fmt.Printf("  Update Type:   %d\n", clusterType)
		fmt.Printf("  Template:      %s\n", cluster.GetTemplate())
		fmt.Printf("  string:        %s\n", cluster.String())
		fmt.Printf("  Tokens:        %v\n", cluster.LogTemplateTokens)
		fmt.Printf("  Messages seen: %d\n\n", cluster.Size)

	}

}
