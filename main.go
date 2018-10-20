package main
import (
	"image"
	//"image/png"
	"image/color"
	"image/jpeg"
	"os"
	"image/draw"
	"github.com/utahta/go-openuri"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	//"fmt"
	"io/ioutil"
	"time"	
    "strconv"
    "github.com/bitly/go-simplejson"
    //"github.com/pixiv/go-libjpeg/jpeg"
    
	) 
   
// Create a struct to deal with pixel
type Pixel struct {
    Point image.Point
    Color color.Color
}

type Images struct {
	Userid string   `json:"userid"`
	Images []string `json:"images"`
}




// Keep it DRY so don't have to repeat opening file and decode
func OpenAndDecode(filepath string) (image.Image, string, error) {
	//imgFile, err := os.Open(filepath)
	imgFile, err := openuri.Open(filepath) 
    if err != nil {
        panic(err)
    }
    defer imgFile.Close()
    img, format, err := image.Decode(imgFile)
    if err != nil {
        panic(err)
    }
    return img, format, nil
}

// Decode image.Image's pixel data into []*Pixel
func DecodePixelsFromImage(img image.Image, offsetX, offsetY int) []*Pixel {
    pixels := []*Pixel{}
    for y := 0; y <= img.Bounds().Max.Y; y++ {
        for x := 0; x <= img.Bounds().Max.X; x++ {
            p := &Pixel{
                Point: image.Point{x + offsetX, y + offsetY},
                Color: img.At(x, y),
            }
            pixels = append(pixels, p)
        }
    }
    return pixels
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/mergeImages", mergImages).Methods("POST")
	
	log.Println(time.Now().Unix())
	log.Println("Server Starting at port 9090")
	log.Fatal(http.ListenAndServe(":9090", router))
	
    
}

func mergImages(w http.ResponseWriter, r *http.Request) {

	
	var images Images
    

	body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        panic(err)
    }
	
	err = json.Unmarshal(body, &images)
	if err != nil {
        panic(err)
    }
    
	//_ = json.NewDecoder(r.Body).Decode(&images)
	//fmt.Println(images)
	
	
	
	println(string("Downloading Image 1"))
	img1, _, err := OpenAndDecode(images.Images[0])
    if err != nil {
        panic(err)
	}
	println(string("Downloaded Image 1"))
    img2, _, err := OpenAndDecode(images.Images[1])
    if err != nil {
        panic(err)
	}
	println(string("Downloaded Image 2"))
    // collect pixel data from each image
    pixels1 := DecodePixelsFromImage(img1, 0, 0)
    // the second image has a Y-offset of img1's max Y (appended at bottom)
    pixels2 := DecodePixelsFromImage(img2, 0, img1.Bounds().Max.Y)
    pixelSum := append(pixels1, pixels2...)

    // Set a new size for the new image equal to the max width
    // of bigger image and max height of two images combined
    newRect := image.Rectangle{
        Min: img1.Bounds().Min,
        Max: image.Point{
            X: img2.Bounds().Max.X,
            Y: img2.Bounds().Max.Y + img1.Bounds().Max.Y,
        },
    }
    finImage := image.NewRGBA(newRect)
    // This is the cool part, all you have to do is loop through
    // each Pixel and set the image's color on the go
    for _, px := range pixelSum {
            finImage.Set(
                px.Point.X,
                px.Point.Y,
                px.Color,
            )
    }
    draw.Draw(finImage, finImage.Bounds(), finImage, image.Point{0, 0}, draw.Src)
	
    // Create a new file and write to it
    var image_file_ame = images.Userid+"_"+strconv.FormatInt(time.Now().Unix(),10)+"_img.jpg"
    out, err := os.Create("/var/www/html/merged_files/"+image_file_ame)
    if err != nil {
        panic(err)
        os.Exit(1)
    }



    err = jpeg.Encode(out, finImage, &jpeg.Options{Quality: 60})
    if err != nil {
        panic(err)
        os.Exit(1)
	}
    json := simplejson.New()
    json.Set("image_url", "http://www.swan-speed.com/merged_files/"+image_file_ame)
    
    payload, err := json.MarshalJSON()
	if err != nil {
		log.Println(err)
	}

    w.Header().Set("Content-Type", "application/json")
	w.Write(payload)

    //return "http://www.swan-speed.com/merged_files/"+image_file_ame , nil



}
