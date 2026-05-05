package cmd

import (
	"fmt"
	"time"

	"github.com/YnaSolyax/godrain/internal/logparser"
	storage "github.com/YnaSolyax/godrain/storage/db"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var logFormat string

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "parsing a file with logs of a specific format",
	Long: `This command is used for logs with a known format.
	The input is the file name and log format, for example: 
	parse BGL.log -f(or --format) [Label] [Timestamp] [Date] [Node] [Time] [NodeRepeat] [Type] [Component] [Level] [Content]"`,

	Run: func(cmd *cobra.Command, args []string) {

		filename := args[0]

		if logFormat == "" {
			fmt.Println("please write the format of log")
			return
		}

		logger, _ := zap.NewProduction()

		db, err := storage.Conn("host=localhost user=user password=1 dbname=log_analysis port=5432 sslmode=disable", logger)
		if err != nil {
			logger.Error("cant' connect to db")
			return
		}

		st := storage.NewDBStorage(db, logger)
		lp := logparser.NewParser(st, 1000, logger)
		start := time.Now()
		err = lp.ParseLog(filename, logFormat)
		if err != nil {
			logger.Error("cant' parse log")
		}
		duration := time.Since(start)
		fmt.Printf("\n=== Parsing finished in: %v ===\n", duration)
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVarP(
		&logFormat,
		"format",
		"f",
		"",
		"Шаблон формата лога, например: [Label] [Timestamp] [Date] [Node] [Time] [NodeRepeat] [Type] [Component] [Level] [Content]",
	)

	parseCmd.Args = cobra.MinimumNArgs(1)
}
