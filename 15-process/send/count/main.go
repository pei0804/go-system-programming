package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Run")
	for {
		time.Sleep(time.Second * 1)
		fmt.Println("Loop")
	}
}
