package main

import (
	"fmt"
	"sync"
)

func ini() {
	fmt.Println("初期処理")
}

var once sync.Once

func main() {
	// 3回呼んでも1度しか呼ばれない
	once.Do(ini)
	once.Do(ini)
	once.Do(ini)
}
