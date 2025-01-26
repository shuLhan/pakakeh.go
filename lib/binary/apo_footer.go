// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"encoding/binary"
	"io"
)

// apoFooter contains dynamic meta data for single Apo file.
type apoFooter struct {
	// idxMetaOff contains the offset of ApoMeta.
	idxMetaOff []int64
}

func (foot *apoFooter) WriteTo(w io.Writer) (n int64, err error) {
	var nidx int64 = int64(len(foot.idxMetaOff))
	err = binary.Write(w, binary.BigEndian, nidx)
	if err != nil {
		return 0, err
	}
	var sizei64 = int64(binary.Size(nidx))
	n = sizei64
	for _, off := range foot.idxMetaOff {
		err = binary.Write(w, binary.BigEndian, off)
		if err != nil {
			return n, err
		}
		n += sizei64
	}
	return n, nil
}

func (foot *apoFooter) ReadFrom(r io.Reader) (n int64, err error) {
	var nidx int64
	err = binary.Read(r, binary.BigEndian, &nidx)
	if err != nil {
		return 0, err
	}
	var (
		off  int64
		size = int64(binary.Size(off))
	)
	for range nidx {
		err = binary.Read(r, binary.BigEndian, &off)
		if err != nil {
			return n, err
		}
		foot.idxMetaOff = append(foot.idxMetaOff, off)
		n += size
	}
	return n, nil
}
