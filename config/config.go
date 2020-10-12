package config

import "github.com/spf13/viper"

//Version Api version
var Version string

//RunDevMode  Sets it to allow developer niceties
func RunDevMode() bool {
	return viper.GetString("DEVEL") == "DEVEL"
}

//HostName external hostname of  server
func HostName() string {
	return viper.GetString("HOSTNAME")
}

//HTTPPort redirects from here
func HTTPPort() int {
	return viper.GetInt("HTTP_PORT")
}

//DropboxKey Dropbox secret
func DropboxKey() string {
	return viper.GetString("DROPBOX_KEY")
}

func RedisHost() string {
	return viper.GetString("REDIS_HOST")
}

func RedisPassword() string {
	return viper.GetString("REDIS_PASSWORD")
}
