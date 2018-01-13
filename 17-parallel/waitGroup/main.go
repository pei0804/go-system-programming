package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	// ジョブ数をあらかじめ登録
	wg.Add(2)

	go func() {
		// 非同期で仕事をする
		fmt.Println("仕事1")
		// Doneで完了を通知
		wg.Done()
	}()

	go func() {
		// 非同期で仕事をする
		fmt.Println("仕事2")
		// Doneで完了を通知
		wg.Done()
	}()

	// 全ての処理が終わるまで待つ
	wg.Wait()
	fmt.Println("終了")
}
