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
		return opts, fmt.Errorf("cannot create server without hostname")
	}

	dbxBlog := blog.NewDropboxBlogStore(
		&http.Client{},
		config.DropboxKey(),
		"/blog",
		"ptid:vjStHN01QQQAAAAAAABF4g")

	var bs cache.CachedBlogStore
	if len(config.RedisHost()) > 0 {
		cachedBlogStore, err := cache.NewRedisBlogCache(dbxBlog, config.RedisHost())

		if err != nil {
			return opts, err
		}
		bs = cachedBlogStore

	} else {
		cachedBlogStore, err := cache.NewInMemoryCache(dbxBlog)
		if err != nil {
			return opts, err
		}
		bs = cachedBlogStore

	}

	go func() {
		for id := range dbxBlog.UpdatesChan {
			err := bs.Invalidate(context.Background(), id)
			if err != nil {
				log.Println("warn: invalidating", err)
			}
		}
	}()

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

	return opts, nil
}

func serve(opts []servers.Option) error {

	hs, err := servers.NewHTTPServer(opts...)

	if err != nil {
		return fmt.Errorf("error configuring http server %w", err)
	}
	err = hs.Serve()

	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error running http server %w", err)
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
