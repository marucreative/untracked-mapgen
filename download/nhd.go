package download

import (
	"regexp"
)

func Nhd() {
	re, _ := regexp.Compile(`.*\d\.zip`)
	NewFtpDownloader(
		"nhdftp.usgs.gov",
		"/DataSets/Staged/SubRegions/FileGDB/HighResolution",
		"data/src/nhd",
	).Scan(1, re)
}
