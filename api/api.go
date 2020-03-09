package OasisAPI

import (
	"io"
	"time"
)

type Server interface {
	ConsolePrinter
	PluginManager
	CommandManager
	TaskManager
	GetCreateTime() time.Time
	RunningTime() time.Duration
}

type ConsolePrinter interface {
	RegisterFormatter(Formatter)
	ClearFormatter()
}

type Formatter interface {
	Format(b string) string
}

type CommandEntry struct {
	Command string
	Plugin
	CommandExcutor
}

type CommandCaller interface{}

type CommandManager interface {
	// ExcuteCommand returns false if command doesn't exist.
	// caller is ConsoleCaller if called from console.
	// so caller shoudn't be nil
	// if caller isn't a ConsoleCaller or a Plugin
	// the command will be ignored and print a warning
	ExcuteCommand(caller CommandCaller, sentence string) bool
	// RegisterCommand returns false if command has already been registered.
	RegisterCommand(string, Plugin, CommandExcutor) bool
	// UnregisterCommand returns false if command doesn't exist.
	UnregisterCommand(string) bool
	UnregisterPluginCommand(Plugin)
	// GetPredict returns empty []CommandEntry if result > PredictionThreshold
	// and force=false. It returns all prediction anyway if force=true
	GetPrediction(command string, force bool) (int, []CommandEntry)
	GetPluginCommands(Plugin) []CommandEntry
}

type CommandExcutor interface {
	// OnCommand Plugin is nil if called from console.
	OnCommand(Plugin, string, []string)
}

type Runnable interface {
	Run()
}

type TaskManager interface {
	// RegisterTask string defines when to run the task
	// returns a taskID if succeeded
	// Seconds       | 0-59            | * / , -
	// Minutes       | 0-59            | * / , -
	// Hours         | 0-23            | * / , -
	// Day of month  | 1-31            | * / , - ?
	// Month         | 1-12 or JAN-DEC | * / , -
	// Day of week   | 0-6 or SUN-SAT  | * / , - ?
	RegisterTask(Plugin, string, Runnable) (int, bool)
	UnregisterPluginTask(Plugin)
	UnregisterTask(int)
}

type PluginManager interface {
	GetPlugin(string) Plugin
	//GetPlugins equals GetEnabledPlugins
	GetPlugins() []Plugin
	GetEnabledPlugins() []Plugin
	GetDisabledPlugins() []Plugin
	GetAllPlugins() []Plugin
	LoadPlugin(...string)
	LoadPlugins()
}

type Plugin interface {
	Load() bool
	Enable() bool
	Disable() bool
	GetName() string
	GetVersion() string
	GetDescription() string
	GetAuthor() string
	IsEnabled() bool
	IsLoaded() bool
	GetDetailedInfo() string
	GetPluginAPI() interface{}
	GetDependencies() []PluginDependency
	GetSoftDependencies() []PluginDependency
}

type UserPlugin interface {
	OnLoad() bool
	OnEnable() bool
	OnDisable() bool
	GetDescription() PluginDescription
	EntryPoint(PluginProperty)
	GetPluginAPI() interface{}
}

type PluginProperty interface {
	GetPlugin() Plugin
	GetServer() Server
	GetLogger() Logger
	GetConfig() Configuration
	GetFolder() string
}

type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Debugw(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Infow(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Warnw(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Errorw(string, ...interface{})
}

type Configuration interface {
	AllSettings() map[string]interface{}
	AllKeys() []string
	IsSet(key string) bool
	Get(key string) interface{}
	SubConfig(key string) Configuration
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetSizeInBytes(key string) uint
	RegisterAlias(alias string, key string)
	InConfig(key string) bool
	SetDefault(key string, value interface{})
	Set(key string, value interface{})
	SetAndWrite(key string, value interface{}) error
	SetConfigType(in string)
	ReadInConfig() error
	MergeInConfig() error
	ReadConfig(in io.Reader) error
	MergeConfig(in io.Reader) error
	MergeConfigMap(cfg map[string]interface{}) error
	WriteConfig() error
	SafeWriteConfig() error
	Unmarshal(interface{}) error
	UnmarshalKey(string, interface{}) error
}

type PluginDescription struct {
	Name                string
	Version             string
	Description         string
	Author              string
	Dependencies        []PluginDependency
	SoftDependencies    []PluginDependency
	DefaultConfigFields map[string]interface{}
}

type PluginDependency struct {
	Name       string
	Version    string
	Comparator COMPARATOR
}

type COMPARATOR string

const (
	GREATER       COMPARATOR = ">"
	GREATER_EQUAL COMPARATOR = ">="
	LESS          COMPARATOR = "<"
	LESS_EQUAL    COMPARATOR = "<="
	EQUAL         COMPARATOR = "="
	UNEQUAL       COMPARATOR = "!="
	ANY           COMPARATOR = "*"
)
