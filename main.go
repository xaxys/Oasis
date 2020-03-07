package main

func main() {
	startReader()
	myserver := getServer()
	myserver.LoadPlugins()
	myserver.Wait()
}
