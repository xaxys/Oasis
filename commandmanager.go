package main

import (
	"fmt"
	"strings"
	"sync"

	. "github.com/xaxys/oasis/api"
)

type ConsoleCaller struct{}

var consoleCaller ConsoleCaller

const PredictionThreshold = 10

var commandManagerLock sync.Mutex
var commandManager *oasisCommandManager

type oasisCommandManager struct {
	commandMap *trie
	pluginMap  map[Plugin]map[string]*trieNode
}

func getCommandManager() *oasisCommandManager {
	if commandManager == nil {
		commandManagerLock.Lock()
		if commandManager == nil {
			commandManager = newCommandManager()
		}
		commandManagerLock.Unlock()
	}
	return commandManager
}

func newCommandManager() *oasisCommandManager {
	return &oasisCommandManager{
		commandMap: newTrie(),
		pluginMap:  map[Plugin]map[string]*trieNode{},
	}
}

func (cm *oasisCommandManager) ExcuteCommand(caller CommandCaller, sentence string) bool {
	if sentence == "" {
		return false
	}

	defer func() {
		if r := recover(); r != nil {
			getLogger().Errorf("Recovered from panic: %v", r)
		}
	}()

	var callerName string
	var p Plugin = nil
	if _, ok := caller.(ConsoleCaller); ok {
		callerName = "[Console]"
	} else if p, ok = caller.(Plugin); ok {
		callerName = fmt.Sprint(p)
	} else {
		return false
	}

	args := strings.Split(sentence, " ")
	command := strings.ToLower(args[0])
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	getLogger().Infof("%s issued command: %s", callerName, sentence)

	c, ok := cm.commandMap.Get(command)
	if ok {
		c.OnCommand(p, command, args)
		return true
	} else {
		getLogger().Infof("Command: %s is not found.", command)
		return false
	}
}

func (cm *oasisCommandManager) RegisterCommand(command string, p Plugin, ce CommandExcutor) bool {
	command = strings.ToLower(command)
	c := &CommandEntry{
		Command:        command,
		Plugin:         p,
		CommandExcutor: ce,
	}
	node, ok := cm.commandMap.Insert(c)
	if ok {
		if cm.pluginMap[p] == nil {
			cm.pluginMap[p] = map[string]*trieNode{}
		}
		cm.pluginMap[p][command] = node
		return true
	} else {
		return false
	}
}
func (cm *oasisCommandManager) UnregisterCommand(command string) bool {
	c, ok := cm.commandMap.Delete(command)
	if ok {
		delete(cm.pluginMap[c.Plugin], command)
		return true
	} else {
		return false
	}
}
func (cm *oasisCommandManager) UnregisterPluginCommand(p Plugin) {
	for _, v := range cm.pluginMap[p] {
		v.content = nil
		cm.commandMap.Update(v)
	}
	delete(cm.pluginMap, p)
}
func (cm *oasisCommandManager) GetPrediction(command string, force bool) (int, []CommandEntry) {
	num := cm.commandMap.Count(command)
	if num > PredictionThreshold && !force || num == 0 {
		return num, nil
	} else {
		var list []CommandEntry
		list = append(list, cm.commandMap.Contents(command)...)
		return num, list
	}
}
func (cm *oasisCommandManager) GetPluginCommands(p Plugin) []CommandEntry {
	var list []CommandEntry
	if cm.pluginMap[p] != nil {
		for _, v := range cm.pluginMap[p] {
			list = append(list, *(v.content))
		}
	}
	return list
}

// Trie

type trie struct {
	root *trieNode
}

type trieNode struct {
	count    int
	father   *trieNode
	children map[byte]*trieNode
	content  *CommandEntry
}

func newTrieNode(father *trieNode) *trieNode {
	return &trieNode{
		count:    0,
		father:   father,
		children: make(map[byte]*trieNode),
	}
}

func newTrie() *trie {
	return &trie{
		root: newTrieNode(nil),
	}
}

func (t *trie) get(key string) (*trieNode, bool) {
	b := []byte(key)
	x := t.root
	for _, c := range b {
		next, ok := x.children[c]
		if !ok {
			return nil, false
		}
		x = next
	}
	return x, true
}

func (t *trie) Get(key string) (*CommandEntry, bool) {
	x, ok := t.get(key)
	if ok && x.content != nil {
		return x.content, true
	} else {
		return nil, false
	}
}

func (t *trie) Insert(value *CommandEntry) (*trieNode, bool) {
	b := []byte(value.Command)
	x := t.root
	for _, c := range b {
		next, ok := x.children[c]
		if !ok {
			n := newTrieNode(x)
			x.children[c] = n
			next = n
		}
		x = next
	}
	if x.content != nil {
		return nil, false
	}
	x.content = value
	t.Update(x)
	return x, true
}

func (t *trie) Update(node *trieNode) {
	x := node
	for x != nil {
		x.count = 0
		for _, v := range x.children {
			x.count += v.count
		}
		if x.content != nil {
			x.count++
		}
		x = x.father
	}
}

func (t *trie) Delete(key string) (*CommandEntry, bool) {
	x, ok := t.get(key)
	if ok && x.content != nil {
		content := x.content
		x.content = nil
		t.Update(x)
		return content, true
	} else {
		return nil, false
	}

}

func (t *trie) Count(key string) int {
	x, ok := t.get(key)
	if !ok {
		return 0
	} else {
		return x.count
	}
}

func (t *trie) Contents(key string) []CommandEntry {
	x, ok := t.get(key)
	if !ok {
		return nil
	} else {
		return t.contents(x)
	}
}

func (t *trie) contents(node *trieNode) []CommandEntry {
	var list []CommandEntry
	for _, v := range node.children {
		list = append(list, t.contents(v)...)
	}
	list = append(list, *node.content)
	return list
}
