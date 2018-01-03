package main

import "os"

func main() {
	os.Create("old_name.txt")
	os.Mkdir("newdir", 0644)

	os.Rename("old_name.txt", "new_name.txt")
	os.Rename("new_name.txt", "newdir/new_name.txt")
	// os.Rename("new_name.txt", "newdir/") ディレクトリ名だけではエラー
}
