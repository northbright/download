package download

import (
	"context"
	"io"
	"os"
	"path"
	"time"

	"github.com/northbright/httputil"
	"github.com/northbright/iocopy"
	"github.com/northbright/iocopy/progress"

	"github.com/northbright/pathelper"
)

type downloader struct {
	downloaded int64
	fn         OnDownloadFunc
	interval   time.Duration
}

// Option sets optional parameters to report download progress.
type Option func(dl *downloader)

// Downloaded returns an option to set the number of bytes downloaded previously.
// It's used to calculate the percent of downloading.
func Downloaded(downloaded int64) Option {
	return func(dl *downloader) {
		dl.downloaded = downloaded
	}
}

// OnDownloadFunc is the callback function when bytes are copied successfully.
// See [progress.OnWrittenFunc].
type OnDownloadFunc progress.OnWrittenFunc

// OnDownload returns an option to set callback to report progress.
func OnDownload(fn OnDownloadFunc) Option {
	return func(dl *downloader) {
		dl.fn = fn
	}
}

// OnDownloadInterval returns an option to set interval of the callback.
func OnDownloadInterval(d time.Duration) Option {
	return func(dl *downloader) {
		dl.interval = d
	}
}

// DownloadBuffer downloads content of remote URL to local file.
// It returns the number of bytes downloaded.
// ctx: [context.Context].
// url: remote URL.
// dst: local file.
// buf: buffer used to download.
// options: [Option] used to report progress.
func DownloadBuffer(ctx context.Context, url, dst string, buf []byte, options ...Option) (n int64, err error) {
	// Get info of remote URL.
	resp, size, rangeIsSupported, err := httputil.GetResp(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Create parent dir of dst if it does not exist.
	dir := path.Dir(dst)
	if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
		return 0, err
	}

	// Set optional parameters.
	dl := &downloader{}
	for _, option := range options {
		option(dl)
	}

	var f *os.File
	var reader io.Reader = resp.Body

	// Check if downloaded > 0.
	if dl.downloaded > 0 {
		if rangeIsSupported {
			// Get new response by range.
			resp2, _, err := httputil.GetRespOfRangeStart(url, dl.downloaded)
			if err != nil {
				return 0, err
			}
			defer resp2.Body.Close()

			// Update reader.
			reader = resp2.Body

			// Open dst file to with O_APPEND flag.
			if f, err = os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
				return 0, err
			}
			defer f.Close()

			// Set offset for dst file.
			if _, err = f.Seek(dl.downloaded, 0); err != nil {
				return 0, err
			}
		} else {
			// Reset download to 0 if range is not supported.
			dl.downloaded = 0
		}
	} else {
		// Set downloaded to 0 if it's negative.
		if dl.downloaded < 0 {
			dl.downloaded = 0
		}

		// Create dst file.
		if f, err = os.Create(dst); err != nil {
			return 0, err
		}
		defer f.Close()
	}

	var writer io.Writer = f

	// Check if callers need to report progress during IO copy.
	if dl.fn != nil {
		// Create a progress.
		p := progress.New(
			// Total size.
			size,
			// OnDownloadFunc
			progress.OnWrittenFunc(dl.fn),
			// Number of bytes copied previously.
			progress.Prev(dl.downloaded),
			// Interval to report progress.
			progress.Interval(dl.interval),
		)

		// Create a multiple writer and dupllicates writes to p.
		writer = io.MultiWriter(f, p)

		// Create a channel.
		// Send an empty struct to it to make progress goroutine exit.
		chExit := make(chan struct{}, 1)
		defer func() {
			chExit <- struct{}{}
		}()

		// Starts a new goroutine to report progress until ctx.Done() and chExit receive an empty struct.
		p.Start(ctx, chExit)
	}

	if buf != nil && len(buf) != 0 {
		return iocopy.CopyBuffer(ctx, writer, reader, buf)

	} else {
		return iocopy.Copy(ctx, writer, reader)
	}
}

// Download downloads content of remote URL to local file.
// It returns the number of bytes downloaded.
// ctx: [context.Context].
// url: remote URL.
// dst: local file.
// options: [Option] used to report progress.
func Download(ctx context.Context, url, dst string, options ...Option) (n int64, err error) {
	return DownloadBuffer(ctx, url, dst, nil, options...)
}
