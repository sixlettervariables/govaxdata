//
// Copyright 2016 Christopher A. Watford. All rights reserved.
// Use of this source code is governed by the MIT License that
// can be found in the LICENSE file. No claim is made to the
// original USGS work released into the public domain.
//

// Provides tests for the conversions to-and-from VAX F_ and G_floats
package vaxdata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"testing"
)

func TestVaxFFloat(t *testing.T) {
	cases := []struct {
		in   float32
		want string
	}{
		{1.000000, "00004080"},
		{-1.000000, "0000C080"},
		{3.500000, "00004160"},
		{-3.500000, "0000C160"},
		{3.141590, "0FD04149"},
		{-3.141590, "0FD0C149"},
		{9.9999999E+36, "BDC27DF0"},
		{-9.9999999E+36, "BDC2FDF0"},
		{9.9999999E-38, "1CEA0308"},
		{-9.9999999E-38, "1CEA8308"},
		{1.234568, "0653409E"},
		{-1.234568, "0653C09E"},
	}

	buf := make([]byte, 4)
	for _, c := range cases {
		got, err := VaxFFloatfromFloat32(c.in)
		if err != nil {
			t.Errorf("VaxFFloat32fromFloat32(%g) raised unexpected error: %q", c.in, err)
		} else if fmt.Sprintf("%08X", got) != c.want {
			t.Errorf("VaxFFloat32fromFloat32(%g) == %08X, want %s", c.in, got, c.want)
		}

		binary.BigEndian.PutUint32(buf, uint32(got))
		reverse, err := Float32fromVaxFFloat(buf)
		if err != nil {
			t.Errorf("Float32fromVaxFFloat(%08X) raised unexpected error: %q", got, err)
		} else if reverse != c.in {
			t.Errorf("Float32fromVaxFFloat(%08X) == %g, want %g", got, reverse, c.in)
		}
	}
}

func TestVaxFFloatbits(t *testing.T) {
	cases := []struct {
		ieee []float32
		vaxf []byte
	}{
		{[]float32{1.000000, -1.000000}, []byte{0x00, 0x00, 0x40, 0x80, 0x00, 0x00, 0xC0, 0x80}},
		{[]float32{3.500000, -3.500000}, []byte{0x00, 0x00, 0x41, 0x60, 0x00, 0x00, 0xC1, 0x60}},
		{[]float32{3.141590, -3.141590}, []byte{0x0F, 0xD0, 0x41, 0x49, 0x0F, 0xD0, 0xC1, 0x49}},
		{[]float32{9.9999999E+36, -9.9999999E+36}, []byte{0xBD, 0xC2, 0x7D, 0xF0, 0xBD, 0xC2, 0xFD, 0xF0}},
		{[]float32{9.9999999E-38, -9.9999999E-38}, []byte{0x1C, 0xEA, 0x03, 0x08, 0x1C, 0xEA, 0x83, 0x08}},
		{[]float32{1.234568, -1.234568}, []byte{0x06, 0x53, 0x40, 0x9E, 0x06, 0x53, 0xC0, 0x9E}},
	}
	for _, c := range cases {
		// test using a reader
		n := 0
		r := NewVaxFFloatReader(bytes.NewBuffer(c.vaxf))
		v, err := r.Read()
		for err == nil {
			if c.ieee[n] != v {
				t.Errorf("VaxFFloatReader.Read(%v) == %v, want %v", c.vaxf[n*4:n*4+4], v, c.ieee[n])
			}
			v, err = r.Read()
			n++
		}
		if err != nil && err != io.EOF {
			t.Errorf("Unexpected error %q", err)
		}

		// test the underlying conversion routine
		for i, f := range c.ieee {
			got := c.vaxf[(i * 4) : (i*4)+4]
			reverse, err := Float32fromVaxFFloat(got)
			if err != nil {
				t.Errorf("Float32fromVaxFFloat(%v) raised unexpected error: %q", got, err)
			} else if f != reverse {
				t.Errorf("Float32fromVaxFFloat(%v) == %v, want %v", got, reverse, f)
			}
		}

		var buf bytes.Buffer
		for _, f := range c.ieee {
			err := WriteFFloat(&buf, f)
			if err != nil {
				t.Errorf("WriteFFloat(%v) raised unexpected error: %q", f, err)
			}
		}
		if !sliceByteEquals(buf.Bytes(), c.vaxf) {
			t.Errorf("WriteFFloat(%v) == %v, want %v", c.ieee, buf.Bytes(), c.vaxf)
		}
	}
}

func TestVaxGFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{1.000000000000000, "0000000000004010"},
		{-1.000000000000000, "000000000000C010"},
		{3.500000000000000, "000000000000402C"},
		{-3.500000000000000, "000000000000C02C"},
		{3.141592653589793, "2D18544421FB4029"},
		{-3.141592653589793, "2D18544421FBC029"},
		{1.0000000000000000E+37, "691B435717B847BE"},
		{-1.0000000000000000E+37, "691B435717B8C7BE"},
		{9.9999999999999999E-38, "8B8F428A039D3861"},
		{-9.9999999999999999E-38, "8B8F428A039DB861"},
		{1.23456789012345, "59DD428CC0CA4013"},
		{-1.23456789012345, "59DD428CC0CAC013"},
	}
	buf := make([]byte, 8)
	for _, c := range cases {
		got, err := VaxGFloatfromFloat64(c.in)
		if err != nil {
			t.Errorf("VaxGFloat64fromFloat64(%g) raised unexpected error: %q", c.in, err)
		} else if fmt.Sprintf("%016X", got) != c.want {
			t.Errorf("VaxGFloat64fromFloat64(%g) == %016X, want %s", c.in, got, c.want)
		}

		binary.BigEndian.PutUint64(buf, uint64(got))
		reverse, err := Float64fromVaxGFloat(buf)
		if err != nil {
			t.Errorf("Float64fromVaxGFloat64(%016X) raised unexpected error: %q", got, err)
		} else if reverse != c.in {
			t.Errorf("Float64fromVaxGFloat64(%016X) == %g, want %g", got, reverse, c.in)
		}
	}
}

func TestVaxGFloatbits(t *testing.T) {
	cases := []struct {
		ieee []float64
		vaxg []byte
	}{
		{[]float64{1.000000000000000, -1.000000000000000}, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0, 0x10}},
		{[]float64{3.500000000000000, -3.500000000000000}, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x2C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0, 0x2C}},
		{[]float64{3.141592653589793, -3.141592653589793}, []byte{0x2D, 0x18, 0x54, 0x44, 0x21, 0xFB, 0x40, 0x29, 0x2D, 0x18, 0x54, 0x44, 0x21, 0xFB, 0xC0, 0x29}},
		{[]float64{1.0000000000000000E+37, -1.0000000000000000E+37}, []byte{0x69, 0x1B, 0x43, 0x57, 0x17, 0xB8, 0x47, 0xBE, 0x69, 0x1B, 0x43, 0x57, 0x17, 0xB8, 0xC7, 0xBE}},
		{[]float64{9.9999999999999999E-38, -9.9999999999999999E-38}, []byte{0x8B, 0x8F, 0x42, 0x8A, 0x03, 0x9D, 0x38, 0x61, 0x8B, 0x8F, 0x42, 0x8A, 0x03, 0x9D, 0xB8, 0x61}},
		{[]float64{1.23456789012345, -1.23456789012345}, []byte{0x59, 0xDD, 0x42, 0x8C, 0xC0, 0xCA, 0x40, 0x13, 0x59, 0xDD, 0x42, 0x8C, 0xC0, 0xCA, 0xC0, 0x13}},
	}
	for _, c := range cases {
		// test using a reader
		n := 0
		r := NewVaxGFloatReader(bytes.NewBuffer(c.vaxg))
		v, err := r.Read()
		for err == nil {
			if c.ieee[n] != v {
				t.Errorf("VaxGFloatReader.ReadFloat64(%v) == %v, want %v", c.vaxg[n*8:n*8+8], v, c.ieee[n])
			}
			v, err = r.Read()
			n++
		}
		if err != nil && err != io.EOF {
			t.Errorf("Unexpected error %q", err)
		}

		// test the underlying conversion routine
		for i, f := range c.ieee {
			got := c.vaxg[(i * 8) : (i*8)+8]
			reverse, err := Float64fromVaxGFloat(got)
			if err != nil {
				t.Errorf("Float64fromVaxGFloat(%v) raised unexpected error: %q", got, err)
			} else if f != reverse {
				t.Errorf("Float64fromVaxGFloat(%v) == %v, want %v", got, reverse, f)
			}
		}

		var buf bytes.Buffer
		for _, f := range c.ieee {
			err := WriteGFloat(&buf, f)
			if err != nil {
				t.Errorf("VaxGFloatWriter.Write(%v) raised unexpected error: %q", f, err)
			}
		}
		if !sliceByteEquals(buf.Bytes(), c.vaxg) {
			t.Errorf("VaxGFloatWriter.Write(%v) == %v, want %v", c.ieee, buf.Bytes(), c.vaxg)
		}
	}
}

func sliceByteEquals(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func sliceEquals(a, b []float32) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func slice64Equals(a, b []float64) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
