package download_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/download"
	"github.com/northbright/iocopy/progress"
)

func ExampleDownload() {
	// Example 1. Download a remote file with reporting progress.
	url := "https://golang.google.cn/dl/go1.23.1.darwin-amd64.pkg"
	dst := filepath.Join(os.TempDir(), "go1.23.1.darwin-amd64.pkg")

	ctx := context.Background()

	log.Printf("download.Download() starts...\nurl: %v\ndst: %v", url, dst)

	n, err := download.Download(
		// Context.
		ctx,
		// URL to download.
		url,
		// Destination.
		dst,
		// Number of bytes downloaded previously.
		0,
		// OnWrittenFunc to report progress.
		progress.OnWritten(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		}),
	)

	if err != nil {
		log.Printf("download.Download() error: %v", err)
		return
	}

	log.Printf("download.Download() OK, %v bytes downloaded", n)

	// Remove the files after test's done.
	os.Remove(dst)

	// Example 2. Stop a download and resume it.
	ctx2, cancel := context.WithTimeout(context.Background(), time.Millisecond*800)
	defer cancel()

	log.Printf("download.Download() starts...\nurl: %v\ndst: %v", url, dst)

	n, err = download.Download(
		// Context.
		ctx2,
		// URL to download.
		url,
		// Destination.
		dst,
		// Number of bytes downloaded previously.
		0,
		// OnWrittenFunc to report progress.
		progress.OnWritten(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("download.Download() error: %v", err)
			return
		}
		log.Printf("download.Download() stopped, cause: %v. %v bytes downloaded", err, n)
	}

	log.Printf("call download.Download again to resume the download, set downloaded to %v", n)

	// Resume the download by set downloaded to n.
	n2, err := download.Download(
		// Context.
		context.Background(),
		// URL to download.
		url,
		// Destination.
		dst,
		// Number of bytes downloaded previously.
		n,
		// OnWrittenFunc to report progress.
		progress.OnWritten(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("download.Download() error: %v", err)
			return
		}
		log.Printf("download.Download() stopped, cause: %v. %v bytes downloaded", err, n2)
		return
	}

	log.Printf("download.Download() OK, %v bytes downloaded, total: %v bytes downloaded", n2, n+n2)

	// Remove the files after test's done.
	os.Remove(dst)

	// Output:
}
