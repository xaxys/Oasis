# Oasis

Oasis is a go plugin server which provides basic plugin operation. You can easily create a plugin with its API and register it to the server  to avoid write config operation or timely task operation etc.

# Plugins

You can find some useful plugin in this repository https://github.com/xaxys/Oasis_Plugins

# Update

* 2020/03/07 Initialized the repositories and released v 0.1.0

# Start

To build a oasis plugin, you need import `github.com/xaxys/oasis/api`. Recommand use `import . "github.com/xaxys/oasis/api"`. And create a struct containing `PluginBase` or rather `OasisAPI.PluginBase` for your plugin. And you shoud claim a variable named `PLUGIN` containing your struct which implements `UserPlugin`for server to load.

Here is a examle:



 ```go
package main

import (
	"fmt"

	. "github.com/xaxys/oasis/api"
)

var PLUGIN UserPlugin = &WhateverPlugin{
	PluginBase: PluginBase{
		PluginDescription: PluginDescription{
			Name:    "whatever",
			Author:  "xaxys",
			Version: "0.1.2",
			DefaultConfigFields: map[string]interface{}{
				"name": "whatever",
			},
			Dependencies: []PluginDependency{
				PluginDependency{
					Name:       "whatever3",
					Version:    "anyversion",
					Comparator: ANY,
				},
			},
		},
	},
}

type WhateverPlugin struct {
	PluginBase
}

func (p *WhateverPlugin) OnLoad() bool {
	p.GetLogger().Infof("Exercuting OnLoad %s", p.Name)
	plist := p.GetServer().GetPlugins()
	for _, p := range plist {
		fmt.Println(p)
	}
	return true
}

// Custom your OnEnable
func (p *WhateverPlugin) OnEnable() bool {
	return true
}

// If you don't implement your OnDisable
// it Will use default Ondisable() in PluginBase
// and return true directly.
func (p *WhateverPlugin) OnDisable() bool {
	return true
}
 ```

