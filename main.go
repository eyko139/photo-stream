package main

import (
	"log"
	"net/http"

	"github.com/eyko139/photo-stream/internal/api"
	"github.com/eyko139/photo-stream/internal/env"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {

    env := env.NewEnv()

    mainEntry := NewHello(env)

    api := api.NewApi(env)

	app.Route("/", func() app.Composer {
		return mainEntry	
    })

    app.Route("/slide", func() app.Composer {
        return &SlideShow{IsLoading: true }
    })

	app.RunWhenOnBrowser()

	// Finally, launching the server that serves the app is done by using the Go
	// standard HTTP package.
	//
	// The Handler is an HTTP handler that serves the client and all its
	// required resources to make it work into a web browser. Here it is
	// configured to handle requests with a path that starts with "/".
	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Styles: []string{
			"/web/main.css",
		},
	})

	http.HandleFunc("/fetchAlbums", api.FetchAlbums())
	http.HandleFunc("/fetchThumbnails", api.FetchThumbnails())
	http.HandleFunc("/downloadAlbum", api.DownloadAlbum())

	if err := http.ListenAndServe(":8001", nil); err != nil {
		log.Fatal(err)
	}
}
