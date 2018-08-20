package main

import "github.com/hantaowang/dispatch/pkg/cmd"

func main() {
	cmd.Start(make(chan struct{}))
}
