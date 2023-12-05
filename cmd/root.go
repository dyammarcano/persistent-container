package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "dataStore",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultDatabasePath := os.Getenv("DATABASE_PATH")
	if defaultDatabasePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Can't bind database flag:", err)
			os.Exit(1)
		}
		defaultDatabasePath = filepath.Join(wd, "dataStore.db")
	}

	rootCmd.PersistentFlags().String("database", defaultDatabasePath, "full path to database file")
	if err := viper.BindPFlag("database", rootCmd.PersistentFlags().Lookup("database")); err != nil {
		fmt.Println("Can't bind database flag:", err)
		os.Exit(1)
	}
}
