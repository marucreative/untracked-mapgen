package prepare

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/marucreative/untracked-mapgen/util"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const precision = "13"
const basePath = "data/src/ned/"
const extractTo = basePath + "tmp/"
const outputTo = "data/processed/ned/"
const colorDefinitions = "config/color_relief.txt"

type Ned struct{}

func (n Ned) imageName(fileinfo os.FileInfo) string {
	src := extractTo + "img" + fileinfo.Name()
	return strings.Replace(src, ".zip", "_"+precision+".img", 1)
}

func (n Ned) extractImage(fileinfo os.FileInfo) bool {
	// read file into memory
	fmt.Println("\t", "reading file", fileinfo.Name())
	b, err := ioutil.ReadFile(basePath + fileinfo.Name())
	if err != nil {
		fmt.Printf("ERROR:\n%v", err)
		return false
	}
	reader := bytes.NewReader(b)

	// read the zip
	r, err := zip.NewReader(reader, fileinfo.Size())
	if err != nil {
		fmt.Printf("ERROR:\n%v", err)
		return false
	}

	// find the .img file
	var img string
	for _, zfile := range r.File {
		if strings.Contains(zfile.Name, ".img") {
			fmt.Println("\t\t", zfile.Name, zfile.FileInfo().Size(), img)

			dst, err := os.Create(extractTo + zfile.Name)
			if err != nil {
				fmt.Printf("ERROR:\n%v", err)
				return false
			}
			defer dst.Close()
			src, err := zfile.Open()
			if err != nil {
				fmt.Printf("ERROR:\n%v", err)
				return false
			}
			defer src.Close()

			io.Copy(dst, src)
		}
	}
	return true
}

func (n Ned) process(fileinfo os.FileInfo) {
	src := n.imageName(fileinfo)
	dest := outputTo + fileinfo.Name()

	fmt.Println("\t\t converting image to tif")
	tif := strings.Replace(src, ".img", "_raw.tif", 1)
	exec.Command("gdal_translate", "-of", "GTiff", src, tif).Run()

	fmt.Println("\t\t converting to mercator projection")
	warped := strings.Replace(src, ".img", ".tif", 1)
	exec.Command("gdalwarp", "-t_srs", "EPSG:3857", "-r", "bilinear", tif, warped).Run()

	// Use the tif file, not the img
	src = strings.Replace(src, ".img", ".tif", 1)

	fmt.Println("\t\t color relief")
	color := strings.Replace(dest, ".zip", "_color.tif", 1)
	exec.Command("gdaldem", "color-relief", src, colorDefinitions, color).Run()

	fmt.Println("\t\t hillshade")
	hillshade := strings.Replace(dest, ".zip", "_hillshade.tif", 1)
	exec.Command("gdaldem", "hillshade", src, hillshade, "-z", "5").Run() //, "-s", "111120")

	fmt.Println("\t\t contour")
	contour := strings.Replace(dest, ".zip", "_contour_50ft.shp", 1)
	exec.Command("gdal_contour", src, contour, "-a", "height", "-i", "15.24").Run()
}

func (n Ned) cleanup(fileinfo os.FileInfo) {
	if err := os.Remove(n.imageName(fileinfo)); err != nil {
		fmt.Println("ERROR:", err)
	}
}

func (n Ned) Run() {
	os.MkdirAll(extractTo, 0777)
	os.MkdirAll(outputTo, 0777)

	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	p := util.NewPool(3)
	for _, fileinfo := range files {
		if !strings.Contains(fileinfo.Name(), ".zip") {
			continue
		}
		wg.Add(1)
		go func(fileinfo os.FileInfo) {
			x := p.Borrow()
			if n.extractImage(fileinfo) {
				n.process(fileinfo)
				n.cleanup(fileinfo)
			}
			p.Return(x)
			wg.Done()
		}(fileinfo)
	}
	wg.Wait()
}
