package download

import (
	"regexp"
)

func Ned() {
	re, _ := regexp.Compile(`.*\d\.zip`)
	NewFtpDownloader(
		"rockyftp.cr.usgs.gov",
		"/vdelivery/Datasets/Staged/NED/13/IMG",
		"data/src/ned",
	).Scan(1, re)
}
