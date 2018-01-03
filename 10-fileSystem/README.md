# 10 File system

# ファイルシステムの基礎

ストレージに対するファイルシステムでは、まず、ストレージ全領域を512バイト〜4キロバイトの固定長のデータ（セクタ）の配列として扱います。そして、そこにファイルの中身だけを格納していくのではなく、ファイルの管理情報を格納する領域も用意しておきます。この管理情報は、現在のLinuxで、inodeと呼ばれてます。(WindowsではNTFS、MacOSではHFS+)

inodeに格納されるファイルの管理情報には、実際のファイルの中身の物理的な配置情報も含まれます。また、inodeにはユニークな識別番号がついています。その識別番号が分かればinodeにアクセスでき、inodeにアクセスできれば実際のファイルの配置場所がわかり、その中身にアクセス出来るという仕組みです。

私達が普段見ているディレクトリで整理されたファイル構造は、この仕掛を使って実現されています。ディレクトリというのは、実は配下に含まれるファイル名とそのinodeのインデックスの一覧表が格納されている特別なファイルです。そして、ルートディレクトリは必ず決まった番号の2番のinodeに格納されています。

# 複雑なファイルシステムとVFS

先の説明だと、ストレージ上にファイルシステムがひとつしかないように読めるかもしれませんが、実際のストレージはもっと複雑に入り組んでいます。ファイルシステムに他のファイルシステムをぶら下げたり（マウント）、仮想的なファイルシステムがあったりします。仮想的なファイルシステムは、物理的なストレージと対応するものではありません。仮想的なファイルシステムとして、例えば/proc以下の情報があります。/proc以下は、各プロセスの詳細情報がファイルとして見られるように、カーネルが動的に作り出したファイルシステムになっています。  
また最近では、ジャーナリングファイルシステムといって、書き込み中に瞬断が発生してもストレージ管理領域と実際の内容に不整合が起きにくくする仕組みが当たり前に量されています。  
最近ではDockerなどがよく使われますが、こうしたコンテナではファイルシステムを一部切り出し、特定のプロセスに対して、あたかもそれがファイルシステム全体であるかのように見せる仕掛けがあったりします。具体的にはchrootというシステムコールを使っています。  
これらの様々なファイルシステムはLinuxでは、VFS（Virtual File System）というAPIで統一的に扱えるようになっています。

# ファイル・ディレクトリを扱うGo言語の関数達

ファイルシステムはパフォーマンスと柔軟性を両立するために複雑になっていますが、アプリケーションからはVFSだけしか見えないのでシンプルに見えます。今回はosパッケージの基本的なところを見ていきます。

## ファイル作成・読み込み

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func open() {
	file, err := os.Create("textfile.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	io.WriteString(file, "New File content\n")
}

func read() {
	file, err := os.Open("textfile.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Println("Read file:")
	io.Copy(os.Stdout, file)
}

func append() {
	file, err := os.OpenFile("textfile.txt", os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	io.WriteString(file, "Appended content\n")
}

func main() {
	open()
	append()
	read()
}
```

## ディレクトリの作成

```go
package main

import "os"

func main() {
	os.Mkdir("setting", 0644)
	os.MkdirAll("setting/myapp/network", 0644)
}
```

## 削除

```go
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
```

Goのos.Removeは先にファイルを削除し、失敗したらディレクトリ削除を呼び出すので、対象関係なく削除可能です。

## リネーム、移動

```go
package main

import "os"

func main() {
	os.Create("old_name.txt")
	os.Mkdir("newdir", 0644)

	os.Rename("old_name.txt", "new_name.txt")
	os.Rename("new_name.txt", "newdir/new_name.txt")
	// os.Rename("new_name.txt", "newdir/") ディレクトリ名だけではエラー
}
```

POSIX系OSであっても、マウントされていて、元のデバイスが異なる場合には rename システムコールでの移動はできません。 下記のエラーメッセージは、macOSで tmpfs というオンメモリの一時ファイルシステム（昔の人はRAMディスクと呼んでいました）を作って os.Rename() を実行したときに返されるエラーです。

```go
err := os.Rename("sample.rst", "/tmp/sample.rst")
if err != nil {
    panic(err)
    // ここが実行され、コンソールに次のエラーが表示される
    // rename sample.rst /tmp/sample.rst: cross-device link
}
```

デバイスやドライブが異なる場合にはファイルを開いてコピーする必要があります。FreeBSDのmvコマンドも最初にrenamesシステムコールしてみて、失敗したら、以下のようなコードで移動させています。

```go
oldFile, err := os.Open("old_name.txt")
if err != nil {
    panic(err)
}
newFile, err := os.Create("/other_device/new_file.txt")
if err != nil {
    panic(err)
}
defer newFile.Close()
_, err = io.Copy(newFile, oldFile)
if err != nil {
    panic(err)
}
oldFile.Close()
os.Remove("old_name.txt")
```

## ファイルの属性取得

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("%s [exec file name]", os.Args[0])
		os.Exit(1)
	}
	info, err := os.Stat(os.Args[1])
	if err == os.ErrNotExist {
		fmt.Printf("file not found: %s\n", os.Args[1])
	} else if err != nil {
		panic(err)
	}
	fmt.Println("FileInfo")
	fmt.Printf("  ファイル名: %v\n", info.Name())
	fmt.Printf("  サイズ: %v\n", info.Size())
	fmt.Printf("  変更日時 %v\n", info.ModTime())
	fmt.Println("Mode()")
	fmt.Printf("  ディレクトリ？ %v\n", info.Mode().IsDir())
	fmt.Printf("  読み書き可能な通常ファイル？ %v\n", info.Mode().IsRegular())
	fmt.Printf("  Unixのファイルアクセス権限ビット %o\n", info.Mode().Perm())
	fmt.Printf("  モードのテキスト表現 %v\n", info.Mode().String())
}
```


```console
./fileStatus main.go
FileInfo
  ファイル名: main.go
  サイズ: 810
  変更日時 2018-01-03 21:43:47 +0900 JST
Mode()
  ディレクトリ？ false
  読み書き可能な通常ファイル？ true
  Unixのファイルアクセス権限ビット 644
  モードのテキスト表現 -rw-r--r--
```

## ファイルの存在チェック

```go
info, err := os.Stat(ファイルパス)
if err == os.ErrNotExist {
    // ファイルが存在しない
} else if err != nil {
    // それ以外のエラー
} else {
    // 正常ケース
}
```

上記の書き方はイディオムとしてよく使われます。この方法でしか存在チェック出来ないわけではないです。  
存在チェックそのもののシステムコールは提供されていません。  
仮に存在チェックを行ってファイルがあることを確認しても、その後のファイル操作までの間に他のプロセスやスレッドがファイルを消してしまうことも考えられます。 ファイル操作関数を直接使い、エラーを正しく扱うコードを書くことが推奨されています。

## OS固有のファイル属性を取得する

ファイル属性にはOS固有のものがあります。それらを取得するにはos.FileInfo.Sys()を使います。os.FileInfo.Sys()はドキュメントにも、interface{}を返すとしか書かれておらず、使い方に関する情報がいっさいない機能です。基本的には下記のようにOS固有の構造体にキャストして利用します。

```go
// Windows
internalStat := info.Sys().(syscall.Win32FileAttributeData)
 
// Windows以外
internalStat := info.Sys().(*syscall.Stat_t)
```

## ファイルの属性の設定

```go
// ファイルのモードを変更
os.Chmod("setting.txt", 0644)
 
// ファイルのオーナーを変更
os.Chown("setting.txt", os.Getuid(), os.Getgid())
 
// ファイルの最終アクセス日時と変更日時を変更
os.Chtimes("setting.txt", time.Now(), time.Now())
```

## リンク

```go
// ハードリンク
os.Link("oldfile.txt", "newfile.txt")
 
// シンボリックリンク
os.Symlink("oldfile.txt", "newfile-symlink.txt")
 
// シンボリックリンクのリンク先を取得
link, err := os.ReadLink("newfile-sysmlink.txt")
```

## ディレクトリ情報の取得

```go
package main
 
import (
    "fmt"
    "os"
)
 
func main() {
    dir, err := os.Open("/")
    if err != nil {
        panic(err)
    }
    fileInfos, err := dir.Readdir(-1)
    if err != nil {
        panic(err)
    }
    for _, fileInfo := range fileInfos {
        if fileInfo.IsDir() {
            fmt.Printf("[Dir]  %s\n", fileInfo.Name())
        } else {
            fmt.Printf("[File] %s\n", fileInfo.Name())
        }
    }
}
```

```console
[File] .dbfseventsd
[Dir]  .DocumentRevisions-V100
[File] .DS_Store
[File] .file
[Dir]  .fseventsd
[Dir]  .MobileBackups
[Dir]  .PKInstallSandboxManager
[Dir]  .Spotlight-V100
[Dir]  .TemporaryItems
[Dir]  .Trashes
[Dir]  .vol
[Dir]  Applications
[Dir]  bin
[Dir]  cores
[Dir]  dev
[Dir]  Developer
[File] etc
[Dir]  home
[File] installer.failurerequests
[Dir]  Library
[Dir]  macOS Install Data
[Dir]  net
[Dir]  Network
[Dir]  opt
[Dir]  private
[Dir]  sbin
[Dir]  System
[File] tmp
[Dir]  Users
[Dir]  usr
[File] var
[Dir]  Volumes

[Process exited 0]
```

ディレクトリ一覧の取得はosパッケージ直下の関数として提供されていません。  
ディレクトリをos.Open()で開き、os.File()のメソッドを使って、ディレクトリ内のファイル一覧を取得しています。  
Readdir()メソッドはos.FileInfoの配列を返します。 ファイル名しか必要がないときはReaddirnames()メソッドを使えます。 このメソッドは文字列の配列を返します。  
Readdir()とReaddirnames()は数値を引数に取ります。正の整数を与えると、最大でその個数の要素だけを返します。 0以下の数値を渡すと、ディレクトリ内の全要素を返します。

## OS内部におけるファイル操作の高速化

CPUにとってディスク読み書きはとても遅い処理であり、なるべく最後までやりたくないタスクです。そこで、Linuxでは、VFSの内部に設けられているバッファを利用することで、ディスクに対する操作をなるべく回避しています。Linuxでファイルを読み書きする場合には、バッファにデータが格納されます。そのため、ファイルへデータを書き込むとバッファに蓄えられた時点でアプリケーション処理が返ります。ファイルからデータを読み込む時も一旦バッファに蓄えられますし、既にバッファに乗っており、そのファイルに対する書き込みが行われていないなら、アクセスしません。したがって、アプリケーションによる入出力は、実際にはLinuxが用意したバッファとの入出力になります。バッファと実際のストレージとの同期はアプリケーションの知らないとこで非同期に行われています。  

このストレージ書き込みを確実にしたい場合は、os.FileのSyncメソッドを呼びます。  
```go
file.Sync()
```
