package main

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// parse input, type multipart/form-data
	r.ParseMultipartForm(10 << 20)
	// retrive file from posted form-data
	file, header, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("couldn't retrieve file ", err)
		return
	}
	defer file.Close()

	//check, whether file exists or not
	OpenFile, err := header.Open()
	defer OpenFile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		fmt.Println(err)
		http.Error(w, "File not found", 404)
		return
	}
	FileName := header.Filename
	ext := filepath.Ext(FileName)
	var Output *os.File
	if ext == ".gif" {
		gifImage, err := gif.DecodeAll(OpenFile)
		if err != nil {
			fmt.Println(err)
		}
		Output = resizeGif(gifImage, FileName)
	} else {
		img, _, err := image.Decode(OpenFile)
		if err != nil {
			fmt.Println(err)
		}

		Output = resizeImage(img, FileName)
	}

	FileContentType, FileSize := createHeaders(Output)
	defer Output.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+FileName)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Output.Seek(0, 0)
	_, err = io.Copy(w, Output)
	if err != nil {
		fmt.Println(err)
	} //'Copy' the file to the client
	return
}

func resizeImage(img image.Image, FileName string) *os.File {
	res := resize.Resize(1000, 0, img, resize.Lanczos3)
	// open file to save
	out, err := os.Create(FileName)
	if err != nil {
		log.Fatal(err)
	}
	//get file exteinsion from file name
	ext := filepath.Ext(FileName)
	// encode as jpeg/png/gif to the file
	jpegOptions := jpeg.Options{Quality: 100}
	switch ext {
	case ".jpeg":
		err = jpeg.Encode(out, res, &jpegOptions)
	case ".jpg":
		err = jpeg.Encode(out, res, &jpegOptions)
	case ".png":
		err = png.Encode(out, res)
	default:
		fmt.Println("unsupported format:", ext, "Only jpeg, png ang gif are supported")
	}
	if err != nil {
		fmt.Println(err)
	}
	return out
}

func resizeGif(gifImage *gif.GIF, FileName string) *os.File {
	for index, frame := range gifImage.Image {

		rect := frame.Bounds()

		tmpImage := frame.SubImage(rect)

		resizedImage := resize.Resize(1000, 0, tmpImage, resize.Lanczos3)

		var tmpPalette color.Palette
		for x := 1; x <= rect.Dx(); x++ {

			for y := 1; y <= rect.Dy(); y++ {

				if !contains(tmpPalette, gifImage.Image[index].At(x, y)) {

					tmpPalette = append(tmpPalette, gifImage.Image[index].At(x, y))

				}
			}
		}

		resizedBounds := resizedImage.Bounds()

		resizedPalette := image.NewPaletted(resizedBounds, tmpPalette)

		draw.Draw(resizedPalette, resizedBounds, resizedImage, image.ZP, draw.Src)

		gifImage.Image[index] = resizedPalette
	}
	gifImage.Config.Width = 1000

	gifImage.Config.Height = 1000
	out, err := os.Create(FileName)
	if err != nil {
		log.Fatal(err)
	}
	err = gif.EncodeAll(out, gifImage)
	return out
}

func contains(colorPalette color.Palette, c color.Color) bool {

	for _, tmpColor := range colorPalette {

		if tmpColor == c {

			return true
		}
	}
	return false
}

func createHeaders(Output *os.File) (string, string) {
	FileHeader := make([]byte, 512)

	contentType := http.DetectContentType(FileHeader)

	fileStat, _ := Output.Stat() //Get info from file

	size := strconv.FormatInt(fileStat.Size(), 10)

	return contentType, size

}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("Upload your image to resize")
	setupRoutes()
}
