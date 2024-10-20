package main

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type SlideShow struct {
	app.Compo
	Images       []string
	CurrentImage string
}

func (sh *SlideShow) Render() app.UI {
	return app.Div().Body(
        app.Img().Src("web/pics/" + sh.CurrentImage),
    )
}
