package ui

import "github.com/maxence-charriere/go-app/v10/pkg/app"


type LoadingSpinner struct {
    app.Compo
}

func (ls *LoadingSpinner) Render() app.UI {
    return app.Div().Class("loading-spinner")
}

type StartButton struct {
    app.Compo
}

func (sb *StartButton) Render() app.UI {
    return app.Button().Class("start-button").OnClick(sb.onClickSlide)
}

func (sb *StartButton) onClickSlide(ctx app.Context, e app.Event) {
	ctx.Navigate("/slide")
}

type DownloadButton struct {
    app.Compo
    OnClick func(app.Context, app.Event)
}

func (db *DownloadButton) Render() app.UI {
    return app.Button().Class("download-button").Text("Download").OnClick(db.OnClick)
}


type Controls struct {
    app.Compo
}

func (c *Controls) Render() app.UI {
    return app.Div().Class("controls")
}
