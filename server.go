package main

import (
	"sync"
	"time"

	. "github.com/xaxys/oasis/api"
)

var serverLock sync.Mutex
var server *oasisServer

func GetServer() Server {
	return getServer()
}

func getServer() *oasisServer {
	if server == nil {
		serverLock.Lock()
		if server == nil {
			server = newServer()
		}
		serverLock.Unlock()
	}
	return server
}

func newServer() *oasisServer {
	time := time.Now()
	initServerConfig()
	initServerCommands()
	server := &oasisServer{
		createTime:     time,
		running:        true,
		PluginManager:  getPluginManager(),
		CommandManager: getCommandManager(),
		ConsolePrinter: getConsolePrinter(),
		TaskManager:    getTaskManager(),
	}
	server.wg.Add(1)
	return server
}

type oasisServer struct {
	wg sync.WaitGroup
	ConsolePrinter
	PluginManager
	CommandManager
	TaskManager
	createTime time.Time
	running    bool
}

func (server *oasisServer) GetCreateTime() time.Time {
	return server.createTime
}

// RunningTime 获取已运行时间
func (server *oasisServer) RunningTime() time.Duration {
	now := time.Now()
	return now.Sub(server.createTime)
}

func (server *oasisServer) Wait() {
	server.wg.Wait()
}

func (server *oasisServer) Stop() {
	getLogger().Info("Stopping the server...")
	getLogger().Debug("Stopping TaskManager...")
	getTaskManager().Stop()
	getLogger().Debug("Stopping PluginManager...")
	getPluginManager().Stop()
	getLogger().Debug("Stopping ConsolePrinter...")
	getConsolePrinter().Stop()
	server.wg.Done()
}
