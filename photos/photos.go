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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tjarratt/babble"
)

type Item struct {
	Image       string         `json:"image"`
	Href        string         `json:"href"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Date        SimpleJsonDate `json:"date"`
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
	yearMonth := fmt.Sprintf("%s/%s", c.Param("year"), c.Param("month"))

	var item = Item{}
	var itemNext = Item{}
	var itemPrevious = Item{}
	for i, value := range config {
		if value.Date.Format("2006/01") == yearMonth && value.Href == photo {
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

	return c.Render(http.StatusOK, "view.html", map[string]interface{}{
		"item":         item,
		"itemNext":     itemNext,
		"itemPrevious": itemPrevious,
	})
}

func upload(c echo.Context) error {
	href := c.FormValue("href")
	title := c.FormValue("title")
	description := c.FormValue("description")
	alphanum, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		return err
	}
	if title == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "You are missing a title")
	}

	if href == "" {
		href = strings.Trim(
			alphanum.ReplaceAllString(
				strings.ReplaceAll(
					strings.ToLower(
						strings.TrimSpace(title)),
					" ",
					"-"),
				""),
			"-")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	timef := time.Now()
	baseDirDate := timef.Format("2006/01")
	fname := href + filepath.Ext(file.Filename)
	fpath := filepath.Join(htmlDir, "content", "images", baseDirDate, fname)

	if _, err := os.Stat(fpath); err == nil {
		babbler := babble.NewBabbler()
		babbler.Count = 1
		randomword := strings.ToLower(alphanum.ReplaceAllString(babbler.Babble(), ""))
		fpath = filepath.Join(htmlDir, "content", "images",
			fmt.Sprintf("%s/%s-%s%s",
				baseDirDate,
				href,
				randomword,
				filepath.Ext(file.Filename)))

		fname = filepath.Base(fpath)
		href = filepath.Base(strings.TrimSuffix(fname, filepath.Ext(fname)))
	}
	fmt.Printf("%s %s %s \n", fpath, fname, href)

	err = os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return err
	}

	dst, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	items, err := readConfig()
	if err != nil {
		return err
	}

	newitem := Item{
		Image:       fname,
		Href:        href,
		Description: description,
		Date:        SimpleJsonDate{timef},
		Title:       title,
	}
	items = append([]Item{newitem}, items...)
	configFP, _ := json.MarshalIndent(items, "", " ")
	err = ioutil.WriteFile(filepath.Join(htmlDir, "config.json"), configFP, 0644)
	if err != nil {
		return err
	}

	err = Generate()
	if err != nil {
		return err
	}

	// return c.HTML(http.StatusOK, "<b>Uploaded!<b>")
	return c.Redirect(http.StatusMovedPermanently, "/")
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

func getOrEnv(env string, def string) string {
	if os.Getenv(env) != "" {
		def = os.Getenv(env)
	}
	return def
}

func Server() (err error) {
	htmlDir = getOrEnv("PHOTOS_HTML_DIRECTORY", htmlDir)
	e := echo.New()
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob(filepath.Join(htmlDir, "html", "*.html"))),
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{}))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	if os.Getenv("PHOTOS_DEBUG") != "" {
		e.Debug = true
	}
	e.Static("/assets", filepath.Join(htmlDir, "assets"))
	e.Static("/content", filepath.Join(htmlDir, "content"))
	e.GET("/upload", func(c echo.Context) error {
		return c.Render(http.StatusOK, "upload.html", map[string]interface{}{})
	})
	e.POST("/upload", upload)
	e.GET("/page/:page", page)
	e.GET("/:year/:month/:photo", view)
	e.GET("/", index)

	return (e.Start(
		fmt.Sprintf("%s:%s",
			getOrEnv("PHOTOS_HOST", host),
			getOrEnv("PHOTOS_PORT", port))))
}
