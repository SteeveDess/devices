// Package monochromeoled contains an Adafruit Monochrome OLED (SSD1306)
// display driver.
package monochromeoled

import (
	"fmt"
	"image"

	"golang.org/x/exp/io/i2c"
	"golang.org/x/exp/io/i2c/driver"
)

const (
	ssd1306_LCDWIDTH  = 128
	ssd1306_LCDHEIGHT = 64

	addr = 0x3C // addr is the I2C address of the device.

	// On or off registers.
	ssd1306_DISPLAY_ON  = 0xAF
	ssd1306_DISPLAY_OFF = 0xAE

	// Scrolling registers.
	ssd1306_ACTIVATE_SCROLL                      = 0x2F
	ssd1306_DEACTIVATE_SCROLL                    = 0x2E
	ssd1306_SET_VERTICAL_SCROLL_AREA             = 0xA3
	ssd1306_RIGHT_HORIZONTAL_SCROLL              = 0x26
	ssd1306_LEFT_HORIZONTAL_SCROLL               = 0x27
	ssd1306_VERTICAL_AND_RIGHT_HORIZONTAL_SCROLL = 0x29
	ssd1306_VERTICAL_AND_LEFT_HORIZONTAL_SCROLL  = 0x2A
)

// OLED represents an SSD1306 OLED display.
type OLED struct {
	dev *i2c.Device

	w   int    // width of the display
	h   int    // height of the display
	buf []byte // each pixel is represented by a bit
}

var initSeq = []byte{
	0xae,
	0x00 | 0x00, // row offset
	0x10 | 0x00, // column offset
	0xd5, 0x40,
	0xa8, ssd1306_LCDHEIGHT - 1,
	0xd3, 0x00, // set display offset to no offset
	0x40 | 0,
	0x8d, 0x14,
	0x20, 0x0,

	0xA0 | 0x1,
	0xC8,
	0xda, 0x12,
	0x81, 0xcf, // set contrast
	0x9d, 0xf1,
	0xdb, 0x40,
	0xa4, 0xa6,

	0x2e,
	0xaf,
}

// Open opens an SSD1306 OLED display. Once not in use, it needs to
// be close by calling Close.
// The default width is 128, height is 64 if zero values are given.
func Open(o driver.Opener) (*OLED, error) {
	w := ssd1306_LCDWIDTH
	h := ssd1306_LCDHEIGHT
	dev, err := i2c.Open(o, addr)
	if err != nil {
		return nil, err
	}
	if err := dev.Write(initSeq); err != nil {
		return nil, err
	}
	buf := make([]byte, w*(h/8)+1)
	buf[0] = 0x40 // start frame of pixel data
	return &OLED{dev: dev, w: w, h: h, buf: buf}, nil
}

// OpenWithI2c create an OLED object using a giving i2cDevice . Once not in use, it needs to
// be close by calling Close.
// The default width is 128, height is 64 if zero values are given.
func OpenWithI2c(i2cDevice *i2c.Device, height int) (*OLED, error) {
	w := ssd1306_LCDWIDTH
	h := height

	buf := make([]byte, w*(h/8)+1)
	buf[0] = 0x40 // start frame of pixel data
	return &OLED{dev: i2cDevice, w: w, h: h, buf: buf}, nil
}

// On turns on the display if it is off.
func (o *OLED) On() error {
	return o.dev.Write([]byte{ssd1306_DISPLAY_ON})
}

// Off turns off the display if it is on.
func (o *OLED) Off() error {
	return o.dev.Write([]byte{ssd1306_DISPLAY_OFF})
}

// Clear clears the entire display.
func (o *OLED) Clear() error {
	for i := 1; i < len(o.buf); i++ {
		o.buf[i] = 0
	}
	return o.Draw()
}

func (o *OLED) SetPixel(x, y int, v byte) error {
	if x >= o.w || y >= o.h {
		return fmt.Errorf("(x=%v, y=%v) is out of bounds on this %vx%v display", x, y, o.w, o.h)
	}
	if v > 1 {
		return fmt.Errorf("value needs to be either 0 or 1; given %v", v)
	}
	i := 1 + x + (y/8)*o.w
	if v == 0 {
		o.buf[i] &= ^(1 << uint((y & 7)))
	} else {
		o.buf[i] |= 1 << uint((y & 7))
	}
	return nil
}

// SetImage draws an image on the display buffer starting from x, y.
// A call to Draw is required to display it on the OLED display.
func (o *OLED) SetImage(x, y int, img image.Image) error {
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()

	endX := x + imgW
	endY := y + imgH

	if endX >= o.w {
		endX = o.w
	}
	if endY >= o.h {
		endY = o.h
	}

	var imgI, imgY int
	for i := x; i < endX; i++ {
		imgY = 0
		for j := y; j < endY; j++ {
			r, g, b, _ := img.At(imgI, imgY).RGBA()
			var v byte
			if r+g+b > 0 {
				v = 0x1
			}
			if err := o.SetPixel(i, j, v); err != nil {
				return err
			}
			imgY++
		}
		imgI++
	}
	return nil
}

// Draw draws the intermediate pixel buffer on the display.
// See SetPixel and SetImage to mutate the buffer.
func (o *OLED) Draw() error {
	if err := o.dev.Write([]byte{
		0xa4,     // write mode
		0x40 | 0, // start line = 0
		0x21, 0, ssd1306_LCDWIDTH,
		0x22, 0, 7,
	}); err != nil { // the write mode
		return err
	}
	return o.dev.Write(o.buf)
}

// EnableScroll starts scrolling in the horizontal direction starting from
// startY column to endY column.
func (o *OLED) EnableScroll(startY, endY int) error {
	panic("not implemented")
}

// DisableScroll stops the scrolling on the display.
func (o *OLED) DisableScroll() error {
	panic("not implemented")
}

// Width returns the display width.
func (o *OLED) Width() int { return o.w }

// Height returns the display height.
func (o *OLED) Height() int { return o.h }

// Close closes the display.
func (o *OLED) Close() error {
	return o.dev.Close()
}
