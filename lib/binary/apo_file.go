// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// ApoFile implement append-only writer that encode the data using
// [binary.Write] with [binary.BigEndian] order.
// The data to be written must support [binary.Write] (must contains
// fixed-size type).
// Type like string or map will not supported, so does struct with that
// field.
// To do that one need to implement [io.WriterTo] in the type.
//
// The file that writen by ApoFile have the following structure,
//
//	Apohead
//	* ApoMeta data
//	Apofoot
//
// Each data prepended by [ApoMeta] as metadata that contains the write
// operation, the time when its written, and kind of data being writen.
//
// The [ApoMeta] define the operation to data:
//
//   - [ApoOpInsert] operation insert new data, which should be unique among
//     others.
//   - [ApoOpUpdate] operation update indicates that the next data contains
//     update for previous inserted data.
//     The data being updated can be partial or all of it.
//   - [ApoOpReplace] operation replace indicated that the next data replace
//     whole previous inserted data.
//   - [ApoOpDelete] operation delete the previous inserted data.
//     Which record being deleted is defined inside the data (probably by
//     using some ID).
//
// The update and replace may seems duplicate.
// The update operation is provided to help the writer to write partial data
// when needed.
type ApoFile struct {
	file *os.File `noequal:""`

	name string
	foot apoFooter
	head apoHeader

	mtx sync.Mutex
}

// OpenApo open file for writing in append mode.
// If the file does not exist, it will be created.
// Once the file is opened it is ready for write-only.
//
// To open a file for reading use [ReadAofile].
func OpenApo(name string) (apo *ApoFile, err error) {
	var logp = `OpenApo`
	var isNew bool
	var openFlag = os.O_RDWR

	apo = &ApoFile{
		name: name,
	}
	_, err = os.Stat(name)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		openFlag |= os.O_CREATE
		isNew = true
	}

	apo.file, err = os.OpenFile(name, openFlag, 0600)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	if isNew {
		err = apo.init()
		if err != nil {
			_ = apo.Close()
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	} else {
		err = binary.Read(apo.file, binary.BigEndian, &apo.head)
		if err != nil {
			return nil, fmt.Errorf(`%s: read header: %w`, logp, err)
		}

		_, err = apo.file.Seek(apo.head.OffFoot, 0)
		if err != nil {
			return nil, fmt.Errorf(`%s: seek footer: %w`, logp, err)
		}

		_, err = apo.foot.ReadFrom(apo.file)
		if err != nil {
			return nil, fmt.Errorf(`%s: read footer: %w`, logp, err)
		}
	}

	return apo, nil
}

// Close the file.
func (apo *ApoFile) Close() (err error) {
	apo.mtx.Lock()
	err = apo.file.Close()
	apo.mtx.Unlock()
	return err
}

// ReadAll read all meta and data from file where all data has the same
// type.
// If data implement [io.ReaderFrom] it will use [io.ReaderFrom.ReadForm],
// otherwise it will use [binary.Read].
func (apo *ApoFile) ReadAll(data any) (list []ApoMetaData, err error) {
	var logp = `ReadAll`

	var hdrSize = int64(binary.Size(apo.head))
	_, err = apo.file.Seek(hdrSize, 0)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var meta ApoMeta
	for x := range apo.head.TotalData {
		err = binary.Read(apo.file, binary.BigEndian, &meta)
		if err != nil {
			return nil, fmt.Errorf(`%s: at %d: %w`, logp, x, err)
		}

		switch v := data.(type) {
		case io.ReaderFrom:
			_, err = v.ReadFrom(apo.file)
		default:
			err = binary.Read(apo.file, binary.BigEndian, data)
		}
		if err != nil {
			return nil, fmt.Errorf(`%s: at %d: %w`, logp, x, err)
		}

		list = append(list, ApoMetaData{
			Meta: meta,
			Data: data,
		})
	}
	return list, nil
}

// Write the meta and data into file.
// If the data is a type with non-fixed size, like slice,
// string, or map (or struct with non-fixed size field); it should implement
// [io.WriterTo], otherwise the write will fail.
func (apo *ApoFile) Write(meta ApoMeta, data any) (err error) {
	var (
		logp = `Write`
		buf  bytes.Buffer
	)

	if meta.At <= 0 {
		meta.At = timeNow().UnixNano()
	}

	switch v := data.(type) {
	case io.WriterTo:
		_, err = v.WriteTo(&buf)
		if err != nil {
			return fmt.Errorf(`%s: using io.WriterTo: %w`,
				logp, err)
		}

	default:
		err = binary.Write(&buf, binary.BigEndian, data)
		if err != nil {
			return fmt.Errorf(`%s: encode data: %w`, logp, err)
		}
	}

	apo.mtx.Lock()
	defer apo.mtx.Unlock()

	apo.head.TotalData++

	// Remember the current footer offset as the new meta-data index.
	apo.foot.idxMetaOff = append(apo.foot.idxMetaOff, apo.head.OffFoot)

	err = apo.commit(meta, buf.Bytes())
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

func (apo *ApoFile) commit(meta ApoMeta, data []byte) (err error) {
	// Move back to the offset of footer ...
	_, err = apo.file.Seek(apo.head.OffFoot, 0)
	if err != nil {
		return fmt.Errorf(`seek back %d: %w`, apo.head.OffFoot, err)
	}

	// write meta and data ...
	err = binary.Write(apo.file, binary.BigEndian, meta)
	if err != nil {
		return fmt.Errorf(`write meta: %w`, err)
	}

	_, err = apo.file.Write(data)
	if err != nil {
		return fmt.Errorf(`write data: %w`, err)
	}

	// get the current offset for new footer ...
	apo.head.OffFoot, err = apo.file.Seek(0, 1)
	if err != nil {
		return fmt.Errorf(`seek current: %w`, err)
	}

	// write footer ...
	_, err = apo.foot.WriteTo(apo.file)
	if err != nil {
		return fmt.Errorf(`write footer: %w`, err)
	}

	// ... and finally write the header.
	_, err = apo.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf(`seek header: %w`, err)
	}

	err = binary.Write(apo.file, binary.BigEndian, apo.head)
	if err != nil {
		return fmt.Errorf(`write header: %w`, err)
	}

	err = apo.file.Sync()
	if err != nil {
		return err
	}

	return nil
}

func (apo *ApoFile) init() (err error) {
	var logp = `init`

	apo.head.init()
	apo.head.OffFoot = int64(binary.Size(apo.head))

	err = binary.Write(apo.file, binary.BigEndian, apo.head)
	if err != nil {
		return fmt.Errorf(`%s: writing header: %w`, logp, err)
	}

	_, err = apo.foot.WriteTo(apo.file)
	if err != nil {
		return fmt.Errorf(`%s: writing footer: %w`, logp, err)
	}

	err = apo.file.Sync()
	if err != nil {
		return fmt.Errorf(`%s: on Sync: %w`, logp, err)
	}
	return nil
}
