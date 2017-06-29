package main

import (
	"bytes"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

var font *truetype.Font
var imageBytes []byte
var imageMutex sync.Mutex

func init() {
	// Read the font data.
	fontBytes, err := ioutil.ReadFile("luxisr.ttf")
	if err != nil {
		log.Fatal(err)
	}
	font, err = freetype.ParseFont(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	// To work around bug in image capture vgrabbj that sometimes
	// returns a green screen (RasPi specific library location).
	os.Setenv("LD_PRELOAD", "/usr/lib/arm-linux-gnueabihf/libv4l/v4l1compat.so")
}

func main() {
	log.Println("River Starting.")

	go makeImages(30)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	image := getImage()
	if image == nil {
		fmt.Fprintf(w, "Could not obtain image")
	} else {
		w.Write(image)
	}
}

func getVideo() string {
	if _, err := os.Stat("/dev/video0"); err == nil {
		return "/dev/video0"
	}
	if _, err := os.Stat("/dev/video1"); err == nil {
		return "/dev/video1"
	}
	if _, err := os.Stat("/dev/video2"); err == nil {
		return "/dev/video2"
	}
	if _, err := os.Stat("/dev/video3"); err == nil {
		return "/dev/video3"
	}
	if _, err := os.Stat("/dev/video4"); err == nil {
		return "/dev/video4"
	}
	if _, err := os.Stat("/dev/video5"); err == nil {
		return "/dev/video5"
	}
	if _, err := os.Stat("/dev/video6"); err == nil {
		return "/dev/video6"
	}
	if _, err := os.Stat("/dev/video7"); err == nil {
		return "/dev/video7"
	}
	if _, err := os.Stat("/dev/video8"); err == nil {
		return "/dev/video8"
	}
	return "/dev/video9"
}

func getImage() []byte {
	imageMutex.Lock()
	defer imageMutex.Unlock()
	return imageBytes
}

func PointToInt26_6(x, dpi float64) fixed.Int26_6 {
	return fixed.Int26_6(x * dpi * (64.0 / 72.0))
}

func makeImages(secondsDelay int) {
	delay := 0
	for {
		// Delay between making an image
		time.Sleep(time.Duration(delay) * time.Second)
		delay = secondsDelay

		// Generate new image.
		now := time.Now()
		out, err := exec.Command("/usr/bin/vgrabbj", "-d", getVideo(), "-i", "svga", "-q", "90").Output()
		if err != nil {
			log.Println(err)
			continue
		}

		// Decode the image.
		src, _, err := image.Decode(bytes.NewReader(out))
		if err != nil {
			log.Println(err)
			continue
		}

		// Time
		label := now.Format("2/1/2006 3:04PM")

		// Colour of text
		fg := image.NewUniform(color.RGBA{255, 255, 0, 255})

		// Copy in original image
		rgba := image.NewRGBA(src.Bounds())
		draw.Draw(rgba, rgba.Bounds(), src, image.ZP, draw.Src)

		// Draw the text.
		size := 24.0
		c := freetype.NewContext()
		c.SetDPI(72)
		c.SetFont(font)
		c.SetFontSize(size)
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetSrc(fg)
		pt := freetype.Pt(10, 10+int(PointToInt26_6(size, 72)>>6))
		_, err = c.DrawString(label, pt)
		if err != nil {
			log.Println(err)
			continue
		}

		// Produce jpeg of new image.
		newImage := new(bytes.Buffer)
		err = jpeg.Encode(newImage, rgba, &jpeg.Options{Quality: 90})
		if err != nil {
			log.Println(err)
			continue
		}

		imageMutex.Lock()
		imageBytes = newImage.Bytes()
		imageMutex.Unlock()
	}
}
