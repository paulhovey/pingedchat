package pcDatabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	giphyPrefix = "@giphy "
)

type GiphyImageData struct {
	URL    string `json:"url"`
	Width  string `json:"width"`
	Height string `json:"height"`
	Size   string `json:"size"`
	Frames string `json:"frames"`
	Mp4    string `json:"mp4"`
}

type GiphyGif struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	URL  string `json:"url"`
	// Tags               string
	BitlyGifURL        string `json:"bitly_gif_url"`
	BitlyFullscreenURL string `json:"bitly_fullscreen_url"`
	BitlyTiledURL      string `json:"bitly_tiled_url"`
	Images             struct {
		Original               GiphyImageData `json:"original"`
		FixedHeight            GiphyImageData `json:"fixed_height"`
		FixedHeightStill       GiphyImageData `json:"fixed_height_still"`
		FixedHeightDownsampled GiphyImageData `json:"fixed_height_downsampled"`
		FixedWidth             GiphyImageData `json:"fixed_width"`
		FixedwidthStill        GiphyImageData `json:"fixed_width_still"`
		FixedwidthDownsampled  GiphyImageData `json:"fixed_width_downsampled"`
	} `json:"images"`
}

type GiphyTranslateResponse struct {
	Data GiphyGif `json:"data"`
}

func rockGiphy(q string) string {
	TRACE.Printf("Searching for %q", q)
	url := fmt.Sprintf("http://api.giphy.com/v1/gifs/translate?s=%s&api_key=dc6zaTOxFJmzC", url.QueryEscape(q)) // default testing API key
	resp, err := http.Get(url)
	if err != nil {
		ERROR.Println(err)
		return ""
	}
	defer resp.Body.Close()
	giphyResp := GiphyTranslateResponse{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&giphyResp); err != nil {
		ERROR.Println("err in decoding giphyresp")
		ERROR.Println(err)
		return q // return original message back
	}
	TRACE.Println("giphyResp: ")
	TRACE.Println(giphyResp)
	msg := "NO RESULTS. I’M A MONSTER."
	msg = fmt.Sprintf(`@giphy %s:<br> <img src="%s" />`, q, giphyResp.Data.Images.Original.URL)
	TRACE.Println("msg from rockGiphy = " + msg)
	return msg
}
