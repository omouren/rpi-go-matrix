package gomatrix

import (
	"github.com/stianeikeland/go-rpio"
	"github.com/toelsiba/fopix"
	"github.com/disintegration/imaging"
	"image"
	"time"
	"sync"
)

type MatrixPins struct {
	r1 rpio.Pin
	r2 rpio.Pin
	g1 rpio.Pin
	g2 rpio.Pin
	b1 rpio.Pin
	b2 rpio.Pin
	a rpio.Pin
	b rpio.Pin
	c rpio.Pin
	d rpio.Pin
	oe rpio.Pin
	clk rpio.Pin
	lat rpio.Pin
}

type Matrix struct {
	sync.Mutex
	Pixels [32][64]*Color
	pins *MatrixPins
}

type Color struct {
	R uint32
	G uint32
	B uint32
	A uint32
}

type Text struct {
	Content string
	Font *fopix.Font
	Image *image.NRGBA
	Scale float32
	X int
	Y int
}

type Picture struct {
	Path string
	Image *image.NRGBA
	Scale float32
	X int
	Y int
}

func GetAdafruitHatPins() (pins *MatrixPins, err error) {
	err = rpio.Open()

	if err != nil {
		return
	}

	pins = &MatrixPins {
		r1: rpio.Pin(5),
		r2: rpio.Pin(12),
		g1: rpio.Pin(13),
		g2: rpio.Pin(16),
		b1: rpio.Pin(6),
		b2: rpio.Pin(23),
		a: rpio.Pin(22),
		b: rpio.Pin(26),
		c: rpio.Pin(27),
		d: rpio.Pin(20),
		oe: rpio.Pin(4),
		clk: rpio.Pin(17),
		lat: rpio.Pin(21),
	}

	return
}

func NewMatrixPins(config struct {
		r1 uint8
		r2 uint8
		g1 uint8
		g2 uint8
		b1 uint8
		b2 uint8
		a uint8
		b uint8
		c uint8
		d uint8
		oe uint8
		clk uint8
		lat uint8
	}) (pins *MatrixPins, err error) {
	err = rpio.Open()

	if err != nil {
		return
	}

	pins = &MatrixPins{
		r1: rpio.Pin(config.r1),
		r2: rpio.Pin(config.r2),
		g1: rpio.Pin(config.g1),
		g2: rpio.Pin(config.g2),
		b1: rpio.Pin(config.b1),
		b2: rpio.Pin(config.b2),
		a: rpio.Pin(config.a),
		b: rpio.Pin(config.b),
		c: rpio.Pin(config.c),
		d: rpio.Pin(config.d),
		oe: rpio.Pin(config.oe),
		clk: rpio.Pin(config.clk),
		lat: rpio.Pin(config.lat),
	}

	return
}


func (pins *MatrixPins) setRow(x uint) {
	aBit := x & 1
	bBit := x & 2
	cBit := x & 4
	dBit := x & 8

	if (aBit != 0) {
		pins.a.High()
	} else {
		pins.a.Low()
	}

	if (bBit != 0) {
		pins.b.High()
	} else {
		pins.b.Low()
	}

	if (cBit != 0) {
		pins.c.High()
	} else {
		pins.c.Low()
	}

	if (dBit != 0) {
		pins.d.High()
	} else {
		pins.d.Low()
	}
}

func (pins *MatrixPins) clock() {
	pins.clk.High()
	pins.clk.Low()
}

func (pins *MatrixPins) latch() {
	pins.lat.High()
	pins.lat.Low()
}

func (pins *MatrixPins) setPixels(topColor *Color, bottomColor *Color) {
	if (topColor != nil) {
		if (topColor.R > 0) {
			pins.r1.High()
		} else {
			pins.r1.Low()
		}
		
		if (topColor.G > 0) {
			pins.g1.High()
		} else {
			pins.g1.Low()
		}
		
		if (topColor.B > 0) {
			pins.b1.High()
		} else {
			pins.b1.Low()
		}
	} else {
		pins.r1.Low()
		pins.g1.Low()
		pins.b1.Low()
	}

	if (bottomColor != nil) {
		if (bottomColor.R > 0) {
			pins.r2.High()
		} else {
			pins.r2.Low()
		}

		if (bottomColor.G > 0) {
			pins.g2.High()
		} else {
			pins.g2.Low()
		}

		if (bottomColor.B > 0) {
			pins.b2.High()
		} else {
			pins.b2.Low()
		}
	} else {
		pins.r2.Low()
		pins.g2.Low()
		pins.b2.Low()
	}
}


func NewMatrix(pins *MatrixPins) (matrix *Matrix, err error) {
	pins.r1.Output()
	pins.r2.Output()
	pins.g1.Output()
	pins.g2.Output()
	pins.b1.Output()
	pins.b2.Output()
	pins.a.Output()
	pins.b.Output()
	pins.c.Output()
	pins.d.Output()
	pins.oe.Output()
	pins.clk.Output()
	pins.lat.Output()

	pins.r1.Low()
	pins.r2.Low()
	pins.g1.Low()
	pins.g2.Low()
	pins.b1.Low()
	pins.b2.Low()
	pins.a.Low()
	pins.b.Low()
	pins.c.Low()
	pins.d.Low()
	pins.oe.Low()
	pins.clk.Low()
	pins.lat.Low()

	c := make(chan int)

	matrix = &Matrix{pins: pins}

	go matrix.refreshDisplay(c)

	return
}

func (matrix *Matrix) Fill(color Color) {
	for row := 0; row < 32; row++ {
		for col := 0; col < 64; col++ {
			matrix.Pixels[row][col] = matrix.Pixels[row][col].merge(color)
		}
	}
}

func (matrix *Matrix) PrintText(text *Text, color Color) {
	rectMatrix := image.Rectangle{image.Point{0,0}, image.Point{64,32}}
	rectText := image.Rectangle{image.Point{text.X,text.Y}, image.Point{text.Image.Rect.Max.X,text.Image.Rect.Max.Y}}

	if (rectText.In(rectMatrix) || rectText.Overlaps(rectMatrix)) {
		startRow := 0
		startCol := 0
		startImageX := 0
		startImageY := 0

		if (text.Y >= 0) {
			startRow = text.Y
		} else {
			startImageY = text.Image.Rect.Max.Y - (text.Y + text.Image.Rect.Max.Y)
		}

		if (text.X >= 0) {
			startCol = text.X
		} else {
			startImageX = text.Image.Rect.Max.X - (text.X + text.Image.Rect.Max.X)
		}

		for row := startRow; row < 32; row++ {
			for col := startCol; col < 64; col++ {
				_,_,_,a := text.Image.At(col-startCol+startImageX, row-startRow+startImageY).RGBA()
				if (a > 0) {
					matrix.Pixels[row][col] = matrix.Pixels[row][col].merge(color)
				}
			}
		}
	}
}

func (matrix *Matrix) PrintPicture(picture *Picture) {
	rectMatrix := image.Rectangle{image.Point{0,0}, image.Point{64,32}}
	rectText := image.Rectangle{image.Point{picture.X,picture.Y}, image.Point{picture.Image.Rect.Max.X,picture.Image.Rect.Max.Y}}

	if (rectText.In(rectMatrix) || rectText.Overlaps(rectMatrix)) {
		startRow := 0
		startCol := 0
		startImageX := 0
		startImageY := 0

		if (picture.Y >= 0) {
			startRow = picture.Y
		} else {
			startImageY = picture.Image.Rect.Max.Y - (picture.Y + picture.Image.Rect.Max.Y)
		}

		if (picture.X >= 0) {
			startCol = picture.X
		} else {
			startImageX = picture.Image.Rect.Max.X - (picture.X + picture.Image.Rect.Max.X)
		}

		for row := startRow; row < 32; row++ {
			for col := startCol; col < 64; col++ {
				r,g,b,a := picture.Image.At(col-startCol+startImageX, row-startRow+startImageY).RGBA()
				if (a > 0) {
					matrix.Pixels[row][col] = matrix.Pixels[row][col].merge(Color{r,g,b,a})
				}
			}
		}
	}
}

func (matrix *Matrix) DrawSquare(x, y, width, height int, color Color) {
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			matrix.Pixels[row][col] = matrix.Pixels[row][col].merge(color)
		}
	}
}

func (matrix *Matrix) refreshDisplay(c chan int) {
	for {
		matrix.Lock()
		for row := uint(0); row < 32; row++ {
			matrix.pins.oe.Low()
			for col := uint(0); col < 64; col++ {
				matrix.pins.setPixels(matrix.Pixels[row][col], matrix.Pixels[row+16][col])
				matrix.pins.clock()
			}

			matrix.pins.oe.High()
			matrix.pins.setRow(row)
			matrix.pins.latch()
		}
		matrix.Unlock()
	}
	
	c <- 1
}


func (matrix *Matrix) NewText(content string, scale float32, x, y int, font *fopix.Font) *Text {
	return &Text{
		Content: content,
		Image: matrix.NewTextImage(content, scale, font),
		Scale: scale,
		X: x,
		Y: y,
		Font: font,
	}
}

func (matrix *Matrix) NewTextImage(content string, scale float32, font *fopix.Font) *image.NRGBA {
	if (scale >= 1.0) {
		font.Scale(int(scale))
	} else {
		font.Scale(1)
	}
	
	m := image.NewNRGBA(font.GetTextBounds(content))
	font.DrawText(m, image.ZP, content)

	if (scale < 1) {
		m = imaging.Resize(m, int(float32(m.Rect.Max.X) * scale), int(float32(m.Rect.Max.Y) * scale), imaging.NearestNeighbor)
	}

	return m
}

func (matrix *Matrix) NewPicture(path string, scale float32, x, y int) (picture *Picture, err error) {
	m, err := imaging.Open(path)
	if err != nil {
		return
	}

	if (scale != 1) {
		m = imaging.Resize(m.(*image.NRGBA), int(float32(m.(*image.NRGBA).Rect.Max.X) * scale), int(float32(m.(*image.NRGBA).Rect.Max.Y) * scale), imaging.NearestNeighbor)
	}

	picture = &Picture{
		Path: path,
		Image: m.(*image.NRGBA),
		Scale: scale,
		X: x,
		Y: y,
	}

	return
}

func (text *Text) Scroll(horizontal, vertical int, interval time.Duration) {
	x := text.X
	y := text.Y

	for {
		time.Sleep(interval)

		x += horizontal
		y += vertical

		if (horizontal > 0 && x > 64) {
			x = -text.Image.Rect.Max.X
		} else if (horizontal < 0 && x < -text.Image.Rect.Max.X) {
			x = 64 + text.Image.Rect.Max.X
		}

		if (vertical > 0 && y > 32) {
			y = -text.Image.Rect.Max.Y
		} else if (vertical < 0 && y < -text.Image.Rect.Max.Y) {
			y = 32 + text.Image.Rect.Max.Y
		}

		text.X = x
		text.Y = y
	}
}

func (text *Text) Center(x, y bool) {
	if (x) {
		text.X = int(64 / 2 - text.Image.Rect.Max.X / 2)
	}

	if (y) {
		text.Y = int(32 / 2 - text.Image.Rect.Max.Y / 2)
	}
}


func (color *Color) merge(color2 Color) *Color{
	if (color == nil) {
		return &Color{color2.R, color2.G, color2.B, color2.A}
	}

	if (color2.A > 0) {
		color.R = color2.R
		color.G = color2.G
		color.B = color2.B
		color.A = color2.A

		return color;
	}

	if (color.R > 0 || color2.R > 0) {
		color.R = 1
	}
	if (color.G > 0 || color2.G > 0) {
		color.G = 1
	}
	if (color.B > 0 || color2.B > 0) {
		color.B = 1
	}

	return color
}