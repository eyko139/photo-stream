package main

import (
	"encoding/base64"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Gallery struct {
	app.Compo
	ThumbNails [][]byte
}

func (g *Gallery) Render() app.UI {
    app.Logf("tbums %s", g.ThumbNails)
	return app.Ul().Body(
		app.Range(g.ThumbNails).Slice(func(i int) app.UI {
			return app.Img().Src(fmt.Sprintf("data:image/jpg;base64, %s", base64.StdEncoding.EncodeToString(g.ThumbNails[i])))
		}),
	)

}
