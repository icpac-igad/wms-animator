package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/icpac-igad/wms-animator/internal/conf"
	"github.com/icpac-igad/wms-animator/internal/service"

	"github.com/pborman/getopt/v2"
	log "github.com/sirupsen/logrus"
)

var flagDebugOn bool
var flagVersion bool
var flagConfigFilename string
var flagWmsFontPath string

func init() {
	initCommandOptions()
}

func initCommandOptions() {
	getopt.FlagLong(&flagConfigFilename, "config", 'c', "", "config file name")
	getopt.FlagLong(&flagDebugOn, "debug", 'd', "Set logging level to TRACE")
	getopt.FlagLong(&flagWmsFontPath, "font", 'f', "Font file path")
	getopt.FlagLong(&flagVersion, "version", 'v', "Output the version information")

}

func main() {
	getopt.Parse()

	if flagVersion {
		fmt.Printf("%s %s\n", conf.AppConfig.Name, conf.AppConfig.Version)
		os.Exit(1)
	}

	log.Infof("----  %s - Version %s ----------\n", conf.AppConfig.Name, conf.AppConfig.Version)

	conf.InitConfig(flagConfigFilename)

	if flagDebugOn {
		log.SetLevel(log.TraceLevel)
		log.Debugf("Log level = DEBUG\n")
	}

	if flagWmsFontPath != "" {
		conf.Configuration.Wms.FontPath = flagWmsFontPath
	}

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	// start the web service
	service.Serve()
}
