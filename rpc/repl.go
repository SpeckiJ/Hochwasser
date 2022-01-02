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
	textSize := 10.0
	var textCol image.Image = image.White
	var bgCol image.Image = image.Transparent
	var taskStore = make(map[string]pixelflut.FlutTask)

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

			case "toggle", ".":
				t.Paused = !t.Paused
				if t.Paused {
					f.stopTask()
					printTask = false
				}

			case "store", "save":
				printTask = false
				if len(args) == 0 {
					fmt.Println("must specify name")
				} else {
					taskStore[strings.Join(args, " ")] = t
				}
				continue

			case "load", "l":
				if len(args) == 0 {
					fmt.Println("must specify name")
				} else {
					t = taskStore[strings.Join(args, " ")]
				}

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

			case "host", "address", "a":
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
						textSize = float64(size)
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

			case "scale", "s":
				quality := true
				var facX, facY float64
				var err error
				if len(args) >= 1 {
					facX, err = strconv.ParseFloat(args[0], 64)
					if err != nil {
						fmt.Println(err)
						continue
					}
					facY = facX
					if len(args) >= 2 {
						facY, err = strconv.ParseFloat(args[1], 64)
						if err != nil {
							fmt.Println(err)
							continue
						}
					}
					if len(args) > 2 {
						quality = false
					}
					t.Img = render.ScaleImage(t.Img, facX, facY, quality)
				}

			case "rotate", "r":
				t.Img = render.RotateImage90(t.Img)

			// the commands below don't affect the task, so we don't need to apply it to clients -> continue

			case "metrics":
				f.toggleMetrics()
				continue

			case "status":
				fmt.Println(t)
				continue

			case "help":
				printHelp()
				continue

			default:
				fmt.Print("[rán] unknown command. ")
				printHelp()
				continue
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
	tasks
		start                                start fluting
		stop                                 pause fluting
		status                               print current task
		save <name>                          store current task
		load <name>							 load previously stored task
	content
		i <filepath>                         set image
		txt <scale> <color <bgcolor> <txt>   send text
		txt [<scale> [<color> [<bgcolor>]]   enter interactive text mode
		scale <facX> [<facY> [lofi]]         scale content
		rotate                               rotate content 90°
	draw modes
		o                                    set order (l,r,t,b,random)
		of <x> <y>                           set top-left offset
		of rand                              random offset for each draw
		rgbsplit                             toggle RGB split effect
	networking
		c <n>                                set number of connections per client
		a <host>:<port>                      set target server
		metrics                              toggle bandwidth reporting (may cost some performance)`)
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
