package main

import (
    "bufio"
    "bytes"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "golang.org/x/image/bmp"
    "golang.org/x/image/font"
    "golang.org/x/image/font/basicfont"
    "golang.org/x/image/math/fixed"
    "github.com/BurntSushi/graphics-go/graphics"
    "image"
    "image/color"
    "image/draw"
    "image/png"
    "math"
    "net"
    "net/http"
    "os"
    "time"
)

const size_x = 12
const size_y = 6
const DST_IP = "192.168.10.240"
const DST_PORT = 9998
const brightness = 0.3
const image_count = 5000

type Frame struct {
    frametype uint8
    frame uint8
    xpos int16
    ypos int16
    wat uint16
    data [16*16*3*2]uint8
}

var conn net.Conn;
var speed uint64 = 0;

func main() {

    var err error
    conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", DST_IP, DST_PORT))
    if err != nil {
        fmt.Printf("Some error %v", err)
        return
    }
    videoStream()
    //
    //go getSpeed()
    //speedStream()
    //textStream()
}


func getSpeed() {
    speed = 0

    for {

        resp, err := http.Get("https://valas.netcrew.fi/opus/ifstats/v1/stream/ixmon0,ixmon1")
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
                    fmt.Printf("data %+v\n", m["bps"])
                    speed = uint64(m["bps"].(float64) / 1024 / 1024)
                }
            }

        }
    }
}

func speedStream() {
    var factor = 0
    var xVector = 1*factor
    var yVector = 1*factor
    var ypos = 87
    var xpos = 15
    var r = 32
    var b = 32
    var g = 32
    var i = 0
    fd, err := os.Open(fmt.Sprintf("varoituskolmio.bmp", ))
    if err != nil {
        panic(err)
    }
    baseimg, err := bmp.Decode(fd)
    if err != nil {
        panic(err)
    }
    fd, err = os.Open(fmt.Sprintf("mopo.png", ))
    if err != nil {
        panic(err)
    }
    basemopo, err := png.Decode(fd)
    if err != nil {
        panic(err)
    }
    fd, err = os.Open(fmt.Sprintf("lisakyltti.bmp", ))
    if err != nil {
        panic(err)
    }
    basekyltti, err := bmp.Decode(fd)
    if err != nil {
        panic(err)
    }
    for {
            img := image.NewRGBA(image.Rect(0, 0, size_x*16, size_y*16))
            draw.Draw(img, img.Bounds(), baseimg, image.Point{0,0}, draw.Src)
            draw.Draw(img, img.Bounds(), basekyltti, image.Point{0,-72}, draw.Src)
            rotatedMopo := image.NewRGBA(image.Rect(0, 0, basemopo.Bounds().Dy(), basemopo.Bounds().Dx()))
            graphics.Rotate(rotatedMopo, basemopo, &graphics.RotateOptions{math.Pi / 2.0 * float64(speed) / 10240})
            draw.Draw(img, img.Bounds(), rotatedMopo, image.Point{-25 ,-30}, draw.Over)
            col := color.RGBA{uint8(r%255), uint8(g%255), uint8(b%255), 255}
            addLabel(img, xpos, ypos, fmt.Sprintf("%003d", speed), col)
            drawImage(img, uint8(i%255))
            <-time.After(10*time.Millisecond)
            i += 1
            if xpos >= 16*size_x - 70 || xpos == 0 {
                xVector = -1*factor
                b += 16
                r += 32
                g += 64
            }
            if ypos >= 16*size_y {
                yVector = -1*factor
                b += 64
            }
            if xpos <= 0 {
                xVector = 1*factor
                g += 64
            }
            if ypos <= 10 {
                yVector = 1*factor
                r += 64
            }
            xpos += xVector
            ypos += yVector
    }


}


func textStream() {
    var factor = 2
    var xVector = 1*factor
    var yVector = 1*factor
    var ypos = 20
    var xpos = 20
    var r = 128
    var b = 128
    var g = 128
    var i = 0
    for {
        img := image.NewRGBA(image.Rect(0, 0, size_x*16, size_y*16))
        col := color.RGBA{uint8(r%255), uint8(g%255), uint8(b%255), 255}
        addLabel(img, xpos, ypos, fmt.Sprintf("Assembly <3"), col)
        drawImage(img, uint8(i%255))
        <-time.After(20*time.Millisecond)
        i += 1
        //speed += 1
        if xpos >= 16*size_x - 80 || xpos == 0 {
            xVector = -1*factor
            b += 16
            r += 32
            g += 64
        }
        if ypos >= 16*size_y {
            yVector = -1*factor
            b += 64
        }
        if xpos <= 0 {
            xVector = 1*factor
            g += 64
        }
        if ypos <= 10 {
            yVector = 1*factor
            r += 64
        }
        xpos += xVector
        ypos += yVector
    }


}

func videoStream() {
    var images [image_count]image.Image

    go func () {
        for id := 1; id < image_count; id++ {
            fd, err := os.Open(fmt.Sprintf("/tmp/thumb%05d.bmp", id))
            if err != nil {
                panic(err)
            }
            images[id], err = bmp.Decode(fd)
            if err != nil {
                panic(err)
            }
            if (id % 10) == 0 {
                fmt.Printf("\rLoading: %d/%d", id, image_count)
            }
            //fmt.Printf("Bounds: %+v\n", images[id].Bounds())

        }
        fmt.Printf("\n")
    }()


    <-time.After(1*time.Second)

    c1 := make(chan bool, 1)

    go func () {
        for {
            time.Sleep(33 * time.Millisecond)
            c1 <- true
        }
    }()

    var frameId = 0
    for {
        select {
        case <-c1:
            frameId += 1
            drawImage(images[frameId], uint8(frameId%254))
            if frameId >= image_count {
                break
            }
        }
    }
}

func addLabel(img *image.RGBA, x, y int, label string, col color.RGBA) {

    //point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
    point := fixed.Point26_6{fixed.Int26_6(x*64), fixed.Int26_6(y * 64)}

    d := &font.Drawer{
        Dst:  img,
        Src:  image.NewUniform(col),
        Face: basicfont.Face7x13,
        Dot:  point,
    }
    d.DrawString(label)
}


func drawImage(img image.Image, frameId uint8) {

    var y_tile, x_tile int
    var i int
    var frame Frame
    var err error
    var bin_buf bytes.Buffer
        for y_tile = 0; y_tile < size_y; y_tile++ {
            for x_tile = 0; x_tile < size_x; x_tile++ {
                frame = Frame{}
                frame.frametype = 1
                frame.frame = uint8(frameId % 254)
                frame.xpos = int16(x_tile*16)
                frame.ypos = int16(y_tile*16)
                i = 0
                for y_pixel := y_tile*16; y_pixel < y_tile*16+16; y_pixel += 1 {
                    for x_pixel := x_tile*16; x_pixel < x_tile*16+16; x_pixel += 1 {
                        pixel := img.At(x_pixel, y_pixel)
                        r, g, b, _ := pixel.RGBA()
                        r = uint32(math.Pow(float64(r), 2.2) / 700000 * brightness)
                        b = uint32(math.Pow(float64(b), 2.2) / 700000 * brightness)
                        g = uint32(math.Pow(float64(g), 2.2) / 700000 * brightness)


                        frame.data[i] = uint8(r & 0xff)
                        frame.data[i+1] = uint8(r >> 8)
                        i += 2
                        frame.data[i] = uint8(g & 0xff)
                        frame.data[i+1] = uint8(g >> 8)
                        i += 2
                        frame.data[i] = uint8(b & 0xff)
                        frame.data[i+1] = uint8(b >> 8)
                        i += 2

                    }
                }
                err = binary.Write(&bin_buf, binary.LittleEndian, frame)
                if err != nil {
                    fmt.Printf("Some error %v", err)
                    return
                }
                sendFrame(bin_buf.Bytes())
                bin_buf.Reset()
            }
        }
        frame = Frame{}
        frame.frametype = 2
        frame.frame = uint8(frameId % 254)
        frame.xpos = 0
        frame.ypos = 0
        /*for i= 0; i < 1536; i++ {
            frame.data[i] = 0
        }*/

        binary.Write(&bin_buf, binary.LittleEndian, frame)
        <-time.After(1*time.Millisecond)
        sendFrame(bin_buf.Bytes())

        //fmt.Printf("Frame %d\n", frameId)
        bin_buf.Reset()
}

func sendFrame(frame []byte) {
        conn.Write(frame)
}
