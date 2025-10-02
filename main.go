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

	if len(tokens) > 9 {
		return strings.Join(tokens[9:], " ")
	}
	return line
}

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d, err := drain3.NewDrain(
		drain3.WithDepth(4),
		drain3.WithSimTh(0.3),
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

	for i := 5000; i < 8000 && scanner.Scan(); i++ {

		raw := scanner.Text()
		msg := preprocess(raw)
		tokens := strings.Fields(raw)

		cluster, _, err := d.AddLogMessage(msg)
		if err != nil {
			logger.Error("logMsg")
		}

		fmt.Printf(
			"Line %d\n", i+1,
		)
		fmt.Printf("  Timestamp:     %s\n", tokens[4:5])
		fmt.Printf("  Level:         %s\n", tokens[8:9])
		fmt.Printf("  Component:     %s\n", tokens[5:8])
		fmt.Printf("  Template:      %s\n", cluster.GetTemplate())
		fmt.Printf("  Tokens:        %v\n", cluster.LogTemplateTokens)

	}

}
