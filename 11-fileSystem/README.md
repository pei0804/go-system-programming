# 11 File System

# path/filepathパッケージ

## ディレクトリのぱすとファイル名と連結する

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Printf("Temp File Path: %s\n", filepath.Join(os.TempDir(), "temp.txt"))
}
```

path/filepathをつかえば、動作環境のファイルシステムで2種類のパス表記のどちらが使われていても、その違いを吸収して各プラットフォームに適した結果が得られます。  
pathパッケージの方は、常に/を使います。なので、主にURLを操作する時に使います。

## パスを分割する

パスからファイル名とその親ディレクトリに分割するfilepath.Split()もよく使います。

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir, name := filepath.Split(os.Getenv("GOPATH"))
	fmt.Printf("Dir: %s, Name: %s\n", dir, name)
}
```

```console
Dir: /Users/jumpei/, Name: go
```

ファイルパスの全要素を配列にしたいこともあるでしょう。 "/a/b/c.txt"を[a b c.txt]にするには、次のようにセパレータ文字を取得してきて分割するのが簡単です。

```go
fragments := strings.Split(path, string(filepath.Separator))
```

パス名を分解する関数には他にも次の4種類があります。

| 関数                  | 説明                                               | "folder1/folder2/example.txt"を入力した結果 | "C:\folder1\folder2 \example.txt" を入力した結果 |
|:----------------------|:---------------------------------------------------|:--------------------------------------------|:-------------------------------------------------|
| filepath.Base()       | パスの最後の要素を返す                             | "example.txt"                               | "example.txt"                                    |
| filepath.Dir()        | パスのディレクトリ部を返す                         | "/folder1/folder2"                          | "C:\folder1\folder2"                             |
| filepath.Ext()        | ファイルの拡張子を返す                             | ".txt"                                      | ".txt"                                           |
| filepath.VolumeName() | ファイルのドライブ名を返す（Windows用）            | ""                                          | "C:"                                             |

## 複数のパスからなる文字列を分解する


filepath.SplitList()という名前の関数もあります。名前だけ見るとパスの分割に使えそうですが、これは別の用途で、環境変数などにある複数のパスを一つのテキストにまとめたものを分解するのに使います。

```console
❯ echo $PATH
/Users/jumpei/.anyenv/envs/pyenv/shims:/Users/jumpei/.anyenv/envs/pyenv/bin:/Users/jumpei/.anyenv/envs/ndenv/bin:/Users/jumpei/.anyenv/envs/goenv/shims:/Users/jumpei/.anyenv/env...
```

例えば、次のコードは、Unix系OSにあるwhichコマンドをGoで実装したものです。PATH環境変数のパス一覧を取得してきて、それをfilepath.SplitList()で個々のパスに分割します。その後、各パスの下に最初の引数で指定された実行ファイルがあるかどうかチェックします。

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("%s [exec file name]", os.Args[0])
		os.Exit(1)
	}
	for _, path := range filepath.SplitList(os.Getenv("PATH")) {
		// それぞれのパスに検索したいファイル名をパスの最後に使いする
		// /Users/jumpei/.anyenv/envs/pyenv/bin ->/Users/jumpei/.anyenv/envs/pyenv/bin/go
		execpath := filepath.Join(path, os.Args[1])
		_, err := os.Stat(execpath)
		if !os.IsNotExist(err) {
			fmt.Println(execpath)
			return
		}
	}
	os.Exit(1)
}
```

```console
./splitList go
/Users/jumpei/.anyenv/envs/goenv/shims/go
```

## パスのクリーン化

```go
package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	// 無駄なパス表記を消す
	fmt.Println(filepath.Clean("./path/filepath/../path.go"))
	// path/path.go

	// 絶対パス
	abspath, _ := filepath.Abs("path/filepath/path_unix.go")
	fmt.Println(abspath)
	// /usr/local/go/src/path/ilepath/path_unix.go

	// 基準のパスから相対パスを出す
	relpath, _ := filepath.Rel("/usr/local/go/src", "/usr/local/go/src/path/filepath/path.go")
	fmt.Println(relpath)
	// path/filepath/path.go
}
```

シンボリックリンクを展開した上で、Clean()をかけた状態のパスを返してくれるfilepath.EvalSymlinks()という関数もあります。

## 環境変数などの展開

```go
package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func main() {
	path := os.ExpandEnv("${GOPATH}/src/github.com/pei0804/go-system-programming")
	fmt.Println(path)
	// /Users/jumpei/go/src/github.com/pei0804/go-system-programming

	// ~はOSではなくシェルが提供しているものなので、OSが解釈出来る形にする
	my, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Println(my.HomeDir)
	//  /Users/jumpei

	fmt.Println(pathClean("~/bin/goenv"))
	// /Users/jumpei/bin/goenv
}

// ~ を置換、環境変数の展開、パスのクリーン化
func pathClean(path string) string {
	if len(path) > 1 && path[0:2] == "~/" {
		my, err := user.Current()
		if err != nil {
			panic(err)
		}
		path = my.HomeDir + path[1:]
	}
	path = os.ExpandEnv(path)
	return filepath.Clean(path)
}
```

## ファイル名のパターンにマッチするファイルの抽出

filepath.Match()  
1文字の任意の文字にマッチするワイルドカード（?）と、ゼロ文字以上の文字にマッチするワイルドカード（*）が使えるほか、 マッチする文字範囲（[0-9]）や、マッチしない文字範囲（[^a]）も指定できます。  

Glob()  
一致したファイル名の一覧を取得する  

```go
package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	fmt.Println(filepath.Match("image-*.png", "image-100.png"))
	// true <nil>

	files, err := filepath.Glob("./*.png")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
	// [golang.png image-100.png]
}
```

## ディレクトリのトラバース

```go
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
```

```console
walk ./
dir/image-100.png
golang.png
```

今回はパス表記が冗長にならないようにする処理を入れていますが、そのファイルを読み込む場合などには、そのままのパス情報の方が良いでしょう。  
コールバック関数がfilepath.SkipDir以外のエラーを返すと、即座にトラバースが終了されます。

## ファイルの変更監視

- 監視したファイルをOS側に通知しておいて、変更があったら教えてもらう（パッシブな方式）
- タイマーなどで定期的にフォルダ走査し、os.Stat()などを使って変更を探しに行く（アクティブな方式）

Go言語の標準ライブラリでは、ファイルの監視を行う機能は提供されていません。ゼロから実装するのであれば、コードが短くて分かりやすいアクティブな方式です。しかし、アクティブ方式は高コストです。

パッシブな方式については、ファイルの変更検知が各OSでシステムコールやAPIとして提供されています。しかし、環境ごとのコードは差が大きくなります。なので、サードパーティであるgopkg.in/fsnotify.v1を利用したパッシブな方式の例を説明します。

```go
package main

import (
	"log"

	fsnotify "gopkg.in/fsnotify.v1"
)

func main() {
	counter := 0
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("created file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("removed file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("renamed file:", event.Name)
					counter++
				} else if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					log.Println("chmod file:", event.Name)
					counter++
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
			if counter > 3 {
				done <- true
			}
		}
	}()

	err = watcher.Add(".")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
```
