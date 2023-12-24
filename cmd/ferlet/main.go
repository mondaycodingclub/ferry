package main

import "ferry/pkg/ferlet"

func main() {
	ferlet.NewFerlet(ferlet.Config{
		Name:       "node1",
		ServerHost: "123.207.28.113",
		ServerPort: "8020",
	}).Run()
}
