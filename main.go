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
const WALLPAPER_FILEPATH_TEMPLATE = "{{TEMPLATE_WALLPAPER_FILEPATH}}"

func main() {
	fmt.Println("hello, image-colors!")

	// dir, err := os.ReadDir(WALLPAPER_DIR)
	// if err != nil {
	// 	log.Printf("Error reading image directory: %s", err)
	// 	os.Exit(1)
	// }

	// randInd := rand.Intn(len(dir))

	// fileName := fmt.Sprintf("%s", WALLPAPER_DIR+dir[randInd].Name())

	fileName, err := getRandomWallpaperPath()
	if err != nil {
		log.Printf("[FAILED]: %s", err)
		os.Exit(1)
	}

	m, err := decodeWallpaperForColorAnalysis(fileName)
	if err != nil {
		log.Printf("[FAILED]: %s", err)
		os.Exit(1)
	}

	colorMap := getColorMap(m)

	sortedColors := sortColorMap(colorMap)

	// reader, err := os.Open(fileName)
	// if err != nil {
	// 	log.Printf("Error reading image file: %s", err)
	// 	os.Exit(1)
	// }

	// m, _, err := image.Decode(reader)
	// if err != nil {
	// 	log.Printf("Error decoding image file: %s", err)
	// 	os.Exit(1)
	// }

	// bounds := m.Bounds()

	// colorMap := map[color.Color]int{}

	// for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
	// 	for x := bounds.Min.X; x < bounds.Max.X; x++ {
	// 		colorMap[m.At(x, y)] += 1
	// 	}
	// }

	// colors := make([]color.Color, 0, len(colorMap))
	// for color := range colorMap {
	// 	colors = append(colors, color)
	// }
	// sort.Slice(colors, func(i, j int) bool { return colorMap[colors[i]] > colorMap[colors[j]] })

	outColors := []color.Color{}

	for rank, col := range sortedColors {
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

	genConfig := readConfigTemplate(colorStrings, fileName)

	f3.WriteString(genConfig)
}

func readConfigTemplate(colors []string, wallPaperPath string) string {
	data, err := os.Open(CONFIG_FILE_TEMPLATE)
	if err != nil {
		log.Printf("Error reading config file template: %s", err)
	}

	repCount := 0
	scanner := bufio.NewScanner(data)

	b := strings.Builder{}

	for scanner.Scan() {

		if scanner.Text() == REPLACE_TEMPLATE {
			b.WriteString(colors[repCount])
			b.WriteString("\n")
			repCount += 1
		} else if scanner.Text() == WALLPAPER_FILEPATH_TEMPLATE {
			b.WriteString(fmt.Sprintf("$wallpaper_path = %s", wallPaperPath))
			b.WriteString("\n")
		} else {
			b.WriteString(scanner.Text())
			b.WriteString("\n")
		}

	}

	return b.String()
}

func getRandomWallpaperPath() (string, error) {
	//TODO:
	//function should error (and re-run automatically) if we select an incompatible wallpaper
	//only jpg, jpeg, and png are supported at this time

	dir, err := os.ReadDir(WALLPAPER_DIR)
	if err != nil {
		log.Printf("Error reading image directory: %s", err)
		return "", err
	}

	randInd := rand.Intn(len(dir))

	fileName := fmt.Sprintf("%s", WALLPAPER_DIR+dir[randInd].Name())

	return fileName, nil
}

func decodeWallpaperForColorAnalysis(wallPaperPath string) (image.Image, error) {
	reader, err := os.Open(wallPaperPath)
	if err != nil {
		log.Printf("Error reading image file: %s", err)
		return nil, err
	}

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Printf("Error decoding wallpaper image file: %s", err)
		return nil, err
	}

	return m, nil
}

func getColorMap(m image.Image) map[color.Color]int {
	colorMap := map[color.Color]int{}

	bounds := m.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			colorMap[m.At(x, y)] += 1
		}
	}

	return colorMap
}

func sortColorMap(cm map[color.Color]int) []color.Color {
	colors := make([]color.Color, 0, len(cm))
	for color := range cm {
		colors = append(colors, color)
	}
	sort.Slice(colors, func(i, j int) bool { return cm[colors[i]] > cm[colors[j]] })

	return colors
}

func getTopXColors(allColors []color.Color, count int) []color.Color {
	topColors := []color.Color{}

	for i := 0; i < count; i++ {
		topColors = append(topColors, allColors[i])
	}

	return topColors
}
