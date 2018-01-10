# 14 Process

Go言語のプログラムから他のプロセスを扱う時、プロセスを表す構造体を利用します。そのための構造体には次の2種類があります。

- osパッケージのos.Process: 低レベルな構造体
- os/execパッケージのexec.Cmd: 少し高機能な構造体。内部でos.Processを持つ

# exec.Cmdによるプロセスの起動

exec.Cmd構造体は次の2つの関数で作ることができます。

- exec.Command(名前, 引数...)
- exec.CommandContext(コンテキスト, 名前, 引数...)

両者の違いは、引数としてコンテキストを取れるかどうかです。コンテキストは依存関係が複雑なときでもタイムアウトやキャンセルをきちんと行う仕組みで、Go1.5から標準で利用出来るようになりました。exec.CommandContextにコンテキストとして渡した処理が、exec.Cmdが表すプロセスの終了前に完了した場合、そのプロセスはos.Process.Kill()メソッドを使って強制終了されます。

上記は引数として外部プログラムを指定すると、その外部プログラムの実行にかかった時間を表示する。UNIX系のOSに備わっているtimeコマンドと似た動作ですが、実処理時間は表示せず、システム時間

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) == 1 {
		return
	}
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	state := cmd.ProcessState
	fmt.Printf("%s\n", state.String())
	fmt.Printf(" Pid: %d\n", state.Pid())
	fmt.Printf(" System: %v\n", state.SystemTime())
	fmt.Printf(" User: %v\n", state.UserTime())
}
```

上記は、引数として外部プログラムを指定すると、その外部プログラムの実行にかかった時間を表示するプログラムです。UNIX系のOSに備わっているtimeコマンドと似た動作ですが、実処理時間は表示せず、システム時間とユーザー時間を表示します。

```console
go run cmd/main.go sleep 1
exit status 0
 Pid: 1180
 System: 2.041ms
 User: 2.106ms
```

上記のサンプルプログラムでは、引数として渡された外部プログラムを指定して、exec.Command()を呼び出し、そのプロセスを表すexec.Cmd構造体のRun()メソッドを呼び出して居ます。exec.Cmdには、プロセスの実行を制御するメソッドとして、Run()だけではなく下記のようなものが用意されています。

|メソッド|説明|
|:------------|:------|
|Start() error|実行を開始する|
|Wait() error| 終了を待つ|
|Run() error| Start() + Wait()|
|Output() ([]byte, error)|Run()実行後に標準出力の結果を返す|
|CombinedOutput() ([]byte, error)|Run()実行後に標準出力、標準エラー出力の結果を返す|

今回取得した情報以外にも以下のようなものが取れます。

```go
state := cmd.ProcessState
// 終了コードと状態を文字列で返す
fmt.Printf("%s\n", state.String())
// 子プロセスのプロセスID
fmt.Printf("  Pid: %d\n", state.Pid())
// 終了しているかどうか
fmt.Printf("  Exited: %v\n", state.Exited())
// 正常終了か？
fmt.Printf("  Success: %v\n", state.Success())
// カーネル内で消費された時間
fmt.Printf("  System: %v\n", state.SystemTime())
// ユーザーランドで消費された時間
fmt.Printf("  User: %v\n", state.UserTime())
```

上記は、引数として外部プログラムを指定すると、その外部プログラムの実行にかかった時間を表示するプログラムです。UNIX系のOSに備わっているtimeコマンドと似た動作ですが、実処理時間は表示せず、システム時間とユーザー時間を表示します。

exec.Cmd構造体では、構造体の作成から実行までの間にプロセスの実行に関する情報を変更するためのメンバーも提供されています。

プロセスの実行に関する情報を変更するためのexec.Cmdのメンバー変数

|変数|種類|説明|
|:---|:---|:---|
|Env[] string|入力|環境変数。セットされていない時は親プロセスを引き継ぐ。|
|Dir string|入力|実行時のディレトリ。セットされないと親プロセスと同じ。|
|ExtraFiles []*os.File|入力|子プロセスに追加で渡すファイル。3以降のファイルディスクリプタで参照できる。追加のファイルは、子プロセスからはos.NewFile()を使って開ける|
|SysProcAttr *syscall.SysProcAttr|入力|OS固有の設定|


## リアルタイムな入出力

実行開始する前に下記の表に示すメソッドを使うことで、子プロセスとリアルタイムに通信を行うためのパイプが取得出来ます。このパイプはexec.Cmd構造体が子プロセス終了時に閉じるため、呼び出し側では閉じる必要はありません。

exec.Cmdとリアルタイムの入出力を行うためのメソッド

|メソッド|説明|
|:-------|:---|
|StdinPipe() (io.WriteCloser, error)|子プロセスの標準入力に繋がるパイプを取得|
|StdoutPipe() (io.WriteCloser, error)|子プロセスの標準出力に繋がるパイプを取得|
|StderrPipe() (io.WriteCloser, error)|子プロセスの標準エラー出力に繋がるパイプを取得|

```go
package main

import (
	"bufio"
	"fmt"
	"os/exec"
)

func main() {
	count := exec.Command("./count/count")
	stdout, _ := count.StdoutPipe()
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("(stdout) %s\n", scanner.Text())
		}
	}()
	err := count.Run()
	if err != nil {
		panic(err)
	}
}
```

```console
go run main.go
(stdout) 0
(stdout) 1
(stdout) 2
(stdout) 3
(stdout) 4
(stdout) 5
(stdout) 6
(stdout) 7
(stdout) 8
(stdout) 9
```

標準出力と標準エラー出力を同時にダンプするときは、sync.Mutexなどを使って同時に書き込まないようにしたほうがよいでしょう。

## os.Processによるプロセスの起動・操作

os.Processは低レベルなAPIです。指定したコマンドを実行出来る他、既に起動中のプロセスのIDを指定して作成出来ます。

- os.StartProcess(コマンド, 引数, オプション)
- os.FindProcess(プロセスID)

os.StartProcess()を使って実行ファイルを指定する場合は、exec.Command()とは異なり、PATH環境変数を見て実行ファイルを探すことはしません。そのため、絶対パスや相対パスなどで実行ファイルを指定する必要があります。  

os.StartProcess()を使うときは、wait()メソッドを呼び出すことで、子プロセスが終了するのを待てます。wait()メソッドは、os.ProcessState構造体のインスタンスを返すので、これを使って終了状態を取ることが出来ます。

一方、os.FindProcess()を使って実行中のプロセスにアタッチして作ったos.Processオブジェクトは、wait()メソッドを呼び出すことができず、終了状態を取得できません。kill()メソッドを呼ぶか、次回以降に説明するシグナルを送っる以外出来ることがありません。

## プロセスに関する便利なGo言語ライブラリ

### プロセスの出力に色付けする

```go
package main

import (
	"fmt"
	"io"
	"os"

	colorable "github.com/mattn/go-colorable"
	isatty "github.com/mattn/go-isatty"
)

var data = "\033[34m\033[47m\033[4mB\033[31me\n\033[24m\033[30mOS\033[49m\033[m\n"

func main() {
	var stdOut io.Writer
	if isatty.IsTerminal(os.Stdout.Fd()) {
		stdOut = colorable.NewColorableStdout()
	} else {
		stdOut = colorable.NewNonColorable(os.Stdout)
	}
	fmt.Fprintln(stdOut, data)
}
```

接続先がターミナルのときはエスケープシーケンスを表示し、そうでないときはエスケープシーケンスを除外するフィルタ（colorable.NonColorable）を使い分ける例です。 このコードを実行すると古い昔のソフトウェアの名前を表示します2が、リダイレクトしてファイルに落とすとエスケープシーケンスが出力されないことが確認できます。

エスケープシーケンスを除外していないと、無駄な文字列が入る

```console
[34m[47m[4mB[31me
[24m[30mOS[49m[m
```

### 外部プロセスに対して自分が擬似端末だと詐称する

```go
package main
 
import (
    "fmt"
    "github.com/mattn/go-colorable"
    "github.com/mattn/go-isatty"
    "io"
    "os"
)
 
func main() {
    var out io.Writer
    if isatty.IsTerminal(os.Stdout.Fd()) {
        out = colorable.NewColorableStdout()
    } else {
        out = colorable.NewNonColorable(os.Stdout)
    }
    if isatty.IsTerminal(os.Stdin.Fd()) {
        fmt.Fprintln(out, "stdin: terminal")
    } else {
        fmt.Println("stdin: pipe")
    }
    if isatty.IsTerminal(os.Stdout.Fd()) {
        fmt.Fprintln(out, "stdout: terminal")
    } else {
        fmt.Println("stdout: pipe")
    }
    if isatty.IsTerminal(os.Stderr.Fd()) {
        fmt.Fprintln(out, "stderr: terminal")
    } else {
        fmt.Println("stderr: pipe")
    }
}
```

```go
package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/kr/pty"
)

func main() {
	cmd := exec.Command("./check/check")
	stdpty, stdtty, _ := pty.Open()
	defer stdtty.Close()
	cmd.Stdin = stdpty
	cmd.Stdout = stdpty
	errpty, errtty, _ := pty.Open()
	defer errtty.Close()
	cmd.Stderr = errtty
	go func() {
		io.Copy(os.Stdout, stdpty)
	}()
	go func() {
		io.Copy(os.Stderr, errpty)
	}()
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
```

```console
stdin: terminal
stdout: terminal
stderr: terminal
```

