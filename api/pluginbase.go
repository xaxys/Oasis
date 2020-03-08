package OasisAPI

type PluginBase struct {
	PluginDescription
	API interface{}
	PluginProperty
}

func (pb *PluginBase) OnLoad() bool {
	return true
}

func (pb *PluginBase) OnEnable() bool {
	return true
}

func (pb *PluginBase) OnDisable() bool {
	return true
}

func (pb *PluginBase) EntryPoint(pp PluginProperty) {
	pb.PluginProperty = pp
}

func (pb *PluginBase) GetDescription() PluginDescription {
	return pb.PluginDescription
}

func (pb *PluginBase) GetPluginAPI() interface{} {
	return pb.API
}
