package photos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/flosch/pongo2/v4"
	"github.com/labstack/echo/v4"
)

type Item struct {
	Image string `json:"image"`
	Href  string `json:"href"`
	Desc  string `json:"desc"`
}

var (
	/// HOST where to bind the upload
	host         = "localhost"
	port         = "8483"
	htmlDir      = "/home/www/photos"
	imagePerPage = 9
)

var (
	indexPPTPL = pongo2.Must(pongo2.FromFile(filepath.Join(htmlDir, "html", "indexpp.html")))
	indexTPL   = pongo2.Must(pongo2.FromFile(filepath.Join(htmlDir, "html", "index.html")))
	viewTPL    = pongo2.Must(pongo2.FromFile(filepath.Join(htmlDir, "html", "page.html")))
)

func index(c echo.Context) error {
	config, err := readConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not read config???.")
	}
	indexHTML, err := indexTPL.Execute(pongo2.Context{"items": config[0:imagePerPage]})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not execute template???.")
	}
	return c.HTML(http.StatusOK, indexHTML)
}

func page(c echo.Context) error {
	config, err := readConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not read config???.")
	}

	pageint, err := strconv.Atoi(c.Param("page"))

	var pagePrevious = ""
	if pageint > 2 {
		pagePrevious = fmt.Sprintf("page/%d", pageint-1)
	}

	var pageNext = fmt.Sprintf("page/%d", pageint+1)
	if (pageint*imagePerPage)+imagePerPage > len(config) {
		pageNext = ""
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Is this page a real page??.")
	}
	imageChunk := getchunk(pageint, config)

	indexHTML, err := indexPPTPL.Execute(pongo2.Context{
		"pageNext":     pageNext,
		"pagePrevious": pagePrevious, "pageCurrent": pageint, "items": imageChunk})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not execute template???.")
	}
	return c.HTML(http.StatusOK, indexHTML)
}

func view(c echo.Context) error {
	config, err := readConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not read config???.")
	}
	photo := c.Param("photo")
	var item = Item{}
	var itemNext = Item{}
	var itemPrevious = Item{}
	for i, value := range config {
		if value.Href == photo {
			item = value
			if i+1 < len(config) {
				itemNext = config[i+1]
			} else {
				itemNext = config[0]
			}
			if i == 0 {
				itemPrevious = config[len(config)-1]
			} else {
				itemPrevious = config[i-1]
			}

		}
	}

	if item.Href == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "I no nothing about this photo???.")
	}

	viewHTML, err := viewTPL.Execute(pongo2.Context{"item": item, "itemNext": itemNext, "itemPrevious": itemPrevious})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not execute template???.")
	}

	return c.HTML(http.StatusOK, viewHTML)
}

func getchunk(page int, items []Item) []Item {
	var cnt = 0

	for i := 0; i < len(items); i += imagePerPage {
		end := i + imagePerPage
		if end > len(items) {
			end = len(items)
		}
		if cnt == page {
			return items[i:end]
		}
		cnt++
	}
	return []Item{}
}

func readConfig() ([]Item, error) {
	var items []Item

	configJson, err := os.Open(filepath.Join(htmlDir, "config.json"))
	if err != nil {
		return items, err
	}
	defer configJson.Close()
	byteValue, _ := ioutil.ReadAll(configJson)
	json.Unmarshal(byteValue, &items)
	return items, nil
}

func Server() error {
	if os.Getenv("PHOTOS_HTML_DIRECTORY") != "" {
		htmlDir = os.Getenv("PHOTOS_HTML_DIRECTORY")
	}

	if os.Getenv("PHOTOS_HOST") != "" {
		host = os.Getenv("PHOTOS_HOST")
	}

	if os.Getenv("PHOTOS_PORT") != "" {
		port = os.Getenv("PHOTOS_PORT")
	}

	e := echo.New()

	e.Static("/assets", filepath.Join(htmlDir, "assets"))
	e.Static("/content", filepath.Join(htmlDir, "content"))
	e.GET("/page/:page", page)
	e.GET("/view/:photo", view)
	e.GET("/", index)

	return (e.Start(fmt.Sprintf("%s:%s", host, port)))
}
