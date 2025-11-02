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

	top5Colours := getTopXColors(sortedColors, 5)

	img := createPalettePNG(top5Colours)

	err = writePalettePNGFile(img)
	if err != nil {
		log.Printf("Error writing palette PNG file: %s", err)
	}

	// f3, err := os.Create(CONFIG_FILE_GENERATED)
	// if err != nil {
	// 	log.Printf("Error generating config file: %s", err)
	// }

	colorStrings := getColorStringsForConfigTemplate(top5Colours)

	genConfig := readConfigTemplate(colorStrings, fileName)

	// f3.WriteString(genConfig)

	err = writeGeneratedConfigFile(genConfig)
	if err != nil {
		log.Printf("[ERROR]: %s", err)
		os.Exit(1)
	}

}

func writeGeneratedConfigFile(cfgContent string) error {
	f3, err := os.Create(CONFIG_FILE_GENERATED)
	if err != nil {
		log.Printf("Error generating config file: %s", err)
		return err
	}

	_, err = f3.WriteString(cfgContent)
	if err != nil {
		log.Printf("Error saving config file: %s", err)
		return err
	}
	log.Printf("[INFO]: Generated config file: %s", f3.Name())
	return nil
}

func readConfigTemplate(colors []string, wallPaperPath string) string {
	log.Printf("[INFO]: Generating config file from template...")
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

	log.Printf("[INFO]: Reading wallpaper directory: %s", WALLPAPER_DIR)

	randInd := rand.Intn(len(dir))

	fileName := fmt.Sprintf("%s", WALLPAPER_DIR+dir[randInd].Name())

	log.Printf("[INFO]: Selected random wallpaper: %s", fileName)
	return fileName, nil
}

func decodeWallpaperForColorAnalysis(wallPaperPath string) (image.Image, error) {
	log.Printf("[INFO]: Decoding wallpaper for color analysis...")
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
	log.Printf("[INFO]: Preparing color map...")
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

func createPalettePNG(top5Colors []color.Color) *image.RGBA {
	log.Printf("[INFO]: Creating palette swatch file...")
	width := 200
	height := 1000

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if y < 200 {
				img.Set(x, y, top5Colors[0])
			}
			if y >= 200 && y < 400 {
				img.Set(x, y, top5Colors[1])
			}
			if y >= 400 && y < 600 {
				img.Set(x, y, top5Colors[2])
			}
			if y >= 600 && y < 800 {
				img.Set(x, y, top5Colors[3])
			}
			if y >= 800 && y < 1000 {
				img.Set(x, y, top5Colors[4])
			}

		}
	}

	return img
}

func writePalettePNGFile(data *image.RGBA) error {
	//TODO:
	// accept string parameter for wallpaper filename and save palette with a name that makes sense
	// which directory to write palette info? is it even useful to save the swatch?
	// command line option maybe

	f, err := os.Create("outColors.png")
	if err != nil {
		log.Printf("Error creating palette PNG file: %s", err)
		return err
	}

	err = png.Encode(f, data)
	if err != nil {
		log.Printf("Error encoding palette PNG file: %s", err)
	}
	log.Printf("[INFO]: Encoding palette PNG file...")
	return nil
}

func getColorStringsForConfigTemplate(colors []color.Color) []string {
	colorStrings := []string{}
	for ind, c := range colors {
		// NOTE: maximum 5 color selection (arbitrary limit?)
		if ind < 5 {
			v1, v2, v3, v4 := c.RGBA()

			s1 := hex.EncodeToString([]byte{byte(v1)})
			s2 := hex.EncodeToString([]byte{byte(v2)})
			s3 := hex.EncodeToString([]byte{byte(v3)})
			s4 := hex.EncodeToString([]byte{byte(v4)})

			hexString := fmt.Sprintf("%s%s%s%s", s1, s2, s3, s4)
			// fmt.Printf("%s%s%s%s\n", s1, s2, s3, s4)

			colString := fmt.Sprintf("$color%d = rgba(%s)\n", ind, hexString)
			colorStrings = append(colorStrings, colString)
		}
	}
	log.Printf("[INFO]: Formatting hex colors for config generator...")
	return colorStrings
}
