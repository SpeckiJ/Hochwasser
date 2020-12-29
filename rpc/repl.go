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

	"github.com/SpeckiJ/Hochwasser/render"
)

// Fluter implements flut operations that can be triggered via a REPL
type Fluter interface {
	getTask() FlutTask
	applyTask(FlutTask)
	stopTask()
	toggleMetrics()
}

const commandMode = "CMD"
const textMode = "TXT"

// RunREPL starts reading os.Stdin for commands to apply to the given Fluter
func RunREPL(f Fluter) {
	mode := commandMode
	textSize := 10
	var textCol image.Image = image.White
	var bgCol image.Image = image.Transparent

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inputStr := scanner.Text()

		switch mode {
		case textMode:
			if inputStr == commandMode {
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
			switch cmd {
			case "stop":
				f.stopTask()

			case "start":
				f.applyTask(f.getTask())

			case "offset":
				if len(args) == 2 {
					x, err := strconv.Atoi(args[0])
					y, err2 := strconv.Atoi(args[1])
					if err == nil && err2 == nil {
						t := f.getTask()
						t.Offset = image.Pt(x, y)
						f.applyTask(t)
					}
				}

			case "conns":
				if len(args) == 1 {
					if conns, err := strconv.Atoi(args[0]); err == nil {
						t := f.getTask()
						t.MaxConns = conns
						f.applyTask(t)
					}
				}

			case "shuffle":
				t := f.getTask()
				t.Shuffle = !t.Shuffle
				f.applyTask(t)

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
					fmt.Printf("[rán] text mode, return via %v\n", commandMode)
					mode = textMode
				} else {
					input := strings.Join(args[3:], " ")
					t := f.getTask()
					t.Img = render.RenderText(input, textSize, textCol, bgCol)
					f.applyTask(t)
				}

			case "img":
				if len(args) > 0 {
					path := strings.Join(args, " ")
					t := f.getTask()
					if img, err := render.ReadImage(path); err != nil {
						fmt.Println(err)
					} else {
						t.Img = render.ImgToNRGBA(img)
						f.applyTask(t)
					}
				}

			case "metrics":
				f.toggleMetrics()

			}
		}
	}
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
		return &render.StripePattern{Palette: pal}
	}

	if p, ok := render.DynPatterns[input]; ok {
		return p
	}

	return image.Transparent
}
