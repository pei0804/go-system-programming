# 05 syscall

# システムコールとは何か？

システムコールの正体は、特権モードでOSの機能を呼ぶことです。

## CPUの動作モード

CPUの行う仕事には、各種資源（メモリ、CPU時間、ストレージなど）の管理と外部入出力の機能提供（ネットワーク、ファイル読み書き、プロセス間通信）があります。  
アプリケーションプロセスの全てが行儀よく他プロセスに配慮しながら、権限以上のことを勝手にやらないのであれば、OSは基本的に外部入出力機能の提供だけで済みます。実際にWindows3.0まではアプリケーションプロセスが他のプロセスのメモリ空間まで自由にアクセス出来ました。

しかし、プロセスにバグがあったり、悪意のあるプロセスが他のプロセスのメモリを書き換えたり、コンピュータ全体を停止させることなどが可能になります。現在では、プロセスが自分のことだけを考えるだけでよくなり、メモリ管理や時間管理は全てOSが行うようになっています。その分、OSの仕事は増えていますが、ハードウェアであるCPUにも様々な仕組みが用意されていて、OSのしごとを裏で支えています。

そのCPUの仕組みの一つに、動作モードというものがああります。ほとんどのOSで使われているのは、OSが動作する特権モードと、一般アプリケーションが動作するユーザーモードの二種類です。

### 特権モード

CPUの機能が基本的に全て使えます。OSは配下にある全てのプロセスのために資源を管理したり、必要に応じて取り上げたり（ユーザー操作による終了・メモリ不足によるOOMキラーなど）する必要があるため、通常のプロセスよりも強い特権モードで動作します。

### ユーザーモード

特権モードにあるような機能をCPUでりようできないようになっています。

## システムコールでモードの壁を越える

通常のアプリケーションでも、メモリ割り当てやファイル入出力、インターネット通信などの機能が必要になることは多々あります。むしろ、それらを全く利用しないアプリケーションは意味がないでしょう。そこで必要になるのが、システムコールです。多くのOSでは、システムコールを介して、特権モードでのみ許されている機能をユーザーモードのアプリケーションから利用出来るようにしています。

システムコールの仕組みは何種類化ありますが、現在主流の64ビットのx86系のCPUでは、通常の関数呼び出し（アセンブリ命令のCALL）と似たSYSCALL命令を呼び出し、戻る時も通常の関数からの戻り（アセンブリ命令のRET）に近い、SYSRET命令を使います。ARMの場合は、SVC命令（スーパーバイザーコール）が使われます。これらの命令を使うと、OSが提供する関数を呼び出ししますが、飛ばれた側では特権モードで動作します。そのため、ユーザーモードでは直接行えない、メモリ割り当て、ファイル入出力、インターネット通信などの機能が実行可能になります。

## Go言語におけるシステムコールの実装

- io.Reader
- io.Writer
- io.Seeker
- io.Closer

上記のインターフェース内部では、最終的にsyscallパッケージが定義された関数を呼び出しています。

|システムコール関数                                    |機能                                      |
|:-----------------------------------------------------|:-----------------------------------------|
|func syscall.Open(path string, mode int, perm uint32) |ファイルを開く（作成も含む）              |
|func syscall.Read(fd int, p []byte)                   |ファイルから読み込みを行う                |
|func syscall.Write(fd int, p []byte)                  |ファイルに書き込みを行う                  |
|func syscall.Close(fd int)                            |ファイルを閉じる                          |
|func syscall.Seek(fd int, offset int64, whence int)   |ファイルの書き込み・読み込みイチを移動する|

## 各OSにおけるシステムコールの実装を見てみよう

```go
package main

import "os"

func main() {
	file, err := os.Create("test.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.Write([]byte("system call example\n"))
}
```

os.Createはos.OpenFile()を使いやすくするための便利関数です。

```go
func Create(name string) (*File, error) {
    return OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)
}
```

ここからOSによって動きが変わります。

### MacOSにおけるシステムコール

```go
// Linux MacOSの分岐
r, e = syscall.Open(name, flag|syscall.O_CLOEXEC, syscallMode(perm))
```

syscall.Open()関数から、zsyscall_darwin_amd64.goの中のOpen()関数が呼び出されています。

```go
func Open(path string, mode int, perm uint32) (fd int, err error) {
	var _p0 *byte
	_p0, err = BytePtrFromString(path)
	if err != nil {
		return
	}
	r0, _, e1 := Syscall(SYS_OPEN, uintptr(unsafe.Pointer(_p0)), uintptr(mode), uintptr(perm))
	use(unsafe.Pointer(_p0))
	fd = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}
```

この関数では、まずGo言語形式の文字列をC言語形式の文字列（先頭要素へのポインタ）に変換しています。これはシステムコールに渡せるのが数値だけだからです。真ん中あたりでSyscall()関数でを呼び出しています。  
OSに対してシステムコール経由で仕事をお願いする時はどんな処理をしてほしいかを番号で指定します。「5番の処理を実行してほしい」 などとお願いするわけです。SYS_OPENは、そのための番号として各OS用のヘッダーファイルなどから自動生成された定数です。名前の先頭がzsysnum_になっている各OS用のファイルで定義されています。  
Syscall()の中身は、MacOSの場合、asm_darwin_amd64.sというGo言語の低レベルアセンブリ言語で書かれたコードが定義されています。  

```
TEXT	·Syscall(SB),NOSPLIT,$0-56
	CALL	runtime·entersyscall(SB)
	MOVQ	a1+8(FP), DI
	MOVQ	a2+16(FP), SI
	MOVQ	a3+24(FP), DX
	MOVQ	$0, R10
	MOVQ	$0, R8
	MOVQ	$0, R9
	MOVQ	trap+0(FP), AX	// syscall entry
	ADDQ	$0x2000000, AX
	SYSCALL // SYSCALLが実行
	JCC	ok
	MOVQ	$-1, r1+32(FP)
	MOVQ	$0, r2+40(FP)
	MOVQ	AX, err+48(FP)
	CALL	runtime·exitsyscall(SB)
	RET
```

SYSCALLという命令からは、OS側のコードに処理が渡ります。この中を覗くことは出来ませんが、rumtimeパッケージのentersyscall()関数と、exitsyscall()関数が呼び出されます。

entersyscall()関数は、現在実行中のOSすれっどが時間のかかるシステムコールでブロックされていることを示すマークをつけます。  
exitsyscall()関数はそのマークを外します。  
Go言語では、実行しなければならないタスクが多くあるシステムコールのブロックなどで動けるスレッドが不足すると、OSに依頼して新しい作業スレッドを作成します。スレッド作成は重い処理なので、これを必要に行わないために、これらの関数を使います。これにより実行効率面でメリットがありますが、Go言語の実行モデルに関係するものであり、他のプログラム言語には見られない特徴と言えます。  

なお、これらのスレッド関係の処理を行わないRawSyscall()という関数もあります。ファイル読み書きやネットワークアクセスの場合、物理的にヘッドを動かしたり、役100ミリ秒から数秒程度のレスポンス待ちが発生する可能生が高く、重い処理です。メモリ確保もスワップが発生するとファイル読み書きと同じくらいコストがかかります。それ以外の短時間で終わることが見込まれる処理の場合には、RawSyscall()が使われます。

## POSIXとC言語の標準規格

Go言語に限らずシステムコール一般の話。  
POSIXという名前を聞いたことがある人は多いでしょう。POSIX（Portable Operatins System Interface）は、OS間で共通のシステムコールを決めることで、アプリケーションの移植性を高めるために作られたIEEE規格です。  
最終的にOSにしごとをお願いするのはシステムコールですが、POSIXでさだめられているのはシステムコールそのものではなく、システムコールを呼び出すためのインターフェースです。具体的にはC言語の関数名と引数、返り値が定義されています。  

例えば、ファイル入出力は、POSIXでは５つの基本システムコールで構成されていて、そのためC言語の関数は、open(),read(),write(),close(),lseek()ｆです。

これらの関数は、Cげんごにおける低レベルな共通インターフェースとして用意されていますが、通常のプログラミングで直接扱うことはほとんどありません。

Go言語におけるsyscallの各関数もこのシステムコールの呼び出し口です。それぞれの先頭を小文字にすれば、C言語の関数と同じ名前になっています。呼び出し時に与える情報の意味、引数の順序、返り値なども、引数の型はGo特有のものを使っていますが、基本は同じです。

そして、Go言語でもsyscall関数を直接使うことはなく、基本的にはos.File構造体とそのメソッドを使ってプログラミングします。syscall以下の関数を使う方法はドキュメントにはほとんどなく、C言語用の情報を見る必要があります。


