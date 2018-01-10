package main

import (
	"fmt"
	"path/filepath"
	"time"

	filelock "github.com/zbiljic/go-filelock"
)

func main() {
	path, err := filepath.Abs("main.go")
	if err != nil {
		panic(err)
	}
	l, err := filelock.New(path)
	if err != nil {
		panic(err)
	}
	fmt.Println("try  locking...")
	l.Lock()
	fmt.Println("locked!")
	time.Sleep(3 * time.Second)
	l.Unlock()
	fmt.Println("unlock")
}
