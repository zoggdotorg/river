package main

import (
	"bytes"
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
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
var imageTime time.Time
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
}

func main() {
	log.Println("River Starting.")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	image := getImage()
	if image != nil {
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

	now := time.Now()
	if now.Sub(imageTime).Seconds() > 2 {
		log.Println("Generating Image")

		// Remove existing image file.
		_ = os.Remove("pic.jpeg")

		// Generate new image.
		cmd := exec.Command("/usr/bin/streamer", "-c", getVideo(), "-o", "pic.jpeg", "-s", "640x480", "-j", "90")
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			return nil
		}

		// Open the image file.
		file, err := os.Open("pic.jpeg")
		if err != nil {
			log.Panic(err)
		}
		defer file.Close()

		// Decode the image.
		src, _, err := image.Decode(file)
		if err != nil {
			log.Panic(err)
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
		pt := freetype.Pt(10, 10+int(c.PointToFix32(size)>>8))
		_, err = c.DrawString(label, pt)
		if err != nil {
			log.Panic(err)
		}

		// Produce jpeg of new image.
		newImage := new(bytes.Buffer)
		err = jpeg.Encode(newImage, rgba, &jpeg.Options{Quality: 90})
		if err != nil {
			log.Panic(err)
		}
		imageTime = now
		imageBytes = newImage.Bytes()

	}
	return imageBytes
}
