package processing

import (
	"github.com/disintegration/imaging"
	"image"
)

func Resize(img image.Image, md5 string, extension string, width int, height int) string {
	resizedImage := imaging.Resize(img, width, height, imaging.Lanczos)

	newFileName := "resized_" + md5 + "." + extension
	imaging.Save(resizedImage, newFileName)

	return newFileName
}
