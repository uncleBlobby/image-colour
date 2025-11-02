package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
)

const WALLPAPER_DIR = "/home/dustin/Downloads/aesthetic-wallpapers/images/"
const CONFIG_FILE_TEMPLATE = "/home/dustin/.config/wallpaper/hyprland.conf.template"
const CONFIG_FILE_GENERATED = "/home/dustin/.config/wallpaper/hyprland.conf.gen"
const REPLACE_TEMPLATE = "{{TEMPLATE_COLOUR}}"

func main() {
	fmt.Println("hello, image-colors!")

	dir, err := os.ReadDir(WALLPAPER_DIR)
	if err != nil {
		log.Printf("Error reading image directory: %s", err)
		os.Exit(1)
	}

	randInd := rand.Intn(len(dir))

	fileName := fmt.Sprintf("%s", WALLPAPER_DIR+dir[randInd].Name())

	// for _, f := range dir {
	// 	fmt.Println(f.Info())
	// }

	reader, err := os.Open(fileName)
	if err != nil {
		log.Printf("Error reading image file: %s", err)
		os.Exit(1)
	}

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Printf("Error decoding image file: %s", err)
		os.Exit(1)
	}

	bounds := m.Bounds()

	colorMap := map[color.Color]int{}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			colorMap[m.At(x, y)] += 1
			// fmt.Println(m.At(x, y))
		}
	}

	// for k, c := range colorMap {
	// 	fmt.Printf("%v - %d\n", k, c)
	// }

	colors := make([]color.Color, 0, len(colorMap))
	for color := range colorMap {
		colors = append(colors, color)
	}
	sort.Slice(colors, func(i, j int) bool { return colorMap[colors[i]] > colorMap[colors[j]] })

	outColors := []color.Color{}

	for rank, col := range colors {
		if rank < 1000 {
			fmt.Printf("%v - %d\n", col, colorMap[col])
			outColors = append(outColors, col)
		}
	}

	width := 200
	height := 1000

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if y < 200 {
				img.Set(x, y, outColors[0])
			}
			if y >= 200 && y < 400 {
				img.Set(x, y, outColors[1])
			}
			if y >= 400 && y < 600 {
				img.Set(x, y, outColors[2])
			}
			if y >= 600 && y < 800 {
				img.Set(x, y, outColors[3])
			}
			if y >= 800 && y < 1000 {
				img.Set(x, y, outColors[4])
			}

			// if (x + y) < len(outColors) {
			// 	img.Set(x, y, outColors[x+y])
			// }
		}
	}

	f2, err := os.Create("/home/dustin/.config/wallpaper/cols.wp")
	if err != nil {
		log.Printf("Error writing cols.wp: %s", err)
	}

	f3, err := os.Create(CONFIG_FILE_GENERATED)
	if err != nil {
		log.Printf("Error generating config file: %s", err)
	}

	colorStrings := []string{}

	for ind, c := range outColors {
		if ind < 5 {
			fmt.Println(c)
			// fmt.Printf("%f\n", c)
			v1, v2, v3, v4 := c.RGBA()
			s1 := hex.EncodeToString([]byte{byte(v1)})

			s2 := hex.EncodeToString([]byte{byte(v2)})
			s3 := hex.EncodeToString([]byte{byte(v3)})
			s4 := hex.EncodeToString([]byte{byte(v4)})

			hexString := fmt.Sprintf("%s%s%s%s", s1, s2, s3, s4)
			fmt.Printf("%s%s%s%s\n", s1, s2, s3, s4)

			colString := fmt.Sprintf("$color%d = rgba(%s)\n", ind, hexString)
			colorStrings = append(colorStrings, colString)
			f2.WriteString(colString)
			// f2.WriteString(hexString)
			// f2.WriteString("\n")
		}

	}

	f, _ := os.Create("outColors.png")
	png.Encode(f, img)

	genConfig := readConfigTemplate(colorStrings)

	f3.WriteString(genConfig)
}

func readConfigTemplate(colors []string) string {
	data, err := os.Open(CONFIG_FILE_TEMPLATE)
	if err != nil {
		log.Printf("Error reading config file template: %s", err)
	}

	repCount := 0
	scanner := bufio.NewScanner(data)

	b := strings.Builder{}

	for scanner.Scan() {
		if scanner.Text() != REPLACE_TEMPLATE {
			fmt.Println(scanner.Text())
			b.WriteString(scanner.Text())
			b.WriteString("\n")
		} else {
			fmt.Println(colors[repCount])
			b.WriteString(colors[repCount])
			b.WriteString("\n")
			repCount += 1
		}
	}

	return b.String()
}
