package config

import "github.com/spf13/viper"

var Version string

// 	//IsDebug : Sets it to allow developer niceties
func RunDevMode() bool {
	return viper.GetString("DEVEL") == "DEVEL"
}

// 	//HostName : hostname of  server
func HostName() string {
	return viper.GetString("HOSTNAME")
}

// 	//HTTPPort redirects from here
func HttpPort() int {
	return viper.GetInt("HTTP_PORT")
}

func DropboxKey() string {
	return viper.GetString("DROPBOX_KEY")
}

// 	//DebugSkipper
// func 	DebugSkipper func(echo.Context) bool
