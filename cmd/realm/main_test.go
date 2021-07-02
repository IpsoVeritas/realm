package main

import (
	"flag"
	"os"
	"runtime"
	"testing"

	logger "github.com/Brickchain/go-logger.v1"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	flag.Parse()
	_ = godotenv.Load(".env.test")
	viper.AutomaticEnv()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(viper.GetString("log_formatter"))
	logger.SetLevel("debug")
	viper.SetDefault("gorm_debug", "false")
	viper.SetDefault("base", "http://test.local")
	viper.SetDefault("proxy_domain", "r.integrity.app")

	viper.SetDefault("filestore_dir", ".test-files")
	defer os.RemoveAll(".test-files")

	viper.SetDefault("revocation", "https://revocation.plusintegrity.com")

	w = logger.GetLogger().Logger.Writer()
	defer w.Close()

	viper.SetDefault("cache", "inmem")

	viper.SetDefault("password", "test")

	viper.SetDefault("gorm_dialect", "sqlite3")
	runtime.GOMAXPROCS(1) // this is required when using an inmem sqlite db
	viper.SetDefault("gorm_options", ":memory:")
	viper.SetDefault("config", "test.yml")

	return m.Run()
}

// func TestVersion(t *testing.T) {
// 	handlers := loadHandler()

// 	_, err := testhelper.DoHttpRequest(handlers, http.MethodGet, "/", "")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
