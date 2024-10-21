package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/eyko139/photo-stream/internal/models"
	"github.com/eyko139/photo-stream/internal/env"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// hello is a component that displays a simple "Hello World!". A component is a
// customizable, independent, and reusable UI element. It is created by
// embedding app.Compo into a struct.
type hello struct {
	app.Compo
	env               *env.Env
	Albums            []models.Album
	ShouldReRender    bool
	isUpdateAvailable bool
	Gallery           *Gallery
	Text              string
	DownloadToken     string
	Images            []string
}

type ThumbNail struct {
	Name       string
	B64        string
	ClassName  string
	UID        string
	PhotoCount int
}

func NewHello(env *env.Env) *hello {
	return &hello{
		Text:   "Hello World!",
		env:    env,
		Images: []string{},
	}
}

// The Render method is where the component appearance is defined. Here, a
// "Hello World!" is displayed as a heading.
func (h *hello) Render() app.UI {

	return app.Div().Body(
		app.Div().Class("section-header").Text("Select albums"),
		app.Div().Class("album-list").Body(
			app.Range(h.Albums).Slice(func(i int) app.UI {
				item := h.Albums[i]
				return app.Div().Class("thumbnail-container").Class(item.ClassName).Body(
					app.Div().Class("thumbnail-header").Body(
						app.Div().Text(item.Title),
						app.Div().Text("Photos: "+fmt.Sprint(item.PhotoCount)),
					),
					app.Img().ID(item.UID).Src(fmt.Sprintf("data:image/jpg;base64, %s", item.B64)).OnClick(h.onClickAlbum),
				)
			}),
		),
		app.Button().Text("Download").OnClick(h.onClickPlay),
		app.Button().Text("Start").OnClick(h.onClickSlide),
	)
}

func (h *hello) onClickSlide(ctx app.Context, e app.Event) {
	ctx.Navigate("/slide")
}

func (h *hello) onClickPlay(ctx app.Context, e app.Event) {

	ctx.Async(func() {
		var wg sync.WaitGroup
		var update []string
        var mutex sync.Mutex
		for _, selectedThumbnails := range h.Albums {
			if selectedThumbnails.ClassName == "active" {
				wg.Add(1)
                go fetchSingleAlbum(h.DownloadToken, selectedThumbnails.UID, &wg, &update, &mutex)
			}
		}
		wg.Wait()
		ctx.Dispatch(func(ctx app.Context) {
			h.Images = update
			duration, _ := time.ParseDuration("99999h")
			ctx.SetState("images", h.Images).Persist().Broadcast().ExpiresIn(duration)
		})
	})
}

func fetchSingleAlbum(token, name string, wg *sync.WaitGroup, update *[]string, mutex *sync.Mutex) {
	defer wg.Done()
	var images []string
	res, err := http.Get(fmt.Sprintf("/downloadAlbum?albumId=%s&downloadToken=%s", name, token))
	if err != nil {
		app.Logf("failed to download album")
	}
	imageBytes, err := io.ReadAll(res.Body)
	if err != nil {
		app.Logf("Failed to decode: %s", err)
	}
	json.Unmarshal(imageBytes, &images)

    mutex.Lock()
    *update = append(*update, images...)
    mutex.Unlock()
}

func (h *hello) onClickAlbum(ctx app.Context, e app.Event) {
	albumClicked := ctx.JSSrc().Get("id").String()
	var update []models.Album
	for _, album := range h.Albums {
		if album.UID == albumClicked {
			album.ClassName = "active"
		}
		update = append(update, album)
	}
	ctx.Dispatch(func(ctx app.Context) {
		h.Albums = update
	})
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

		update := removeNonAlbums(albums)

		for idx, album := range update {
			thumbRes, _ := http.Get(fmt.Sprintf("/fetchThumbnails?albumId=%s&downloadToken=%s", album.UID, res.Header.Get("X-Download-Token")))
			thumbBytes, _ := io.ReadAll(thumbRes.Body)
			base64String := base64.StdEncoding.EncodeToString(thumbBytes)
			update[idx].B64 = base64String
			app.Logf("update: %+v", album)
		}

		ctx.Dispatch(func(ctx app.Context) {
			h.Albums = update
			h.DownloadToken = res.Header.Get("X-Download-Token")
		})

	})
}

func removeNonAlbums(albums []models.Album) []models.Album {
	var result []models.Album
	for _, album := range albums {
		if album.Type == "album" {
			app.Logf("base64: %s", album.B64)
			result = append(result, album)
		}
	}
	return result
}
