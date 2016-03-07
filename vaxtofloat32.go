package vaxdata

import (
	"errors"
	"io"
	"math"
)

// VaxFFloatReader reads float32 values from F_Float's in the underlying io.Reader.
type VaxFFloatReader struct {
	r   io.Reader
	buf []byte
}

// NewVaxFFloatReader creates a new VaxFFloatReader. VaxFFloatReader.Read reads
// a float32 from a F_Float in the underlying io.Reader.
func NewVaxFFloatReader(r io.Reader) *VaxFFloatReader {
	vaxin := new(VaxFFloatReader)
	(*vaxin).r = r
	(*vaxin).buf = make([]byte, 4)
	return vaxin
}

// Read takes a F_Float from the underlying io.Reader and returns a float32.
func (vaxin *VaxFFloatReader) Read() (float32, error) {
	if _, err := io.ReadFull(vaxin.r, vaxin.buf); err != nil {
		return 0, err
	}
	return Float32fromVaxFFloat(vaxin.buf)
}

// Float32fromVaxFFloat returns the float32 representation of a VAX F_Float.
func Float32fromVaxFFloat(buf []byte) (float32, error) {
	const (
		MantissaMask                     = VaxFMantissaMask
		MantissaSize                     = VaxFMantissaSize
		HiddenBit                        = VaxFHiddenBit
		ExponentAdjustment        int32  = int32(1 + VaxFExponentBias - IeeeSExponentBias)
		InPlaceExponentAdjustment uint32 = uint32(ExponentAdjustment << IeeeSMantissaSize)
	)

	var (
		result uint32
	)

	vaxpart1 := uint32FromVaxbits(buf)

	if e := int32(vaxpart1 & VaxFExponentMask); e == 0 {
		// If the biased VAX exponent is zero [e=0]

		if (vaxpart1 & SignBit) == SignBit {
			// If negative [s=1]
			// fixup to IEEE zero
			return 0, errors.New("F_Float to S_Float: VAX reserved operand fault")
		}

		// Set VAX dirty [m<>0] or true [m=0] zero to IEEE +zero [s=e=m=0]
		result = 0
	} else {
		// The biased VAX exponent is non-zero [e<>0]
		e >>= MantissaSize // Obtain the biased VAX exponent

		// The  biased  VAX  exponent  has to be adjusted to account for the
		// right shift of the IEEE mantissa binary point and the  difference
		// between  the biases in their "excess n" exponent representations.
		// If the resulting biased IEEE exponent is less than  or  equal  to
		// zero, the converted IEEE S_float must use subnormal form.

		if e -= ExponentAdjustment; e > 0 {
			// Use IEEE normalized form [e>0]
			// Both mantissas are 23 bits; adjust the exponent field in place
			result = vaxpart1 - InPlaceExponentAdjustment
		} else {
			// Use IEEE subnormal form [e=0, m>0]

			// In IEEE subnormal form, even though the biased exponent is 0
			// [e=0], the effective biased exponent is 1.  The mantissa must
			// be shifted right by the number of bits, n, required to adjust
			// the biased exponent from its current value, e, to 1.  I.e.,
			// e + n = 1, thus n = 1 - e.  n is guaranteed to be at least 1
			// [e<=0], which guarantees that the hidden 1.m bit from the ori-
			// ginal mantissa will become visible, and the resulting subnor-
			// mal mantissa will correctly be of the form 0.m.
			result = (vaxpart1 & SignBit) | uint32((HiddenBit|(vaxpart1&MantissaMask))>>uint32(1-e))
		}
	}

	return math.Float32frombits(result), nil
}
