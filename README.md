# vaxdata
**vaxdata** is an implementation of [libvaxdata](http://pubs.usgs.gov/of/2005/1424/) in Go. vaxdata allows conversions to and from VAX floating point formats. The following conversions are supported:

- VAX F_Float to and from `float32`
- VAX G_Float to and from `float64`

## Usage

```go
import "github.com/sixlettervariables/vaxdata"

var (
  f float32
  g float64
)

f, err := Float32fromVaxFFloat([]byte { 0x00, 0x00, 0x40, 0x80 })
// f == 1.0

g, err := Float64fromVaxGFloat([]byte{0x2D, 0x18, 0x54, 0x44, 0x21, 0xFB, 0x40, 0x29 })
// g == 3.141592653589793
```

#### func  Float32fromVaxFFloat

```go
func Float32fromVaxFFloat(buf []byte) (float32, error)
```
Float32fromVaxFFloat returns the float32 representation of a VAX F_Float.

#### func  Float64fromVaxGFloat

```go
func Float64fromVaxGFloat(buf []byte) (float64, error)
```
Float64fromVaxGFloat returns the float64 representation of a VAX G_Float.

#### func  WriteFFloat

```go
func WriteFFloat(w io.Writer, f float32) error
```
WriteFFloat takes a float32 and writes an F_Float to the io.Writer.

#### func  WriteGFloat

```go
func WriteGFloat(w io.Writer, f float64) error
```
WriteGFloat takes a float64 and writes an G_Float to the io.Writer.

#### type VaxFFloat

```go
type VaxFFloat uint32
```

VaxFFloat represents a VAX F_Float 32-bit value

#### func  VaxFFloatfromFloat32

```go
func VaxFFloatfromFloat32(f float32) (VaxFFloat, error)
```
VaxFFloatfromFloat32 returns the VAX F_Float representation of a float32.

#### type VaxFFloatReader

```go
type VaxFFloatReader struct {
}
```

VaxFFloatReader reads float32 values from F_Float's in the underlying io.Reader.

#### func  NewVaxFFloatReader

```go
func NewVaxFFloatReader(r io.Reader) *VaxFFloatReader
```
NewVaxFFloatReader creates a new VaxFFloatReader. VaxFFloatReader.Read reads a
float32 from a F_Float in the underlying io.Reader.

#### func (\*VaxFFloatReader) Read

```go
func (vaxin *VaxFFloatReader) Read() (float32, error)
```
Read takes a F_Float from the underlying io.Reader and returns a float32.

#### type VaxGFloat

```go
type VaxGFloat uint64
```

VaxGFloat represents a VAX G_Float 64-bit value

#### func  VaxGFloatfromFloat64

```go
func VaxGFloatfromFloat64(f float64) (VaxGFloat, error)
```
VaxGFloatfromFloat64 returns the VAX G_Float representation of a float64.

#### type VaxGFloatReader

```go
type VaxGFloatReader struct {
}
```

VaxGFloatReader reads float64 values from G_Float's in the underlying io.Reader.

#### func  NewVaxGFloatReader

```go
func NewVaxGFloatReader(r io.Reader) *VaxGFloatReader
```
NewVaxGFloatReader creates a new VaxGFloatReader. VaxGFloatReader.Read reads a
float64 from a G_Float in the underlying io.Reader.

#### func (\*VaxGFloatReader) Read

```go
func (vaxin *VaxGFloatReader) Read() (float64, error)
```
Read takes a G_Float from the underlying io.Reader and returns a float64

### Constants

```go
// Floating point data format invariants
const (
	SignBit uint32 = 0x80000000

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
```

## Implementation
[Reproduced from `convert_vax_data.h`](http://pubs.usgs.gov/of/2005/1424/).

Most Unix machines implement the ANSI/IEEE 754-1985 floating-point arithmetic
standard. VAX and IEEE formats are similar (after byte-swapping). The high-order
bit is a sign bit (s). This is followed by a biased exponent (e), and a
(usually) hidden-bit normalized mantissa (m). They differ in the number used to
bias the exponent, the location of the implicit binary point for the mantissa,
and the representation of exceptional numbers (e.g., +/-infinity).

VAX floating-point formats: (-1)^s * 2^(e-bias) * 0.1m

                     31              15              0
                      |               |              |
    F_floating        mmmmmmmmmmmmmmmmseeeeeeeemmmmmmm  bias = 128
    D_floating        mmmmmmmmmmmmmmmmseeeeeeeemmmmmmm  bias = 128
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
    G_floating        mmmmmmmmmmmmmmmmseeeeeeeeeeemmmm  bias = 1024
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
    H_floating        mmmmmmmmmmmmmmmmseeeeeeeeeeeeeee  bias = 16384
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm

IEEE floating-point formats: (-1)^s * 2^(e-bias) * 1.m

                     31              15              0
                      |               |              |
    S_floating        seeeeeeeemmmmmmmmmmmmmmmmmmmmmmm  bias = 127
    T_floating        seeeeeeeeeeemmmmmmmmmmmmmmmmmmmm  bias = 1023
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
    X_floating        seeeeeeeeeeeeeeemmmmmmmmmmmmmmmm  bias = 16383
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm
                      mmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm

A VAX floating-point number is converted to IEEE floating-point format by
subtracting (1+VAX_bias-IEEE_bias) from the exponent field to (1) adjust from
VAX 0.1m hidden-bit normalization to IEEE 1.m hidden-bit normalization and (2)
adjust the bias from VAX format to IEEE format. True zero [s=e=m=0] and dirty
zero [s=e=0, m<>0] are special cases which must be recognized and handled
separately. Both VAX zeros are converted to IEEE +zero [s=e=m=0].

Numbers whose absolute value is too small to represent in the normalized IEEE
format illustrated above are converted to subnormal form [e=0, m>0]: (-1)^s *
2^(1-bias) * 0.m. Numbers whose absolute value is too small to represent in
subnormal form are set to 0.0 (silent underflow).

> Note: If the fractional part of the VAX floating-point number is too
large for the corresponding IEEE floating-point format,  bits  are
simply discarded from the right.  Thus, the remaining fractional part
is chopped, not rounded to the lowest-order bit.  This can only occur
when the conversion requires IEEE subnormal form.

A VAX floating-point reserved operand [s=1, e=0, m=any] causes a SIGFPE
exception to be raised. The converted result is set to zero.

Conversely, an IEEE floating-point number is converted to VAX floating-point
format by adding (1+VAX_bias-IEEE_bias) to the exponent field. +zero [s=e=m=0],
-zero [s=1, e=m=0], infinities [s=X, e=all-1's, m=0], and NaNs [s=X, e=all-1's,
m<>0] are special cases which must be recognized and handled separately. Both
IEEE zeros are converted to VAX true zero [s=e=m=0]. Infinities and NaNs cause a
SIGFPE exception to be raised. The result returned has the largest VAX exponent
[e=all-1's] and zero mantissa [m=0] with the same sign as the original.

Numbers whose absolute value is too small to represent in the normalized VAX
format illustrated above are set to 0.0 (silent underflow). (VAX floating-point
format does not support subnormal numbers.) Numbers whose absolute value exceeds
the largest representable VAX-format number cause a SIGFPE exception to be
raised (overflow). (VAX floating-point format does not have reserved bit
patterns for infinities and not-a-numbers [NaNs].) The result returned has the
largest VAX exponent and mantissa [e=m= all-1's] with the same sign as the
original.

# Contributing
All contributions to the project will be considered for inclusion, simply submit a pull request!

1. [Fork](https://github.com/sixlettervariables/vaxdata/fork) the repository ([read more detailed steps](https://help.github.com/articles/fork-a-repo)).
2. [Create a branch](https://help.github.com/articles/creating-and-deleting-branches-within-your-repository#creating-a-branch)
3. Make and commit your changes
4. Push your changes back to your fork.
5. [Submit a pull request](https://github.com/sixlettervariables/vaxdata/compare/) ([read more detailed steps](https://help.github.com/articles/creating-a-pull-request)).

# License
The MIT License

Portions copyright (C) 2016 Christopher A. Watford, no claims made to original
USGS work in public domain.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice, this permission notice, and the disclaimer below
shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

Original libvaxdata Disclaimer

Although this program has been used by the USGS, no warranty, expressed or
implied, is made by the USGS or the United States  Government  as  to  the
accuracy  and functioning of the program and related program material, nor
shall the fact of  distribution  constitute  any  such  warranty,  and  no
responsibility is assumed by the USGS in connection therewith.
