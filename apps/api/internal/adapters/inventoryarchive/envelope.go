package inventoryarchive

import (
	"archive/zip"
	"context"
	"encoding/binary"
	"io"
)

// Preflight bounds actual central-directory allocations before archive/zip runs.
// Version 1 emits no ZIP comments, multi-disk records or unindexed padding.
func preflight(ctx context.Context, r io.ReaderAt, size int64, l Limits) (int64, error) {
	if size < 22 {
		return 0, ErrInvalid
	}
	var end [22]byte
	if _, err := r.ReadAt(end[:], size-22); err != nil {
		return 0, ErrInvalid
	}
	if u32(end[:4]) != 0x06054b50 || u16(end[4:6]) != 0 || u16(end[6:8]) != 0 || u16(end[20:]) != 0 {
		return 0, ErrInvalid
	}
	count := uint64(u16(end[10:12]))
	length := uint64(u32(end[12:16]))
	offset := uint64(u32(end[16:20]))
	endOffset := uint64(size - 22)
	if count == 0xffff || length == 0xffffffff || offset == 0xffffffff {
		if size < 98 {
			return 0, ErrInvalid
		}
		var locator [20]byte
		if _, err := r.ReadAt(locator[:], size-42); err != nil || u32(locator[:4]) != 0x07064b50 || u32(locator[4:8]) != 0 || u32(locator[16:]) != 1 {
			return 0, ErrInvalid
		}
		position := binary.LittleEndian.Uint64(locator[8:16])
		if position > uint64(size-98) {
			return 0, ErrInvalid
		}
		var ext [56]byte
		if _, err := r.ReadAt(ext[:], int64(position)); err != nil || u32(ext[:4]) != 0x06064b50 || binary.LittleEndian.Uint64(ext[4:12]) != 44 || u32(ext[16:20]) != 0 || u32(ext[20:24]) != 0 || position+56 != uint64(size-42) {
			return 0, ErrInvalid
		}
		count = binary.LittleEndian.Uint64(ext[32:40])
		if binary.LittleEndian.Uint64(ext[24:32]) != count {
			return 0, ErrInvalid
		}
		length = binary.LittleEndian.Uint64(ext[40:48])
		offset = binary.LittleEndian.Uint64(ext[48:56])
		endOffset = position
	} else if uint64(u16(end[8:10])) != count {
		return 0, ErrInvalid
	}
	if count > uint64(l.Entries) || length > uint64(l.MetadataBytes) {
		return 0, ErrLimit
	}
	if offset > endOffset || length != endOffset-offset {
		return 0, ErrInvalid
	}
	cursor := offset
	actual := uint64(0)
	for cursor < endOffset {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if actual >= uint64(l.Entries) {
			return 0, ErrLimit
		}
		if endOffset-cursor < 46 {
			return 0, ErrInvalid
		}
		var header [46]byte
		if _, err := r.ReadAt(header[:], int64(cursor)); err != nil || u32(header[:4]) != 0x02014b50 {
			return 0, ErrInvalid
		}
		record := uint64(46) + uint64(u16(header[28:30])) + uint64(u16(header[30:32])) + uint64(u16(header[32:34]))
		if record > endOffset-cursor {
			return 0, ErrInvalid
		}
		cursor += record
		actual++
	}
	if actual != count {
		return 0, ErrInvalid
	}
	return int64(offset), nil
}
func verifyCoverage(ctx context.Context, r io.ReaderAt, z *zip.Reader, directory int64) error {
	cursor := int64(0)
	for _, f := range z.File {
		if err := ctx.Err(); err != nil {
			return err
		}
		if cursor > directory-30 {
			return ErrInvalid
		}
		var header [30]byte
		if _, err := r.ReadAt(header[:], cursor); err != nil || u32(header[:4]) != 0x04034b50 || u16(header[6:8]) != f.Flags || u16(header[8:10]) != f.Method {
			return ErrInvalid
		}
		nameLength := int64(u16(header[26:28]))
		extra := int64(u16(header[28:30]))
		if nameLength != int64(len(f.Name)) || nameLength > 70 {
			return ErrInvalid
		}
		var name [70]byte
		if _, err := r.ReadAt(name[:nameLength], cursor+30); err != nil || string(name[:nameLength]) != f.Name {
			return ErrInvalid
		}
		data, err := f.DataOffset()
		if err != nil || data != cursor+30+nameLength+extra || data > directory || f.CompressedSize64 > uint64(directory-data) {
			return ErrInvalid
		}
		cursor = data + int64(f.CompressedSize64)
		if f.Flags&8 != 0 {
			width := int64(16)
			if f.CompressedSize64 >= 0xffffffff || f.UncompressedSize64 >= 0xffffffff {
				width = 24
			}
			if width > directory-cursor {
				return ErrInvalid
			}
			var descriptor [24]byte
			if _, err = r.ReadAt(descriptor[:width], cursor); err != nil || u32(descriptor[:4]) != 0x08074b50 || u32(descriptor[4:8]) != f.CRC32 {
				return ErrInvalid
			}
			compressed := uint64(u32(descriptor[8:12]))
			expanded := uint64(u32(descriptor[12:16]))
			if width == 24 {
				compressed = binary.LittleEndian.Uint64(descriptor[8:16])
				expanded = binary.LittleEndian.Uint64(descriptor[16:24])
			}
			if compressed != f.CompressedSize64 || expanded != f.UncompressedSize64 {
				return ErrInvalid
			}
			cursor += width
		}
	}
	if cursor != directory {
		return ErrInvalid
	}
	return nil
}
func u16(b []byte) uint16 { return binary.LittleEndian.Uint16(b) }
func u32(b []byte) uint32 { return binary.LittleEndian.Uint32(b) }
