package prepare

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

func (n Ned) extractImage(fileinfo os.FileInfo) {
	// read file into memory
	fmt.Println("\t", "reading file", fileinfo.Name())
	b, err := ioutil.ReadFile(basePath + fileinfo.Name())
	if err != nil {
		fmt.Printf("ERROR:\n%v", err)
		return
	}
	reader := bytes.NewReader(b)

	// read the zip
	r, err := zip.NewReader(reader, fileinfo.Size())
	if err != nil {
		fmt.Printf("ERROR:\n%v", err)
		return
	}

	// find the .img file
	var img string
	for _, zfile := range r.File {
		if strings.Contains(zfile.Name, ".img") {
			fmt.Println("\t\t", zfile.Name, zfile.FileInfo().Size(), img)

			dst, err := os.Create(extractTo + zfile.Name)
			if err != nil {
				fmt.Printf("ERROR:\n%v", err)
				continue
			}
			defer dst.Close()
			src, err := zfile.Open()
			if err != nil {
				fmt.Printf("ERROR:\n%v", err)
				continue
			}
			defer src.Close()

			io.Copy(dst, src)
		}
	}
}

func (n Ned) process(fileinfo os.FileInfo) {
	src := n.imageName(fileinfo)
	dest := outputTo + fileinfo.Name()

	fmt.Println("\t\t color relief")
	color := strings.Replace(dest, ".zip", "_color.tif", 1)
	cmd := exec.Command("gdaldem", "color-relief", src, colorDefinitions, color)
	cmd.Run()

	fmt.Println("\t\t hillshade")
	hillshade := strings.Replace(dest, ".zip", "_hillshade.tif", 1)
	cmd = exec.Command("gdaldem", "hillshade", src, hillshade, "-z", "5", "-s", "111120")
	cmd.Run()

	fmt.Println("\t\t contour")
	contour := strings.Replace(dest, ".zip", "_contour_50ft.shp", 1)
	cmd = exec.Command("gdal_contour", src, contour, "-a", "height", "-i", "15.24")
	cmd.Run()
}

func (n Ned) cleanup(fileinfo os.FileInfo) {
	if err := os.Remove(n.imageName(fileinfo)); err != nil {
		panic(err)
	}
}

func (n Ned) Run() {
	os.MkdirAll(extractTo, 0777)
	os.MkdirAll(outputTo, 0777)

	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		panic(err)
	}

	for _, fileinfo := range files {
		if !strings.Contains(fileinfo.Name(), ".zip") {
			continue
		}
		n.extractImage(fileinfo)
		n.process(fileinfo)
		n.cleanup(fileinfo)
	}
}
