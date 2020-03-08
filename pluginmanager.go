package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	goplugin "plugin"
	"strings"
	"sync"

	. "github.com/xaxys/oasis/api"
)

var pluginManagerLock sync.Mutex
var pluginManager *oasisPluginManager

type oasisPluginManager struct {
	unloadedPlugins []*pluginInfo
	disabledPlugins []*pluginInfo
	enabledPlugins  []*pluginInfo
	pluginTable     map[string]*pluginInfo
	dependencyMap   map[string][]string
}

type pluginInfo struct {
	Plugin
	dependenciesCount int
}

func (p *pluginInfo) String() string {
	return fmt.Sprint(p.Plugin)
}

func getPluginManager() *oasisPluginManager {
	if pluginManager == nil {
		pluginManagerLock.Lock()
		if pluginManager == nil {
			pluginManager = newPluginManager()
		}
		pluginManagerLock.Unlock()
	}
	return pluginManager
}

func newPluginManager() *oasisPluginManager {
	return &oasisPluginManager{
		pluginTable:   map[string]*pluginInfo{},
		dependencyMap: map[string][]string{},
	}
}

// GetPlugins equals GetEnabledPlugins
func (pm *oasisPluginManager) GetPlugins() []Plugin {
	return pm.GetEnabledPlugins()
}

func (pm *oasisPluginManager) GetEnabledPlugins() []Plugin {
	var list []Plugin
	pluginManagerLock.Lock()
	for _, p := range pm.enabledPlugins {
		list = append(list, p)
	}
	pluginManagerLock.Unlock()
	return list
}

func (pm *oasisPluginManager) GetDisabledPlugins() []Plugin {
	var list []Plugin
	pluginManagerLock.Lock()
	for _, p := range pm.disabledPlugins {
		list = append(list, p)
	}
	pluginManagerLock.Unlock()
	return list
}

// GetAllPlugins return a list of enabled and disabled plugins
func (pm *oasisPluginManager) GetAllPlugins() []Plugin {
	var list []Plugin
	pluginManagerLock.Lock()
	for _, p := range pm.enabledPlugins {
		list = append(list, p)
	}
	for _, p := range pm.disabledPlugins {
		list = append(list, p)
	}
	pluginManagerLock.Unlock()
	return list
}

func (pm *oasisPluginManager) GetPlugin(name string) Plugin {
	pluginManagerLock.Lock()
	v, ok := pm.pluginTable[name]
	if !ok {
		getLogger().Debugf("Plugin %s is not found", name)
	}
	pluginManagerLock.Unlock()
	return v.Plugin
}

// checkPluginFile return Plugin if success
func (pm *oasisPluginManager) checkPluginFile(name string) (*pluginInfo, error) {

	if !strings.HasSuffix(name, ".so") {
		name = name + ".so"
	}
	getLogger().Infof("Checking plugin file %s", name)

	pluginpath := filepath.Join(ServerConfig.GetString("PluginPath"), name)
	if _, err := os.Stat(pluginpath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Plugin file %s is not found. Details: %v", name, err)
	}

	goplugin, err := goplugin.Open(pluginpath)
	if err != nil {
		return nil, fmt.Errorf("Plugin file %s isn't a valid go plugin. Details: %v", name, err)
	}

	p, err := NewPlugin(goplugin)
	if err != nil {
		return nil, fmt.Errorf("Plugin file %s isn't a valid oasis plugin. Details: %v", name, err)
	}

	pinfo := &pluginInfo{
		p,
		len(p.GetDependencies()),
	}

	pm.pluginTable[p.GetName()] = pinfo

	return pinfo, nil
}

// checkPluginConfig check config to return it's enable statue
// and generate default plugin configs in PluginManagerConfig
func (pm *oasisPluginManager) checkPluginConfig(p Plugin) bool {
	name := p.GetName()
	enable := true
	if PluginManagerConfig.IsSet(name + ".Enable") {
		enable = PluginManagerConfig.GetBool(name + ".Enable")
	}
	PluginManagerConfig.Set(name, map[string]interface{}{
		"Enable":           enable,
		"Version":          p.GetVersion(),
		"Description":      p.GetDescription(),
		"Author":           p.GetAuthor(),
		"Dependencies":     p.GetDependencies(),
		"SoftDependencies": p.GetSoftDependencies(),
	})
	if err := PluginManagerConfig.WriteConfig(); err != nil {
		getLogger().Warn(err)
	}
	return enable
}

func (pm *oasisPluginManager) LoadPlugin(names ...string) {

	pluginManagerLock.Lock()

	var loadedList []*pluginInfo
	for _, name := range names {
		if p, err := pm.checkPluginFile(name); err != nil {
			getLogger().Warn(err)
		} else {
			loadedList = append(loadedList, p)
		}
	}

	pluginManagerLock.Unlock()

	if len(loadedList) == 0 {
		return
	}

	getLogger().Info("Handling Plugin Dependencies...")
	unsatisfiedCount := 0
	for _, p := range loadedList {
		pName := p.GetName()
		for _, d := range p.GetDependencies() {
			dp, ok := pm.pluginTable[d.Name]
			if ok {
				if !Compare(dp.GetVersion(), d.Version, d.Comparator) {
					unsatisfiedCount++
					getLogger().Warnf("Plugin Dependency not satisfied: [%s] -> [%s version%s%s]. But Found [%s]", pName, d.Name, d.Comparator, d.Version, dp)
				} else {
					getLogger().Debugf("Plugin Dependency satisfied: [%s] -> [%s version%s%s]. Found [%s]", pName, d.Name, d.Comparator, d.Version, dp)
				}
			} else {
				unsatisfiedCount++
				getLogger().Warnf("Plugin Dependency not satisfied: [%s] -> [%s version%s%s]. [%s] is not found", pName, d.Name, d.Comparator, d.Version, d.Name)
			}
			pluginManagerLock.Lock()
			pm.dependencyMap[d.Name] = append(pm.dependencyMap[d.Name], pName)
			pluginManagerLock.Unlock()
		}
	}
	getLogger().Infof("Reported %d unsatisfied dependencies", unsatisfiedCount)

	getLogger().Info("Loading Plugins...")
	//Topological sort
	num := 1
	times := 0
	for num != 0 {
		num = 0
		var tmpList []*pluginInfo
		for _, p := range loadedList {
			if !p.IsLoaded() {
				if p.dependenciesCount == 0 {
					if p.Load() {
						getLogger().Infof("Plugin [%s] successfully loaded.", p)
					} else {
						getLogger().Warnf("Plugin [%s] unsuccessfully loaded.", p)
					}
					if pm.checkPluginConfig(p) {
						if p.Enable() {
							getLogger().Infof("Plugin [%s] successfully enabled.", p)
						} else {
							getLogger().Warnf("Plugin [%s] unsuccessfully enabled.", p)
						}
						pluginManagerLock.Lock()
						pm.enabledPlugins = append(pm.enabledPlugins, p)
						pluginManagerLock.Unlock()
					} else {
						pluginManagerLock.Lock()
						pm.disabledPlugins = append(pm.disabledPlugins, p)
						pluginManagerLock.Unlock()
					}

					if pList, ok := pm.dependencyMap[p.GetName()]; ok {
						for _, dpName := range pList {
							dp := pm.pluginTable[dpName]
							dp.dependenciesCount--
							if !dp.IsLoaded() {
								tmpList = append(tmpList, dp)
							}
							getLogger().Debugf("Plugin [%s]'d unloaded dependencies-1, left %d", dp, dp.dependenciesCount)
						}
					}
					num++
				} else {
					getLogger().Debugf("Plugin [%s] has %d dependencies unloaded, ignored.", p, p.dependenciesCount)
					tmpList = append(tmpList, p)
				}
			}
		}
		getLogger().Debugf("Topological sort[%d] Finished. Loaded %d Plugins", times, num)
		loadedList = tmpList
		times++
	}

	pluginManagerLock.Lock()
	pm.unloadedPlugins = append(pm.unloadedPlugins, loadedList...)
	pluginManagerLock.Unlock()
}

func (pm *oasisPluginManager) LoadPlugins() {
	folder, err := ioutil.ReadDir(ServerConfig.GetString("PluginPath"))
	if err != nil {
		getLogger().Warnf("Fail to access to PluginPath. Details: %v", err)
		return
	}

	var pluginList []string
	for _, file := range folder {
		if !file.IsDir() {
			ok := strings.HasSuffix(file.Name(), ".so")
			if ok {
				pluginList = append(pluginList, file.Name())
			}
		}
	}

	pm.LoadPlugin(pluginList...)
}

func (pm *oasisPluginManager) Stop() {
	pList := pm.GetPlugins()
	for _, p := range pList {
		p.Disable()
	}
}
