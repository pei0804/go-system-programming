package main

import (
	"fmt"
	"sync"
)

func main() {
	// Poolを作成。Newで新規作成時のコードを実装
	var count int
	pool := sync.Pool{
		New: func() interface{} {
			count++
			return fmt.Sprintf("created: %d", count)
		},
	}

	// 追加した要素から受け取れる
	// プールが空だと新規作成
	pool.Put("manualy added: 1")
	pool.Put("manualy added: 2")
	fmt.Println(pool.Get())
	fmt.Println(pool.Get())

	// 新規作成時のコードが実行
	// New: func() interface{} {
	// 	count++
	// 	return fmt.Sprintf("created: %d", count)
	// },
	fmt.Println(pool.Get())
	fmt.Println(pool.Get())
}
