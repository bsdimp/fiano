// Copyright 2019 the LinuxBoot Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package manifest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	bytes2 "github.com/linuxboot/fiano/pkg/bytes"
)

// Refer to: AMD Platform Security Processor BIOS Architecture Design Guide for AMD Family 17h and Family 19h
// Processors (NDA), Publication # 55758 Revision: 1.11 Issue Date: August 2020 (1)

// EmbeddedFirmwareStructureSignature is a special identifier of Firmware Embedded Structure
const EmbeddedFirmwareStructureSignature = 0x55aa55aa

// EmbeddedFirmwareStructure represents Embedded Firmware Structure defined in Table 2 in (1)
type EmbeddedFirmwareStructure struct {
	Signature                uint32
	Reserved1                [16]byte
	PSPDirectoryTablePointer uint32

	BIOSDirectoryTableFamily17hModels00h0FhPointer uint32
	BIOSDirectoryTableFamily17hModels10h1FhPointer uint32
	BIOSDirectoryTableFamily17hModels30h3FhPointer uint32
	Reserved2                                      uint32
	BIOSDirectoryTableFamily17hModels60h3FhPointer uint32

	Reserved3 [30]byte
}

// FindEmbeddedFirmwareStructure locates and parses Embedded Firmware Structure
func FindEmbeddedFirmwareStructure(firmware Firmware) (*EmbeddedFirmwareStructure, bytes2.Range, error) {
	var addresses = []uint64{
		0xfffa0000,
		0xfff20000,
		0xffe20000,
		0xffc20000,
		0xff820000,
		0xff020000,
	}

	image := firmware.ImageBytes()

	for _, addr := range addresses {
		offset := firmware.PhysAddrToOffset(addr)
		if offset+4 > uint64(len(image)) {
			continue
		}

		actualSignature := binary.LittleEndian.Uint32(image[offset:])
		if actualSignature == EmbeddedFirmwareStructureSignature {
			result, length, err := ParseEmbeddedFirmwareStructure(bytes.NewBuffer(image[offset:]))
			return result, bytes2.Range{Offset: offset, Length: length}, err
		}
	}
	return nil, bytes2.Range{}, fmt.Errorf("EmbeddedFirmwareStructure is not found")
}

// ParseEmbeddedFirmwareStructure converts input bytes into EmbeddedFirmwareStructure
func ParseEmbeddedFirmwareStructure(r io.Reader) (*EmbeddedFirmwareStructure, uint64, error) {
	var result EmbeddedFirmwareStructure
	if err := binary.Read(r, binary.LittleEndian, &result); err != nil {
		return nil, 0, err
	}

	if result.Signature != EmbeddedFirmwareStructureSignature {
		return nil, 0, fmt.Errorf("incorrect signature: %d", result.Signature)
	}
	return &result, uint64(binary.Size(result)), nil
}
