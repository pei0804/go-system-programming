package main

import "os"

func main() {
	os.Create("server.log")
	os.Mkdir("workdir", 0644)

	// 先頭100バイトで切る
	os.Truncate("server.log", 100)
	// Truncateメソッドを利用する場合
	file, _ := os.Open("server.log")
	file.Truncate(100)
	// システムコールではunlink()
	os.Remove("server.log")
	// システムコールではrmdir()
	os.RemoveAll("workdir")
}
