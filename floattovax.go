package vaxdata

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// WriteFFloat takes a float32 and writes an F_Float to the io.Writer.
func WriteFFloat(w io.Writer, f float32) error {
	v, err := VaxFFloatfromFloat32(f)
	if err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, uint32(v))
}

// VaxFFloatfromFloat32 returns the VAX F_Float representation of a float32.
func VaxFFloatfromFloat32(f float32) (VaxFFloat, error) {
	const (
		MantissaMask       uint32 = VaxFMantissaMask
		MantissaSize       uint32 = VaxFMantissaSize
		HiddenBit          uint32 = VaxFHiddenBit
		ExponentAdjustment uint32 = (1 + VaxFExponentBias - IeeeSExponentBias)
	)

	var (
		result uint32
		err    error = nil
	)

	ieeepart1 := math.Float32bits(f)
	if (ieeepart1 & ^SignBit) == 0 {
		// Set IEEE +-zero [e=m=0] to VAX zero [s=e=m=0]
		result = 0
	} else if e := (ieeepart1 & IeeeSExponentMask); e == IeeeSExponentMask {
		// VAX's have no equivalents for IEEE +-Infinity and +-NaN [e=all-1's]
		err = errors.New("no VAX equivalent for IEEE +-Infinity and +-NaN")

		// Fixup to VAX +-extrema [e=all-1's] with zero mantissa [m=0]
		result = (ieeepart1 & SignBit) | VaxFExponentMask
	} else {
		e >>= MantissaSize              // Obtain the biased IEEE exponent
		m := (ieeepart1 & MantissaMask) // Obtain the IEEE mantissa

		// Denormalized? [e=0, m<>0]
		if e == 0 {
			// Adjust representation from 2**(1-bias) to 2**(e-bias)
			m <<= 1
			for (m & HiddenBit) == 0 {
				m <<= 1
				e -= 1 // Adjust exponent
			}
			// Adjust mantissa to hidden-bit form
			m &= MantissaMask
		}

		if e += ExponentAdjustment; e <= 0 {
			result = 0 // Silent underflow
		} else if e > (2*VaxFExponentBias - 1) {
			// Overflow; fixup to VAX +-extrema [e=m=all-1's]
			err = errors.New("IEEE S_Float too large for VAX F_Float")
			result = (ieeepart1 & SignBit) | ^SignBit
		} else {
			// VAX normalized form [e>0] (both mantissas are 23 bits)
			result = (ieeepart1 & SignBit) | (e << MantissaSize) | m
		}
	}

	return VaxFFloat(uint32FromVax(result)), err
}

// WriteGFloat takes a float64 and writes an G_Float to the io.Writer.
func WriteGFloat(w io.Writer, f float64) error {
	v, err := VaxGFloatfromFloat64(f)
	if err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, uint64(v))
}

// VaxGFloatfromFloat64 returns the VAX G_Float representation of a float64.
func VaxGFloatfromFloat64(f float64) (VaxGFloat, error) {
	const (
		MantissaMask       uint32 = VaxGMantissaMask
		MantissaSize       uint32 = VaxGMantissaSize
		HiddenBit          uint32 = VaxGHiddenBit
		ExponentAdjustment uint32 = (1 + VaxGExponentBias - IeeeTExponentBias)
	)

	var (
		in       uint64 = math.Float64bits(f)
		vaxpart1 uint32
		err      error = nil
	)

	ieeepart1 := uint32(in >> 32)
	vaxpart2 := uint32(in & 0x0FFFFFFFF)

	if ((ieeepart1 & ^SignBit) | vaxpart2) == 0 {
		// Set IEEE +-zero [e=m=0] to VAX zero [s=e=m=0]
		vaxpart1 = 0
		// vaxpart2 is already zero
	} else if e := (ieeepart1 & IeeeTExponentMask); e == IeeeTExponentMask {
		// VAX's have no equivalents for IEEE +-Infinity and +-NaN [e=all-1's]
		err = errors.New("no VAX equivalent for IEEE +-Infinity and +-NaN")

		// Fixup to VAX +-extrema [e=all-1's] with zero mantissa [m=0]
		vaxpart1 = (ieeepart1 & SignBit) | VaxGExponentMask
		vaxpart2 = 0
	} else {
		e >>= MantissaSize            // Obtain the biased IEEE exponent
		m := ieeepart1 & MantissaMask // Obtain the IEEE mantissa

		// Denormalized? [e=0, m<>0]
		if e == 0 {
			// Adjust representation from 2**(1-bias) to 2**(e-bias)
			m = (m << 1) | (vaxpart2 >> 31)
			vaxpart2 <<= 1
			for (m & HiddenBit) == 0 {
				m = (m << 1) | (vaxpart2 >> 31)
				vaxpart2 <<= 1
				e -= 1 // Adjust exponent
			}

			// Adjust mantissa to hidden-bit form
			m &= MantissaMask
		}

		if e += ExponentAdjustment; e <= 0 {
			vaxpart1 = 0 // Silent underflow
			vaxpart2 = 0
		} else if e > (2*VaxGExponentBias - 1) {
			err = errors.New("IEEE T_Float too large for VAX G_Float")

			// Overflow; fixup to VAX +-extrema [e=m=all-1's]
			vaxpart1 = (ieeepart1 & SignBit) | ^SignBit
			vaxpart2 = 0xFFFFFFFF
		} else {
			// VAX normalized form [e>0] (both mantissas are 52 bits)
			vaxpart1 = (ieeepart1 & SignBit) | (e << MantissaSize) | m
			// vaxpart2 is already correct
		}
	}

	result := (uint64(uint32FromVax(vaxpart2)) << 32) | uint64(uint32FromVax(vaxpart1))
	return VaxGFloat(result), err
}
