package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ProgressTracker interface {
	SetTotal(total int64)
	SetReader(reader io.Reader)
	io.Reader
}

type Option func(*downloadOptions)

type downloadOptions struct {
	httpClient      *http.Client
	progressTracker ProgressTracker
}

func defaultDownloadOptions() *downloadOptions {
	return &downloadOptions{
		httpClient:      http.DefaultClient,
		progressTracker: nil,
	}
}

func WithHTTPClient(client *http.Client) func(*downloadOptions) {
	return func(d *downloadOptions) {
		d.httpClient = client
	}
}

func WithProgress(tracker ProgressTracker) func(*downloadOptions) {
	return func(d *downloadOptions) {
		d.progressTracker = tracker
	}
}

func DownloadFile(ctx context.Context, url string, path string, opt ...Option) error {
	opts := defaultDownloadOptions()
	for _, o := range opt {
		o(opts)
	}

	// Create a new HTTP get request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	// Perform the HTTP request.
	res, err := opts.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("making HTTP request: %w", err)
	}
	defer res.Body.Close()

	// Create the file to write the downloaded content to.
	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}

	var reader io.Reader = res.Body

	// If a progress tracker is provided, wrap the response body in a progress reader.
	if opts.progressTracker != nil {
		opts.progressTracker.SetReader(res.Body)
		opts.progressTracker.SetTotal(res.ContentLength)

		reader = opts.progressTracker
	}

	// Copy the response body to the output file.
	_, err = io.Copy(outFile, reader)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}
	defer outFile.Close()

	return nil
}
