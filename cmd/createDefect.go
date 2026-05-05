package cmd

import (
	"fmt"

	storage "github.com/YnaSolyax/godrain/storage/db"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	description string
	solution    string
)

var defectCmd = &cobra.Command{
	Use:   "add-defect",
	Short: "Add a new defect to the knowledge base",
	Long:  `Manually create a defect with a description and solution for future vector similarity search. example: ./godrain add-defect --description "Hardware failure" --solution "Replace the module"`,

	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewProduction()
		defer logger.Sync()

		db, err := storage.Conn("host=localhost user=user password=1 dbname=log_analysis port=5432 sslmode=disable", logger)
		if err != nil {
			logger.Error("failed to connect to database")
			return
		}

		st := storage.NewDBStorage(db, logger)
		existingID, vec, err := st.FindDefectByText(description, 0.8)
		if err != nil {
			logger.Error("failed to get vector from ollama", zap.Error(err))
			return
		}

		if existingID != 0 {
			logger.Info("defect already exist")
			return
		}
		err = st.CreateDefect(description, solution, vec)
		if err != nil {
			logger.Error("failed to create defect", zap.Error(err))
			return
		}

		fmt.Println("Defect successfully added to the knowledge base")
	},
}

func init() {
	rootCmd.AddCommand(defectCmd)

	defectCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the issue")
	defectCmd.Flags().StringVarP(&solution, "solution", "s", "", "Solution for the issue")

	defectCmd.MarkFlagRequired("description")
	defectCmd.MarkFlagRequired("solution")
}
