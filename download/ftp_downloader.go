package download

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/marucreative/mapgen2/util"
	"io"
	"os"
	"regexp"
	"sync"
)

const nedHost = "rockyftp.cr.usgs.gov:21"
const nedBaseDir = "/vdelivery/Datasets/Staged/NED/13/IMG"

type FtpDownloader struct {
	conn *ftp.ServerConn
	dest string
	host string
}

func (f FtpDownloader) Download(entry *ftp.Entry) {
	filename := f.dest + "/" + entry.Name
	if info, err := os.Stat(filename); err == nil {
		if info.Size() == int64(entry.Size) {
			fmt.Printf("Skipping %s, already downloaded\n", entry.Name)
			return
		}
	}

	fmt.Printf("Downloading %s / %s - %v\n", f.host, entry.Name, entry.Size)
	rc, err := f.conn.Retr(entry.Name)
	if err != nil {
		fmt.Println(err)
	}
	defer rc.Close()

	dst, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer dst.Close()

	io.Copy(dst, rc)
}

func (f FtpDownloader) Scan(threads int, matcher *regexp.Regexp) {
	// Scan for and download proper files
	// Note that this server doesn't like parallel downloads
	p := util.NewPool(threads)
	var i uint64
	var wg sync.WaitGroup
	entries, err := f.conn.List(".")
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloading files")
	for _, entry := range entries {
		if matcher.MatchString(entry.Name) {
			wg.Add(1)
			x := p.Borrow()
			go func() {
				f.Download(entry)
				p.Return(x)
				wg.Done()
			}()
			i += entry.Size
		}
	}
	wg.Wait()
}

func NewFtpDownloader(host string, src string, dest string) FtpDownloader {
	os.MkdirAll(dest, 0777)

	conn, err := ftp.Dial(host + ":21")
	if err != nil {
		panic(err)
	}

	err = conn.Login("anonymous", "")
	if err != nil {
		panic(err)
	}

	err = conn.ChangeDir(src)
	if err != nil {
		panic(err)
	}

	return FtpDownloader{host: host, conn: conn, dest: dest}
}
