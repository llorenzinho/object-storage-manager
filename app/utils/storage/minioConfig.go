package storage

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type StorageConfig struct {
	MinioStorageConfigStruct MinioStorageConfigStruct `mapstructure:"storage"`
}

type MinioStorageConfigStruct struct {
	MinioStorageConfig MinioStorageConfig `mapstructure:"minio"`
}

type MinioStorageConfig struct {
	Url  string           `mapstructure:"url"`
	Port uint16           `mapstructure:"port"`
	Auth MinioStorageAuth `mapstructure:"auth"`
}

type MinioStorageAuth struct {
	AccessKeyID     string `mapstructure:"accessKeyID"`
	SecretAccessKey string `mapstructure:"secretAccessKey"`
}

// Read config into Struct using viper
func ReadConfig() *StorageConfig {
	var config *StorageConfig = &StorageConfig{}
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %w", err))
	}
	log.Default().Println("Storage config loaded")
	return config
}
