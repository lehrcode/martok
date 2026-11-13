package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// pdfTimeout limits how long Chrome may take to load and print a document.
const pdfTimeout = 60 * time.Second

type pdfOptions struct {
	// Margin is the width of the four page margins in inches.
	Margin float64
	// Outline embeds the document outline (bookmarks) of the headings.
	Outline bool
}

// overlayDirs is a file system of several directories. A file is opened from
// the first directory that contains it.
type overlayDirs []http.Dir

func (dirs overlayDirs) Open(name string) (http.File, error) {
	var firstErr error
	for _, dir := range dirs {
		file, err := dir.Open(name)
		if err == nil {
			return file, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return nil, firstErr
}

// filesHandler serves the files of the directories.
func filesHandler(dirs []string) http.Handler {
	var overlay overlayDirs
	for _, dir := range dirs {
		overlay = append(overlay, http.Dir(dir))
	}
	return http.FileServer(overlay)
}

// documentHandler serves the HTML document as the root of the given handler,
// so that the relative URIs of the document are resolved by it.
func documentHandler(document []byte, files http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(document)
	})
}

// printToPDF lets a headless Chrome print the page at the URL path of the
// handler. The handler is served by a temporary web server that is reachable
// from this machine only, on a free port chosen by the operating system.
func printToPDF(handler http.Handler, path string, opts pdfOptions) ([]byte, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting the web server: %w", err)
	}
	host := listener.Addr().String()
	server := &http.Server{Handler: handler}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// The browser lives as long as the context of the first call, so it is
	// started before the timeout is applied.
	if err := chromedp.Do(ctx); err != nil {
		return nil, fmt.Errorf("starting Chrome: %w", err)
	}
	ctx, cancelTimeout := context.WithTimeout(ctx, pdfTimeout)
	defer cancelTimeout()

	// The page is printed when its network has been idle for a moment and not
	// at the load event, because scripts like Mermaid fetch further modules
	// and render the diagrams after the page has loaded.
	lifecycle := chromedp.Events(ctx, page.LifecycleEvent)
	navigation, err := chromedp.Call(ctx, page.Navigate, page.NavigateParams{URL: "http://" + host + path})
	if err != nil {
		return nil, fmt.Errorf("loading the document: %w", err)
	}
	if navigation.ErrorText != "" {
		return nil, fmt.Errorf("loading the document: %s", navigation.ErrorText)
	}
	for event, err := range lifecycle {
		if err != nil {
			return nil, fmt.Errorf("waiting for the document to load: %w", err)
		}
		if event.LoaderID == navigation.LoaderID && event.Name == "networkIdle" {
			break
		}
	}

	printOptions := []chromedp.PDFOption{
		// The default paper size, a stylesheet can set another one with the
		// CSS rule @page.
		chromedp.PDFPaper(chromedp.PaperA4),
		chromedp.PDFMargin(opts.Margin),
		chromedp.PDFPrintBackground(),
		chromedp.PDFPreferCSSPageSize(),
		chromedp.PDFStream(),
	}
	if opts.Outline {
		printOptions = append(printOptions, chromedp.PDFOutlineAndTagged())
	}
	pdf, err := chromedp.Run(ctx, chromedp.PrintToPDF(printOptions...))
	if err != nil {
		return nil, fmt.Errorf("printing the document: %w", err)
	}
	return pdf, nil
}
