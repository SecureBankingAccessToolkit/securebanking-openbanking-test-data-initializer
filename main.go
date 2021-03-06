package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"securebanking-test-data-initializer/pkg/common"
	"securebanking-test-data-initializer/pkg/httprest"
	platform "securebanking-test-data-initializer/pkg/identity-platform"
	"securebanking-test-data-initializer/pkg/rs"
	"securebanking-test-data-initializer/pkg/types"
)

// init function is execute before main to initialize the program,
// this function is called when the package is initialized
func init() {
	fmt.Println("initializing the program.....")
	viper.AutomaticEnv()
	viper.SetDefault("ENVIRONMENT.VERBOSE", false)
	viper.SetDefault("ENVIRONMENT.STRICT", true)
	viper.SetDefault("ENVIRONMENT.VIPER_CONFIG", "default")
	viper.SetDefault("IDENTITY.AM_REALM", "alpha")
	// load default logger
	fmt.Println("initializing the default logger.....")
	loadLogger()
	loadConfiguration()
	// load logger again to update the level set in the configuration file
	loadLogger()
	checks()
	// after call 'loadConfiguration' we have an object with all configuration mapped
	if common.Config.Environment.Verbose {
		verboseProgramInfo()
	}

	if viper.GetBool("ENVIRONMENT.ONLY_CONFIG") {
		os.Exit(0)
	}
}

// verboseProgramInfo is a method to add all additional information about the program to output in the console in verbose/debug mode
func verboseProgramInfo() {
	fmt.Println("IdentityPlatformFQDN:", common.Config.Hosts.IdentityPlatformFQDN)
	zap.S().Infow("Configuration", "config", config)
}

// config to get configuration values
var config types.Configuration

func main() {
	// operations
	checkValidPlatformCert()
	session := getIdentityPlatformSession()

	fmt.Println("Resty initialization....")

	//get IDM auth code
	session.Authenticate()

	//to obtain cookies values
	httprest.InitRestReaderWriter(session.Cookie, session.AuthToken.AccessToken)
	userId := rs.CreatePSU()
	rs.PopulateRSData(userId)
}

func loadLogger() {
	logger, e := common.ConfigureLogger()
	if e != nil {
		panic(e)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {

		}
	}(logger)

	zap.ReplaceGlobals(logger)
}

func loadConfiguration() {
	fmt.Println("Load the [", viper.GetString("ENVIRONMENT.VIPER_CONFIG"), "] configuration.....")
	err := common.LoadConfigurationByEnv(viper.GetString("ENVIRONMENT.VIPER_CONFIG"))
	if err != nil {
		zap.S().Fatalw("Cannot load config:", "error", err)
	}
	config = common.Config
}

func checks() {
	fmt.Println("Making some checks.....")
}

func getIdentityPlatformSession() *common.Session {
	zap.L().Info("Get CookieName")
	c := platform.GetCookieNameFromAm()
	zap.L().Info("Get user session")
	return platform.FromUserSession(c)
}

// Operations
func checkValidPlatformCert() {
	zap.L().Info("Check valid cert")
	if !platform.IsValidX509() {
		zap.L().Fatal("No Valid SSL certificate present in the cdk")
	}
}
