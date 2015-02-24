package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/eavgerinos/image-scaler/processing"
	"github.com/gorilla/mux"
	"github.com/rakyll/magicmime"
)

// Image represents the JSON response of the api.
type Image struct {
	URL string
}

var port int
var directory string

// Handle /resize http request
func Resize(w http.ResponseWriter, req *http.Request) {
	url := req.URL.Query().Get("url")
	paramWidth := req.URL.Query().Get("width")
	paramHeight := req.URL.Query().Get("height")

	log.Println("Downloading url: " + url)

	if url == "" {
		panic("Url is empty")
	}

	var width int
	if paramWidth == "" {
		width = 1280
	} else {
		width, _ = strconv.Atoi(paramWidth)
	}

	var height int
	if paramHeight == "" {
		height = 0
	} else {
		height, _ = strconv.Atoi(paramHeight)
	}

	mm, err := magicmime.New(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_ERROR)
	if err != nil {
		panic(err)
	}

	// Download the url
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Create an md5 checksum for filename
	hasher := md5.New()
	hasher.Write([]byte(url))
	originalFilename := hex.EncodeToString(hasher.Sum(nil))
	originalFile, err := os.Create(originalFilename + ".jpeg")
	io.Copy(originalFile, resp.Body)

	// Determine content-type of downloaded file
	mimetype, err := mm.TypeByFile(originalFilename + ".jpeg")
	isImage, _ := regexp.MatchString("^image/", mimetype)
	extension := strings.Split(mimetype, "/")[1]
	defer originalFile.Close()
	if !isImage {
		fmt.Println("Not an image file")
		return
	}

	// Process the image
	img, err := imaging.Open(originalFilename + ".jpeg")
	if err != nil {
		panic(err)
	}

	newFileName := processing.Resize(img, originalFilename, extension, width, height)

	log.Println(newFileName)

	image := Image{"http://127.0.0.1:" + strconv.Itoa(port) + "/static/" + newFileName}
	js, err := json.Marshal(image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	// Use all available CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.IntVar(&port, "port", 3000, "the port on which the service will listen on")
	flag.StringVar(&directory, "d", ".", "the directory on which the processed images will be stored")
	flag.Parse()

	stringPort := strconv.Itoa(port)

	r := mux.NewRouter()

	r.HandleFunc("/resize", Resize)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(directory))))
	http.Handle("/", r)

	log.Println("Starting server on " + stringPort)

	err := http.ListenAndServe(":"+stringPort, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	log.Println("Ready")
}
