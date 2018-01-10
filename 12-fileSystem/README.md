# 12 File system

# ファイルのロック syscall.Flock()

ファイルロックの最も単順な方法は、リソースが使用中であることをファイル（ロックファイル）によって示す方法です。昔はたいていこの方法が取られていました。

ロックファイルは簡単ですが、確実性という面では劣ります。より確実なのはファイルロックのためのシステムコールを利用して、既に存在するファイルに対してロックをかける方法です。

Go言語では、POSIX系OSの場合、このロックのためのsyscall.Flock()というシステムコールが利用出来ます。ただし、syscall.Flock()によるロック状態は、通常のファイル入出力のためのシステムコールによって確認されます。ロックをまじめに確認しないプロセスが一つでもあると、自由に上書きされてしまう可能生があります。このような強制力のないロックのことをアドバイザリーロック（勧告ロック）と呼びます。

このsyscall.Flock()によるロックは、Windowsでは利用出来ません。WindowsではファイルロックにLockFileEx()という関数を使います。こちらは、syscall.Flock()とは違い。他のプロセスもブロックする強制ロックです。

## syscall.Flock()によるPOSIX系OSでのファイルロック

Go言語本体コード

```go
// +build darwin dragonfly freebsd linux netbsd openbsd
 
package main
 
import (
  "sync"
  "syscall"
)
 
type FileLock struct {
  l  sync.Mutex
  fd int
}
 
func NewFileLock(filename string) *FileLock {
  if filename == "" {
    panic("filename needed")
  }
  fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_RDONLY, 0750)
  if err != nil {
    panic(err)
  }
  return &FileLock{fd: fd}
}
 
func (m *FileLock) Lock() {
  m.l.Lock()
  if err := syscall.Flock(m.fd, syscall.LOCK_EX); err != nil {
    panic(err)
  }
}
 
func (m *FileLock) Unlock() {
  if err := syscall.Flock(m.fd, syscall.LOCK_UN); err != nil {
    panic(err)
  }
  m.l.Unlock()
}
```

syscall.Flock()は引数を２つ取ります。1つはロックしたい対象のファイルディスクリプタです。もう1つは、ロックのモードを指示するフラグです。

|フラグ |説明               |
|:------|:------------------|
|LOCK_SH|共有ロック。他のプロセスからも共有ロックなら可能だが、排他ロックは同時に行えない。|
|LOCK_EX|排他ロック。他のプロセスからも共有ロックも、排他ロックも行えない。|
|LOCK_UN|ロック解除。ファイルをクローズしても解除になる。|
|LOCK_NB|ノンブロッキングモード|

syscall.Flock()によるロックでは、既にロックされているファイルに対してロックをかけようとすると、最初のロック外れるまでずっと待たされます。そのため、定期的に何度もアクセスしてロック出来るかトライするということが出来ません。これを可能にするのがノンブロッキングモードです。

```go
package main

import (
	"fmt"
	"path/filepath"
	"time"

	filelock "github.com/zbiljic/go-filelock"
)

func main() {
	path, err := filepath.Abs("main.go")
	if err != nil {
		panic(err)
	}
	l, err := filelock.New(path)
	if err != nil {
		panic(err)
	}
	fmt.Println("try  locking...")
	l.Lock()
	fmt.Println("locked!")
	time.Sleep(3 * time.Second)
	l.Unlock()
	fmt.Println("unlock")
}
```

### マルチプラットフォームを実現するための手段

Go言語には、マルチプラットフォームを実現する方法が大きく分けて2つあります。

１つ目は、Build Constrainstsと呼ばれるもので、ビルド対象にプラットフォームを指定する方法です。具体的には、コード先頭に // +buildに続けてビルド対象のプラットフォームを列挙したり、ファイル名に_windows.goのようなサフィックスを付けます。

上記のコード例では、POSIX用にはｍ// +buildを指定しています。対象をWindowsに限定する場合は、ファイル名にサフィックスをつけるのが一般的です。（そのため上記のWindows用のコードには// +buildを指定していません）

もう一つは、runtime.GOOS定数を使って実行時に処理を分岐する手法です。ただし、この方法は今回のようにAPI自体がプラットフォームによって異なる場合には、リンクエラーが発生してしまいます。そのため、上記コードでは前者の手段を行なっています。

# ファイルのメモリへのマッピング syscall.Mmap

```go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	mmap "github.com/edsrzf/mmap-go"
)

func main() {
	// テストデータを書き込み
	var testData = []byte("0123456789ABCDEF")
	var testPath = filepath.Join(os.TempDir(), "testdata")
	err := ioutil.WriteFile(testPath, testData, 0644)
	if err != nil {
		panic(err)
	}

	// メモリにマッピング
	// mは[]byteのエイリアスなので添字アクセス可能
	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer m.Unmap()

	// メモリ上のデータを修正して書き込み
	m[9] = 'X'
	m.Flush()

	// 読み込んで見る
	fileData, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("original: %s\n", testData)
	fmt.Printf("mmap: %s\n", m)
	fmt.Printf("file: %s\n", fileData)
}
```

```console
original: 0123456789ABCDEF
mmap: 012345678XABCDEF
file: 012345678XABCDEF
```

- mmap.Map(): 指定したファイルの内容をメモリ上に展開
- mmap.Umap(): メモリ上に展開された内容を削除して閉じる
- mmap.Flush(): 書きかけの内容をファイルに保存する
- mmap.Lock(): 開いているメモリ領域をロックする
- mmap.Unlock(): メモリ領域をアンロックする

上記の例では、まずファイルを読み書きフラグ付きで、os.OpenFileによってオープンし、その結果をmmap.Map()関数に渡して読み書きモードでメモリ上に展開し、そこで内容を書き換え（数字の0のところをXに書き換え）、Flush()メソッドを使ってそれをファイルに書き戻しています。最終的なファイルが書き換わるので、最後にioutil.ReadAll()で読み込んだ内容も元データと異なります。

mmap.Map()関数の2つめの引数には、メモリ領域に対して許可する設定をします。許可した走査に応じて次のような値を指定します。

- mmap.RDONLY: 読み込み専用
- mmap.RDWR: 読み書き可能
- mmap.COPY: コピー・オン・ライト
- mmap.EXEC: 実行可能にする

コピー・オン・ライト時は、通常通りメモリ領域にファイルをマッピングしますが、メモリ領域で書き換え発生するとその領域がまるごとコピーされます。そのため、元のファイルには変更が反映されません。不思議な挙動ですが、書き換えが発生するまでは複数のバリエーションの状態を保持する必要がないので、メモリを節約出来ます。

### mmapの実行速度

通常のFile.Read()メソッドのシステムコールと比べて、mmapの方が実行速度が出そうですが、実はケース・バイ・ケースです。

前から順番に読み込んで逐次処理するのであれば、通常の処理のFile.Read()でも十分に速いでしょう。データベースのファイルなど。全部を一度にメモリ上に読み込んで処理する必要があって、その上でランダムアクセスが必要なケースでは、mmapの方が有利なこともあり、使いやすいと思います。しかし、一度に多くのメモリを確保しなければいけないため、ファイルサイズが大きくなるとI/O待ちが長くなる可能生があります。もちろん、コピー・オン・ライト昨日を使う場合や、確保したメモリの領域にアセンブリ命令が格納されていて、実行を許可する場合にはmmap一択です。

mmapは、メモリ共有の仕組みとしても使えたりします。

[mmapの速度](http://d.hatena.ne.jp/kazuhooku/20131010/1381403041)

## 同期・非同期/ブロッキング・ノンブロッキング

ファイルもネットワークもCPU内部の処理の比べると劇的に遅いタスクです。この重い処理に引きづられて全体が遅くならないようにする仕組みが必要になります。

そのための仕組みを、OSのシステムコールにおいて整備するためのモデルとなるのが、同期、非同期。ブロッキングとノンブロッキングという分類です。
[同期、非同期、ブロッキング、ノンブロッキング](http://blog.matsumoto-r.jp/?p=2030)

- 同期: OSに仕事を投げて、入出力の準備が出来たらアプリケーションに処理が帰ってくる
- 非同期: OSに仕事を投げて、入出力が完了したら通知してもらう
- ブロッキング: お願いした仕事が完了するまで待つ
- ノンブロッキング: お願いした仕事が完了するのを待たない

### 同期・ブロッキング

読み込み・書き込み処理が完了するまで何もせずに待ちます。重い処理があると、そこで全ての処理が止まってしまいます。実行時のパフォーマンスは良くないですが、コードは最もシンプルでわかりやすいです。

### 同期・ノンブロッキング

いわゆるポーリングです。ファイルオープン時にノンブロッキングのフラグを付与することで実現できます。APIを呼ぶと「まだ完了していないか」どうかのフラグと現在得られるデータが得られます。クライアントは、完了が返ってくるまで何度もAPIを呼びます。

### 非同期・ブロッキング

I/O多重化（I/Oマルチプレクサー）とも呼ばれます。準備が完了したものがあれば通知してもらうというイベント駆動モデルです。そのための通知には、select属と呼ばれるシステムコールを使います。

### 非同期・ノンブロッキング

メインプロセスのスレッドとは完全に別のスレッドでタスクを行い、完了したらその通知だけを受け取る処理です。APIとしては、POSIXのAPIで定義されている非同期I/O(aio_*)インターフェイスが有名です。
しかし、全然使われていないらしい。しかも、場合によっては遅くなる。

## Go言語で様々なI/Oモデルを扱う手法

並行処理が得意なGo言語ですが、ベースとなるファイルI/OやネットワークI/Oは、シンプルな同期・ブロッキングインターフェースになっています。同期・ブロッキングのAPIを並行処理するだけでも重い処理で全体が止まることがなくなるため、効率が改善されます。

- goroutineをたくさん実行し、それぞれ同期・ブロッキングI/Oを担当させると、非同期・ノンブロッキングとなります。
- goroutineで並行化させたI/Oの入出力でチャネルを使い、他のgoroutineとのやりとりする箇所のみ同期が行えます。
- このチャネルにバッファがあれば書き込み側もノンブロッキングとなります。
- select構文にdefault節があると、読み込みノンブロッキングで行えるようになり、aio化が行えます。

## select属のシステムコールによるI/O多重化

select属はC10K問題と呼ばれる万の単位の入出力を効率よく扱うための手法として有効です。ネットワークにについては、既にselect族のシステムコールがGo言語のランタイム内部に組み込まれており、サーバーを実装した時に効率よくスケジューラが働くようになっています。他にはたくさんのファイルの変更監視などにも使えます。

```go
```

