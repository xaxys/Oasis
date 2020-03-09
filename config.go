package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	. "github.com/spf13/viper"
	. "github.com/xaxys/oasis/api"
)

const ConfigPath = "./configs"

type oasisConfiguration struct {
	*Viper
}

func (c *oasisConfiguration) SetAndWrite(key string, value interface{}) error {
	c.Set(key, value)
	if err := c.WriteConfig(); err != nil {
		return fmt.Errorf("write config Failed.")
	}
	return nil
}

func (c *oasisConfiguration) UnmarshalKey(key string, rawVal interface{}) error {
	return c.Viper.UnmarshalKey(key, rawVal)
}

func (c *oasisConfiguration) Unmarshal(rawVal interface{}) error {
	return c.Viper.Unmarshal(rawVal)
}

func (c *oasisConfiguration) SubConfig(key string) Configuration {
	return &oasisConfiguration{c.Sub(key)}
}

func NewConfig(file string, defaultFields ...map[string]interface{}) Configuration {
	var c Configuration
	err, updated := InitConfig(&c, file, defaultFields...)
	if updated {
		getLogger().Infof("Found config %s in an old version. Update to latest version.", file)
	}
	if err != nil {
		getLogger().Warn(err)
		getLogger().Infof("Config %s initialized unsuccessfully", file)
	} else {
		getLogger().Infof("Config %s initialized successfully", file)
	}
	return c
}

func InitConfig(config *Configuration, name string, defaultFields ...map[string]interface{}) (Err error, updated bool) {
	Err, updated = initConfig(config, name, ServerConfig.GetString("ConfigType"), ConfigPath, defaultFields...)
	return Err, updated
}

func initConfig(config *Configuration, name string, configType string, configPath string, defaultFields ...map[string]interface{}) (Err error, updated bool) {

	v := New()
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
		if _, ok := err.(ConfigFileAlreadyExistsError); !ok {
			Err = fmt.Errorf("%v; %v", err, Err)
		}
	}
	if err := v.ReadInConfig(); err != nil {
		Err = fmt.Errorf("%v; %v", err, Err)
	}

	// Check config version and Update config if has field "Version"
	if version != "" && v.GetString("Version") < version {
		updated = true
		v.Set("Version", version)
		if err := v.WriteConfig(); err != nil {
			Err = fmt.Errorf("%v; %v", err, Err)
		}
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if ServerConfig.GetBool("NotifyConfigChange") {
			getLogger().Infof("Config file changed: %s", e.Name)
		}
	})

	return Err, updated
}
