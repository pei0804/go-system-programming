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
		go func() {
			// goroutineが起動するときにはループが回りきって
			// 全部のtaskが最後のタスクになってしまう
			// goroutineの起動はループに比べると遅いため
			fmt.Println(task)
		}()
	}
	time.Sleep(time.Second)
}
