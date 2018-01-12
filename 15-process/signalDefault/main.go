package main

import (
	"os/signal"
	"syscall"
)

func main() {
	// 可変長引数で任意の数のシグナルを設定可能
	signal.Reset(syscall.SIGINT, syscall.SIGHUP)
}
