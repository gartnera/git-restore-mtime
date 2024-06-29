package main

import (
	"log"
	"os"
)

func main() {
	m, err := NewManagerFromPath(os.Args[1], WithMaxDepth(1000))
	if err != nil {
		log.Fatal(err)
	}
	err = m.Run()
	if err != nil {
		log.Fatal(err)
	}
}
