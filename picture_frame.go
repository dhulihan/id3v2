// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"io"

	"github.com/bogem/id3v2/util"
)

// PictureFrame structure is used for picture frames (APIC).
// The information about how to add picture frame to tag you can
// see in the docs to tag.AddAttachedPicture function.
//
// Available picture types you can see in constants.
type PictureFrame struct {
	Encoding    Encoding
	MimeType    string
	PictureType byte
	Description string
	Picture     []byte
}

func (pf PictureFrame) Size() int {
	return 1 + len(pf.MimeType) + 1 + 1 + encodedSize(pf.Description, pf.Encoding) +
		len(pf.Encoding.TerminationBytes) + len(pf.Picture)
}

func (pf PictureFrame) WriteTo(w io.Writer) (n int64, err error) {
	var i int
	bw := getBufioWriter(w)
	defer putBufioWriter(bw)

	err = bw.WriteByte(pf.Encoding.Key)
	if err != nil {
		return
	}
	n++

	i, err = bw.WriteString(pf.MimeType)
	if err != nil {
		return
	}
	n += int64(i)

	err = bw.WriteByte(0)
	if err != nil {
		return
	}
	n++

	err = bw.WriteByte(pf.PictureType)
	if err != nil {
		return
	}
	n++

	i, err = encodeWriteText(bw, pf.Description, pf.Encoding)
	if err != nil {
		return
	}
	n += int64(i)

	i, err = bw.Write(pf.Encoding.TerminationBytes)
	if err != nil {
		return
	}
	n += int64(i)

	i, err = bw.Write(pf.Picture)
	if err != nil {
		return
	}
	n += int64(i)

	err = bw.Flush()
	return
}

func parsePictureFrame(rd io.Reader) (Framer, error) {
	bufRd := getUtilReader(rd)
	defer putUtilReader(bufRd)

	encodingKey, err := bufRd.ReadByte()
	if err != nil {
		return nil, err
	}
	encoding := getEncoding(encodingKey)

	mimeType, err := bufRd.ReadTillDelim(0)
	if err != nil {
		return nil, err
	}
	if _, err = bufRd.Discard(1); err != nil {
		return nil, err
	}

	pictureType, err := bufRd.ReadByte()
	if err != nil {
		return nil, err
	}

	description, err := bufRd.ReadTillDelims(encoding.TerminationBytes)
	if err != nil {
		return nil, err
	}
	if _, err = bufRd.Discard(len(encoding.TerminationBytes)); err != nil {
		return nil, err
	}

	picture, err := util.ReadAll(bufRd)
	if err != nil {
		return nil, err
	}

	pf := PictureFrame{
		Encoding:    encoding,
		MimeType:    string(mimeType),
		PictureType: pictureType,
		Description: decodeText(description, encoding),
		Picture:     picture,
	}

	return pf, nil
}
