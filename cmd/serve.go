package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zaker/anachrome-be/config"
	"github.com/zaker/anachrome-be/servers"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve anachrome",
	Long:  `serve anachrome`,
	Run:   runServe,
}

func createHTTPServerOptions() ([]servers.Option, error) {
	opts := []servers.Option{servers.WithAPIVersion(config.Version)}

	// if config.UseAuth() {
	// 	opts = append(opts,
	// 		servers.WithOAuth2(server.OAuth2Option{
	// 			AuthServer: config.AuthServer(),
	// 			Audience:   config.ResourceID(),
	// 			Issuer:     config.Issuer(),
	// 			ApiSecret:  []byte(config.ApiSecret()),
	// 		}))
	// }

	if len(config.HostName()) > 0 {
		opts = append(
			opts,
			servers.WithWebConfig(
				servers.WebConfig{
					HostName:  config.HostName(),
					HttpPort:  config.HttpPort(),
					HttpsPort: config.HttpsPort(),
					Cert:      config.Cert(),
					CertKey:   config.CertKey(),
				}))
	} else {
		return opts, fmt.Errorf("No host name")
	}

	// if config.HTTPOnly() {
	opts = append(
		opts,
		servers.WithHTTPOnly())
	// }

	// if config.UseLetsEncrypt() {
	// 	opts = append(
	// 		opts,
	// 		servers.WithLetsEncrypt(config.DomainList(), config.DomainMail()))
	// }

	// if config.UseTLS() {
	// 	opts = append(
	// 		opts,
	// 		servers.WithTLS(config.CertFile(), config.KeyFile()))
	// }

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
