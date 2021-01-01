package rpc

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/SpeckiJ/Hochwasser/pixelflut"
	"github.com/SpeckiJ/Hochwasser/render"
)

// Fluter implements flut operations that can be triggered via a REPL
type Fluter interface {
	getTask() pixelflut.FlutTask
	applyTask(pixelflut.FlutTask)
	stopTask()
	toggleMetrics()
}

const commandMode = "cmd"
const textMode = "txt"

// RunREPL starts reading os.Stdin for commands to apply to the given Fluter
func RunREPL(f Fluter) {
	mode := commandMode
	textSize := 10
	var textCol image.Image = image.White
	var bgCol image.Image = image.Transparent

	fmt.Print("[rán] REPL is active. ")
	printHelp()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inputStr := scanner.Text()

		switch strings.ToLower(mode) {
		case textMode:
			if strings.ToLower(inputStr) == commandMode {
				fmt.Println("[rán] command mode")
				mode = commandMode
				continue
			}
			t := f.getTask()
			t.Img = render.RenderText(inputStr, textSize, textCol, bgCol)
			f.applyTask(t)

		case commandMode:
			input := strings.Split(inputStr, " ")
			cmd := strings.ToLower(input[0])
			args := input[1:]
			t := f.getTask()
			printTask := true

			switch cmd {
			case "stop":
				t.Paused = true
				f.stopTask()
				printTask = false

			case "start":
				t.Paused = false

			case "offset", "of":
				if len(args) == 1 && args[0] == "rand" {
					t.RandOffset = true
					t.Offset = image.Point{}
				} else if len(args) == 2 {
					t.RandOffset = false
					x, err := strconv.Atoi(args[0])
					y, err2 := strconv.Atoi(args[1])
					if err == nil && err2 == nil {
						t.Offset = image.Pt(x, y)
					}
				}

			case "connections", "c":
				if len(args) == 1 {
					if conns, err := strconv.Atoi(args[0]); err == nil {
						t.MaxConns = conns
					}
				}

			case "address", "a":
				if len(args) == 1 {
					t.Address = args[0]
				}

			case "order", "o":
				if len(args) == 1 {
					t.RenderOrder = pixelflut.NewOrder(args[0])
				}

			case "rgbsplit":
				t.RGBSplit = !t.RGBSplit

			case "txt":
				if len(args) > 0 {
					if size, err := strconv.Atoi(args[0]); err == nil {
						textSize = size
					}
				}
				if len(args) > 1 {
					textCol = parseColorOrPalette(args[1])
				}
				if len(args) > 2 {
					bgCol = parseColorOrPalette(args[2])
				}
				if len(args) < 4 {
					fmt.Printf("[rán] text mode, return via '%v'\n", strings.ToUpper(commandMode))
					mode = textMode
					printTask = false
				} else {
					input := strings.Join(args[3:], " ")
					t.Img = render.RenderText(input, textSize, textCol, bgCol)
				}

			case "img", "i":
				if len(args) > 0 {
					path := strings.Join(args, " ")
					if img, err := render.ReadImage(path); err != nil {
						fmt.Println(err)
						continue
					} else {
						t.Img = img
					}
				}

			case "metrics":
				f.toggleMetrics()
				printTask = false

			default:
				printTask = false
				fmt.Print("[rán] unknown command. ")
				printHelp()
			}

			if printTask {
				fmt.Println(t)
			}
			f.applyTask(t)
		}
	}
}

func printHelp() {
	fmt.Println(`available commands:
	start                                start fluting
	stop                                 pause fluting
	c <n>                                set number of connections per client
	a <host>:<port>                      set target server
	offset <x> <y>                       set top-left offset
	offset rand                          random offset for each draw
	metrics                              toggle bandwidth reporting (may cost some performance)

	i <filepath>                         set image
	txt <scale> <color <bgcolor> <txt>   send text
	txt [<scale> [<color> [<bgcolor>]]   enter interactive text mode
	rgbsplit                             toggle RGB split effect
	o                                    set order (l,r,t,b,shuffle)`)
}

// try to parse as hex-encoded RGB color,
// alternatively treat it as palette name. If both fail,
// give image.Transparent
func parseColorOrPalette(input string) image.Image {
	if input == "w" {
		return image.NewUniform(color.White)
	} else if input == "b" {
		return image.NewUniform(color.Black)
	} else if input == "t" {
		return image.Transparent
	} else if col, err := hex.DecodeString(input); err == nil && len(col) >= 3 {
		var alpha byte = 0xff
		if len(col) == 4 {
			alpha = col[3]
		}
		return image.NewUniform(color.NRGBA{col[0], col[1], col[2], alpha})
	}

	if pal := render.PrideFlags[input]; len(pal) != 0 {
		return &render.StripePattern{Palette: pal, Size: 13}
	}

	if p, ok := render.DynPatterns[input]; ok {
		return p
	}

	return image.Transparent
}
