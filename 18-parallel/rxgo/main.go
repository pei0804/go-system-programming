package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/reactivex/rxgo/observable"
	"github.com/reactivex/rxgo/observer"
)

func main() {
	//observableを作成
	emitter := make(chan interface{})
	source := observable.Observable(emitter)

	// イベントを受け取るobserverを作成
	watcher := observer.Observer{
		NextHandler: func(item interface{}) {
			line := item.(string)
			if strings.HasPrefix(line, "func ") {
				fmt.Println(line)
			}
		},
		ErrHandler: func(err error) {
			fmt.Printf("Encountered error: %v\n", err)
		},
		DoneHandler: func() {
			fmt.Println("Done!")
		},
	}

	// observableとobserverを接続（購読）
	sub := source.Subscribe(watcher)

	// observable
	go func() {
		content, err := ioutil.ReadFile("main.go")
		if err != nil {
			emitter <- err
		} else {
			for _, line := range strings.Split(string(content), "\n") {
				emitter <- line
			}
		}
		close(emitter)
	}()
	// 終了待ち
	<-sub
}
