# 17 Parallel

goroutineは軽量スレッドと呼ばれるものです。では、通常のOSのスレッドとどう違うのか？を掘り下げます。

# スレッドとgoroutineのち外

スレッドとは、プログラムを実行するための「もの」であり、OSによって手配されます。

プログラムから見たスレッドは、「メモリにロードされたプログラムの現在の実行状態を持つ仮想CPU」です。この仮想CPUにスタックメモリが割り当てられます。

一方、OSやCPUから見たスレッドは、「時間が凍結されたプログラムの実行状態」です。この実行状態には、CPUが演算に使ったり計算結果や状態を保持したりするレジスタと呼ばれるメモリとスタックメモリが含まれます。

OSのしごとは、凍結状態のプログラムの実行状態を復元して、各スレッドを順番に短時間ずつ処理を再開させることです。その際の順番や、一回に実行する時間タイムスライスは、スレッドごとに設定されている優先度で決まります。実行予定のスレッドはランキューと呼ばれるリストに入っており、なるべく公平に処理が回るようにスケジューリングされています。複数のプログラムは、このようにして、時間分割しCPUコアにマッピングされて実行されるのです。

スレッドがCPUコアに対してマッピングされるのに対し、goroutineはOSのスレッド（Go製のアプリケーションから見ると１つの仮想CPU）にマッピングされます。この点が通常のスレッドとGo言語のgoroutineとの大きな違いです。

## GoのラインタイムはミニOS

OSが提供するスレッド以外に、プログラミング言語のラインタイムでスレッド相当の機能を持つことには、どんなメリットがあるのでしょうか。Go言語の場合、機能が少ない代わりにシンプルで起動が早いスレッドが提供されています。

- 大量のクライアントを効率よくさばくサーバーを実装する（C10K）時に、クライアントごとに1つのgoroutineを割り当てるような実装であっても、リーズナブルなメモリ使用量で処理出来る。
- OSのスレッドでブロッキングを行う操作をすると、他のスレッドが処理を開始するにはOSがコンテキストスイッチして順番を待つ必要があるが、Goの場合チャネルなどでブロックしたら、残ったタイムスライスでランキューに入った別のgoroutineのタスクを実行出来る。
- プログラムのランタイムが、プログラム全体の中でどのgoroutineがブロック中なのかといった情報をきちんと把握しているため、デッドロックを作ってもランタイムが検知してどこでブロックしているか一覧表示出来る。

## runtimeパッケージのgoroutine関連の機能

軽量スレッドであるgoroutineを使うには、前回説明したように、goを付けて関数呼び出しを行うだけです。しかし、場合によってはベースとなるOSのスレッドに対して何らかの制限を課すといった、より低レベルな操作をしたこともあります。そんな時にruntimeパッケージには低レベルな関数がいくつかあります。

### runtime.LockOSThread() / runtime.UnlockOSThread()

runtime.LockOSThread()を呼ぶと、現在実行中のOSスレッドでのみgoroutineが実行されるように束縛出来ます。さらにそのスレッドが他のgoroutineによって使用されなくなります。これらの束縛は、runtime.UnlockOSThread()を呼んだり、ロックしたgoroutineが終了すると解除されます。

この機能が必要になる状況は、メインスレッドでの実行が強制されるライブラリ（GUIのフレームワークや、OpenGLとその依存ライブラリなど）をGo言語で利用する場合などです。

### runtime.Gosched()

現在実行中のgoroutineを一時中断して、他のgoroutineに処理を回します。goroutineにはOSスレッドとは異なり、タスクをスリープ状態にしたり、復元したりする機能はありません。ランキューの順番が回ってきたら何事なかったように処理が再開されます。

### runtime.GOMAXPROCS(n) / runtime.NumCPU()

同時に実行するOSスレッド数（I/Oのブロック中のスレッドは除く）を制御する関数です。  
現在は、runtime.GOMAXPROCS()がせっていされるようになったので、特別な場合を除いて設定する必要はありません。しかし、最速を狙おうとすると、このデフォルト値の半分に設定する方がスループットが上がる場合があります。現代のCPUのいくつかは、余剰のCPUリソースを使って1コアで2以上のスレッドを同時に実行する機構（ハイパースレッディング）を備えています。そのような機構を利用している場合、1コアで2つのヘビーな計算を同時に実行をすると、CPUのリソースを食い合ってパフォーマンスが上がらないことがあります。

## Race Detector

Go言語には、データ競合を発見する機能があります3。 この機能は、Race Detectorと呼ばれ、go buildやgo runコマンドに-raceオプションを追加するだけで使えます。

Race Detectorを有効にしてGoプログラムを実行すると、次のようなメッセージが表示され、 競合が発生した個所と、競合した書き込みを行ったgoroutine、そのgoroutineの生成場所が分かります

```console
==================
WARNING: DATA RACE
Read at 0x0000011a7118 by goroutine 7:
  main.main.func1()
      /Users/shibu/.../mutex2.go:25 +0x41
 
Previous write at 0x0000011a7118 by goroutine 6:
  main.main.func1()
      /Users/shibu/.../mutex2.go:25 +0x60
 
Goroutine 7 (running) created at:
  main.main()
      /Users/shibu/.../mutex2.go:26 +0x93
 
Goroutine 6 (finished) created at:
  main.main()
      /Users/shibu/.../mutex2.go:26 +0x93
==================
```

## syncパッケージ

### sync.Mutex / sync.RWMutex

マルチスレッドプログラミングでは、「メモリ保護のためにロックを使う」といった説明をされることがあります。これはスレッドが同じメモリ空間で動くためですが、実際に保護するのは実行パスであり、メモリを直接保護するわけではありません。sync.Mutexは実行パスに入ることが可能なgorutineを排他制御によって制限するにに使います。

sync.Mutexを使うと、「メモリを読み込んで書き換える」コードに入るgorutineが１つされるため、不整合を防ぐことが出来ます。この同時に実行されると問題が起きる実行コード行をクリティカルセクションと呼びます。マップや配列に対する操作はアトミックではないため、複数のgorutineからのアクセスする場合には保護が必要です。

```go
package main

import (
	"fmt"
	"sync"
)

var id int

func generateID(mutex *sync.Mutex) int {
	// Lock()/Unlock()をペアで呼び出してブロックする
	mutex.Lock()
	mutex.Unlock()
	id++
	return id
}

func main() {
	// sync.Mutex構造体の変数宣言
	// 次の宣言をしてもポインタ型になるだけで正常に動作します。
	// mutex := new(sync.Mutex)
	var mutex sync.Mutex
	for i := 0; i < 100; i++ {
		go func() {
			fmt.Printf("id: %d\n", generateID(&mutex))
		}()
	}
}
```

なお、上記のコードにはバグがあります。 このバグを直す方法はいくつかありますが、一番簡単なのが、次節で説明するsync.WaitGroupを使う方法です。


Mutexとチャネルの使い分けについては以下の通りです。  

- チャネルが有用な用途：データの所有権を渡す場合、作業を並列化して分散する場合、非同期で結果を受け取る場合
- Mutexが有用な用途：キャッシュ、状態管理

RMutexについては、RLock()とRUnlockというめそっどがあります。Rが付く方は読み込み用のロック取得と解放で、読み込みはいくつものgorutineが並列して行えるが、書き込み時には他のgorutineの実行は許されない。

### sync.WaitGroup

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	// ジョブ数をあらかじめ登録
	wg.Add(2)

	go func() {
		// 非同期で仕事をする
		fmt.Println("仕事1")
		// Doneで完了を通知
		wg.Done()
	}()

	go func() {
		// 非同期で仕事をする
		fmt.Println("仕事2")
		// Doneで完了を通知
		wg.Done()
	}()

	// 全ての処理が終わるまで待つ
	wg.Wait()
	fmt.Println("終了")
}
```

### sync.Once

```go
package main

import (
	"fmt"
	"sync"
)

func ini() {
	fmt.Println("初期処理")
}

var once sync.Once

func main() {
	// 3回呼んでも1度しか呼ばれない
	once.Do(ini)
	once.Do(ini)
	once.Do(ini)
}
```

Go言語には、init()という名前の関数がパッケージ内にあると、それが初期化関数として呼ばれる機能があります6。 sync.Onceではなくinit()を使うほうが、初期化処理を呼び出すコードを書かなくても実行され、コード行数も減るので、シンプルです。 sync.Onceをあえて使うのは、初期化処理を必要なときまで遅延させたい場合でしょう。

### sync.Cond

条件変数と呼ばれる排他制御の仕組みです。

- 先に終わらせなければいけないタスクがあり、それが完了したら待っている全てのgoroutineに通知する（Broadcat()メソッド）
- リソースの準備が出来た次第、そのリソースを待っているgorutineに通知をする（Signal()メソッド）

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mutex sync.Mutex
	cond := sync.NewCond(&mutex)

	for _, name := range []string{"A", "B", "C"} {
		go func(name string) {
			// ロックしてからwaitメソッドを呼ぶ
			mutex.Lock()
			defer mutex.Unlock()
			// Broadcat()が呼ばれるまで待つ
			cond.Wait()
			// 呼ばれた
			fmt.Println(name)
		}(name)
	}
	fmt.Println("よーい")
	time.Sleep(time.Second)
	fmt.Println("どん！")
	// 待っているgoroutineを一斉に起こす
	cond.Broadcast()
	time.Sleep(time.Second)
}
```

### sync.Pool

```go
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
```

```go
package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	var count int
	pool := sync.Pool{
		New: func() interface{} {
			count++
			return fmt.Sprintf("created: %d", count)
		},
	}

	// GCを呼ぶと追加された内容が消える
	pool.Put("remove 1")
	pool.Put("remove 2")
	runtime.GC()

	// Newが実行される
	fmt.Println(pool.Get())
}
```

### sync/atomicパッケージは不可分操作と呼ばれる操作を提供しています。

これはCPUレベルで提供されている「１つで複数の操作を同時に行う命令」などを駆使したり、提供されていなければ正しく処理が行われるまでループするという命令を駆使して、「確実に実行される」ことを保証している関数として提供されています。途中でコンテキストスイッチが入って操作が失敗しないことが保証されます。

```go
var id int64
 
func generateId(mutex *sync.Mutex) int {
    return atomic.AddInt64(&id, 1)
}
```

複数のgorutineがアクセスしてロックされると、コンテキストスイッチが発生します。こちらのロックフリーな関数を使えばコンテキストスイッチが発生しないため、要件に合えば最速です.

コンテキストスイッチ (context switch) とは、複数のプロセスが1つのCPUを共有できるように、CPUの状態(コンテキスト (情報工学))を保存したり復元したりする過程のことである。コンテキストスイッチはマルチタスクオペレーティングシステムに不可欠な機能である。通常コンテキストスイッチは多くの計算機処理を必要とするため、オペレーティングシステムの設計においてはコンテキストスイッチを最適化することが重要である。
