package main

import (
	"fmt"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Ignore Ctrl + C for 3 second")
	signal.Ignore(syscall.SIGINT, syscall.SIGHUP)
	time.Sleep(time.Second * 3)
}
