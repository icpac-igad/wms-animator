package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration for system
var Configuration Config

func setDefaultConfig() {
	viper.SetDefault("Server.HttpHost", "0.0.0.0")
	viper.SetDefault("Server.HttpPort", 9000)
	viper.SetDefault("Server.UrlBase", "")
	viper.SetDefault("Server.BasePath", "")
	viper.SetDefault("Server.CORSOrigins", "*")
	viper.SetDefault("Server.Debug", false)

	viper.SetDefault("Server.ReadTimeoutSec", 5)
	viper.SetDefault("Server.WriteTimeoutSec", 30)

	viper.SetDefault("Metadata.Title", "wms-animator")
	viper.SetDefault("Metadata.Description", "WMS Animator")
}

// Config for system
type Config struct {
	Server   Server
	Metadata Metadata
	Wms      Wms
}

// Server config
type Server struct {
	HttpHost        string
	HttpPort        int
	UrlBase         string
	BasePath        string
	CORSOrigins     string
	Debug           bool
	ReadTimeoutSec  int
	WriteTimeoutSec int
}

type Metadata struct {
	Title       string
	Description string
}

type Wms struct {
	FontPath string
}

// InitConfig initializes the configuration from the config file
func InitConfig(configFilename string) {
	// --- defaults
	setDefaultConfig()

	isExplictConfigFile := configFilename != ""
	confFile := AppConfig.Name + ".toml"

	if configFilename != "" {
		viper.SetConfigFile(configFilename)
		confFile = configFilename
	} else {
		viper.SetConfigName(confFile)
		viper.SetConfigType("toml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/config")
		viper.AddConfigPath("/etc")
	}

	fmt.Println(configFilename)

	err := viper.ReadInConfig() // Find and read the config file

	if err != nil {
		_, isConfigFileNotFound := err.(viper.ConfigFileNotFoundError)
		errrConfRead := fmt.Errorf("fatal error reading config file: %s", err)
		isUseDefaultConfig := isConfigFileNotFound && !isExplictConfigFile
		if isUseDefaultConfig {
			confFile = "DEFAULT" // let user know config is defaulted
			log.Debug(errrConfRead)
		} else {
			log.Fatal(errrConfRead)
		}
	}

	log.Infof("Using config file: %s", viper.ConfigFileUsed())
	viper.Unmarshal(&Configuration)

	if port := os.Getenv("PORT"); port != "" {
		iPort, err := strconv.Atoi(port)

		if err != nil {
			log.Fatal(fmt.Errorf("invalid port:%s", port))
		}
		Configuration.Server.HttpPort = iPort
	}

	// sanitize the configuration
	Configuration.Server.BasePath = strings.TrimRight(Configuration.Server.BasePath, "/")

	//fmt.Printf("Viper: %v\n", viper.AllSettings())
	//fmt.Printf("Config: %v\n", Configuration)
}
