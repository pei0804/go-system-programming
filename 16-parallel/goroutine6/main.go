package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("start sub()")
	// 終了を受け取るためのチャネル
	done := make(chan bool)
	go func() {
		fmt.Println("sub()is finished")
		time.Sleep(time.Second)
		// 終了を通知
		done <- true
	}()
	<-done
	fmt.Println("all tasks are finished")
}
