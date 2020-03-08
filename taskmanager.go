package main

import (
	"sync"

	"github.com/robfig/cron/v3"
	. "github.com/xaxys/oasis/api"
)

var taskManagerLock sync.Mutex
var taskManager *oasisTaskManager

type oasisTaskManager struct {
	taskMap   *cron.Cron
	pluginMap map[Plugin][]int
}

func getTaskManager() *oasisTaskManager {
	if taskManager == nil {
		taskManagerLock.Lock()
		if taskManager == nil {
			taskManager = newTaskManager()
			taskManager.taskMap.Start()
		}
		taskManagerLock.Unlock()
	}
	return taskManager
}

func newTaskManager() *oasisTaskManager {
	return &oasisTaskManager{
		taskMap:   cron.New(cron.WithSeconds()),
		pluginMap: map[Plugin][]int{},
	}
}

func (tm *oasisTaskManager) RegisterTask(p Plugin, time string, r Runnable) (int, bool) {
	eid, err := tm.taskMap.AddJob(time, r)
	if err != nil {
		getLogger().Warnf("Failed to register task for %s. Details: %v", p, err)
		return 0, false
	}
	id := int(eid)
	taskManagerLock.Lock()
	tm.pluginMap[p] = append(tm.pluginMap[p], id)
	taskManagerLock.Unlock()
	return id, true
}

func (tm *oasisTaskManager) UnregisterTask(id int) {
	tm.taskMap.Remove(cron.EntryID(id))
}

func (tm *oasisTaskManager) UnregisterPluginTask(p Plugin) {
	taskManagerLock.Lock()
	for _, id := range tm.pluginMap[p] {
		tm.taskMap.Remove(cron.EntryID(id))
	}
	taskManagerLock.Unlock()
}

func (tm *oasisTaskManager) Stop() {
	tm.taskMap.Stop()
}
