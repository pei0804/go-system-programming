package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var imageSuffix = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
	".tiff": true,
	".eps":  true,
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf(`Find images
Usage:
	%s [path to find]`, os.Args[0])
		return
	}
	root := os.Args[1]
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		// ファイルを探すものなので、ディレクトリに来た時は何もしない
		if info.IsDir() {
			// Sphnixというツールでは、生成したHTMLやPDFを_buildというフォルダに出力するので
			// そのディレクトリより下のトラバースは無用なのでスキップしている
			if info.Name() == "_build" {
				return filepath.SkipDir
			}
			return nil
		}
		// 対象がディレクトリでなかった場合には、filepath.Extで拡張しを取り出して
		// そこから画像ファイルかどうかの参考にする。（JPGなどはjpgに統一される）
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if imageSuffix[ext] {
			// 現在のディレクトリとの位置関係を分かりやすくするためRelを使う
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}
			fmt.Printf("%s\n", rel)
		}
		return nil
	})
	if err != nil {
		fmt.Println(1, err)
	}
}
