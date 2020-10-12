package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/zaker/anachrome-be/stores/cache"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zaker/anachrome-be/config"
	"github.com/zaker/anachrome-be/servers"
	"github.com/zaker/anachrome-be/services"
	"github.com/zaker/anachrome-be/stores"
	"github.com/zaker/anachrome-be/stores/blog"
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

	dbxBlog := blog.NewDropboxBlogStore(
		&http.Client{},
		config.DropboxKey(),
		"/blog",
		"ptid:vjStHN01QQQAAAAAAABF4g")

	var bs blog.BlogStore
	if len(config.RedisHost()) > 0 {
		cachedBlogStore, err := cache.NewRedisBlogCache(dbxBlog, config.RedisHost())

		if err != nil {
			return opts, err
		}
		bs = cachedBlogStore

		go func() {
			for id := range dbxBlog.UpdatesChan {
				cachedBlogStore.Invalidate(context.Background(), id)
			}
		}()
	} else {
		bs = dbxBlog
	}

	opts = append(
		opts,
		servers.WithBlogStore(bs))

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
