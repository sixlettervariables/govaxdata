//
// Portions copyright 2016 Christopher A. Watford. All rights reserved.
// Use of this source code, except those in the public domain, are governed
// by the MIT License that can be found in the LICENSE file. No claim is made
// to the original USGS work released into the public domain.
//
// Original source:
//
// Baker, L.M., 2005, libvaxdata: VAX Data Format Conversion
//    Routines: U.S. Geological Survey Open-File Report 2005-1424
//    http://pubs.usgs.gov/of/2005/1424/
//
// Original libvaxdata Disclaimer:
//
// Although this program has been used by the USGS, no warranty, expressed or
// implied, is made by the USGS or the United States  Government  as  to  the
// accuracy  and functioning of the program and related program material, nor
// shall the fact of  distribution  constitute  any  such  warranty,  and  no
// responsibility is assumed by the USGS in connection therewith.
//

// vaxdata allows conversions to and from VAX floating point formats.
package vaxdata

// Assumes LittleEndian architecture
import (
	"encoding/binary"
)

// VaxFFloat represents a VAX F_Float 32-bit value
type VaxFFloat uint32

// VaxGFloat represents a VAX G_Float 64-bit value
type VaxGFloat uint64

// Floating point data format invariants
//
// (reproduced from convert_vax_data.h)
//
// Most  Unix machines implement the ANSI/IEEE 754-1985 floating-point arith-
// metic standard.  VAX and IEEE formats are similar  (after  byte-swapping).
// The  high-order bit is a sign bit (s).  This is followed by a biased expo-
// nent (e), and a (usually) hidden-bit normalized mantissa (m).  They differ
// in  the number used to bias the exponent, the location of the implicit bi-
// nary point for the mantissa, and the representation of exceptional numbers
// (e.g., +/-infinity).
//
// VAX floating-point formats:  (-1)^s * 2^(e-bias) * 0.1m
//
// 	                 31              15              0
// 	                  |               |              |
// 	F_floating        mmmmmmmmmmmmmmmmseeeeeeeemmmmmmm  bias = 128
// 	D_floating        mmmmmmmmmmmmmmmmseeeeeeeemmmmmmm  bias = 128
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
// 	G_floating        mmmmmmmmmmmmmmmmseeeeeeeeeeemmmm  bias = 1024
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
// 	H_floating        mmmmmmmmmmmmmmmmseeeeeeeeeeeeeee  bias = 16384
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
//
// IEEE floating-point formats:  (-1)^s * 2^(e-bias) * 1.m
//
// 	                 31              15              0
// 	                  |               |              |
// 	S_floating        seeeeeeeemmmmmmmmmmmmmmmmmmmmmmm  bias = 127
// 	T_floating        seeeeeeeeeeemmmmmmmmmmmmmmmmmmmm  bias = 1023
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
// 	X_floating        seeeeeeeeeeeeeeemmmmmmmmmmmmmmmm  bias = 16383
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
// 	                  mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
//
// A  VAX floating-point number is converted to IEEE floating-point format by
// subtracting (1+VAX_bias-IEEE_bias) from the exponent field to  (1)  adjust
// from  VAX  0.1m hidden-bit normalization to IEEE 1.m hidden-bit normaliza-
// tion and (2) adjust the bias from VAX format to IEEE  format.   True  zero
// [s=e=m=0]  and  dirty  zero  [s=e=0, m<>0] are special cases which must be
// recognized and handled separately.  Both VAX zeros are converted  to  IEEE
// +zero [s=e=m=0].
//
// Numbers  whose  absolute value is too small to represent in the normalized
// IEEE format illustrated above are converted to subnormal form [e=0,  m>0]:
// (-1)^s * 2^(1-bias) * 0.m.   Numbers  whose absolute value is too small to
// represent in subnormal form are set to 0.0 (silent underflow).
//
// 	Note: If the fractional part of the VAX floating-point number is too
// 	large for the corresponding IEEE floating-point format,  bits  are
// 	simply discarded from the right.  Thus, the remaining fractional part
// 	is chopped, not rounded to the lowest-order bit.  This can only occur
// 	when the conversion requires IEEE subnormal form.
//
// A  VAX  floating-point  reserved operand [s=1, e=0, m=any] causes a SIGFPE
// exception to be raised.  The converted result is set to zero.
//
// Conversely,  an  IEEE  floating-point number is converted to VAX floating-
// point format by  adding  (1+VAX_bias-IEEE_bias)  to  the  exponent  field.
// +zero [s=e=m=0], -zero [s=1, e=m=0], infinities [s=X, e=all-1's, m=0], and
// NaNs [s=X, e=all-1's, m<>0] are special cases which must be recognized and
// handled  separately.   Both  IEEE  zeros  are  converted  to VAX true zero
// [s=e=m=0].  Infinities and NaNs cause a SIGFPE  exception  to  be  raised.
// The  result  returned  has  the  largest VAX exponent [e=all-1's] and zero
// mantissa [m=0] with the same sign as the original.
//
// Numbers  whose  absolute value is too small to represent in the normalized
// VAX format illustrated above are set  to  0.0  (silent  underflow).   (VAX
// floating-point  format does not support subnormal numbers.)  Numbers whose
// absolute value exceeds the largest representable VAX-format number cause a
// SIGFPE exception to be raised (overflow).  (VAX floating-point format does
// not have reserved bit patterns for infinities and  not-a-numbers  [NaNs].)
// The  result  returned  has  the  largest  VAX  exponent and mantissa [e=m=
// all-1's] with the same sign as the original.
//
const (
	SignBit uint32 = 0x80000000

	//  VAX floating point data formats (see VAX Architecture Reference Manual)

	VaxFExponentMask uint32 = 0x7F800000
	VaxFExponentSize uint32 = 8
	VaxFExponentBias uint32 = (1 << (VaxFExponentSize - 1))
	VaxFMantissaMask uint32 = 0x007FFFFF
	VaxFMantissaSize uint32 = 23
	VaxFHiddenBit    uint32 = (1 << VaxFMantissaSize)

	VaxGExponentMask uint32 = 0x7FF00000
	VaxGExponentSize uint32 = 11
	VaxGExponentBias uint32 = (1 << (VaxGExponentSize - 1))
	VaxGMantissaMask uint32 = 0x000FFFFF
	VaxGMantissaSize uint32 = 20
	VaxGHiddenBit    uint32 = (1 << VaxGMantissaSize)

	// IEEE floating point data formats (see Alpha Architecture Reference Manual)

	IeeeSExponentMask uint32 = 0x7F800000
	IeeeSExponentSize uint32 = 8
	IeeeSExponentBias uint32 = ((1 << (IeeeSExponentSize - 1)) - 1)
	IeeeSMantissaMask uint32 = 0x007FFFFF
	IeeeSMantissaSize uint32 = 23
	IeeeSHiddenBit    uint32 = (1 << IeeeSMantissaSize)

	IeeeTExponentMask uint32 = 0x7FF00000
	IeeeTExponentSize uint32 = 11
	IeeeTExponentBias uint32 = ((1 << (IeeeTExponentSize - 1)) - 1)
	IeeeTMantissaMask uint32 = 0x000FFFFF
	IeeeTMantissaSize uint32 = 20
	IeeeTHiddenBit    uint32 = (1 << IeeeTMantissaSize)
)

// uit32FromVaxbits reads a 32-bit VAX doubleword into LittleEndian
func uint32FromVaxbits(b []byte) uint32 {
	return uint32(binary.BigEndian.Uint16(b[0:2])) |
		uint32(binary.BigEndian.Uint16(b[2:4]))<<16
}

// uit32FromVax swaps the 16-bit words in a 32-bit doubleword
func uint32FromVax(v uint32) uint32 {
	return uint32(v>>16) | (uint32(v&0x0FFFF) << 16)
}
