package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "golang.org/x/image/bmp"
    "image"
    "math"
    "net"
    "os"
    "time"
)

const size_x = 12
const size_y = 6
const DST_IP = "192.168.10.240"
const brightness = 0.2
const image_count = 10000

type Frame struct {
    frametype uint8
    frame uint8
    xpos int16
    ypos int16
    data [16*16*3*2]uint8
}

var conn net.Conn;

func main() {

    var err error
    conn, err = net.Dial("udp", "192.168.10.240:9998")
    if err != nil {
        fmt.Printf("Some error %v", err)
        return
    }
    var images [image_count]image.Image

    //fd, err := os.Open("/tmp/thumb01000.bmp")
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
            fmt.Printf("Loading: %d/%d\n", id, image_count)
        }
        //fmt.Printf("Bounds: %+v\n", images[id].Bounds())

    }


    var y_tile, x_tile int
    var i int
    var frame Frame
    var bin_buf bytes.Buffer
    for frame_id := 1; frame_id < len(images); frame_id += 1 {
        for y_tile = 0; y_tile < size_y; y_tile++ {
            for x_tile = 0; x_tile < size_x; x_tile++ {
                frame = Frame{}
                frame.frametype = 1
                frame.frame = uint8(frame_id % 254)
                frame.xpos = int16(x_tile*16)
                frame.ypos = int16(y_tile*16)
                i = 0
                //fmt.Printf("%03d(9) %03d(12)\n", y_tile, x_tile)
                /*for i = 0; i < 1536; i++ {
                    frame.data[i]  = uint8(x_tile*y_tile)
                }*/
                for y_pixel := y_tile*16; y_pixel < y_tile*16+16; y_pixel += 1 {
                    for x_pixel := x_tile*16; x_pixel < x_tile*16+16; x_pixel += 1 {
                        //fmt.Printf("%03d(9) %03d(12) %03d(255) %03d(255) %03d\n", y_tile, x_tile, y_pixel, x_pixel, i)
                        pixel := images[frame_id].At(x_pixel, y_pixel)
                        //fmt.Printf("%+v", img.At(x_pixel, y_pixel))
                        r, g, b, _ := pixel.RGBA()
                        //fmt.Printf("%03d %03d %03d\n", r, b, g)
                        r = uint32(math.Pow(float64(r), 2.2) / 700000 * brightness)
                        b = uint32(math.Pow(float64(b), 2.2) / 700000 * brightness)
                        g = uint32(math.Pow(float64(g), 2.2) / 700000 * brightness)
                        //fmt.Printf("%03d %03d %03d\n", r, b, g)
                        frame.data[i] = uint8(b & 0xff)
                        frame.data[i+1] = uint8(b >> 8)
                        i += 2
                        frame.data[i] = uint8(r & 0xff)
                        frame.data[i+1] = uint8(r >> 8)
                        i += 2
                        frame.data[i] = uint8(g & 0xff)
                        frame.data[i+1] = uint8(g >> 8)
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
        frame.frame = uint8(frame_id % 254)
        frame.xpos = 0
        frame.ypos = 0
        for i= 0; i < 1536; i++ {
            frame.data[i] = 0
        }

        binary.Write(&bin_buf, binary.LittleEndian, frame)
        <-time.After(1*time.Millisecond)
        sendFrame(bin_buf.Bytes())
        <-time.After(10*time.Millisecond)
        fmt.Printf("Frame %d\n", frame_id)
        bin_buf.Reset()
    }
}

func sendFrame(frame []byte) {
        conn.Write(frame)
        //conn.Write(frame[1000:len(frame)])
}
