package main

import (
	"fmt"
	"golang.org/x/image/draw"
	"image"
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

	fmt.Printf("Uploaded file: %+v\n", header.Filename)
	fmt.Printf("Size: %+v\n", header.Size)
	fmt.Printf("MIME header: %+v\n", header.Header)

	//check, whether file exists or not
	OpenFile, err := header.Open()
	defer OpenFile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		fmt.Println(err)
		http.Error(w, "File not found", 404)
		return
	}

	img, _, err := image.Decode(OpenFile)
	if err != nil {
		log.Fatal(err)
	}

	FileName := header.Filename

	Output := resizeImage(img, FileName)

	FileHeader := make([]byte, 512)

	FileContentType := http.DetectContentType(FileHeader)

	FileStat, _ := Output.Stat() //Get info from file

	FileSize := strconv.FormatInt(FileStat.Size(), 10)

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
	src := img
	// new size of image
	dr := image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2)
	// resize using given scaler
	var res image.Image
	// perform resizing
	res = scaleTo(src, dr, draw.NearestNeighbor)
	// open file to save
	dstFile, err := os.Create(FileName)
	if err != nil {
		log.Fatal(err)
	}
	//get file exteinsion from file name
	ext := filepath.Ext(FileName)

	jpegOptions := jpeg.Options{Quality: 100}

	gifOptions := gif.Options{NumColors: 256, Quantizer: nil, Drawer: nil}
	// encode as jpeg/png/gif to the file
	switch ext {

	case ".jpeg":
		err = jpeg.Encode(dstFile, res, &jpegOptions)
	case ".jpg":
		err = jpeg.Encode(dstFile, res, &jpegOptions)
	case ".png":
		err = png.Encode(dstFile, res)
	case ".gif":
		err = gif.Encode(dstFile, res, &gifOptions)
	default:
		fmt.Println("unsupported format:", ext, "Only jpeg, png ang gif are supported")
	}
	if err != nil {
		fmt.Println(err)
	}
	return dstFile
}

// for RGBA images
// src   - source image
// rect  - size we want
// scale - scaler
func scaleTo(src image.Image,
	rect image.Rectangle, scale draw.Scaler) image.Image {
	dst := image.NewRGBA(rect)
	scale.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("Upload your image to resize")
	setupRoutes()
}
