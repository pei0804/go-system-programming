package main

import (
	"fmt"
	"time"
)

func sub1(c int) {
	fmt.Println("share by args", c*c)
}

func main() {
	// 引数渡し
	go sub1(10)

	// クロージャのキャプチャ渡し
	c := 20
	go func() {
		fmt.Println("share by args", c*c)
	}()
	time.Sleep(time.Second)
}
