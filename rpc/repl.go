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
	textSize := 4
	textCol := color.NRGBA{0xff, 0xff, 0xff, 0xff}

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
			t.Img = render.RenderText(inputStr, textSize, textCol)
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
				fmt.Printf("[rán] text mode, return via %v\n", commandMode)
				mode = textMode
				if len(args) > 0 {
					if size, err := strconv.Atoi(args[0]); err == nil {
						textSize = size
					}
				}
				if len(args) > 1 {
					if col, err := hex.DecodeString(args[1]); err == nil {
						textCol = color.NRGBA{col[0], col[1], col[2], 0xff}
					}
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
