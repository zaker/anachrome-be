package config

import (
	"encoding/json"
	"fmt"
	"github.com/zaker/anachrome-be/cert"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
)

//WebConfig contains settings for webserver
type WebConfig struct {
	//IsDebug : Sets it to allow developer niceties
	IsDebug bool
	//HostName : hostname of  server
	HostName string
	//HTTPPort redirects from here
	HTTPPort int
	//HTTPSPort to here
	HTTPSPort int
	//AppDir path to SPA
	AppDir string

	//DebugSkipper
	DebugSkipper func(echo.Context) bool
}

// Load loads the config file or creates it
func Load(fileName string) (WebConfig, error) {
	if len(fileName) < 0 {
		fileName = "webConf.json"
	}

	conf := WebConfig{}

	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return conf, fmt.Errorf("Couldn't open config File %s \n  %v ", fileName, err)
	}
	err = json.Unmarshal(content, &conf)

	if err != nil {
		log.Println("Error reading config file", err)
		conf.IsDebug = true
		conf.HostName = "localhost"
		conf.HTTPPort = 8080
		conf.HTTPSPort = 8443
		m, err := json.Marshal(conf)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("webConf.json", m, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	conf.DebugSkipper = func(c echo.Context) bool {
		return !conf.IsDebug
	}
	return conf, nil
}

func fileExist(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

//HostURI returns normalized uri for host
func (w *WebConfig) HostURI() string {
	uri := "https://" + w.HostName
	if w.HTTPSPort != 443 {
		uri += ":" + strconv.Itoa(w.HTTPSPort)
	}
	return uri
}

func (w *WebConfig) SkipIfDebug(c echo.Context) bool {
	return w.IsDebug
}

// GenerateCert generates local cert or uses letsEncrypt
func (w *WebConfig) GenerateCert(e *echo.Echo) {
	if w.HostName == "localhost" {
		if !fileExist(".tmp/cert.pem") || !fileExist(".tmp/key.pem") {
			cert.GenerateCertFiles([]string{"localhost"}, 365*24*time.Hour, true)
		}
	} 
}
