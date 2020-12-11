package screen

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"os"
	"testing"
)

func TestCaptureScreen(t *testing.T) {
	img := CaptureScreen(300, 400, 100)
	data := bytes.NewBuffer(nil)
	err := jpeg.Encode(data, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
	if err != nil {
		t.Fatal(err)
	}
	img.Bounds()
	file, err := os.OpenFile("1.jpeg", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	n, err := file.Write(data.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(n)
	r, g, b := AverageColor(img)
	fmt.Printf("%d, %d, %d\n", r, g, b)
}
