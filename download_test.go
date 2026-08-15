package download_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/download"
)

func ExampleDownloadBufferWithProgress() {
	url := "https://golang.google.cn/dl/go1.23.1.darwin-amd64.pkg"

	c := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	// Set User-Agent.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36 Edg/118.0.2088.76")

	dst := filepath.Join(os.TempDir(), "go1.23.1.darwin-amd64.pkg")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*800)
	defer cancel()

	log.Printf("download.DownloadBufferWithProgress() starts...\nurl: %v\ndst: %v", url, dst)

	buf := make([]byte, 1024*640)

	n, err := download.DownloadBufferWithProgress(
		ctx,
		c,
		req,
		dst,
		buf,
		// Number of bytes downloaded previously.
		0,
		// Callback to report progress.
		func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		},
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("download.DownloadBufferWithProgress() error: %v", err)
			return
		}
		log.Printf("download.DownloadBufferWithProgress() stopped, cause: %v. %v bytes downloaded", err, n)
	} else {
		log.Printf("download.DownloadBufferWithProgress() OK, %v bytes downloaded", n)
		fmt.Printf("download successfully, total %v bytes downloaded", n)
		return
	}

	log.Printf("download.DownloadBufferWithProgress() starts again to resume downloading...\nurl: %v\ndst: %v\ndownloaded: %v", url, dst, n)

	// Resume the download by setting downloaded to n.
	n2, err := download.DownloadBufferWithProgress(
		context.Background(),
		c,
		req,
		dst,
		buf,
		// Number of bytes downloaded.
		n,
		// Callback to report progress.
		func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		},
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("download.DownloadBufferWithProgress() error: %v", err)
			return
		}
		log.Printf("download.DownloadBufferWithProgress() stopped, cause: %v. %v bytes downloaded", err, n2)
	} else {
		log.Printf("download.DownloadBufferWithProgress() OK, %v bytes downloaded", n2)
	}

	log.Printf("total %v bytes downloaded", n+n2)
	fmt.Printf("download successfully, total %v bytes downloaded", n+n2)

	// Remove the files after test's done.
	os.Remove(dst)

	// Output:
	// download successfully, total 75917313 bytes downloaded
}
