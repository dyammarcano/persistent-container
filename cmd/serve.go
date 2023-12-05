package cmd

import (
	"context"
	"fmt"
	"github.com/caarlos0/log"
	"github.com/dyammarcano/persistent-container/internal/monitoring"
	"github.com/dyammarcano/persistent-container/internal/store"
	"github.com/spf13/viper"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: server,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().Int("port", 8080, "port to the server")
	if err := viper.BindPFlag("port", serveCmd.Flags().Lookup("port")); err != nil {
		fmt.Println("Can't bind port flag:", err)
		os.Exit(1)
	}
}

func server(cmd *cobra.Command, args []string) error {
	databasePath := viper.GetString("database")
	log.Infof("database path is %s", databasePath)

	ctx := context.Background()
	databasePath, err := filepath.Abs(databasePath)
	if err != nil {
		return err
	}

	db, err := store.NewStore(ctx, databasePath)
	if err != nil {
		return err
	}

	port := viper.GetInt("port")
	log.Infof("Starting server on port %d", port)

	newMonitoring := monitoring.NewMonitoring(ctx, db, port)

	callback := func(err error) {
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	newMonitoring.StartServer(callback)

	return nil
}
