package main

import (
	"fmt"

	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
)

// メッセージ
type hello struct{ Who string }

// アクター
type helloActor struct{}

func (state *helloActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *hello:
		fmt.Printf("Hello %v\n", msg.Who)
	}
}

func main() {
	props := actor.FromInstance(&helloActor{})
	pid := actor.Spawn(props)
	pid.Tell(&hello{Who: "Roger"})
	pid.Tell(&hello{Who: "World"})
	console.ReadLine()
}
