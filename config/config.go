package config

import "github.com/spf13/viper"

//WebConfig contains settings for webserver
// type WebConfig struct {
// 	//IsDebug : Sets it to allow developer niceties
// 	IsDebug bool
// 	//HostName : hostname of  server
// 	HostName string
// 	//HTTPPort redirects from here
// 	HTTPPort int
// 	//HTTPSPort to here
// 	HTTPSPort int
// 	//AppDir path to SPA
// 	AppDir string

// 	Cert    string
// 	CertKey string
// 	//DebugSkipper
// 	DebugSkipper func(echo.Context) bool
// }

var Version string

// 	//IsDebug : Sets it to allow developer niceties
func RunDevMode() bool {
	return viper.GetString("DEVEL") == "DEVEL"
}

// 	//HostName : hostname of  server
func HostName() string {
	return viper.GetString("HOSTNAME")
}

func HTTPOnly() bool {
	return viper.GetBool("HTTP_ONLY")
}

// 	//HTTPPort redirects from here
func HttpPort() int {
	return viper.GetInt("HTTP_PORT")
}

// 	//HTTPSPort to here
func HttpsPort() int {
	return viper.GetInt("HTTPS_PORT")
}

// 	//AppDir path to SPA
func AppDir() string {
	return viper.GetString("APP_DIR")
}

func Cert() string {
	return viper.GetString("CERT_PATH")
}
func CertKey() string {
	return viper.GetString("KEY_PATH")
}

func TWFile() string {
	return viper.GetString("TIDDLY_PATH")
}

// 	//DebugSkipper
// func 	DebugSkipper func(echo.Context) bool
