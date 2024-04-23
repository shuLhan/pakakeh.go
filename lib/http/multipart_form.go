// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"sort"
	"strings"
)

// CreateMultipartFileHeader create [multipart.FileHeader] from raw content
// with optional filename.
func CreateMultipartFileHeader(filename string, content []byte) (fh *multipart.FileHeader, err error) {
	var (
		logp      = `NewMultipartFormFile`
		boundary  = `__boundary__`
		fieldname = `__fieldname__`

		buf bytes.Buffer
	)

	var wrt = multipart.NewWriter(&buf)

	err = wrt.SetBoundary(boundary)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var formfile io.Writer

	formfile, err = wrt.CreateFormFile(fieldname, filename)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	_, err = formfile.Write(content)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	err = wrt.Close()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		rdr = multipart.NewReader(&buf, boundary)

		form *multipart.Form
	)

	form, err = rdr.ReadForm(0)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		listFH []*multipart.FileHeader
		ok     bool
	)

	listFH, ok = form.File[fieldname]
	if !ok {
		return nil, fmt.Errorf(`%s: missing generated file %s`, logp, filename)
	}
	if len(listFH) == 0 {
		return nil, fmt.Errorf(`%s: empty generated FileHeader`, logp)
	}

	fh = listFH[0]

	return fh, nil
}

// GenerateFormData generate "multipart/form-data" body from mpform.
func GenerateFormData(mpform *multipart.Form) (contentType, body string, err error) {
	if mpform == nil {
		return ``, ``, nil
	}

	var (
		logp    = `GenerateFormData`
		sb      = new(strings.Builder)
		w       = multipart.NewWriter(sb)
		listKey = make([]string, 0, len(mpform.File))

		k string
	)

	// Generate files part.

	for k = range mpform.File {
		listKey = append(listKey, k)
	}
	sort.Strings(listKey)

	var (
		listFH  []*multipart.FileHeader
		fh      *multipart.FileHeader
		part    io.Writer
		file    multipart.File
		content []byte
	)
	for k, listFH = range mpform.File {
		for _, fh = range listFH {
			part, err = w.CreateFormFile(k, fh.Filename)
			if err != nil {
				return ``, ``, fmt.Errorf(`%s: %w`, logp, err)
			}

			file, err = fh.Open()
			if err != nil {
				return ``, ``, fmt.Errorf(`%s: %w`, logp, err)
			}

			content, err = io.ReadAll(file)
			if err != nil {
				return ``, ``, fmt.Errorf(`%s: %w`, logp, err)
			}

			_, err = part.Write(content)
			if err != nil {
				return ``, ``, fmt.Errorf(`%s: %w`, logp, err)
			}
		}
	}

	// Generate values part.

	listKey = listKey[:0]
	for k = range mpform.Value {
		listKey = append(listKey, k)
	}
	sort.Strings(listKey)

	var (
		listValue []string
		v         string
	)
	for _, k = range listKey {
		listValue = mpform.Value[k]
		for _, v = range listValue {
			part, err = w.CreateFormField(k)
			if err != nil {
				return ``, ``, err
			}

			_, err = part.Write([]byte(v))
			if err != nil {
				return ``, ``, err
			}
		}
	}

	err = w.Close()
	if err != nil {
		return ``, ``, err
	}

	contentType = w.FormDataContentType()

	return contentType, sb.String(), nil
}
