package download

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/northbright/httputil"
	"github.com/northbright/iocopy"
	"github.com/northbright/iocopy/progress"

	"github.com/northbright/pathelper"
)

// Download downloads content of remote URL to local file.
// It returns the number of bytes downloaded.
// ctx: [context.Context].
// url: remote URL.
// dst: local file.
// downloaded: number of bytes downloaded previously. It's used to resume previous download.
// options: [progress.Option] used to report progress.
func Download(ctx context.Context, url, dst string, downloaded int64, options ...progress.Option) (n int64, err error) {
	resp, sizeIsKnown, size, rangeIsSupported, err := httputil.GetResp(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	dir := path.Dir(dst)
	if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
		return 0, err
	}

	f, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	if downloaded > 0 {
		if rangeIsSupported {
			// Range is supported.
			// Close d.resp.Body()
			resp.Body.Close()

			// Get new response by range.
			resp, _, err = httputil.GetRespOfRangeStart(url, downloaded)
			if err != nil {
				return 0, err
			}

			// Close dst file.
			f.Close()

			// Reopen dst file to with O_APPEND flag.
			if f, err = os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
				return 0, err
			}

			// Set offset for dst file.
			if _, err = f.Seek(downloaded, 0); err != nil {
				return 0, err
			}
		}
	}

	// Check if callers need to report progress during IO copy.
	if len(options) > 0 {
		// Pass -1 as size to progress.New() when total size is unknown.
		if !sizeIsKnown {
			size = -1
		}
		// Create a progress.
		p := progress.New(
			// Total size.
			size,
			// Number of bytes copied previously.
			downloaded,
			// Options: OnWrittenFunc, Interval.
			options...,
		)

		// Create a multiple writen and dupllicates writes to p.
		mw := io.MultiWriter(f, p)

		// Create a channel.
		// Send an empty struct to it to make progress goroutine exit.
		chExit := make(chan struct{}, 1)
		defer func() {
			chExit <- struct{}{}
		}()

		// Starts a new goroutine to report progress until ctx.Done() and chExit receive an empty struct.
		p.Start(ctx, chExit)
		return iocopy.Copy(ctx, mw, resp.Body)
	} else {
		return iocopy.Copy(ctx, f, resp.Body)
	}
}
