package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/eyko139/photo-stream/internal/models"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// hello is a component that displays a simple "Hello World!". A component is a
// customizable, independent, and reusable UI element. It is created by
// embedding app.Compo into a struct.
type hello struct {
	app.Compo
    env *Env
	ThumbNails        []ThumbNail
	ShouldReRender    bool
	isUpdateAvailable bool
	Gallery           *Gallery
	Text              string
	DownloadToken     string
	Images            []string
	ImagesAvailable   bool
}

type ThumbNail struct {
	Name      string
	B64       string
	ClassName string
	UID       string
}

func NewHello(env *Env) *hello {
	return &hello{
		Text: "Hello World!",
        env: env,
	}
}

// The Render method is where the component appearance is defined. Here, a
// "Hello World!" is displayed as a heading.
func (h *hello) Render() app.UI {

	app.Logf("images: %s", h.Images)

	return app.Div().Body(
		app.Div().Class("section-header").Text("Select albums"),
		app.Div().Body(
			app.Range(h.ThumbNails).Slice(func(i int) app.UI {
				item := h.ThumbNails[i]
				return app.Div().Class("thumbnail-container").Class(item.ClassName).Body(
					app.Div().Class("thumbnail-header").Text(item.Name),
					app.Img().ID(item.UID).Src(fmt.Sprintf("data:image/jpg;base64, %s", item.B64)).OnClick(h.onClickAlbum),
				)
			}),
		),
		app.Button().Text("Play").OnClick(h.onClickPlay),
		app.If(h.ImagesAvailable, func() app.UI {
			return &SlideShow{
				CurrentImage: h.Images[0],
			}
		}),
	)
}

func (h *hello) onClickPlay(ctx app.Context, e app.Event) {
	var images []string
	for _, selectedThumbnails := range h.ThumbNails {
		if selectedThumbnails.ClassName == "active" {

			ctx.Async(func() {
				res, err := http.Get(fmt.Sprintf("/downloadAlbum?albumId=%s&downloadToken=%s", selectedThumbnails.UID, h.DownloadToken))
				if err != nil {
					app.Logf("failed to download album")
				}
				imageBytes, err := io.ReadAll(res.Body)
				if err != nil {
					app.Logf("Failed to decode: %s", err)
				}
				json.Unmarshal(imageBytes, &images)

				ctx.Dispatch(func(ctx app.Context) {
					h.Images = images
					h.ImagesAvailable = true
				})

			})
		}
	}

}
func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (h *hello) onClickAlbum(ctx app.Context, e app.Event) {
	albumClicked := ctx.JSSrc().Get("id").String()
	var update []ThumbNail
	for _, album := range h.ThumbNails {
		if album.UID == albumClicked {
			app.Logf("albumClicked: %s,  albumId: %s", albumClicked, album.UID)
			album.ClassName = "active"
		}
		update = append(update, album)
	}
	ctx.Dispatch(func(ctx app.Context) {
		h.ThumbNails = update
	})
}

func (h *hello) FetchAlbums(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", h.env.PrismURL + "/api/v1/albums?count=10", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-Auth-Token", h.env.PrismAuthToken)
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
	}

	dt := res.Header.Get("X-Download-Token")
	defer res.Body.Close()
	bytes, _ := io.ReadAll(res.Body)
	w.Header().Add("X-Download-Token", dt)
	w.Write(bytes)
}

func (h *hello) FetchThumbnails(w http.ResponseWriter, r *http.Request) {

	albumId := r.URL.Query().Get("albumId")
	downloadToken := r.URL.Query().Get("downloadToken")

	if albumId == "" || downloadToken == "" {
		app.Logf("missing token or ID")
		return
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	thumbReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/albums/%s/t/%s/tile_500", h.env.PrismURL, albumId, downloadToken), nil)
	thumbReq.Header.Set("X-Auth-Token", "jqGHgo-aXoKSV-muuhiF-Djh1Zu")

	app.Logf("req %+v", thumbReq)
	thumbRes, err := client.Do(thumbReq)
	if err != nil {
		app.Logf("Error fetching thumbnails %s", err)
	}

	defer thumbRes.Body.Close()
	thumbBytes, err := io.ReadAll(thumbRes.Body)

	os.WriteFile(fmt.Sprintf("web/thumbs/%s.jpg", albumId), thumbBytes, 0666)
	w.Write(thumbBytes)
}

func (h *hello) OnMount(ctx app.Context) {
	ctx.Async(func() {
		var albums []models.Album
		res, _ := http.Get("/fetchAlbums")
		bytes, _ := io.ReadAll(res.Body)
		err := json.Unmarshal(bytes, &albums)

		if err != nil {
			app.Logf("cant unmarshall albums")
		}

		app.Logf("response: %+v", albums)

		var thumbNailUpdate []ThumbNail

		for _, album := range albums {

			if album.Type == "album" {

				thumbNail := &ThumbNail{
					Name: album.Title,
					UID:  album.UID,
				}

				thumbRes, _ := http.Get(fmt.Sprintf("/fetchThumbnails?albumId=%s&downloadToken=%s", album.UID, res.Header.Get("X-Download-Token")))
				thumbBytes, _ := io.ReadAll(thumbRes.Body)
				base64String := base64.StdEncoding.EncodeToString(thumbBytes)
				thumbNail.B64 = base64String
				thumbNailUpdate = append(thumbNailUpdate, *thumbNail)
			}
		}

		ctx.Dispatch(func(ctx app.Context) {
			h.ThumbNails = thumbNailUpdate
			h.DownloadToken = res.Header.Get("X-Download-Token")
		})
	})
}

func (h *hello) DownloadAlbum(w http.ResponseWriter, r *http.Request) {

	albumId := r.URL.Query().Get("albumId")
	downloadToken := r.URL.Query().Get("downloadToken")

	if albumId == "" || downloadToken == "" {
		app.Logf("missing token or ID")
		return
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	albumReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/albums/%s/dl?t=%s", h.env.PrismURL, albumId, downloadToken), nil)
	albumReq.Header.Set("X-Auth-Token", h.env.PrismAuthToken)

	app.Logf("req %+v", albumReq)
	thumbRes, err := client.Do(albumReq)
	if err != nil {
		app.Logf("Error downloading album %+v", err)
	}

	defer thumbRes.Body.Close()
	thumbBytes, err := io.ReadAll(thumbRes.Body)

	zipReader, err := zip.NewReader(bytes.NewReader(thumbBytes), int64(len(thumbBytes)))
	if err != nil {
		app.Logf("failed to write file %s: ", err)
	}
	images := []string{}
	for _, zipFile := range zipReader.File {
		fmt.Println("Reading file:", zipFile.Name)
		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			app.Logf("%s", err)
			continue
		}

		err = os.WriteFile(fmt.Sprintf("web/pics/%s", zipFile.Name), unzippedFileBytes, 0666)
		if err != nil {
			app.Logf("failed to write: %s", err)
		} else {
			images = append(images, zipFile.Name)
		}
	}
	app.Logf("Returning images: %s", images)

	imageBytes, _ := json.Marshal(images)

	w.Write(imageBytes)
}
