package main

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

func readImage(path string) (img image.Image) {
	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err2 := image.Decode(reader)
	if err2 != nil {
		log.Fatal(err2)
	}
	return img
}

func writeImage(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}
}

func imgToNRGBA(img image.Image) *image.NRGBA {
	b := img.Bounds()
	r := image.NewNRGBA(b)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			r.Set(x, y, img.At(x, y))
		}
	}
	return r
}
