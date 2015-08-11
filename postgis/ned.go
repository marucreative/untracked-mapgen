package postgis

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const basePath = "data/processed/ned/"
const outputTo = "data/postgis/"

func GenerateSQL(tableName, matcher, filename string) {
	fmt.Println(filename)
	list := []string{"-I", "-C"}
	files, _ := ioutil.ReadDir(basePath)
	for _, file := range files {
		if !strings.Contains(file.Name(), matcher) {
			continue
		}
		list = append(list, basePath+file.Name())
	}
	list = append(list, tableName)

	fmt.Println(list)
	cmd := exec.Command("raster2pgsql", list...)

	out, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	writer := bufio.NewWriter(out)

	sqlPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	go io.Copy(writer, sqlPipe)
	cmd.Wait()
}

func Ned() {
	os.MkdirAll(outputTo, 0777)

	// Generate color relief SQL
	fmt.Println("Generating color relief SQL")
	GenerateSQL("public.color_relief", "_color.tif", outputTo+"color_relief.sql")

	// Generate hill shading SQL
	fmt.Println("Generating hill shading SQL")
	GenerateSQL("public.hillshade", "_hillshade.tif", outputTo+"hillshade.sql")
}
