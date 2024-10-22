package ui

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

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
	OnClick   func(app.Context, app.Event)
	isLoading bool
}

func (db *DownloadButton) Render() app.UI {
	return app.Div().Body(
		app.If(db.isLoading, func() app.UI {
			return &LoadingSpinner{}
		}).Else(func() app.UI {
			return app.Button().Class("download-button").Text("Download").OnClick(db.OnClick)
		}),
	)
}

func (db *DownloadButton) OnMount(ctx app.Context) {
	ctx.Handle("downloadingPictures", db.HandleDownloadingPictures)
}

func (db *DownloadButton) HandleDownloadingPictures(ctx app.Context, a app.Action) {
	isDownloading, ok := a.Value.(bool)
	if !ok {
        app.Logf("Ex %s:", ok)
		return
	}
    db.isLoading = isDownloading
}

type Controls struct {
	app.Compo
}

func (c *Controls) Render() app.UI {
	return app.Div().Class("controls")
}

type AlbumSkeleton struct {
	app.Compo
}

func (as *AlbumSkeleton) Render() app.UI {
	return app.Div().Class("album-list").Body(
		app.Div().Class("album-skeleton skeleton-loader"),
		app.Div().Class("album-skeleton skeleton-loader"),
		app.Div().Class("album-skeleton skeleton-loader"),
	)
}
