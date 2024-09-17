package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultRetries         = 3
	defaultFilePermissions = 0o644
	defaultBufferSize      = 32 * 1024
	defaultInterval        = 500 * time.Millisecond
)

// ProgressTracker defines an interface for tracking download progress.
type ProgressTracker interface {
	Update(downloaded int64, total int64)
}

// Option defines a function type for setting download options.
type Option func(*downloadOptions)

type downloadOptions struct {
	httpClient      *http.Client
	progressTracker ProgressTracker
	bufferSize      int
	filePerm        os.FileMode
	retries         int
}

// defaultDownloadOptions returns a new downloadOptions with default values.
func defaultDownloadOptions() *downloadOptions {
	return &downloadOptions{
		httpClient:      http.DefaultClient,
		progressTracker: nil,
		bufferSize:      defaultBufferSize,
		filePerm:        defaultFilePermissions,
		retries:         defaultRetries,
	}
}

// WithHTTPClient sets a custom HTTP client for the download.
func WithHTTPClient(client *http.Client) Option {
	return func(d *downloadOptions) {
		d.httpClient = client
	}
}

// WithProgress sets a progress tracker for the download.
func WithProgress(tracker ProgressTracker) Option {
	return func(d *downloadOptions) {
		d.progressTracker = tracker
	}
}

// WithFilePermissions sets the file permissions for the downloaded file.
func WithFilePermissions(perm os.FileMode) Option {
	return func(d *downloadOptions) {
		d.filePerm = perm
	}
}

// WithBufferSize sets a custom buffer size for the copy operation.
func WithBufferSize(size int) Option {
	return func(d *downloadOptions) {
		d.bufferSize = size
	}
}

// WithRetries sets the number of retries for transient errors.
func WithRetries(retries int) Option {
	return func(d *downloadOptions) {
		d.retries = retries
	}
}

// DownloadFile downloads a file from the specified URL to the given path.
func DownloadFile(ctx context.Context, url string, path string, opt ...Option) error {
	if err := validateInputParameters(ctx, url, path); err != nil {
		return err
	}

	opts := defaultDownloadOptions()
	for _, o := range opt {
		o(opts)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("getting absolute path for %s: %w", path, err)
	}

	// Create the output directory if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating directories for %s: %w", absPath, err)
	}

	// Perform the HTTP request with retries.
	res, err := executeRequestWithRetries(ctx, url, opts)
	if err != nil {
		return fmt.Errorf("performing HTTP request: %w", err)
	}

	// Create a temporary file for downloading
	tempFile, err := os.CreateTemp(filepath.Dir(absPath), "downloader-*")
	if err != nil {
		res.Body.Close()

		return fmt.Errorf("creating temporary file for %s: %w", absPath, err)
	}

	// Ensure the temp file is closed properly even if an error occurs.
	defer func() {
		tempFile.Close()
		// Remove the temp file if an error occurred.
		if err != nil {
			os.Remove(tempFile.Name())
		}
	}()

	// Perform the download
	if err := performDownload(ctx, res, tempFile, opts); err != nil {
		// Cleanup temp file on error
		tempFile.Close()
		os.Remove(tempFile.Name())

		return err
	}
	defer res.Body.Close()

	// Finalize the download
	if err := finalizeDownload(tempFile, absPath, opts); err != nil {
		return err
	}

	return nil
}

// validateInputParameters checks if the input parameters are valid.
func validateInputParameters(ctx context.Context, url, path string) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}

	if url == "" {
		return errors.New("url must not be empty")
	}

	if path == "" {
		return errors.New("path must not be empty")
	}

	return nil
}

// executeRequestWithRetries performs the HTTP request with retry logic.
//
// nolint:cyclop
func executeRequestWithRetries(ctx context.Context, url string, opts *downloadOptions) (*http.Response, error) {
	var res *http.Response

	var err error

	// Create a new HTTP GET request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	for attempt := range opts.retries {
		// Perform the HTTP request
		res, err = opts.httpClient.Do(req)
		if err == nil && (res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices) {
			err = fmt.Errorf("received non-success HTTP status code %d: %s", res.StatusCode, res.Status)
		}

		if err == nil {
			return res, nil
		}

		// Close response body if it was opened
		if res != nil {
			res.Body.Close()
		}

		// Check if we should retry
		if attempt < opts.retries-1 {
			// Exponential backoff
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second //nolint:mnd
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return nil, ctx.Err() //nolint:wrapcheck
			}
		}
	}

	return nil, fmt.Errorf("failed to download %s after %d attempts: %w", url, opts.retries, err)
}

// performDownload handles the actual downloading of the file content.
func performDownload(ctx context.Context, res *http.Response, tempFile *os.File, opts *downloadOptions) error {
	length := res.ContentLength // Can be -1 if the length is unknown.

	// Wrap the response body with a context-aware reader
	bodyReader := newContextReader(ctx, res.Body)

	// Create a progress writer
	progress := &progressWriter{
		total:    length,
		callback: opts.progressTracker,
	}

	// Create a TeeReader to track progress
	reader := io.TeeReader(bodyReader, progress)

	// Copy the data to the temp file
	if _, err := io.CopyBuffer(tempFile, reader, make([]byte, opts.bufferSize)); err != nil {
		return fmt.Errorf("downloading content: %w", err)
	}

	return nil
}

// finalizeDownload renames the temporary file and sets file permissions.
func finalizeDownload(tempFile *os.File, absPath string, opts *downloadOptions) error {
	// Close the temp file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("closing temporary file %s: %w", tempFile.Name(), err)
	}

	// Rename the temporary file to the final path.
	if err := os.Rename(tempFile.Name(), absPath); err != nil {
		return fmt.Errorf("moving temporary file to %s: %w", absPath, err)
	}

	// Set file permissions after renaming
	if err := os.Chmod(absPath, opts.filePerm); err != nil {
		return fmt.Errorf("setting permissions for %s: %w", absPath, err)
	}

	return nil
}

// contextReader wraps an io.Reader and checks for context cancellation before each read.
type contextReader struct {
	// Context is fine because it's lifecycle is managed by the caller.
	ctx    context.Context //nolint:containedctx
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err() //nolint:wrapcheck
	default:
		return r.reader.Read(p) //nolint:wrapcheck
	}
}

// newContextReader returns a context-aware reader that checks for context cancellation before each read.
func newContextReader(ctx context.Context, reader io.Reader) io.Reader {
	return &contextReader{
		ctx:    ctx,
		reader: reader,
	}
}

// progressWriter tracks the number of bytes written and reports progress at specified intervals.
type progressWriter struct {
	downloaded   int64
	total        int64
	lastReported time.Time
	callback     ProgressTracker
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.downloaded += int64(n)

	if time.Since(w.lastReported) >= defaultInterval {
		if w.callback != nil {
			w.callback.Update(w.downloaded, w.total)
		}

		w.lastReported = time.Now()
	}

	return n, nil
}

type SimpleTracker struct {
	w     io.Writer
	title string
}

func NewSimpleTracker(w io.Writer, title string) ProgressTracker {
	return &SimpleTracker{w: w, title: title}
}

func (t *SimpleTracker) Update(downloaded int64, total int64) {
	dl := float64(downloaded)
	tl := float64(total)

	percentage := dl / tl * 100      //nolint:mnd
	downloadedMB := dl / 1024 / 1024 //nolint:mnd
	totalMB := tl / 1024 / 1024      //nolint:mnd

	fmt.Fprintf(t.w, "%s: %.2f%% (%.2f MB / %.2f MB)\r", t.title, percentage, downloadedMB, totalMB)
}
