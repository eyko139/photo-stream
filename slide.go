package main

import (
	"time"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type SlideShow struct {
	app.Compo
	CurrentImage int
	Images       []string
	IsLoading    bool
}

func (sh *SlideShow) Render() app.UI {
	app.Logf("images slide", sh.Images)
	return app.If(sh.IsLoading, func() app.UI {
		return app.Div().Text("Loading...")
	}).ElseIf(len(sh.Images) > 0, func() app.UI {
		return app.Div().Body(
			app.Div().Body(
				app.Img().Class("slideshow-image").Src("web/pics/" + sh.Images[sh.CurrentImage]),
			),
		)
	},
	).Else(func() app.UI {
		return app.Div().Text("Whats going on")
	})
}

func (sh *SlideShow) OnMount(ctx app.Context) {
	var images []string
	sh.IsLoading = true

	ctx.GetState("images", &images)

    if len(images) == 0 {
        app.Logf("No pictures in local storage")
    }

	if len(images) > 0 {
		sh.Images = images
		sh.IsLoading = false
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		currentImage := 0
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
                if currentImage == len(images)-1 {
                    currentImage = -1
                }
				currentImage++
				ctx.Dispatch(func(ctx app.Context) {
					sh.CurrentImage = currentImage
					app.Logf("current image", currentImage)
				})
			}
		}
	}()

}
