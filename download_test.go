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

/*
func ExampleNew() {
	var (
		savedData []byte
	)

	url := "https://golang.google.cn/dl/go1.22.2.darwin-amd64.pkg"
	dst := filepath.Join(os.TempDir(), "go1.22.2.darwin-amd64.pkg")

	// Create a new downloader.
	d, err := download.New(
		// Url
		url,
		// Destination
		dst,
	)
	if err != nil {
		log.Printf("download.New() error: %v", err)
		return
	}

	// Use a timeout to emulate that users stop the downloading.
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	log.Printf("start downloading...\nurl: %v\ndst: %v", url, dst)

	// Call task.Do to do the download task.
	task.Do(
		// Context.
		ctx,
		// task.Task. download.Downloader implements task.Task interface.
		d,
		// On bytes written.
		func(isTotalKnown bool, total, copied, written uint64, percent float32) {
			log.Printf("on written: %d/%d(%.2f%%)", copied, total, percent)
		},
		// On stop.
		func(isTotalKnown bool, total, copied, written uint64, percent float32, cause error) {
			log.Printf("on stop(%v): %d/%d(%.2f%%)", cause, copied, total, percent)
			// Save the state for resuming downloading.
			if savedData, err = d.Save(); err != nil {
				log.Printf("d.Save() error: %v", err)
				return
			}
			log.Printf("d.Save() successfully, savedData: %s", string(savedData))
		},
		// On ok.
		func(isTotalKnown bool, total, copied, written uint64, percent float32) {
			log.Printf("on ok: %d/%d(%.2f%%)", copied, total, percent)
		},
		// On error.
		func(err error) {
			log.Printf("on error: %v", err)
		},
		// Buffer size option.
		iocopy.BufSize(uint(64*1024)),
		// Refresh rate option for on written.
		iocopy.RefreshRate(time.Millisecond*100),
	)

	// Load downloader from the saved data.
	d, err = download.Load(savedData)
	if err != nil {
		log.Printf("download.Load() error: %v", err)
		return
	}
	log.Printf("load downloader from saved data successfully")

	ctx = context.Background()

	log.Printf("resume downloading...\nurl: %v\ndst: %v", url, dst)

	// Call task.Do to do the download task.
	task.Do(
		// Context.
		ctx,
		// task.Task. download.Downloader implements task.Task interface.
		d,
		// On bytes written.
		func(isTotalKnown bool, total, copied, written uint64, percent float32) {
			log.Printf("on written: %d/%d(%.2f%%)", copied, total, percent)
		},
		// On stop.
		func(isTotalKnown bool, total, copied, written uint64, percent float32, cause error) {
			log.Printf("on stop(%v): %d/%d(%.2f%%)", cause, copied, total, percent)
		},
		// On ok
		func(isTotalKnown bool, total, copied, written uint64, percent float32) {
			log.Printf("on ok: %d/%d(%.2f%%)", copied, total, percent)
		},
		// On error
		func(err error) {
			log.Printf("on error: %v", err)
		},
		// Buffer size option.
		iocopy.BufSize(uint(64*1024)),
		// Refresh rate option for on written.
		iocopy.RefreshRate(time.Millisecond*100),
	)

	// Remove the files after test's done.
	os.Remove(dst)

	// Output:
}

func ExampleDo() {
	url := "https://golang.google.cn/dl/go1.22.2.darwin-amd64.pkg"
	dst := filepath.Join(os.TempDir(), "go1.22.2.darwin-amd64.pkg")

	ctx := context.Background()

	log.Printf("download.Do() starts...\nurl: %v\ndst: %v", url, dst)

	if err := download.Do(ctx, url, dst); err != nil {
		log.Printf("download.Do() error: %v", err)
		return
	}

	log.Printf("download.Do() ok")

	// Remove the files after test's done.
	os.Remove(dst)

	// Output:
}
*/

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
