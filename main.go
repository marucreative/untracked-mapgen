package main

import (
	"flag"
	"github.com/marucreative/untracked-mapgen/download"
	"github.com/marucreative/untracked-mapgen/postgis"
	"github.com/marucreative/untracked-mapgen/prepare"
	"sync"
)

var cmd string

func dld() {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() { download.Nhd(); wg.Done() }()
	go func() { download.Ned(); wg.Done() }()
	wg.Wait()
}

func prep() {
	prepare.Ned{}.Run()
}

func pg() {
	postgis.Ned()
}

func main() {
	flag.Parse()
	cmd := flag.Args()[0]
	switch cmd {
	case "download":
		dld()
	case "prepare":
		prep()
	case "postgis":
		pg()
	default:
		dld()
	}
}
