package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

type BaseConfig struct {
	Env string `yaml:"env"` // Dev or Prod
}

type WebConfig struct {
	Port int `yaml:"port"`
}

type TelegramConfig struct {
	Token string `yaml:"token"`
}

type MongoConfig struct {
	Source   string `yaml:"source"`
	DataBase string `yaml:"data_base"`
}

type Log struct {
	Path string `yaml:"path"`
}

type AwsPinPoint struct {
	KeyId  string `yaml:"key_id"`
	Secret string `yaml:"secret"`
}

type AdminConfig struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

type Config struct {
	Base     *BaseConfig     `yaml:"base"`
	Mongo    *MongoConfig    `yaml:"mongo"`
	Web      *WebConfig      `yaml:"web"`
	TG       *TelegramConfig `yaml:"telegram"`
	Log      *Log            `yaml:"log"`
	PinPoint *AwsPinPoint    `yaml:"aws_pinpoint"`
	AdminCfg *AdminConfig    `yaml:"admin_cfg"`
}

var config *Config
var secrets map[string]string

func GetConfig() *Config {
	return config
}

func IsProd() bool {
	return GetConfig().Base.Env == "Prod"
}

func IsTest() bool {
	return GetConfig().Base.Env == "Test"
}

func IsDev() bool {
	return GetConfig().Base.Env == "Dev"
}

func init() {
	if config == nil {
		load()
	}
}

func load() {
	secrets = make(map[string]string)
	bytes, err := os.ReadFile("./conf/secret.yml")
	if err != nil {
		log.Printf("secret.yml not found: %v\n", err)
	} else {
		yerr := yaml.Unmarshal(bytes, &secrets)
		if yerr != nil {
			log.Fatalf("unmarshal secrets file error: %v\n", err)
		}
	}

	cfg := &Config{}
	bytes, err = os.ReadFile("./conf/config.yml")
	if err != nil {
		log.Fatalf("load config file failed: %v\n", err)
	}

	cfgStr := string(bytes)

	for key, secret := range secrets {
		r := "${" + key + "}"
		cfgStr = strings.ReplaceAll(cfgStr, r, secret)
	}

	bytes = []byte(cfgStr)

	yerr := yaml.Unmarshal(bytes, cfg)
	if yerr != nil {
		log.Fatalf("unmarshal config file failed: %v\n", yerr)
		return
	}

	config = cfg
}
