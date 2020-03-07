package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	. "github.com/xaxys/oasis/api"
)

const ConfigPath = "./configs"

type oasisConfiguration struct {
	*viper.Viper
}

func (c *oasisConfiguration) SetAndWrite(key string, value interface{}) error {
	c.Set(key, value)
	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("write config Failed.")
	}
	return nil
}

func (c *oasisConfiguration) SubConfig(key string) Configuration {
	return &oasisConfiguration{c.Sub(key)}
}

func NewConfig(file string, defaultFields ...map[string]interface{}) Configuration {
	var c Configuration
	InitConfig(&c, file, defaultFields...)
	return c
}

func InitConfig(config *Configuration, name string, defaultFields ...map[string]interface{}) {
	initConfig(config, name, ServerConfig.GetString("ConfigType"), ConfigPath, defaultFields...)
}

func initConfig(config *Configuration, name string, configType string, configPath string, defaultFields ...map[string]interface{}) {
	v := viper.New()
	*config = &oasisConfiguration{v}

	v.SetConfigName(name)
	v.AddConfigPath(configPath)
	for _, m := range defaultFields {
		for key, value := range m {
			v.SetDefault(key, value)
		}
	}
	v.SetConfigType(configType)

	version := v.GetString("Version")

	// Create default config file
	CheckFolder(ConfigPath)
	if err := v.SafeWriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
			getLogger().Warn(err)
		}
	}
	if err := v.ReadInConfig(); err != nil {
		getLogger().Warn(err)
	}

	// Check config version and Update config if has field "Version"
	if version != "" && v.GetString("Version") < version {
		getLogger().Infof("Found config %s in an old version. Update to lasted version.", name)
		v.Set("Version", version)
		if err := v.WriteConfig(); err != nil {
			getLogger().Warn(err)
		}
	}

	if ServerConfig.GetBool("NotifyConfigChange") {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			getLogger().Infof("Config file changed: %s", e.Name)
		})
	}

	getLogger().Infof("Config %s initialized successfully", name)
}
