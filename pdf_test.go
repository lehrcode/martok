package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func writeTestFile(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDocumentHandler(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	writeTestFile(t, filepath.Join(first, "img", "a.txt"), "first a")
	writeTestFile(t, filepath.Join(first, "both.txt"), "first both")
	writeTestFile(t, filepath.Join(second, "img", "b.txt"), "second b")
	writeTestFile(t, filepath.Join(second, "both.txt"), "second both")

	server := httptest.NewServer(documentHandler([]byte("<p>document</p>"), filesHandler([]string{first, second})))
	defer server.Close()

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", http.StatusOK, "<p>document</p>"},
		{"/img/a.txt", http.StatusOK, "first a"},
		{"/img/b.txt", http.StatusOK, "second b"},
		{"/both.txt", http.StatusOK, "first both"},
		{"/missing.txt", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && string(body) != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

// TestPrintToPDF prints with a real browser and is skipped without one.
func TestPrintToPDF(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "img", "dot.svg"),
		`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"><rect width="10" height="10"/></svg>`)
	document := []byte(`<!DOCTYPE html><html><head><title>Test</title></head>` +
		`<body><h1>Heading</h1><p><img src="img/dot.svg" alt=""></p></body></html>`)

	for _, outline := range []bool{false, true} {
		var imageRequests atomic.Int32
		files := filesHandler([]string{dir})
		handler := documentHandler(document, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/img/dot.svg" {
				imageRequests.Add(1)
			}
			files.ServeHTTP(w, r)
		}))

		pdf, err := printToPDF(handler, "/", pdfOptions{Margin: 0.4, Outline: outline})
		if errors.Is(err, exec.ErrNotFound) {
			t.Skipf("no browser found: %v", err)
		}
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
			t.Fatalf("outline %v: output is not a PDF: %q", outline, pdf[:min(len(pdf), 16)])
		}
		if imageRequests.Load() == 0 {
			t.Errorf("outline %v: the embedded image was not requested", outline)
		}
		if got := bytes.Contains(pdf, []byte("/Outlines")); got != outline {
			t.Errorf("PDF has outline = %v, want %v", got, outline)
		}
	}
}
