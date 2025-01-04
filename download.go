package download

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/northbright/httputil"
	"github.com/northbright/iocopy"

	"github.com/northbright/pathelper"
)

// DownloadBufferWithProgress downloads content of remote URL to local file and returns the number of bytes downloaded.
// It accepts [context.Context] to make download cancalable.
// It also accepts callback function on bytes written to report progress.
// downloaded: number of bytes downloaded previously.
// It can be used to resume the download.
// 1. Set downloaded to 0 when call DownloadBufferWithProgress for the first time.
// 2. User stops the download and DownloadBufferWithProgress returns the number of bytes downloaded and error.
// 3. Check if err == context.Canceled || err == context.DeadlineExceeded.
// 4. Set downloaded to the "n" return value of previous DownloadBufferWithProgress when make next call to resume the download.
// fn: callback on bytes written.
func DownloadBufferWithProgress(
	ctx context.Context,
	url string,
	dst string,
	buf []byte,
	downloaded int64,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
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

	var f *os.File
	var reader io.Reader = resp.Body

	// Check if downloaded > 0.
	if downloaded > 0 {
		if rangeIsSupported {
			// Get new response by range.
			resp2, _, err := httputil.GetRespOfRangeStart(url, downloaded)
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
			if _, err = f.Seek(downloaded, 0); err != nil {
				return 0, err
			}
		} else {
			// Reset download to 0 if range is not supported.
			downloaded = 0
		}
	} else {
		// Set downloaded to 0 if it's negative.
		if downloaded < 0 {
			downloaded = 0
		}

		// Create dst file.
		if f, err = os.Create(dst); err != nil {
			return 0, err
		}
		defer f.Close()
	}

	return iocopy.CopyBufferWithProgress(ctx, f, reader, buf, size, downloaded, fn)
}

// Download downloads content of remote URL to local file and returns the number of bytes downloaded.
// It accepts [context.Context] to make download cancalable.
func Download(ctx context.Context, url, dst string) (n int64, err error) {
	return DownloadBufferWithProgress(ctx, url, dst, nil, 0, nil)
}

// DownloadBuffer is the buffered version of [Download].
func DownloadBuffer(ctx context.Context, url, dst string, buf []byte) (n int64, err error) {
	return DownloadBufferWithProgress(ctx, url, dst, buf, 0, nil)
}

// DownloadWithProgress is the non-buffered version of [DownloadBufferWithProgress].
func DownloadWithProgress(
	ctx context.Context,
	url string,
	dst string,
	downloaded int64,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	return DownloadBufferWithProgress(ctx, url, dst, nil, downloaded, fn)
}
