package wms

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/icpac-igad/wms-animator/internal/conf"
)

type result struct {
	url        string
	paramValue string
	image      image.Image
	err        error
}

type Input struct {
	BaseUrl     string            `json:"url"`
	Parameter   Parameter         `json:"parameter"`
	WmsParams   map[string]string `json:"wms_params"`
	Title       string            `json:"title"`
	Attribution string            `json:"attribution"`
}

type Parameter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type ParamUrl struct {
	Url         string
	Value       string
	Attribution string
	Title       string
}

const maxDigesters = 50

// generate urls from input
func generateUrls(done <-chan struct{}, input Input) (<-chan ParamUrl, <-chan error) {
	urls := make(chan ParamUrl)
	errc := make(chan error, 1)

	go func() {
		// Close the urls channel after generator returns.
		defer close(urls)

		baseUrl, err := url.Parse(input.BaseUrl)

		if err != nil {
			errc <- err
		}

		for _, value := range input.Parameter.Values {
			url := getUrl(baseUrl, input.Parameter.Name, value, input.WmsParams)

			select {
			case urls <- ParamUrl{Url: url, Value: value, Title: input.Title, Attribution: input.Attribution}:
			case <-done:
				errc <- errors.New("generator canceled")
			}
		}

		// no error. close error channel
		errc <- nil
	}()

	return urls, errc
}

// get wms image for url
func getWmsImage(done <-chan struct{}, urls <-chan ParamUrl, c chan<- result) {
	for url := range urls {
		resImage, err := requestHttpImage(url)

		select {
		case c <- result{url: url.Url, paramValue: url.Value, image: resImage, err: err}:
		case <-done:
			return
		}

	}
}

// get all wms images
func GetAllWmsImages(input Input) (map[string]image.Image, error) {

	done := make(chan struct{})
	defer close(done)

	numImages := len(input.Parameter.Values)

	urls, errc := generateUrls(done, input)

	// Start a fixed number of goroutines to request urls
	c := make(chan result) // HLc
	var wg sync.WaitGroup

	numDigesters := maxDigesters

	// use less go routines if we have less than maxDigesters
	if numImages <= maxDigesters {
		numDigesters = numImages
	}

	wg.Add(numDigesters)

	for i := 0; i < numDigesters; i++ {
		go func() {
			getWmsImage(done, urls, c) // HLc
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c) // HLc
	}()
	// End of pipeline. OMIT

	m := make(map[string]image.Image)

	for r := range c {
		if r.err != nil {
			return nil, r.err
		}

		m[r.paramValue] = r.image
	}

	// Check whether the generator failed.
	if err := <-errc; err != nil { // HLerrc
		return nil, err
	}

	return m, nil
}

func getUrl(baseUrl *url.URL, paramName string, value string, wmsParams map[string]string) string {

	if baseUrl.Scheme == "" {
		baseUrl.Scheme = "https"
	}
	q := baseUrl.Query()

	// common wms params
	for k, v := range wmsParams {
		q.Set(k, v)
	}

	// dynamic param
	q.Set(paramName, value)

	baseUrl.RawQuery = q.Encode()

	return baseUrl.String()
}

func requestHttpImage(url ParamUrl) (image.Image, error) {

	layoutISO := "2006-01-02T15:04:05.000Z"
	layoutISOString := "UTC 2006-01-02 15:04"

	res, err := http.Get(url.Url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resImage, _, err := image.Decode(res.Body)

	if err != nil {
		return nil, err
	}

	// add texts
	dc := gg.NewContextForImage(resImage)

	err = dc.LoadFontFace(conf.Configuration.Wms.FontPath, 18)

	if err != nil {
		return nil, fmt.Errorf("error loading font file: %s", conf.Configuration.Wms.FontPath)
	}

	textPadding := float64(15)
	bgPadding := float64(10)
	bgRadius := float64(4)
	yPosition := float64(dc.Height()) - textPadding
	dcWidth := dc.Width()

	// add title
	if url.Title != "" {
		pTextWidth, pTextHeight := dc.MeasureString(url.Title)
		dc.DrawRoundedRectangle(textPadding, textPadding, pTextWidth+bgPadding, pTextHeight+bgPadding, bgRadius)
		dc.SetColor(color.RGBA{0, 0, 0, 100})
		dc.Fill()

		dc.SetColor(color.White)
		dc.DrawString(url.Title, textPadding+(bgPadding/2), textPadding+(pTextHeight/2)+bgPadding)

	}

	t, err := time.Parse(layoutISO, url.Value)

	if err != nil {
		return nil, fmt.Errorf("invalid date: %s ", url.Value)
	}

	tString := t.Format(layoutISOString)

	pTextWidth, pTextHeight := dc.MeasureString(tString)

	dc.DrawRoundedRectangle(float64(dcWidth)-(pTextWidth+float64(textPadding)), yPosition-pTextHeight, pTextWidth+bgPadding, pTextHeight+bgPadding, bgRadius)
	dc.SetColor(color.RGBA{0, 0, 0, 100})
	dc.Fill()

	dc.SetColor(color.White)
	dc.DrawString(tString, float64(dcWidth)-(pTextWidth+textPadding)+(bgPadding/2), yPosition+(bgPadding/2))

	if url.Attribution != "" {

		pTextWidth, pTextHeight := dc.MeasureString(url.Attribution)

		dc.DrawRoundedRectangle(float64(textPadding), yPosition-pTextHeight, pTextWidth+bgPadding, pTextHeight+bgPadding, bgRadius)
		dc.SetColor(color.RGBA{0, 0, 0, 100})
		dc.Fill()

		dc.SetColor(color.White)
		dc.DrawString(url.Attribution, textPadding+(bgPadding/2), yPosition+(bgPadding/2))

	}

	return dc.Image(), nil
}
