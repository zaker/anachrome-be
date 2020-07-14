package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zaker/anachrome-be/config"
	"github.com/zaker/anachrome-be/servers"
	"github.com/zaker/anachrome-be/services"
	"github.com/zaker/anachrome-be/stores"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve anachrome",
	Long:  `serve anachrome`,
	Run:   runServe,
}

func createHTTPServerOptions() ([]servers.Option, error) {
	opts := []servers.Option{servers.WithAPIVersion(config.Version)}

	if len(config.HostName()) > 0 {
		opts = append(
			opts,
			servers.WithWebConfig(
				servers.WebConfig{
					HostName: config.HostName(),
					HTTPPort: config.HTTPPort(),
				}))
	} else {
		return opts, fmt.Errorf("Cannot create server without hostname")
	}
	opts = append(
		opts,
		servers.WithBlogStore(stores.NewDropboxBlogStore(config.DropboxKey())))

	opts = append(
		opts,
		servers.WithGQL())

	if config.RunDevMode() {
		opts = append(
			opts,
			servers.WithDevMode())
	}

	authn, err := services.NewAuthN(
		&stores.UserFileStore{},
		&services.AuthNConfig{})
	if err != nil {
		return opts, err
	}
	opts = append(
		opts,
		servers.WithAuthN(authn))
	return opts, nil
}

func serve(opts []servers.Option) error {

	hs, err := servers.NewHTTPServer(opts...)

	if err != nil {
		return fmt.Errorf("Error configuring http server %w", err)
	}
	err = hs.Serve()

	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("Error running http server %w", err)
	}
	return nil
}

func runServe(cmd *cobra.Command, args []string) {

	if viper.ConfigFileUsed() == "" {
		log.Println("Config from environment variables")
	}

	opts, err := createHTTPServerOptions()
	if err != nil {
		log.Fatal("Creating http server options", err)

	}

	err = serve(opts)
	if err != nil {
		log.Fatal("Error starting http server", err)

	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
