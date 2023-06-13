package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/cmplx"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

const width float64 = 180.0
const height float64 = 60.0

const pngWidth float64 = 1920 * 4
const pngHeight float64 = 1080 * 4
const pngName string = "binted.png"
const pngIterations int = 800

var updatingIterations = []int{150, 150, 250, 400, 600}

var passFx = []int{3, 2, 2, 1, 1}

const maxSqPasses = 4

var crossHair = true
var hqRender = false

// const single_rune = true

var ux, uy, uz float64

const boundCheck complex128 = 4 + 0i

var power = 2 + 0i
var usePower = false

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	quit := func() {
		s.Fini()
		os.Exit(0)
	}

	var pass = 0

	var renderWg sync.WaitGroup

	// event handling
	go func(s tcell.Screen) {
		for {
			event := s.PollEvent()
			pass = 0

			switch ev := event.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyESC:
					quit()

				case tcell.KeyLeft:
					ux -= 0.2 / uz

				case tcell.KeyRight:
					ux += 0.2 / uz

				case tcell.KeyUp:
					uy -= 0.2 / uz

				case tcell.KeyDown:
					uy += 0.2 / uz

				case tcell.KeyUpLeft:
					ux -= 0.2 / uz
					uy -= 0.2 / uz

				case tcell.KeyUpRight:
					ux += 0.2 / uz
					uy -= 0.2 / uz

				case tcell.KeyDownLeft:
					ux -= 0.2 / uz
					uy += 0.2 / uz

				case tcell.KeyDownRight:
					ux += 0.2 / uz
					uy += 0.2 / uz

				case tcell.KeyRune:
					switch ev.Rune() {
					case 'Z', 'z':
						uz *= 1.05
					case 'X', 'x':
						uz /= 1.05
					case 'C', 'c':
						crossHair = !crossHair
					case 'Q', 'q':
						go func() {
							ux = 0
							uy = 0
							uz = 1
							pass = 0
						}()
					case 'R', 'r':
						if !hqRender {
							hqRender = true
							pass = maxSqPasses
						}
					case 'Y', 'y':
						go func() {
							usePower = true
							power = 0.05 + 0i
							// hq_render = true
							for cmplx.Abs(power) < 5.0 {
								power += 0.05
								pass = 0
								time.Sleep(time.Millisecond * 200)
							}
							time.Sleep(time.Millisecond * 500)
							power = 2
							usePower = false
							pass = 0
						}()

					case 'P', 'p':
						go func(_cx float64, _cy float64, _cz float64, _width float64, _height float64) {
							var pngWg sync.WaitGroup
							printImage := image.NewNRGBA(image.Rect(0, 0, int(_width), int(_height)))
							for ryo := 0; ryo < int(_height); ryo++ {
								pngWg.Add(1)
								go func(_y int, _pass int, _iteration int, _ux float64, _uy float64, _uz float64) {
									defer pngWg.Done()
									for rxo := 0; rxo < int(_width); rxo++ {
										z := getAtPoint(float64(rxo), float64(_y), _ux, _uy, _uz, _width, _height, _iteration, false, false)
										fz := float64(z) / float64(_iteration)
										z = int(fz * 255) //int(fz*0xFFFFFF)
										printImage.SetNRGBA(rxo, _y, color.NRGBA{R: uint8(z), G: uint8(z), B: uint8(z), A: 255})
									}
								}(ryo, 1, pngIterations, _cx, _cy, _cz)
							}
							pngWg.Wait()
							f, err := os.Create(pngName)
							if err != nil {
								log.Fatal(err)
							}
							err = png.Encode(f, printImage)
							err2 := f.Close()
							if err != nil {
								log.Fatal(err)
							}
							if err2 != nil {
								log.Fatal(err2)
							}
						}(ux, uy, uz, pngWidth, pngHeight)
					}
				}
			}
		}
	}(s)

	uz = 1.0
	for {
		if pass < maxSqPasses || hqRender {
			currentIteration := updatingIterations[int(math.Min(float64(maxSqPasses), float64(pass)))]
			if hqRender {
				currentIteration = updatingIterations[maxSqPasses]
			}

			for y := 0; y < int(height); y += passFx[pass] {
				renderWg.Add(1)
				go func(_y int, _pass int, _currentIteration int, _ux float64, _uy float64, _uz float64, _hq bool, _up bool) {
					//fmt.Printf("%d: on duty!", _y)
					defer renderWg.Done()
					for _x := 0; _x < int(width); _x += passFx[_pass] {

						iteration := getAtPoint(float64(_x), float64(_y), _ux, _uy, _uz, width, height, _currentIteration, _up, true)

						hqIter := 0
						if _hq {
							hqy := float64(_y) + float64(passFx[_pass])/2.0

							hqIter = getAtPoint(float64(_x), hqy, _ux, _uy, _uz, width, height, _currentIteration, _up, true)

							// Characters that slap
							// ▄ ░ ▒ ▓
							s.SetContent(_x, _y, '▄', nil, tcell.StyleDefault.Background(tcell.PaletteColor(iteration%256)).Foreground(tcell.PaletteColor(hqIter%256)))
						} else {
							// if !single_rune {
							// 	ending_rune = rune(iteration+'0') % 127
							// 	if ending_rune < '0' {
							// 		ending_rune += '0'
							// 	}
							// } else {
							// 	ending_rune = ' '
							// }

							endingRune := ' '
							// ┃╋━

							if crossHair {
								if _x == int(width/2) {
									endingRune = '┃'
								}

								if _y == int(height/2) {
									if endingRune == '┃' {
										endingRune = '╋'
									} else {
										endingRune = '━'
									}
								}
							}
							s.SetContent(_x, _y, endingRune, nil, tcell.StyleDefault.Background(tcell.PaletteColor(iteration%256)).Foreground(tcell.ColorWhite))
						}

						// if x+_y == 0 {
						// ending_rune = '0' + rune(iteration)

						// s.SetContent(_x, _y, ending_rune, nil, tcell.StyleDefault.Background(tcell.PaletteColor(iteration)).Foreground(tcell.ColorWhite))
						// }
					}
				}(y, pass, currentIteration, ux, uy, uz, hqRender, usePower)
			}
			pass++
			renderWg.Wait()
			hqRender = false
			s.Show()
		}
	}
}

func getAtPoint(x float64, y float64, ux float64, uy float64, uz float64, _width float64, _height float64, maxIterations int, usePower bool, maxOverwrite bool) (iteration int) {
	y0 := -1.2/uz + (2.47*(y/_height))/uz + uy
	x0 := -2.0/uz + (4.00*(x/_width))/uz + ux

	var point = complex(x0, y0)

	z := 0 + 0i

	checkWith := cmplx.Abs(boundCheck)
	for cmplx.Abs(z) <= checkWith && iteration < maxIterations {
		if !usePower {
			z = z*z + point
		} else {
			z = cmplx.Pow(z, power) + point
		}
		iteration++
	}

	if iteration >= maxIterations && maxOverwrite {
		iteration = 0
	}

	return iteration
}

// func vanilla() {
// 	screen := make([]string, height)

// 	for y := range screen {
// 		line := ""
// 		for x := 0; x < width; x++ {
// 			y0 := -1.2 + (2.47 * (float64(y) / height))
// 			x0 := -2.0 + (4 * (float64(x) / width))
// 			iteration := 0

// 			mx := 0.0
// 			my := 0.0

// 			x2 := mx * mx
// 			y2 := my * my

// 			for x2+y2 <= 4.0 && iteration < max_iteration {
// 				xt := x2 - y2 + x0
// 				my = 2*mx*my + y0
// 				mx = xt

// 				x2 = mx * mx
// 				y2 = my * my
// 				iteration++
// 			}

// 			ending_rune := rune(iteration + '0')
// 			ending_rune %= 127
// 			if ending_rune < '0' {
// 				ending_rune += '0'
// 			}

// 			line += string(ending_rune)
// 		}
// 		screen[y] = line
// 	}

// 	for _, line := range screen {
// 		fmt.Println(line)
// 	}
// }
