package main

import (
	"context"

	"github.com/ysoding/spidey/cmd"
	"github.com/ysoding/spidey/spidey"
)

func test() {
	var events cmd.Events
	cnf := spidey.NewDefaultConfig(events)
	spidey.Run(context.Background(), &cnf)
}

func main() {
	// test()
	cmd.Execute()
}
