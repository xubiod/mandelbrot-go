package main

import (
	"log"
	"math"
	"math/cmplx"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
)

const width float64 = 180.0
const height float64 = 60.0

// const max_iteration = 200

var updating_iterations = []int{25, 50, 100, 150, 200, 300, 500}

var pass_fx = []int{3, 2, 2, 1, 1, 1, 1}

const max_sq_passes = 6

var crosshair = true
var hq_render = false

// const single_rune = true

var ux, uy, uz float64

// const frames = 600

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	quit := func() {
		// time.Sleep(time.Second * 3)
		s.Fini()
		os.Exit(0)
	}

	// var _x, y, iteration int
	// var x0, y0, mx, my, x2, y2, xt float64
	var pass int = 0
	// var current_iteration int

	var render_wg sync.WaitGroup

	// done := make(chan int, int(height))
	// fmt.Printf("!!%d!!!! ! ? ", cap(done))

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
						crosshair = !crosshair
					case 'Q', 'q':
						go func() {
							ux = 0
							uy = 0
							uz = 1
							pass = 0
						}()
					case 'R', 'r':
						if !hq_render {
							hq_render = true
							pass = max_sq_passes
						}
					}
				}
			}
		}
	}(s)

	uz = 1.0
	for /*f := 0; f < frames; f++*/ {
		if pass < max_sq_passes || hq_render {
			current_iteration := updating_iterations[int(math.Min(float64(max_sq_passes), float64(pass)))]
			if hq_render {
				current_iteration = updating_iterations[max_sq_passes]
			}

			for y := 0; y < int(height); y += pass_fx[pass] {
				render_wg.Add(1)
				go func(_y int, _pass int, _current_iteration int, _ux float64, _uy float64, _uz float64, _hq bool) {
					//fmt.Printf("%d: on duty!", _y)
					defer render_wg.Done()
					for _x := 0; _x < int(width); _x += pass_fx[_pass] {

						iteration := getAtPoint(float64(_x), float64(_y), _ux, _uy, _uz, _current_iteration)

						hq_iter := 0
						if _hq {
							hqy := float64(_y) + float64(pass_fx[_pass])/2.0

							hq_iter = getAtPoint(float64(_x), hqy, _ux, _uy, _uz, _current_iteration)

							// Characters that slap
							// ▄ ░ ▒ ▓
							s.SetContent(_x, _y, '▄', nil, tcell.StyleDefault.Background(tcell.PaletteColor(iteration)).Foreground(tcell.PaletteColor(hq_iter)))
						} else {
							// if !single_rune {
							// 	ending_rune = rune(iteration+'0') % 127
							// 	if ending_rune < '0' {
							// 		ending_rune += '0'
							// 	}
							// } else {
							// 	ending_rune = ' '
							// }

							var ending_rune rune = ' '
							// ┃╋━

							if crosshair {
								if _x == int(width/2) {
									ending_rune = '┃'
								}

								if _y == int(height/2) {
									if ending_rune == '┃' {
										ending_rune = '╋'
									} else {
										ending_rune = '━'
									}
								}
							}
							s.SetContent(_x, _y, ending_rune, nil, tcell.StyleDefault.Background(tcell.PaletteColor(iteration)).Foreground(tcell.ColorWhite))
						}

						// if x+_y == 0 {
						// ending_rune = '0' + rune(iteration)

						// s.SetContent(_x, _y, ending_rune, nil, tcell.StyleDefault.Background(tcell.PaletteColor(iteration)).Foreground(tcell.ColorWhite))
						// }
					}
					//fmt.Printf("%d: i'm done! ", _y)
					//done <- _y
				}(y, pass, current_iteration, ux, uy, uz, hq_render)
			}
			pass++
			render_wg.Wait()
			hq_render = false
			// amount := cap(done)
			// for q := 0; q < int(height); q++ {
			// 	fmt.Printf("main: y %d is done, waiting for %d more, even with %d around! ", <-done, amount-1, len(done))
			// 	amount--
			// }
			s.Show()
		}
		//i *= 1.05
	}
}

const boundCheck complex128 = 2 + 0i

func getAtPoint(x float64, y float64, ux float64, uy float64, uz float64, max_iterations int) int {
	y0 := -1.2/uz + (2.47*(y/height))/uz + uy
	x0 := -2.0/uz + (4.00*(x/width))/uz + ux

	var point complex128 = complex(x0, y0)
	iteration := 0

	// mx := 0.0
	// my := 0.0

	// x2 := mx * mx
	// y2 := my * my
	z := 0 + 0i

	// var xt float64

	for cmplx.Abs(z) <= cmplx.Abs(boundCheck) && iteration < max_iterations {
		// xt = real(cn2) - imag(cn2) + real(cn)
		// my = 2*cn + complex(0, imag(cn))
		// mx = complex(xt, 0)
		// x2 = mx * mx
		// y2 = my * my
		z = z*z + point
		iteration++
	}

	if iteration >= max_iterations {
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
