// Package scan walks a library folder and keeps the index in sync with it.
package scan

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"findbooks/internal/extract"
	"findbooks/internal/fb2"
	"findbooks/internal/index"
)

type Progress struct{ Done, Total int }

type FileError struct {
	RelPath string
	Err     error
}

type Report struct {
	Added, Updated, Removed, Unchanged int
	Errors                             []FileError
}

const batchSize = 200

var formats = map[string]string{
	".fb2": "fb2", ".pdf": "pdf", ".djvu": "djvu", ".djv": "djvu", ".doc": "doc",
	".docx": "docx", ".rtf": "rtf", ".txt": "txt", ".epub": "epub", ".mobi": "mobi",
}

const zipFB2 = ".fb2.zip"

func hasZipFB2Suffix(name string) bool {
	return len(name) >= len(zipFB2) && strings.EqualFold(name[len(name)-len(zipFB2):], zipFB2)
}

// formatOf returns the book format of a file name, or false for non-books.
func formatOf(name string) (string, bool) {
	if hasZipFB2Suffix(name) {
		return "fb2", true
	}
	f, ok := formats[strings.ToLower(filepath.Ext(name))]
	return f, ok
}

// stem strips the book extension (".fb2.zip" counts as one).
func stem(name string) string {
	if hasZipFB2Suffix(name) {
		return name[:len(name)-len(zipFB2)]
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}

type job struct {
	rel, abs, format string
	size, mtime      int64
	existed          bool
}

type result struct {
	rec     index.BookRecord
	existed bool
	err     error
}

// Run brings the index of one library in line with the folder at root:
// new and changed files (by size and mtime) are parsed in parallel and
// written in batches, vanished files are deleted. Unreadable files are
// indexed by file name and listed in Report.Errors.
func Run(ctx context.Context, st *index.Store, libraryID int64, root string, onProgress func(Progress)) (Report, error) {
	var rep Report
	stored, err := st.Files(libraryID)
	if err != nil {
		return rep, err
	}
	jobs, seen, err := collect(ctx, root, stored, &rep)
	if err != nil {
		return rep, err
	}
	var del []int64
	for rel, f := range stored {
		if !seen[rel] {
			del = append(del, f.BookID)
		}
	}
	rep.Removed = len(del)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := parseAll(ctx, jobs)

	batch := make([]index.BookRecord, 0, batchSize)
	flush := func() error {
		if len(batch) == 0 && len(del) == 0 {
			return nil
		}
		err := st.WriteBatch(libraryID, batch, del)
		batch, del = batch[:0], nil
		return err
	}
	done := 0
	for r := range results {
		done++
		if r.err != nil {
			rep.Errors = append(rep.Errors, FileError{RelPath: r.rec.RelPath, Err: r.err})
		}
		if r.existed {
			rep.Updated++
		} else {
			rep.Added++
		}
		batch = append(batch, r.rec)
		if len(batch) >= batchSize {
			if err := flush(); err != nil {
				cancel()
				for range results {
				}
				return rep, err
			}
		}
		if onProgress != nil {
			onProgress(Progress{Done: done, Total: len(jobs)})
		}
	}
	if err := ctx.Err(); err != nil {
		return rep, err
	}
	if err := flush(); err != nil {
		return rep, err
	}
	return rep, st.MarkScanned(libraryID, time.Now())
}

func collect(ctx context.Context, root string, stored map[string]index.FileStamp, rep *Report) ([]job, map[string]bool, error) {
	var jobs []job
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			if p == root {
				return err
			}
			rep.Errors = append(rep.Errors, FileError{RelPath: relPath(root, p), Err: err})
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if p != root && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		format, ok := formatOf(d.Name())
		if !ok {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			rep.Errors = append(rep.Errors, FileError{RelPath: relPath(root, p), Err: err})
			return nil
		}
		rel := relPath(root, p)
		seen[rel] = true
		old, existed := stored[rel]
		size, mtime := info.Size(), info.ModTime().Unix()
		if existed && old.Size == size && old.MTime == mtime {
			rep.Unchanged++
			return nil
		}
		jobs = append(jobs, job{rel: rel, abs: p, format: format, size: size, mtime: mtime, existed: existed})
		return nil
	})
	return jobs, seen, err
}

func relPath(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(r)
}

func parseAll(ctx context.Context, jobs []job) <-chan result {
	jobCh := make(chan job)
	results := make(chan result)
	go func() {
		defer close(jobCh)
		for _, j := range jobs {
			select {
			case jobCh <- j:
			case <-ctx.Done():
				return
			}
		}
	}()
	var wg sync.WaitGroup
	for range runtime.NumCPU() {
		wg.Go(func() {
			for j := range jobCh {
				select {
				case results <- parse(j):
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

// failedSize is stored as the size of a book whose parse failed, so the
// next scan sees a size mismatch and retries it (e.g. after a transient
// read error or permission problem).
const failedSize = -1

func parse(j job) (r result) {
	title := stem(path.Base(j.rel))
	fallback := index.BookRecord{
		RelPath: j.rel, Format: j.format, Size: j.size, MTime: j.mtime, Title: title,
		Works: []index.WorkRecord{{Title: title, TreePath: title}},
	}
	r = result{rec: fallback, existed: j.existed}
	if j.format != "fb2" {
		return r
	}
	failed := fallback
	failed.Size = failedSize
	defer func() {
		if p := recover(); p != nil {
			r = result{rec: failed, existed: j.existed, err: fmt.Errorf("parser panic: %v", p)}
		}
	}()
	b, err := fb2.ParseFile(j.abs)
	if err != nil {
		r.rec, r.err = failed, err
		return r
	}
	if b.Title == "" {
		b.Title = title
	}
	works, isCollection := extract.Works(b)
	rec := index.BookRecord{
		RelPath: j.rel, Format: j.format, Size: j.size, MTime: j.mtime,
		Title: b.Title, Authors: b.AuthorNames(), Year: b.Year, IsCollection: isCollection,
	}
	for _, w := range works {
		rec.Works = append(rec.Works, index.WorkRecord{Title: w.Title, Author: w.Author, TreePath: w.TreePath})
	}
	r.rec = rec
	return r
}
