package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/eyko139/photo-stream/internal/env"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Api struct {
	env *env.Env
}

func NewApi(env *env.Env) *Api {
	return &Api{
		env: env,
	}
}

func (a *Api) DownloadAlbum() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		albumId := r.URL.Query().Get("albumId")
		downloadToken := r.URL.Query().Get("downloadToken")

		if albumId == "" || downloadToken == "" {
			app.Logf("missing token or ID")
			return
		}

		client := http.Client{
			Timeout: 30 * time.Second,
		}

		albumReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/albums/%s/dl?t=%s", a.env.PrismURL, albumId, downloadToken), nil)
		albumReq.Header.Set("X-Auth-Token", a.env.PrismAuthToken)

		thumbRes, err := client.Do(albumReq)
		if err != nil {
			app.Logf("Error downloading album %+v", err)
		}

		defer thumbRes.Body.Close()
		thumbBytes, err := io.ReadAll(thumbRes.Body)

		zipReader, err := zip.NewReader(bytes.NewReader(thumbBytes), int64(len(thumbBytes)))
		if err != nil {
			app.Logf("failed to write file %s: ", err)
		}
		images := []string{}
		for _, zipFile := range zipReader.File {
			unzippedFileBytes, err := readZipFile(zipFile)
			if err != nil {
				app.Logf("%s", err)
				continue
			}

			err = os.WriteFile(fmt.Sprintf("web/pics/%s", zipFile.Name), unzippedFileBytes, 0666)
			if err != nil {
				app.Logf("failed to write: %s", err)
			} else {
				images = append(images, zipFile.Name)
			}
		}

		imageBytes, _ := json.Marshal(images)

		w.Write(imageBytes)
	}
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (a *Api) FetchThumbnails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		albumId := r.URL.Query().Get("albumId")
		downloadToken := r.URL.Query().Get("downloadToken")

		if albumId == "" || downloadToken == "" {
			app.Logf("missing token or ID")
			return
		}

		client := http.Client{
			Timeout: 30 * time.Second,
		}

		thumbReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/albums/%s/t/%s/tile_500", a.env.PrismURL, albumId, downloadToken), nil)
		thumbReq.Header.Set("X-Auth-Token", "jqGHgo-aXoKSV-muuhiF-Djh1Zu")

		// app.Logf("req %+v", thumbReq)
		thumbRes, err := client.Do(thumbReq)
		if err != nil {
			app.Logf("Error fetching thumbnails %s", err)
		}

		defer thumbRes.Body.Close()
		thumbBytes, err := io.ReadAll(thumbRes.Body)

		os.WriteFile(fmt.Sprintf("web/thumbs/%s.jpg", albumId), thumbBytes, 0666)
		w.Write(thumbBytes)
	}
}

func (a *Api) FetchAlbums() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequest("GET", a.env.PrismURL+"/api/v1/albums?count=10", nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("X-Auth-Token", a.env.PrismAuthToken)
		client := http.Client{
			Timeout: 30 * time.Second,
		}

		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("client: error making http request: %s\n", err)
		}

		dt := res.Header.Get("X-Download-Token")
		defer res.Body.Close()
		bytes, _ := io.ReadAll(res.Body)
		w.Header().Add("X-Download-Token", dt)
		w.Write(bytes)
	}
}
