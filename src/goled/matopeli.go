package main

import (
	"fmt"
	"github.com/pkg/term"
	"image"
	"image/color"
	"math/rand"
	"time"
)

type XY struct {
	x int
	y int
}

func matopeli() {

	<-time.After(1*time.Second)

	c1 := make(chan bool, 1)

	go func () {
		for {
			time.Sleep(33 * time.Millisecond)
			c1 <- true
		}
	}()


	var xdir = 0
	var ydir = 1

	go func() {
		var err error
		t, _ := term.Open("/dev/tty")
		term.RawMode(t)
		bytes_read := make([]byte, 3)
		var numRead int
		for {
			numRead, err = t.Read(bytes_read)
			if err != nil {
				return
			}
			if numRead == 3 && bytes_read[0] == 27 && bytes_read[1] == 91 {
				if bytes_read[2] == 65 {
					if xdir != 0 {
						ydir = -1
						xdir = 0
					}
					fmt.Printf("UP\n")
				} else if bytes_read[2] == 66 {
					if xdir != 0 {
						ydir = 1
						xdir = 0
					}
					fmt.Printf("Down\n")
				} else if bytes_read[2] == 67 {
					if ydir != 0 {
						ydir = 0
						xdir = 1
					}
					fmt.Printf("Right\n")
				} else if bytes_read[2] == 68 {
					if ydir != 0 {
						ydir = 0
						xdir = -1
					}
					fmt.Printf("Left\n")
				}
			}
			fmt.Printf("Key: %+v", numRead)
		}
	}()

	var apples [512]XY
	var mato [512]XY
	var len = 1
	mato[0] = XY{0, 0}
	var i int
	var dead = false
	var a = 0
Game:
	for  i = 0; i < 511; i++{
		randx := rand.Int() % size_x*16
		randy := rand.Int() % size_y*16

		apples[i] = XY{randx, randy}
		var found = false
	Mato:
		for {
			fmt.Printf("Len: %d\n", len)
			found = false
			select {
			case <-c1:
				img := image.NewRGBA(image.Rect(0, 0, size_x*16, size_y*16))
				img.Set(randx, randy, color.RGBA{255, 0, 0, 255})

				if mato[0].y == randy && mato[0].x == randx {
					found = true
				}

				var firstx = 0
				var firsty = 0
				for idx := len; idx > 0; idx-- {
					img.Set(mato[idx].x, mato[idx].y, color.RGBA{uint8(255+(25*idx) % 255), uint8(255-(25*idx) % 255), 255, 255})

					if idx == len {
						firstx = mato[idx].x
						firsty = mato[idx].y
					}

					if idx == 0 {
						continue
					}

					mato[idx].x = mato[idx-1].x
					mato[idx].y = mato[idx-1].y
					fmt.Printf("(%d, %d, %d)", idx, mato[idx].x, mato[idx].y)
				}
				mato[0].x += xdir
				if mato[0].x >= size_x*16 {
					mato[0].x = 0
				} else if mato[0].x < 0 {
					mato[0].x = size_x*16
				}
				mato[0].y += ydir
				if mato[0].y >= size_y*16 {
					mato[0].y = 0
				} else if mato[0].y < 0 {
					mato[0].y = size_y*16
				}
				drawImage(img, uint8(a%254))
				a += 1

				if len > 2 {
					for idx := 1; idx < len; idx++ {
						if mato[0].y == mato[idx].y && mato[0].x == mato[idx].x {
							dead = true
							break Game
						}
					}
				}

				if found {
					mato[len] = XY{firstx, firsty}
					fmt.Printf("Old %d %d, new %d %d", firstx, firsty, mato[len].x, mato[len].x)
					len += 1
					break Mato
				}
			}
		}
	}
	if dead {
		<-time.After(1*time.Second)
		img := image.NewRGBA(image.Rect(0, 0, size_x*16, size_y*16))
		col := color.RGBA{255, 255, 255, 255}
		addLabel(img, 70, 48, fmt.Sprintf("Game over!"), col)
		drawImage(img, uint8(a%254))
		//<-time.After(10*time.Second)
	} else {
		<-time.After(1*time.Second)
		img := image.NewRGBA(image.Rect(0, 0, size_x*16, size_y*16))
		col := color.RGBA{255, 255, 255, 255}
		addLabel(img, 70, 48, fmt.Sprintf("Victory!!"), col)
		drawImage(img, uint8(a%254))
		//<-time.After(10*time.Second)
	}
}

