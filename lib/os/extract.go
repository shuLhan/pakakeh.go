// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	archiveKindTar int = 1 << iota
	archiveKindZip
)

const (
	compressKindBzip2 int = 1 << iota
	compressKindGzip
)

var (
	ErrExtractInputExt = errors.New("unknown extract input extension")
)

type extractor struct {
	dirOutput string

	compFile string // Path to compressed file.
	archFile string // Path to archived file.

	uncompName string // The base name of uncompressed file.
	uncompFile string // Path to uncompressed file (dirOutput + uncompName).

	archivedWith   int
	compressedWith int
}

// archiveTimes store the file access and modified times inside the archive.
type archiveTimes struct {
	accTime  time.Time
	modTime  time.Time
	filePath string
}

//
// Extract uncompress and/or unarchive file from fileInput into directory
// defined by dirOutput.
// This is the high level API that combine standard archive/zip, archive/tar,
// compress/bzip2, and/or compress/gzip.
//
// The compression and archive format is detected automatically based on the
// following fileInput extension:
//
//	* .bz2: decompress using compress/bzip2.
//	* .gz: decompress using compress/gzip.
//	* .tar: unarchive using archive/tar.
//	* .zip: unarchive using archive/zip.
//	* .tar.bz2: decompress using compress/bzip2 and unarchive using
//	archive/tar.
//	* .tar.gz: decompresss using compress/gzip and unarchive using
//	archive/tar.
//
// The output directory, dirOutput, where the decompressed and/or unarchived
// file stored will be created if not exist.
// If its empty, it will set to current directory.
//
// On success, the compressed and/or archived file will be removed from the
// file system.
//
func Extract(fileInput, dirOutput string) (err error) {
	var (
		logp     = "Extract"
		baseName = filepath.Base(fileInput)
		ext      = strings.ToLower(filepath.Ext(baseName))
		xtrk     = extractor{
			dirOutput: dirOutput,
		}
	)

	baseName = strings.TrimSuffix(baseName, ext)

	switch ext {
	case ".bz2":
		xtrk.compressedWith = compressKindBzip2
		xtrk.uncompName = baseName
		xtrk.compFile = fileInput

	case ".gz":
		xtrk.compressedWith = compressKindGzip
		xtrk.uncompName = baseName
		xtrk.compFile = fileInput

	case ".tar":
		xtrk.archivedWith = archiveKindTar
		xtrk.archFile = fileInput

	case ".zip":
		xtrk.archivedWith = archiveKindZip
		xtrk.archFile = fileInput

	default:
		return fmt.Errorf("%s: %s: %s: %w", logp, fileInput, ext, ErrExtractInputExt)
	}

	ext = strings.ToLower(filepath.Ext(baseName))
	if ext == ".tar" {
		xtrk.archivedWith = archiveKindTar
		xtrk.archFile = filepath.Join(filepath.Dir(fileInput), baseName)
	}

	if xtrk.compressedWith == 0 && xtrk.archivedWith == 0 {
		return fmt.Errorf("%s: %s: %w", logp, fileInput, ErrExtractInputExt)
	}

	if len(dirOutput) == 0 {
		xtrk.dirOutput, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}

	var (
		fin  *os.File
		fout *os.File
	)

	fin, err = os.Open(fileInput)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	switch xtrk.compressedWith {
	case compressKindBzip2:
		fout, err = xtrk.bunzip2(fin)
		if err != nil {
			_ = fin.Close()
			return fmt.Errorf("%s: %w", logp, err)
		}

		err = fin.Close()
		if err != nil {
			_ = fout.Close()
			return fmt.Errorf("%s: %w", logp, err)
		}

		fin = fout

	case compressKindGzip:
		fout, err = xtrk.gunzip(fin)
		if err != nil {
			_ = fin.Close()
			return fmt.Errorf("%s: %w", logp, err)
		}

		err = fin.Close()
		if err != nil {
			_ = fout.Close()
			return fmt.Errorf("%s: %w", logp, err)
		}

		fin = fout
	}

	switch xtrk.archivedWith {
	case archiveKindTar:
		err = xtrk.untar(fin)
	case archiveKindZip:
		err = xtrk.unzip(fin)
	}
	if err != nil {
		_ = fin.Close()
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = fin.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	// Remove the archive file and/or compress files.
	if len(xtrk.archFile) > 0 {
		err = os.Remove(xtrk.archFile)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}
	if len(xtrk.compFile) > 0 {
		err = os.Remove(xtrk.compFile)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}

	return nil
}

//
// bunzip2 uncompress the file input fin using bzip2.
// Since we did not know how large the output file and how much the caller
// memory, we store the output into the temporary file.
//
func (xtrk *extractor) bunzip2(fin *os.File) (fout *os.File, err error) {
	var (
		logp      = "uncompressWithBzip"
		bz2Reader io.Reader
	)

	xtrk.uncompFile = filepath.Join(xtrk.dirOutput, xtrk.uncompName)

	if xtrk.archivedWith == archiveKindTar {
		// Replace the archive file path with the output of
		// uncompress.
		xtrk.archFile = xtrk.uncompFile
	}

	fout, err = os.Create(xtrk.uncompFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	bz2Reader = bzip2.NewReader(bufio.NewReader(fin))

	_, err = io.Copy(fout, bz2Reader)
	if err != nil {
		_ = fout.Close()
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	// Reset the file output descriptor to the beginning.
	// Let the caller Close it later.
	_, err = fout.Seek(0, 0)
	if err != nil {
		_ = fout.Close()
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return fout, nil
}

func (xtrk *extractor) gunzip(fin *os.File) (fout *os.File, err error) {
	var (
		logp     = "gunzip"
		gzReader *gzip.Reader
	)

	gzReader, err = gzip.NewReader(bufio.NewReader(fin))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	xtrk.uncompFile = filepath.Join(xtrk.dirOutput, xtrk.uncompName)

	if xtrk.archivedWith == archiveKindTar {
		// Replace the archive file path with the output of
		// uncompress.
		xtrk.archFile = xtrk.uncompFile
	}

	fout, err = os.Create(xtrk.uncompFile)
	if err != nil {
		_ = gzReader.Close()
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	_, err = io.Copy(fout, gzReader)
	if err != nil {
		_ = fout.Close()
		_ = gzReader.Close()
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	err = gzReader.Close()
	if err != nil {
		_ = fout.Close()
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	// Reset the file output descriptor to the beginning.
	_, err = fout.Seek(0, 0)
	if err != nil {
		_ = fout.Close()
		return nil, err
	}

	return fout, nil
}

//
// untar the tar archive from fin, store the result into directory
// dirOutput.
//
func (xtrk *extractor) untar(fin io.Reader) (err error) {
	var (
		logp      = "untar"
		tarReader = tar.NewReader(fin)

		hdr       *tar.Header
		fi        fs.FileInfo
		outFile   *os.File
		filePath  string
		fileTimes []archiveTimes
	)

	for {
		hdr, err = tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("%s: %w", logp, err)
		}

		fi = hdr.FileInfo()
		filePath = filepath.Join(xtrk.dirOutput, hdr.Name)

		if fi.IsDir() {
			err = os.Mkdir(filePath, fi.Mode())
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}
		} else {
			outFile, err = os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}

			_, err = io.CopyN(outFile, tarReader, fi.Size())
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}

			err = outFile.Close()
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}
		}

		fileTimes = append(fileTimes, archiveTimes{
			filePath: filePath,
			accTime:  hdr.AccessTime,
			modTime:  fi.ModTime(),
		})
	}

	// Now that all files has been extracted, update the access and
	// modification times in reverse order.
	for x := len(fileTimes) - 1; x > 0; x-- {
		f := fileTimes[x]
		err = os.Chtimes(f.filePath, f.accTime, f.modTime)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}

	return nil
}

//
// unzip extract the zip archive from fin, store the result into directory
// dirOutput.
//
func (xtrk *extractor) unzip(fin *os.File) (err error) {
	var (
		logp = "unzip"

		fi        fs.FileInfo
		zipReader *zip.Reader
		inFile    io.ReadCloser
		filePath  string
		outFile   *os.File
		fileTimes []archiveTimes
	)

	fi, err = fin.Stat()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	zipReader, err = zip.NewReader(fin, fi.Size())
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	for _, zipFile := range zipReader.File {
		inFile, err = zipFile.Open()
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}

		fi = zipFile.FileInfo()
		filePath = filepath.Join(xtrk.dirOutput, zipFile.Name)

		if fi.IsDir() {
			err = os.Mkdir(filePath, fi.Mode())
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}
		} else {
			outFile, err = os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}

			_, err = io.CopyN(outFile, inFile, fi.Size())
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}

			err = outFile.Close()
			if err != nil {
				return fmt.Errorf("%s: %w", logp, err)
			}
		}

		fileTimes = append(fileTimes, archiveTimes{
			filePath: filePath,
			accTime:  fi.ModTime(),
			modTime:  fi.ModTime(),
		})
	}

	// Now that all files has been extracted, update the access and
	// modification times in reverse order.
	for x := len(fileTimes) - 1; x > 0; x-- {
		f := fileTimes[x]
		err = os.Chtimes(f.filePath, f.accTime, f.modTime)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}

	return nil
}
