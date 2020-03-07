package main

import (
	"fmt"

	. "github.com/xaxys/oasis/api"
)

// Default Configs

func initServerConfig() {
	initConfig(&ServerConfig, ServerConfigName, "yml", ".", serverConfigDefault)
	initConfig(&PluginManagerConfig, PluginManagerConfigName, "yml", ".", pluginManagerConfigDefault)
}

const ServerConfigName = "server"
const PluginManagerConfigName = "plugin"

var ServerConfig Configuration
var PluginManagerConfig Configuration

var serverConfigDefault = map[string]interface{}{
	"Version":            "0.1.4",
	"LogLevel":           "info",
	"LogPath":            "./logs/log.log",
	"PluginResourcePath": "./resources",
	"PluginPath":         "./plugins",
	"ConfigType":         "yml",
	"DebugMode":          false,
	"NotifyConfigChange": true,
}

var pluginManagerConfigDefault = map[string]interface{}{}

// Default Commands

func initServerCommands() {
	getCommandManager().RegisterCommand("stop", nil, stopCommandExcutor)
	getCommandManager().RegisterCommand("exit", nil, stopCommandExcutor)
	getCommandManager().RegisterCommand("quit", nil, stopCommandExcutor)

	getCommandManager().RegisterCommand("pm", nil, pluginCommandExcutor)
	getCommandManager().RegisterCommand("plugin", nil, pluginCommandExcutor)
	getCommandManager().RegisterCommand("pluginmanager", nil, pluginCommandExcutor)
}

var stopCommandExcutor StopCommandExcutor

type StopCommandExcutor struct{}

func (StopCommandExcutor) OnCommand(p Plugin, command string, args []string) {
	getServer().Stop()
}

var pluginCommandExcutor PluginCommandExcutor

type PluginCommandExcutor struct{}

func (PluginCommandExcutor) OnCommand(p Plugin, command string, args []string) {
	if len(args) == 1 && (args[0] == "l" || args[0] == "list") {
		plugin := GetServer().GetEnabledPlugins()
		fmt.Printf("Found %d Enabled plugins:", len(plugin))
		for i, v := range plugin {
			if i%5 == 0 {
				fmt.Println()
			}
			fmt.Printf("[%s] \t", v)
		}
		fmt.Println()
		plugin = GetServer().GetDisabledPlugins()
		fmt.Printf("Found %d Disabled plugins:", len(plugin))
		for i, v := range plugin {
			if i%5 == 0 {
				fmt.Println()
			}
			fmt.Printf("[%s] \t", v)
		}
		fmt.Println()
		return
	}
	if len(args) > 1 && (args[0] == "i" || args[0] == "info") {
		for _, v := range args[1:] {
			plugin := GetServer().GetPlugin(v)
			if plugin == nil {
				fmt.Printf("No such a plugin Named: %s", v)
			} else {
				fmt.Println(plugin.GetDetailedInfo())
			}
		}
		return
	}
	if len(args) > 1 && (args[0] == "e" || args[0] == "enable") {
		for _, v := range args[1:] {
			plugin := GetServer().GetPlugin(v)
			if plugin == nil {
				fmt.Printf("No such a plugin Named: %s", v)
			} else {
				plugin.Enable()
			}
		}
		return
	}
	if len(args) > 1 && (args[0] == "d" || args[0] == "disable") {
		for _, v := range args[1:] {
			plugin := GetServer().GetPlugin(v)
			if plugin == nil {
				fmt.Printf("No such a plugin Named: %s", v)
			} else {
				plugin.Disable()
			}
		}
		return
	}
	if len(args) > 1 && (args[0] == "r" || args[0] == "restart") {
		for _, v := range args[1:] {
			plugin := GetServer().GetPlugin(v)
			if plugin == nil {
				fmt.Printf("No such a plugin Named: %s", v)
			} else {
				plugin.Disable()
				plugin.Enable()
			}
		}
		return
	}
	if len(args) > 1 && (args[0] == "u" || args[0] == "usage") {
		for _, v := range args[1:] {
			plugin := GetServer().GetPlugin(v)
			if plugin == nil {
				fmt.Printf("No such a plugin Named: %s", v)
			} else {
				usages := GetServer().GetPluginCommands(plugin)
				fmt.Printf("Usages of [%s]:", plugin)
				for i, v := range usages {
					if i%5 == 0 {
						fmt.Println()
					}
					fmt.Printf("%s \t", v.Command)
				}
				fmt.Println()
			}
		}
		return
	}

	fmt.Println("------------[PluginManager Usage]------------")
	fmt.Println(">>> l[ist] --------	| list plugins and statues")
	fmt.Println(">>> i[nfo] <plugin>	| Show plugin info")
	fmt.Println(">>> e[nable] <plugin>	| Enable plugin")
	fmt.Println(">>> d[isable] <plugin>	| Disable plugin")
	fmt.Println(">>> u[sage] <plugin>	| Check registed commands")
}
