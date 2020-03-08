package main

import (
	"bufio"
	"fmt"
	"os"
)

func startReader() {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")
		fmt.Scanln()
		for scanner.Scan() {
			line := scanner.Text()
			GetServer().ExcuteCommand(consoleCaller, line)
			fmt.Print("> ")
		}
	}()
}
