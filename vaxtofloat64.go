package vaxdata

import (
	"errors"
	"io"
	"math"
)

// VaxGFloatReader reads float64 values from G_Float's in the underlying io.Reader.
type VaxGFloatReader struct {
	r   io.Reader
	buf []byte
}

// NewVaxGFloatReader creates a new VaxGFloatReader. VaxGFloatReader.Read reads
// a float64 from a G_Float in the underlying io.Reader.
func NewVaxGFloatReader(r io.Reader) *VaxGFloatReader {
	vaxin := new(VaxGFloatReader)
	(*vaxin).r = r
	(*vaxin).buf = make([]byte, 8)
	return vaxin
}

// Read takes a G_Float from the underlying io.Reader and returns a float64
func (vaxin *VaxGFloatReader) Read() (float64, error) {
	if _, err := io.ReadFull(vaxin.r, vaxin.buf); err != nil {
		return 0, err
	}
	return Float64fromVaxGFloat(vaxin.buf)
}

// Float64fromVaxGFloat returns the float64 representation of a VAX G_Float.
func Float64fromVaxGFloat(buf []byte) (float64, error) {
	const (
		MantissaMask              uint32 = VaxGMantissaMask
		MantissaSize              uint32 = VaxGMantissaSize
		HiddenBit                 uint32 = VaxGHiddenBit
		ExponentAdjustment        int32  = int32(1 + VaxGExponentBias - IeeeTExponentBias)
		InPlaceExponentAdjustment uint32 = uint32(ExponentAdjustment << IeeeTMantissaSize)
	)

	var (
		ieeepart1, ieeepart2 uint32
		err                  error = nil
	)

	vaxpart2 := uint32FromVaxbits(buf[:4])
	vaxpart1 := uint32FromVaxbits(buf[4:8])

	if e := int32(vaxpart1 & VaxGExponentMask); e == 0 {
		// If the biased VAX exponent is zero [e=0]

		if (vaxpart1 & SignBit) == SignBit {
			// If negative [s=1]
			// fixup to IEEE zero
			err = errors.New("G_Float to T_Float: VAX reserved operand fault")
		}

		// Set VAX dirty [m<>0] or true [m=0] zero to IEEE +zero [s=e=m=0]
		ieeepart1 = 0
		ieeepart2 = 0

	} else {
		// The biased VAX exponent is non-zero [e<>0]

		e >>= MantissaSize // Obtain the biased VAX exponent

		// The  biased  VAX  exponent  has to be adjusted to account for the
		// right shift of the IEEE mantissa binary point and the  difference
		// between  the biases in their "excess n" exponent representations.
		// If the resulting biased IEEE exponent is less than  or  equal  to
		// zero, the converted IEEE T_float must use subnormal form.

		if e -= ExponentAdjustment; e > 0 {
			// Use IEEE normalized form [e>0]

			// Both mantissas are 52 bits; adjust the exponent field in place
			ieeepart1 = vaxpart1 - InPlaceExponentAdjustment
			ieeepart2 = vaxpart2

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
			//

			vaxpart1 = (vaxpart1 & (SignBit | MantissaMask)) | HiddenBit
			ieeepart1 = (vaxpart1 & SignBit) | ((vaxpart1 & (HiddenBit | MantissaMask)) >> uint32(1-e))
			ieeepart2 = (vaxpart1 << uint32(31+e)) | (vaxpart2 >> uint32(1-e))

		}
	}

	result := uint64(uint64(ieeepart1)<<32) | uint64(ieeepart2)
	return math.Float64frombits(result), err
}
