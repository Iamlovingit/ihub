package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// DBConfig .
type DBConfig struct {
	Name           string `yaml:"NAME"`
	Host           string `yaml:"HOST"`
	Port           int    `yaml:"PORT"`
	User           string `yaml:"USER"`
	Passwd         string `yaml:"PASSWD"`
	Charset        string `yaml:"CHARSET"`
	SM2PrivateFile string `yaml:"SM2PRIVATEFILE"`
	MaxOpenConns   int    `yaml:"MAXOPENCONNS"`
	MaxIdleConns   int    `yaml:"MAXIDLECONNS"`
}

// LogConfig ...
type LogConfig struct {
	Level             string `yaml:"level"`
	DefaultConfigName string `yaml:"defaultConfigName"`
}

// ServerConfig ...
type ServerConfig struct {
	Port int `yaml:"port"`
}

// CacheConfig ...
type CacheConfig struct {
	CachePath string `yaml:"cachePath"`
}

type MidwareConfig struct {
	Midware string `yaml:"midware"`
}

// Configuration ...
type Configuration struct {
	DB       DBConfig        `yaml:"DB"`
	LOG      LogConfig       `yaml:"log"`
	SERVER   ServerConfig    `yaml:"server"`
	CACHE    CacheConfig     `yaml:"cache"`
	Midwares []MidwareConfig `yaml:"midwares"`
	Runmode  string          `yaml:"runmode"`
}

var gConfig Configuration

var configInstance *viper.Viper

// 初始化
// 配置文件路径
func Init(filename string) error {
	// 设置配置文件的名字
	viper.SetConfigFile(filename)
	// 设置配置文件的类型
	// viper.SetConfigType("yaml")
	// 添加配置文件所在的路径
	// viper.AddConfigPath("../hfproxy-config.yaml")
	// 启动自动检索配置文件中的环境变量并导入viper功能
	viper.AutomaticEnv()
	// 读取并载入配置文件至内存
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	// 检查配置文件是否合法
	if err := validateConfig(viper.GetViper()); err != nil {
		return err
	}

	// 将配置文件的数据反序列化到结构体gConfig中
	viper.Unmarshal(&gConfig)
	configInstance = viper.GetViper()

	// 监听配置文件的变化，并将变化更新到gConfig中
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		viper.Unmarshal(&gConfig)
	})
	viper.WatchConfig()
	return nil
}

// 返回配置文件的实例
func GetConfig() *Configuration {
	return &gConfig
}

// 待完成
func validateConfig(viper *viper.Viper) error {
	return nil
}
