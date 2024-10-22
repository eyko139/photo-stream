package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/eyko139/photo-stream/internal/env"
	"github.com/eyko139/photo-stream/internal/models"
	"github.com/eyko139/photo-stream/ui"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"io"
	"net/http"
	"sync"
	"time"
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
	images            []string
	IsLoadingAlbums   bool
	FetchingError     bool
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
		images: []string{},
	}
}

func (h *hello) Render() app.UI {
	return app.Div().Body(
		app.Div().Class("section-header").Text("Photo Slideshow"),
		app.Div().Class("").Text("Select album(s)"),
		app.If(h.IsLoadingAlbums, func() app.UI {
			return &ui.AlbumSkeleton{}
		}).Else(func() app.UI {
			return app.Div().Class("album-list").Body(
				app.Range(h.Albums).Slice(func(i int) app.UI {
					item := h.Albums[i]
					return app.Div().Class("thumbnail-container").Class(item.ClassName).Body(
						app.Div().Class("thumbnail-header").Body(
							app.Div().Class("thumbnail-title").Text(item.Title),
							app.Div().Text("Photos: "+fmt.Sprint(item.PhotoCount)),
						),
						app.Img().ID(item.UID).Src(fmt.Sprintf("data:image/jpg;base64, %s", item.B64)).OnClick(h.onClickAlbum),
					)
				}),
			)
		}),
		app.Div().Class("controls").Body(
			&ui.DownloadButton{OnClick: h.onClickDownload},
			app.If(len(h.images) > 0, func() app.UI {
				return &ui.StartButton{}
			}),
		),
	)
}

func (h *hello) onClickDownload(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue("downloadingPictures", true)
	ctx.Async(func() {
		var wg sync.WaitGroup
		var update []string
		var mutex sync.Mutex
		for _, selectedThumbnails := range h.Albums {
			if selectedThumbnails.ClassName == "active" {
				wg.Add(1)
				go h.fetchSingleAlbum(h.DownloadToken, selectedThumbnails.UID, &wg, &update, &mutex)
			}
		}
		wg.Wait()
		h.images = update
		duration, _ := time.ParseDuration("99999h")
		ctx.SetState("images", h.images).Persist().Broadcast().ExpiresIn(duration)
		ctx.NewActionWithValue("downloadingPictures", false)
	})
}

func (h *hello) fetchSingleAlbum(token, name string, wg *sync.WaitGroup, update *[]string, mutex *sync.Mutex) {
	defer wg.Done()
	var images []string
	res, err := http.Get(fmt.Sprintf("%s/downloadAlbum?albumId=%s&downloadToken=%s", h.env.BaseUrl, name, token))
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
			if album.ClassName == "active" {
				album.ClassName = ""
			} else {
				album.ClassName = "active"
			}
		}
		update = append(update, album)
	}
	ctx.Dispatch(func(ctx app.Context) {
		h.Albums = update
	})
}

func (h *hello) OnMount(ctx app.Context) {
	ctx.Async(func() {
		ctx.Dispatch(func(ctx app.Context) {
			h.IsLoadingAlbums = true
		})
		var albums []models.Album
		res, err := http.Get(h.env.BaseUrl + "/fetchAlbums")
		bytes, err := io.ReadAll(res.Body)

		if err != nil {
			app.Logf("Failed to fetch albums")
			ctx.Dispatch(func(ctx app.Context) {
				h.FetchingError = true
				h.IsLoadingAlbums = false
			})
		}

		err = json.Unmarshal(bytes, &albums)

		if err != nil {
			app.Logf("cant unmarshall albums")
		}

		update := removeNonAlbums(albums)

		for idx, album := range update {
			thumbRes, _ := http.Get(fmt.Sprintf("%s/fetchThumbnails?albumId=%s&downloadToken=%s", h.env.BaseUrl, album.UID, res.Header.Get("X-Download-Token")))
			thumbBytes, _ := io.ReadAll(thumbRes.Body)
			base64String := base64.StdEncoding.EncodeToString(thumbBytes)
			update[idx].B64 = base64String
		}

		ctx.Dispatch(func(ctx app.Context) {
			h.IsLoadingAlbums = false
			h.Albums = update
			h.DownloadToken = res.Header.Get("X-Download-Token")
		})

	})
}

func removeNonAlbums(albums []models.Album) []models.Album {
	var result []models.Album
	for _, album := range albums {
		if album.Type == "album" {
			result = append(result, album)
		}
	}
	return result
}
