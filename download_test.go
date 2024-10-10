package download_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/download"
)

func ExampleDownload() {
	// Example 1. Download a remote file with reporting progress.
	log.Printf("\n============ Example 1 Begin ============")

	url := "https://golang.google.cn/dl/go1.23.1.darwin-amd64.pkg"
	dst := filepath.Join(os.TempDir(), "go1.23.1.darwin-amd64.pkg")

	log.Printf("download.Download() starts...\nurl: %v\ndst: %v", url, dst)
	n, err := download.Download(
		// Context.
		context.Background(),
		// URL to download.
		url,
		// Destination.
		dst,
		// Option to set OnDownloadFunc to report progress.
		download.OnDownload(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("download.Download() error: %v", err)
			return
		}
		log.Printf("download.Download() stopped, cause: %v. %v bytes downloaded", err, n)
	} else {
		log.Printf("download.Download() OK, %v bytes downloaded", n)
	}

	// Remove the files after test's done.
	os.Remove(dst)

	log.Printf("\n------------ Example 1 End ------------")

	// Example 2. Stop a download and resume it.
	log.Printf("\n============ Example 2 Begin ============")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*800)
	defer cancel()

	log.Printf("download.Download() starts...\nurl: %v\ndst: %v", url, dst)
	n, err = download.Download(
		// Context.
		ctx,
		// URL to download.
		url,
		// Destination.
		dst,
		// Option to set OnDownloadFunc to report progress.
		download.OnDownload(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("download.Download() error: %v", err)
			return
		}
		log.Printf("download.Download() stopped, cause: %v. %v bytes downloaded", err, n)
	} else {
		log.Printf("download.Download() OK, %v bytes downloaded", n)
	}

	log.Printf("download.Download() starts again to resume downloading...\nurl: %v\ndst: %v\ndownloaded: %v", url, dst, n)
	// Resume the download by set downloaded to n.
	n2, err := download.Download(
		// Context.
		context.Background(),
		// URL to download.
		url,
		// Destination.
		dst,
		// Option to set number of bytes downloaded previously.
		download.Downloaded(n),
		// Option to set OnDownloadFunc to report progress.
		download.OnDownload(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) downloaded", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("download.Download() error: %v", err)
			return
		}
		log.Printf("download.Download() stopped, cause: %v. %v bytes downloaded", err, n2)
	} else {
		log.Printf("download.Download() OK, %v bytes downloaded", n2)
	}

	log.Printf("total %v bytes downloaded", n+n2)

	// Remove the files after test's done.
	os.Remove(dst)

	log.Printf("\n------------ Example 2 End ------------")

	// Output:
}
