package config

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	routes "github.com/llorenzinho/object-storage-manager/routes/files"
	"github.com/llorenzinho/object-storage-manager/utils/storage"
	"github.com/spf13/viper"
)

func Init() {
	initConfig()
	initStorage()
	initGin()
}

// Load configurations from config.json file using viper
func initConfig() {
	viper.SetConfigName("config")
	viper.SetEnvPrefix("PROD")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetConfigType("json")
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("unable to initialize viper: %w", err))
	}
	log.Println("Config loaded")
}

// Init Gin app
func initGin() {
	port := viper.GetString("server.port")
	mode := viper.GetString("server.mode")

	if mode == "" || (mode != gin.DebugMode && mode != gin.ReleaseMode) {
		mode = gin.DebugMode
	}

	gin.SetMode(mode)
	r := gin.Default()
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	routes.FilesRouter(r)
	if port == "" {
		panic(fmt.Errorf("unable to parse port"))
	}
	r.Run(fmt.Sprintf(":%s", port))

}

func initStorage() {
	storage.Instance()
}
