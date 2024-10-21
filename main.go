package main

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"log"
	"net/http"
)

func main() {

    env := NewEnv()

    mainEntry := NewHello(env)


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

	http.HandleFunc("/fetchAlbums", mainEntry.FetchAlbums)
	http.HandleFunc("/fetchThumbnails", mainEntry.FetchThumbnails)
	http.HandleFunc("/downloadAlbum", mainEntry.DownloadAlbum)

	if err := http.ListenAndServe(":8001", nil); err != nil {
		log.Fatal(err)
	}
}
