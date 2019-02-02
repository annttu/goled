package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/graphics-go/graphics"
	"golang.org/x/image/bmp"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"image/draw"
	"os"
	"time"
)

func getSpeed() {
	speed = 0

	for {
		timeout := time.Duration(24*7 * time.Hour)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get("https://valas.netcrew.fi/opus/ifstats/v1/stream/ixmon0,ixmon1")
		if err != nil {
			panic(err)
		}
		reader := bufio.NewReader(resp.Body)
		var m map[string]interface{}
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Printf("Error: %v", err)
				break
			} else {
				if len(line) > 15 {
					//fmt.Printf("data %+v", line[5:])
					json.Unmarshal(line[5:], &m)
					//fmt.Printf("data %+v\n", m["bps"])
					speed = uint64(m["bps"].(float64) / 1024 / 1024)
				}
			}

		}
	}
}

func loadSpriteBMP(filename string) image.Image {
	fd, err := os.Open(fmt.Sprintf(filename, ))
	if err != nil {
		panic(err)
	}
	baseimg, err := bmp.Decode(fd)
	if err != nil {
		panic(err)
	}
	return baseimg
}

func loadSpritePNG(filename string) image.Image {
	fd, err := os.Open(fmt.Sprintf(filename, ))
	if err != nil {
		panic(err)
	}
	baseimg, err := png.Decode(fd)
	if err != nil {
		panic(err)
	}
	return baseimg
}

func speedStream() {
	var ypos = 87
	var xpos = 15
	var r = 32
	var b = 32
	var g = 32
	var i = 0
	var history [192]int
	//baseimg := loadSpriteBMP("varoituskolmio.bmp")
	basemopo := loadSpritePNG("mopo_white.png")
	basekyltti := loadSpriteBMP("lisakyltti.bmp")
	//bowsette := loadSpritePNG("bowsette.png")
	tatti := loadSpritePNG("tatti.png")
	for {
		img := image.NewRGBA(image.Rect(0, 0, size_x*16, size_y*16))
		draw.Draw(img, img.Bounds(), basekyltti, image.Point{0,-72}, draw.Src)
		rotatedMopo := image.NewRGBA(image.Rect(0, 0, basemopo.Bounds().Dy(), basemopo.Bounds().Dx()))
		graphics.Rotate(rotatedMopo, basemopo, &graphics.RotateOptions{math.Pi / -2.5 * float64(speed) / 10240})
		col := color.RGBA{uint8(r%255), uint8(g%255), uint8(b%255), 255}
		addLabel(img, xpos, ypos, fmt.Sprintf("%003d", speed), col)


		for i := 0; i < 191; i++ {
			history[i] = history[i+1]
			img.Set(0+i, 96 - history[i], color.RGBA{181,101,29, 255})
			img.Set(0+i, 96 - history[i]+1, color.RGBA{64,255,64, 255})
			img.Set(0+i, 96 - history[i]+2, color.RGBA{64,255,64, 255})
			img.Set(0+i, 96 - history[i]+3, color.RGBA{64,255,64, 255})
		}
		history[191] = int(speed / 160)

		//draw.Draw(img, img.Bounds(), baseimg, image.Point{0,0}, draw.Over)
		draw.Draw(img, img.Bounds(), rotatedMopo, image.Point{-77 ,-96 + history[75+32]+29}, draw.Over)

		//draw.Draw(img, img.Bounds(), bowsette, image.Point{-130,-96 + history[155]+70}, draw.Over)
		draw.Draw(img, img.Bounds(), tatti, image.Point{-150,-96 + history[155]+30}, draw.Over)


		drawImage(img, uint8(i%255))


		<-time.After(10*time.Millisecond)
		i += 1
	}


}


func mopo() {
	go getSpeed()
	speedStream()
}