package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math/rand"
)

func randomColor() color.RGBA {
	c := uint8(rand.Intn(255))
	return color.RGBA{c, c, c, 255}
}

// ランダムな画像の生成
func randomImage() ([]byte, error) {
	size := image.Rect(0, 0, 640, 480)
	img := image.NewRGBA(size)

	for x := 0; x < size.Dx(); x++ {
		for y := 0; y < size.Dy(); y++ {
			img.SetRGBA(x, y, randomColor())
		}
	}

	buf := bytes.NewBuffer([]byte{})
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var (
	randomStringPrefixes = []string{
		"Hello",
		"Hi",
		"Yay",
		"Oh",
		"Wow",
	}
	randomStringSuffixes = []string{
		"World",
		"Baby",
		"Image",
		"My Photo",
		"Great Picture",
	}
)

func randomText() string {
	prefix := randomStringPrefixes[rand.Intn(len(randomStringPrefixes))]
	suffix := randomStringSuffixes[rand.Intn(len(randomStringSuffixes))]

	return prefix + ", " + suffix
}
