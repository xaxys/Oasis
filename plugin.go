package main

import (
	"fmt"
	goplugin "plugin"

	. "github.com/xaxys/oasis/api"
)

type oasisPlugin struct {
	PluginDescription
	pluginProperty
	UserPlugin
	goPlugin *goplugin.Plugin
	enabled  bool
	loaded   bool
}

type pluginProperty struct {
	this   Plugin
	logger Logger
	config Configuration
	folder string
}

func (pp *pluginProperty) GetPlugin() Plugin {
	return pp.this
}

func (pp *pluginProperty) GetServer() Server {
	return GetServer()
}

func (pp *pluginProperty) GetLogger() Logger {
	return pp.logger
}

func (pp *pluginProperty) GetConfig() Configuration {
	return pp.config
}

func (pp *pluginProperty) GetFolder() string {
	return pp.folder
}

func NewPlugin(goPlugin *goplugin.Plugin) (*oasisPlugin, error) {

	up, err := goPlugin.Lookup("PLUGIN")
	if err != nil {
		return nil, fmt.Errorf("variable PLUGIN is not found. Details: %v", err)
	}
	userPlugin, ok := up.(*UserPlugin)
	if !ok {
		return nil, fmt.Errorf("variable PLUGIN isn't a (*UserPlugin) interface. Found %T", up)
	}

	description := (*userPlugin).GetDescription()

	if description.Name == "" {
		return nil, fmt.Errorf("Plugin name is empty")
	}

	p := &oasisPlugin{
		goPlugin:          goPlugin,
		PluginDescription: description,
		pluginProperty:    pluginProperty{},
		UserPlugin:        *userPlugin,
		enabled:           false,
		loaded:            false,
	}

	return p, nil
}

func (p *oasisPlugin) Load() bool {
	if p.loaded {
		return false
	}
	getLogger().Infof("Loading Plugin [%s]...", p)

	p.folder = CheckFolder(ServerConfig.GetString("PluginResourcePath"), p.GetName())
	p.logger = GetPluginLogger(p.GetName())
	p.config = NewConfig(p.GetName(), p.DefaultConfigFields)
	p.this = p

	p.EntryPoint(&p.pluginProperty)
	res := p.OnLoad()
	p.loaded = true
	return res
}

func (p *oasisPlugin) Enable() bool {
	if p.enabled {
		return false
	}
	getLogger().Infof("Enabling Plugin [%s]...", p)

	p.enabled = true
	res := p.OnEnable()
	return res
}

func (p *oasisPlugin) Disable() bool {
	if !p.enabled || !p.loaded {
		return false
	}
	getLogger().Infof("Disabling Plugin [%s]...", p)
	getTaskManager().UnregisterPluginTask(p)

	p.enabled = true
	res := p.OnDisable()
	return res
}

func (p *oasisPlugin) GetName() string {
	return p.Name
}

func (p *oasisPlugin) GetVersion() string {
	return p.Version
}

func (p *oasisPlugin) GetDescription() string {
	return p.Description
}

func (p *oasisPlugin) GetAuthor() string {
	return p.Author
}

func (p *oasisPlugin) GetFolder() string {
	return p.folder
}

func (p *oasisPlugin) GetConfig() Configuration {
	return p.config
}

func (p *oasisPlugin) IsEnabled() bool {
	return p.enabled
}

func (p *oasisPlugin) IsLoaded() bool {
	return p.loaded
}

func (p *oasisPlugin) GetDependencies() []PluginDependency {
	return p.Dependencies
}

func (p *oasisPlugin) GetSoftDependencies() []PluginDependency {
	return p.SoftDependencies
}

func (p *oasisPlugin) GetDetailedInfo() string {
	info := fmt.Sprintf(`
	[Plugin]
		[Name]: %s
		[Version]: %s
		[Author]: %s
		[Description]: %s
		[Enabled]: %v
	`,
		p.Name,
		p.Version,
		p.Author,
		p.Description,
		p.enabled)
	return info
}

func (p *oasisPlugin) String() string {
	return fmt.Sprintf("%s version=%s", p.GetName(), p.GetVersion())
}
