package photos

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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

func index(c echo.Context) error {
	config, err := readConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not read config???.")
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"items": config[0:imagePerPage],
	})
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

	return c.Render(http.StatusOK, "indexpp.html", map[string]interface{}{
		"pageNext":     pageNext,
		"pagePrevious": pagePrevious,
		"pageCurrent":  pageint,
		"items":        getchunk(pageint, config),
	})
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

	return c.Render(http.StatusOK, "indexpp.html", map[string]interface{}{
		"item":         item,
		"itemNext":     itemNext,
		"itemPrevious": itemPrevious,
	})
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

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Server() (err error) {
	if os.Getenv("PHOTOS_HTML_DIRECTORY") != "" {
		htmlDir = os.Getenv("PHOTOS_HTML_DIRECTORY")
	}

	if os.Getenv("PHOTOS_HOST") != "" {
		host = os.Getenv("PHOTOS_HOST")
	}

	if os.Getenv("PHOTOS_PORT") != "" {
		port = os.Getenv("PHOTOS_PORT")
	}

	templates := &TemplateRenderer{
		templates: template.Must(template.ParseGlob(filepath.Join(htmlDir, "html", "*.html"))),
	}

	e := echo.New()
	e.Renderer = templates
	e.Debug = true
	e.Static("/assets", filepath.Join(htmlDir, "assets"))
	e.Static("/content", filepath.Join(htmlDir, "content"))
	e.GET("/page/:page", page)
	e.GET("/view/:photo", view)
	e.GET("/", index)

	return (e.Start(fmt.Sprintf("%s:%s", host, port)))
}
