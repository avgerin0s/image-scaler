package main

import (
  "github.com/rakyll/magicmime"
  "github.com/disintegration/imaging"
  "github.com/gorilla/mux"
  "runtime"
  "fmt"
  "os"
  "io"
  "log"
  "strings"
  "strconv"
  "regexp"
  "net/http"
  "crypto/md5"
  "encoding/hex"
  "encoding/json"
)

var chttp = http.NewServeMux()

type Image struct {
  Url string
}

func Resize(w http.ResponseWriter, req *http.Request) {
  if (strings.Contains(req.URL.Path, "static/")) {
    chttp.ServeHTTP(w, req)
    return
  }

  url := req.URL.Query().Get("url")
  param_width := req.URL.Query().Get("width")
  param_height := req.URL.Query().Get("height")

  log.Println("Downloading url: " + url)

  var width int
  var height int

  if url == "" {
    panic("Url is empty")
  }

  if param_width == "" {
    width = 1280
  } else {
    width, _ = strconv.Atoi(param_width)
  }

  if param_height == "" {
    height = 0
  } else {
    height, _ = strconv.Atoi(param_height)
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

  resizedImage := imaging.Resize(img, width, height, imaging.Lanczos)
  newFileName := "resized_" + originalFilename + "." + extension

  imaging.Save(resizedImage, newFileName)

  image := Image{"http://127.0.0.1:3000/static/" + newFileName}
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

  r := mux.NewRouter()

  r.HandleFunc("/resize", Resize)
  http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("."))))
  http.Handle("/", r)

  err := http.ListenAndServe(":3000", nil)
  if err != nil {
    log.Fatal("ListenAndServe:", err)
  }
}
