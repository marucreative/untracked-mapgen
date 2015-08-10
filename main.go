package main

import (
	"github.com/marucreative/untracked-mapgen/download"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() { download.Nhd(); wg.Done() }()
	go func() { download.Ned(); wg.Done() }()

	wg.Wait()
}
