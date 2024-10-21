package main

import (
	"time"

	"fmt"

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
	var test []string
	sh.IsLoading = true
	ctx.GetState("images", &test)
	fmt.Println("images slide observed", test)
	// ctx.ObserveState("images", &test).OnChange(func() {
	// fmt.Println("images slide observed")
	// app.Logf("images slide observed", test)
	// 	if len(test) > 0 {
	// 		// sh.Images = test
	// 		sh.IsLoading = false

	if len(test) > 0 {
		sh.Images = test
		sh.IsLoading = false
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		currentImage := 0
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				currentImage++
				ctx.Dispatch(func(ctx app.Context) {
					sh.CurrentImage = currentImage
					app.Logf("current image", currentImage)
				})
			}
		}
	}()

}
