package main

import (
	"fmt"
	"time"
)

func main() {
	tasks := []string{
		"cmake ...",
		"cmake . --build Release",
		"cpack",
	}
	for _, task := range tasks {
		go func(task string) {
			fmt.Println(task)
		}(task)
	}
	time.Sleep(time.Second)
}
