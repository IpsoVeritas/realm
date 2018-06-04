package cache

import "github.com/spf13/viper"

func LoadFromEnv() (Cache, error) {
	var s Cache
	var err error

	switch viper.GetString("cache") {
	case "redis":
		s = NewRedisCache(viper.GetString("redis"))
	}

	return s, err
}
