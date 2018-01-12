# 15 Process

シグナルには、大きく２つ用途があります。

- プロセス間通信: あるプロセス、別のプロセスに対し、シグナルを送ることが出来ます。場合によっては、あるプロセスから自分自身に対してシグナルを送ることも出来ます。
- ソフトウェア割り込み： システムで発生したイベントは、シグナルとしてプロセスに送られます。シグナルを受け取ったプロセスは、現在行っているタスクを中断して、あらかじめ登録しておいた登録ルーチンを実行します。

# シグナルのライフサイクル

シグナルは様々なタイミングで発生（raise）します。0除算エラーやメモリ範囲外アクセス（船具面と違反）は、CPUレベルで発生し、それを受けてカーネルがシグナルを生成します。アプリケーションプロセスで生成（generate）されるシグナルもあります。

生成されたシグナルは、対象となるプロセスに送信（send）されます。プロセスは、シグナルを受け取ると、現在の処理を中断して受け取ったシグナルの処理を行います。プロセスは受け取ったシグナルを無視するか、補足して処理（handle）します。デフォルトの処理は無視かプロセスの終了です。

プロセスがシグナルを受け取った場合の処理内容は、事前に登録してカスタマイズできます。プロセスを終了しない場合は、シグナルを受け取る前に行なっていたタスクを継続します。

なお、プロセス側でシグナルをハンドルするコードが自由に書けるといっても、シグナルにはそれぞれ決められた役割があります。本来の役割から書けな離れた処理は実装するべきではありません。

## シグナルの種類

Unix系では次のコマンドを実行するとシグナルが一覧できます。

[signal](https://linuxjm.osdn.jp/html/LDP_man-pages/man7/signal.7.html)

```console
man 7 signal
```

## ハンドルできないシグナル

- SIGKILL: プロセスを強制終了
- SIGSTOP: プロセスを一時停止して、バックグラウンドジョブにする

これらのシグナルは、それぞれ「SIG」を除外した文字列（KILL、STOP）をkillコマンドにオプションとして指定することで、コマンドラインからプロセスに対して送信できます。

```console
# プロセスIDを指定してSIGKILLシグナル送信
kill -KILL 35698
 
# プロセス名を指定してSIGSTOPシグナル送信
pkill -STOP ./sample
```

SIGSTOPでは、ジョブがバックグランドに回っていったんターミナルが戻ってきて、別のプロセスを実行出来るようになります。fgコマンドを使うとサイドフォアグラウンドとして戻すことが出来ます。

```console
# バックグラウンドジョブになっているジョブを呼び戻して仮想端末に再接続
fg ./sample ⏎
```

## サーバーアプリケーションでハンドルするシグナル

- SIGTERM: kill()システムコールやkillコマンドが送信するシグナルで、プロセスを終了させるもの
- SIGHUP: 通常は後述するコンソールアプリケーション用のシグナルだが、ターミナルを持たないデーモンでは絶対に受け取ることは出来ないので、サーバアプリケーションでは別の意味で使われる。具体的には、設定ファイルの再接続を外部から指示する用途で使われることがデファクトスタンダートとなっている。

上記はデフォルトではどちらもプロセス終了です。

## コンソールアプリケーションでハンドルするシグナル

- SIGINT: ユーザーがctrl + Cでプログラムを終了（ハンドル出来るバージョンのSIGKILL）
- SIGQUIT: ユーザーがctrl + \ でコアダンプを生成して終了
- SIGTSTP: ユーザーがctrl + Zで停止させ、バックグランド動作させる（ハンドルできるバージョンはSIGSTOP）
- SIGCOUNT: バックグランド動作から戻される指令
- SIGWINCH: ウィンドウサイズ変更
- SIGHUP: バックグランド動作になったり、往路セスが終了したりして、擬似ターミナルから切断される時に呼ばれるシグナル

## たまに使うかもしれない、その他のシグナル

- SIGUSR1とSIGUSR2: ユーザー定義のシグナル。アプリケーションが任意の用途で使える。
- SIGPWR: 外部電源が切断し、無停電電源装置が使われたものの、バッテリー残量が低下してシステムを終了する必要がある時にOSから送信されるシグナル。

## Go言語におけるシグナルの種類

```console
var (
  Interrupt Signal = syscall.SIGINT
  Kill      Signal = syscall.SIGKILL
)
```

- ハンドル不可・外部からのシグナルは無視: SIGFPE, SIESEGV, SIGBUSが該当。算術エラー、メモリ範囲外アクセス、その他のハードウェア例外を表す。知名度の高いシグナル。Go言語では、自分のコード中で発生した場合にはpanicに変換して処理される。外部から送付することはできず、ハンドラ定義しても呼ばれない。
- ハンドル不可: SIGKILL, SIGSTOPが該当。Go言語に限らず、C言語でもハンドルできないシグナル。
- ハンドル可能・終了ステータス1: SIGQUIT, SIGABRT が該当
- ハンドル可能・パニック、レジスター覧表示、終了ステータス2: SIGILL, SIGTRAP, SIGEMT, SIGSYSが該当
- ハンドル可能・無視 ：SIGPIPE, SIGALRM, SIGURG, SIGIO, SIGXCPU, SIGXFZ, SIGVTALRM, SIGWINCH, SIGINFO, SIGUSR1, SIGUSR2, SIGCHLD, SIGPROFが該当
- ハンドル可能・OSデフォルト動作（macOSは無視） ：SIGTSTP, SIGCONT, SIGTTIN, SIGTTOUが該当

## シグナルのハンドラを書く

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// サイズが1より大きいチャネルを作成
	signals := make(chan os.Signal, 1)
	// 最初のチャネル以降は、可変長引数で任意の数のシグナルを設定可能
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	s := <-signals

	switch s {
	case syscall.SIGINT:
		fmt.Println("SIGINT")
	case syscall.SIGTERM:
		fmt.Println("SIGTERM")
	}
}
```

```console
❯ ./signal // ctrl + c
^CSIGINT
❯ ./signal // kill hoge
SIGTERM
```

シグナルに関する設定はプロセス全体に及ぶグローバルな設定です。またシグナルはフォアグラウンドのプロセスに最初に送信されます。したがって、自作のコードでシグナルのハンドラを下記、それをgo runを使って実行すると、シグナルは自作のコードのプロセスではなくgoコマンドのプロセスに送信されるので、go buildして実行ファイルを作成してから実行してください。

## シグナルを無視する

```go
package main

import (
	"fmt"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Ignore Ctrl + C for 3 second")
	signal.Ignore(syscall.SIGINT, syscall.SIGHUP)
	time.Sleep(time.Second * 3)
}
```

## シグナルのハンドラをデフォルトに戻す

```go
package main

import (
	"os/signal"
	"syscall"
)

func main() {
	// 可変長引数で任意の数のシグナルを設定可能
	signal.Reset(syscall.SIGINT, syscall.SIGHUP)
}
```

## シグナルの送付

```go
signal.Stop(signals)
```

これを呼び出すと、それ以降Notify()で指定したシグナルを受け取らなくなるわけではなく、デフォルトに戻るようです。 Notify()でSIGINT（Ctrl + C）を指定していた場合、呼び出し後はブロックせずにデフォルトでプロセスを終了するようになります。

## シグナルを他のプロセスに送る

```go
package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s [pid]\n", os.Args[0])
		return
	}
	// 第一引数で指定されたプロセスIDを数値に変換
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		panic(err)
	}
	// シグナルを送る
	process.Signal(os.Kill)
	// KILLの場合は次のショートカットも利用可能
	process.Kill()
}
```

```console
send 1903
```

```console
Loop
Loop
Loop
Loop
Loop
Loop
[1]    1903 killed     ./coun
```

os.execパッケージを使った高級なインターフェースの場合は以下のようになります

```console
cmd := exec.Command("child")
cmd.Start()
 
// シグナル送信
cmd.Process.Signal(os.Interrupt)
```

## シグナル応用例

サーバ系のプログラムは、CUIやGUIのツールとは異なり、複数のユーザーが同時にアクセスして利用できます。そのため、バージョンアップや設定修正などで再起動する際に、正しく終了するのが難しいという問題があります。いきなりシャットダウンや自然にユーザーが離れることを待つわけにはいきません。それらに対応するのがグレイスフルリスタートです。

グレイスフル・リスタートを実現するための補助ツールとして広く利用されている仕組みに、奥一穂さんが作成したServer::Starterがあります。動きとしては、新しいサーバーを起動して新しいリクエストはそちらに流しつつ、古いサーバーのリクエストが完了したら正しく終了させる仕組みです。これをすることでダウンタイムが0に出来ます。

## Server::Starterの使い方

install

```console
go get github.com/lestrrat/go-server-starter/cmd/start_server
```

実行方法

```console
start_server --port 8080 --pid-file app.pid -- ./server
```

- ポート8080番を開く
- 現在のプロセスIDをapp.pidファイルにかき出す
- 開いたポートを渡し、serverを子プロセスとして起動する

(親)start_server -> (子)./server

### 再起動の仕組み

Server::Starterで起動したサーバープロセスを再起動する時は、シグナルを利用します。デフォルでは、SIGUPを使って再起動を依頼します。（ただし、どのシグナルを利用するかはコマンドで指定できます。）  
再起動した時は、親プロセスであるServer::StaterにSIGHUPシグナルを送ります。Unix系OSの場合、次のようにすれば良いでしょう。

```console
kill -HUP `cat app.pid`
```

多重起動していなければ、以下のようにしれも良いです。

```console
pkill -HUP start_server
```

SIGHUPを受け取ったServer::Starterは、新しいプロセスを起動し、起動済みの子プロセスにはSIGTERMを送ります。 子プロセスであるサーバーが、「SIGTERMシグナルを受け取ったら、新規のリクエスト受け付けを停止し、現在処理中のリクエストが完了するまで待って終了する」という実装になっていれば、 再起動によるエラーに遭遇するユーザーを一人も出すことなく、ダウンタイムなしでサービスを更新できます。

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lestrrat/go-server-starter/listener"
)

func main() {
	// シグナル初期化
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	// Server::Staterからもらったソケットを確認
	listeners, err := listener.ListenAll()
	if err != nil {
		panic(err)
	}
	// ウェブサーバーをgoroutineで起動
	// ウェブサーバーをgoroutineで起動
	server := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "server pid: %d %v\n", os.Getpid(), os.Environ())
		}),
	}
	go server.Serve(listeners[0])

	// SIGTERMを受け取ったら終了させる
	<-signals
	server.Shutdown(context.Background())
}
```

```console
start_server --port 8080 --pid-file app.pid -- ./greceful
starting new worker 3574
```

SIGHUPを送信する

```console
kill -HUP `cat app.pid`
```

```console
--port 8080 --pid-file app.pid -- ./greceful
starting new worker 4582
received HUP (num_old_workers=TODO)
spawning a new worker (num_old_workers=TODO)
starting new worker 4589
new worker is now running, sending TERM to old workers:4582
sleep 0 secs
killing old workers
old worker 4582 died, status:0
```
