# 13 Process

オペレーティングシステムが実行ファイルを読み込んで実行するには、そのためのリソースを用意しなければなりません。そのようなリソースをまとめたプログラムの実行単位がプロセスです。

# プロセスに含まれるもの（Go言語視点）

- プロセスID
- プロセスグループID, セッショングループID
- ユーザーID, グループID
- 実効ユーザーID, 実効グループID
- カレントディレクトリ
- ファイルディスクリプタ

## プロセスID

プロセスには必ずプロセスごとにユニークな識別子があります。それがプロセスIDです。Go言語では、os.Getpid()を使って現在のプロセスIDを取得できます。

また、ほとんどのプロセスはすでに存在している別のプロセスから作成された子プロセスとなっているので、親のプロセスIDを知りたい場合もあります。 親のプロセスIDはos.Getppid()で取得できます。

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("プロセスID: %d\n", os.Getpid())
	fmt.Printf("親プロセスID: %d\n", os.Getppid())
}
```

```console
プロセスID: 6717
親プロセスID: 6688
```

## プロセスグループ・セッショングループ

プロセスを束ねたグループというものがあり、プロセスはそのグループを示すID情報を持っています。次のようにパイプでつなげて実行された仲間が、１つのプロセスグループになります。

```console
$ cat sample.go | echo
```

プロセスグループと似た概念として、セッショングループがあります。同じターミナルから起動したアプリケーションであれば、同じセッショングループになります。同じキーボードに繋がって同じ端末に出力するプロセスも同じセッショングループとなります。

```go
package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	sid, _ := syscall.Getsid(os.Getpid())
	fmt.Fprintf(os.Stderr, "グループID: %d セッションID: %d\n", syscall.Getpgrp(), sid)
}
```

```console
グループID: 7335 セッションID: 7335
```

## ユーザーIDとグループID

プロセスは誰かしらのユーザー権限で動作します。また、ユーザーはいくつかのグループに所属しています。ユーザーは、メインのグループには１つだけしか所属できませんが、サブのグループには複数入れます。

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("ユーザーID: %d\n", os.Getuid())
	fmt.Printf("グループID: %d\n", os.Getgid())
	groups, _ := os.Getgroups()
	fmt.Printf("サブグループID: %v\n", groups)
}
```

```console
ユーザーID: 501
グループID: 20
サブグループID: [20 701 12 61 79 80
81 98 33 100 204 395 398 399]
```

## 実効ユーザーIDと実効グループID

POSIX系OSでは、SUID,SGIDフラグを付与することで、実行ファイルに設定された所有者（実効ユーザーID）と所有グループ（実効グループID）でプロセスが実行されるようになります。これらのフラグが無い時は、実効ユーザーIDも実効グループIDも、元のユーザーIDとグループIDと同じです。これらのフラグが付与されているときは、ユーザーIDとグループIDはそのままですが、実効ユーザーIDと実効グループIDは変更されます。

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("ユーザーID: %d\n", os.Getuid())
	fmt.Printf("グループID: %d\n", os.Getuid())
	fmt.Printf("実効ユーザーID: %d\n", os.Geteuid())
	fmt.Printf("実効グループID: %d\n", os.Geteuid())
}
```

```console
./uid
ユーザーID: 501
ユーザーID: 501
グループID: 501
実効ユーザーID: 501
実効グループID: 501
```

```console
sudo chmod u+s uid
sudo chown root uid

./uid
ユーザーID: 501
グループID: 501
実効ユーザーID: 0
実効グループID: 0
```

POSIX系OSでは、ケーパビリティ（capability）という、権限だけ付与する仕組みが提案されました。それまで、ルート権限が必要な情報の設定・取得を行うツールでは、SUIDを付けてルートユーザーの所有にしたプログラムを用意し、ユーザー権限からも利用可能にするということが行われてきましたが、これでは与えられる権限が大きすぎるため、権限はスーパーユーザーのみが利用でいる権限を細かく分け必要なツールに必要なだけの権限を与える仕組みであり、リスクを減らしています。

## 作業フォルダ

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Println(wd)
}
```


```console
/Users/jumpei/go/src/github.com/pei0
804/go-system-programming/13-process
/wd
```

## ファイルディスクリプタ

子プロセスを起動した時に、他のプロセスの標準入力にデータを流し込んだり、他のプロセスが出力する標準出力や標準エラー出力の内容を読む事もできます。

```go
Stdin  = os.NewFile(0, "/dev/stdin")
Stdout = os.NewFile(1, "/dev/stdout")
Stderr = os.NewFile(2, "/dev/stderr")
```

# プロセスの入出力


プロセスには入力があって、プログラムがそれを処理し、最後出力を行います。その意味では、プロセスはGo言語や他の言語の関数やサブルーチンのようなものだと言えます。

プロセスは次の３つの入出力データを待っています。

- コマンドライン引数
- 環境変数
- 終了コード

## コマンドライン引数

プログラムに設定を与える一般的な手法として使われています。Go言語は、os.Args引数に文字列の配列としてコマンドライン引数が格納されています。

通常はこの配列をそのまま使わず、オプションパーサーを使います。

## 環境変数

- os.Environ() 文字列のリストで全取得
- os.ExpandEnv() 環境変数が埋め込まれた文字列を展開
- os.Setenv() キーに対する値を設定
- os.LookupEnv() キーに対する値を取得（有無をboolで返す）
- os.Getenv() キーに対する値を取得
- os.Unsetenv() 指定されたキーを削除する
- os.Clearenv() 全部クリアする

多言語にはないExpandEnv

```go
package main
 
import (
    "fmt"
    "os"
)
 
func main() {
    fmt.Println(os.ExpandEnv("${HOME}/gobin"))
}
```

```console
/Users/jumpei/gobin
```

## 終了コード

```go
package main

import "os"

func main() {
	os.Exit(1)
}
```

プロセス終了時にはos.Exit()関数を呼びます。この関数は引数として数値を取り、この数値がプログラムの返り値として親プロセスに返されます。この数値が終了コードです。

終了コードは非負の整数です。一般的な慣習として、0が正常終了、1以上がエラー終了ということになっています。安心して使える数値の上限については諸説ありますが、Windowsでは恐らく32bitの数値の範囲で使えます。POSIX系OSでは、子プロセスの終了を待つシステムコールが5種類あります。（wait, waitpid, waiid, wait3, wait4）このうちwaitidを使えば32bitの範囲で扱えるはずです。それ以外の関数は、シグナル受信状態とセットで同じ数値の値にまとめられて返され、その時に8ビットにまとめられてしまうため、255までしか使えません。

- wait: 子プロセスどれか（選択できない）の終了を待つ
- waitpid: 指定されたプロセスIDを持つ子プロセスの終了を待つ
- waitid: プロセスグループ内といった柔軟なプロセス指定ができ、32ビット対応

シェルやPythonなどを親プロセスにして試した限りでは、256以上は扱えなかったので、ポータビリティを考えると255までにしておくのが無難でしょう。

## プロセスの名前や資源情報の取得

タスクマネージャーのようなツールでは、プロセスIDと一緒にアプリケーション名が表示されます。しかし、あるプロセスIDが何者なのか知る方法は標準APIにはありません。

LinuxやBSD系OSの場合、/procディレクトリの情報が取得出来ます。このディレクトリはカーネル内部の情報をファイルシステムとして表示したものです。GNU系のpsコマンドは、このディレクトリをパースして情報を得ています。以下に示すように、/proc/プロセスID/cmdlineというテキストファイルの中にコマンドの引数が格納されているように見えます。

```console
cat /proc/2862/cmdline
bash
```


```go
package main

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/process"
)

func main() {
	p, _ := process.NewProcess(int32(os.Getppid()))
	name, _ := p.Name()
	cmd, _ := p.Cmdline()
	fmt.Printf("parent pid: %d name: '%s' cmd: '%s'\n", p.Pid, name, cmd)
}
```

```console
parent pid: 2757 name: 'go' cmd: 'go run /Users/jumpei/go/src/github.com/pei0804/go-system-programming/13-process/gopsutil/main.go'
```

上記のサンプルでは、プロセスの実行で使われた実行ファイル名と、実行時のプロセスの引数情報を表示しています。 これ以外にも、ホストのOS情報、CPU情報、プロセス情報、ストレージ情報など、数多くの情報が取得できます。

## OSから見たプロセス

プロセスから見た世界と比べると、OSから見た世界の方が、やっていることが少し複雑です。

OSから見たプロセスは、CPU時間を消費して予め用意してあったプログラムに従って動く「タスク」です。OSのしごとは、たくさんあるプロセスを効率よく仕事をさせることです。

Linuxではプロセスごとにtask_strcut型のプロセスディスクリプタと呼ばれる構造体を持っています。プロセスを構成する全ての要素は、この構造体に含まれています。基本的にプロセスから見た各種属性と同じ内容ですが、それには含まれていない要素もいくつかあります。

