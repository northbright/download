package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/northbright/httputil"
	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

// Downloader implements [github.com/northbright/iocopy.Task].
type Downloader struct {
	Url              string         `json:"url"`
	Dst              string         `json:"dst"`
	IsSizeKnown      bool           `json:"is_size_known"`
	Size             uint64         `json:"size,string"`
	IsRangeSupported bool           `json:"is_range_supported"`
	Downloaded       uint64         `json:"downloaded,string"`
	resp             *http.Response `json:"-"`
	f                *os.File       `json:"-"`
}

func New(url, dst string) (*Downloader, error) {
	resp, isSizeKnown, size, isRangeSupported, err := httputil.GetResp(url)
	if err != nil {
		return nil, err
	}

	dir := path.Dir(dst)
	if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
		return nil, err
	}

	f, err := os.Create(dst)
	if err != nil {
		return nil, err
	}

	d := &Downloader{
		Url:              url,
		Dst:              dst,
		IsSizeKnown:      isSizeKnown,
		Size:             size,
		IsRangeSupported: isRangeSupported,
		Downloaded:       0,
		resp:             resp,
		f:                f,
	}

	return d, nil
}

func Load(data []byte) (*Downloader, error) {
	var err error

	d := &Downloader{}

	if err = json.Unmarshal(data, d); err != nil {
		return nil, err
	}

	dir := path.Dir(d.Dst)
	if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
		return nil, err
	}

	d.f, err = os.OpenFile(d.Dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Check if it can resume downloading.
	if !d.IsRangeSupported {
		d.resp, d.IsSizeKnown, d.Size, d.IsRangeSupported, err = httputil.GetResp(d.Url)
		if err != nil {
			return nil, err
		}

		// Reset number of bytes downloaded to 0.
		d.Downloaded = 0
	} else {
		d.resp, _, err = httputil.GetRespOfRangeStart(d.Url, d.Downloaded)
		if err != nil {
			return nil, err
		}

		if _, err = d.f.Seek(int64(d.Downloaded), 0); err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (d *Downloader) Total() (bool, uint64) {
	return d.IsSizeKnown, d.Size
}

func (d *Downloader) Copied() uint64 {
	return d.Downloaded
}

func (d *Downloader) SetCopied(copied uint64) {
	d.Downloaded = copied
}

func (d *Downloader) Writer() io.Writer {
	return d.f
}

func (d *Downloader) Reader() io.Reader {
	return d.resp.Body
}

func (d *Downloader) Save() ([]byte, error) {
	return json.MarshalIndent(d, "", "    ")
}

func Do(ctx context.Context, url, dst string, bufSize uint) error {
	var (
		err = fmt.Errorf("unexpected behavior")
	)
	d, err := New(url, dst)
	if err != nil {
		return err
	}

	if bufSize == 0 {
		bufSize = iocopy.DefaultBufSize
	}

	iocopy.Do(
		ctx,
		d,
		bufSize,
		iocopy.DefaultInterval,
		func(isTotalKnown bool, total, copied, written uint64, percent float32) {
		},
		func(isTotalKnown bool, total, copied, written uint64, percent float32, cause error) {
			err = cause
		},
		func(isTotalKnown bool, total, copied, written uint64, percent float32) {
			err = nil
		},
		func(e error) {
			err = e
		},
	)
	return err
}
