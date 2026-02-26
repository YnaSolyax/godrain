package cmd

import (
	"fmt"

	"github.com/YnaSolyax/godrain/logparser"
	storage "github.com/YnaSolyax/godrain/storage/db"
	"github.com/spf13/cobra"
)

var logFormat string

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "parsing a file with logs of a specific format",
	Long: `This command is used for logs with a known format.
	The input is the file name and log format, for example: 
	parse BGL.log [Label] [Timestamp] [Date] [Node] [Time] [NodeRepeat] [Type] [Component] [Level] [Content]"`,

	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		format := logFormat

		if format == "" {
			fmt.Println("Ошибка: Необходимо указать формат лога c помощью флага --format.")
			return
		}

		logparser.ParseLog(&storage.DBStorage{}, filename, format)
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
