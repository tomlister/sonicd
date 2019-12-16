package main

import (
	"image"
	//"log"
	"os"
	"fmt"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	//"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/BurntSushi/freetype-go/freetype/truetype"
	"github.com/itchyny/volume-go"
	"sync"
	"time"
)

func e(err error) {
	if err != nil {
		panic(err.Error())
	}
}

var (
	canvasWidth, canvasHeight = 200, 45

	bg = xgraphics.BGRA{B: 0x00, G: 0x00, R: 0x00, A: 0x00}

	fg = xgraphics.BGRA{B: 0xff, G: 0xff, R: 0xff, A: 0xff}

	fontPath = "/usr/share/fonts/truetype/noto/NotoMono-Regular.ttf"

	iconFontPath = "/usr/share/fonts/truetype/font-awesome/fontawesome-webfont.ttf"

	size = 20.0

)

func close_window(win *xwindow.Window) {
	win.Unmap()
}

func clear_canvas(ximg *xgraphics.Image, win *xwindow.Window) {
	ximg.For(func(x,y int) xgraphics.BGRA {
		return bg
	})
	ximg.XDraw()
	ximg.XPaint(win.Id)
}

func draw_volume(font *truetype.Font, win *xwindow.Window, canvas *xgraphics.Image) {
	clear_canvas(canvas, win)
	vol, err := volume.GetVolume()
	e(err)

	displaytext := fmt.Sprintf("Volume: %d%%", vol)

	firstw, firsth := xgraphics.Extents(font, size, displaytext)
	bounds := image.Rect(10, 10+firsth, 10+firstw, 10+firsth+firstw)

	text_sub_image := canvas.SubImage(bounds).(*xgraphics.Image)
	_, _, err = canvas.Text(10, 10, fg, size, font, displaytext)
	_, _, err = text_sub_image.Text(10, 10, fg, size, font, displaytext)
	e(err)
	win.Map()
	canvas.XDraw()
	canvas.XPaint(win.Id)
}

func main() {
	X, err := xgbutil.NewConn()
	e(err)

	fontReader, err := os.Open(fontPath)
	e(err)

	font, err := xgraphics.ParseFont(fontReader)
	e(err)

	command_queue := make([]string, 0)

	command_queue_mutex := &sync.Mutex{}

	go usock_server(&command_queue, command_queue_mutex)

	visible := true
	timeout := 10

	canvas := xgraphics.New(X, image.Rect(0, 0, canvasWidth, canvasHeight))

	//win := canvas.XShowExtra("sonicd", true)
	win, err := xwindow.Generate(X)
	e(err)
	win.Create(canvas.X.RootWin(), 1700, 1015, canvasWidth, canvasHeight, 0)
	//clear_canvas(canvas, win)
	canvas.XSurfaceSet(win.Id)
	canvas.XDraw()
	canvas.XPaint(win.Id)
	win.Map()

	for {
		command_queue_mutex.Lock()
		temp_queue := command_queue
		for i := 0; i < len(command_queue); i++ {
			if (command_queue[i] == "show") {
				draw_volume(font, win, canvas)
				visible = true
				timeout = 1000
			} else if (command_queue[i] == "up") {
				vol, err := volume.GetVolume()
				e(err)
				if vol+5 >= 100 {
					err = volume.SetVolume(100)
					e(err)
				} else {
					err = volume.SetVolume(vol+5)
					e(err)
				}
				draw_volume(font, win, canvas)
				visible = true
				timeout = 1000
			} else if (command_queue[i] == "down") {
				vol, err := volume.GetVolume()
				e(err)
				if vol-5 <= 0 {
					err = volume.SetVolume(0)
					e(err)
				} else {
					err = volume.SetVolume(vol-5)
					e(err)
				}
				draw_volume(font, win, canvas)
				visible = true
				timeout = 1000
			} else if (command_queue[i] == "hide") {
				visible = false
				timeout = 0
				close_window(win)
			}
			temp_queue = append(temp_queue[:i], temp_queue[i+1:]...)
		}
		command_queue = temp_queue
		command_queue_mutex.Unlock()
		if (timeout == 0 && visible == true) {
			visible = false
			close_window(win)
		} else if (timeout != 0) {
			timeout--
		}
		time.Sleep(time.Millisecond)
	}

	xevent.Main(X)
}