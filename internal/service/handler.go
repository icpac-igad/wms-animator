package service

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/gocraft/web"
	"github.com/icpac-igad/wms-animator/internal/conf"
	"github.com/icpac-igad/wms-animator/internal/wms"
)

type Context struct {
}

func Error(rw web.ResponseWriter, req *web.Request, err interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recovered panic:", err)
			return
		}
		fmt.Println("no panic recovered")
	}()
}

func (c *Context) HandleGet(rw web.ResponseWriter, req *web.Request) {

	// ready request body
	body, err := ioutil.ReadAll(req.Body)

	defer req.Body.Close()

	if err != nil {
		err := appError{Status: http.StatusBadRequest, Message: err.Error()}
		JSONHandleError(rw, err)
		return
	}

	var input wms.Input
	err = json.Unmarshal(body, &input)

	if err != nil {
		JSONHandleError(rw, appError{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if input.FramesPerSecond == 0 {
		if conf.Configuration.Wms.FramesPerSecond != 0 {
			input.FramesPerSecond = conf.Configuration.Wms.FramesPerSecond
		} else {
			input.FramesPerSecond = 3
		}
	}

	wmsImages, err := wms.GetAllWmsImages(input)

	if err != nil {
		err := appError{Status: http.StatusBadRequest, Message: err.Error()}
		JSONHandleError(rw, err)
		return
	}

	var keys []string
	for key := range wmsImages {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var wmsImagesSorted []image.Image

	for _, key := range keys {
		wmsImagesSorted = append(wmsImagesSorted, wmsImages[key])
	}

	buff, err := wms.GenerateGif(wmsImagesSorted, input.FramesPerSecond)

	if err != nil {
		err := appError{Status: http.StatusBadRequest, Message: err.Error()}
		JSONHandleError(rw, err)
		return
	}

	rw.Header().Set("Content-Type", "image/gif")
	rw.Write(buff.Bytes())
}

func initRouter(basePath string) *web.Router {
	// create router
	router := web.New(Context{})

	// ovveride gocraft defualt error handler
	router.Error(Error)

	// add middlewares
	router.Middleware(web.LoggerMiddleware)
	// router.Middleware(web.ShowErrorsMiddleware)

	// handle routes
	router.Post("/wms", (*Context).HandleGet)

	return router
}
