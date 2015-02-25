package storage

import (
	"bytes"
	"image"
	"image/jpeg"
	"sync"
)

var mutex = &sync.Mutex{}

type Storage struct {
	numberOfFiles int
	pointer       int
}

func InitializeStorage() {
}

func StoreImage(img image.Image) {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		panic(err)
	}
}
