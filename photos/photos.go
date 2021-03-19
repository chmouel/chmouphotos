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

var (
	/// HOST where to bind the upload
	host         = "localhost"
	port         = "8483"
	htmlDir      = "/home/www/photos"
	dbPath       = "/home/www/photos/config.db"
	imagePerPage = 9
)

func index(c echo.Context) error {
	var items []Item
	db.Limit(imagePerPage).Find(&items)
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"items": items,
	})
}

func page(c echo.Context) error {
	var items []Item
	var allitemscount int64

	pageint, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Is this page a real page??.")
	}

	db.Find(&[]Item{}).Count(&allitemscount)
	db.Limit(imagePerPage).Offset(pageint * imagePerPage).Find(&items)

	var pagePrevious = ""
	if pageint > 2 {
		pagePrevious = fmt.Sprintf("page/%d", pageint-1)
	}

	var pageNext = fmt.Sprintf("page/%d", pageint+1)
	if (pageint*imagePerPage)+imagePerPage > int(allitemscount) {
		pageNext = ""
	}

	return c.Render(http.StatusOK, "indexpp.html", map[string]interface{}{
		"pageNext":     pageNext,
		"pagePrevious": pagePrevious,
		"pageCurrent":  pageint,
		"items":        items,
	})
}

func view(c echo.Context) error {
	var item Item
	var itemPrevious Item
	var items []Item
	var itemNext Item

	db.Raw("select photo.* from photos as photo where photo.id >= (select id-1 from photos where href=?) order by photo.id asc limit 3;", c.Param("href")).Scan(&items)

	if len(items) == 3 && items[0].Href == c.Param("href") {
		db.Raw("select * from photos where id = (select max(id) from photos)").Scan(&itemPrevious)
		item = items[0]
		itemNext = items[1]
	} else if len(items) == 3 {
		itemPrevious = items[0]
		item = items[1]
		itemNext = items[2]
	} else if len(items) == 2 && items[1].Href == c.Param("href") {
		itemPrevious = items[0]
		item = items[1]
		db.Where("id=1").First(&itemNext)
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

	db, err = NewDB(getOrEnv("PHOTOS_DB", dbPath))
	if err != nil {
		return
	}
	if os.Getenv("PHOTOS_DEBUG") != "" {
		db.Debug()
	}
	db.AutoMigrate(&Item{})

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
	e.GET("/:year/:month/:href", view)
	e.GET("/", index)

	return (e.Start(
		fmt.Sprintf("%s:%s",
			getOrEnv("PHOTOS_HOST", host),
			getOrEnv("PHOTOS_PORT", port))))
}
