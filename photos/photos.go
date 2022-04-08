package photos

import (
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	// / HOST where to bind the upload
	host         = "localhost"
	port         = "8483"
	htmlDir      = "/home/www/photos"
	dbPath       = "/home/www/photos/photos.db"
	imagePerPage = 9
)

func index(c echo.Context) error {
	var items []Item
	db.Order("created_at desc").Limit(imagePerPage).Find(&items)
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"items": items,
	})
}

func dopage(pagenum int) (string, map[string]interface{}) {
	var items []Item
	var allitemscount int64

	db.Order("created_at desc").Find(&[]Item{}).Count(&allitemscount)
	db.Order("created_at desc").Limit(imagePerPage).Offset(pagenum * imagePerPage).Find(&items)

	pagePrevious := ""
	if pagenum > 2 {
		pagePrevious = fmt.Sprintf("page/%d", pagenum-1)
	}

	pageNext := fmt.Sprintf("page/%d", pagenum+1)
	if (pagenum*imagePerPage)+imagePerPage > int(allitemscount) {
		pageNext = ""
	}

	return "indexpp.html", map[string]interface{}{
		"pageNext":     pageNext,
		"pagePrevious": pagePrevious,
		"pageCurrent":  pagenum,
		"items":        items,
	}
}

func MakeStatic(outputdir string) error {
	var allitemscount int64
	var allitems []Item

	db, err := NewDB(getOrEnv("PHOTOS_DB", dbPath))
	if err != nil {
		return err
	}
	defer func() {
		dbConn, _ := db.DB()
		dbConn.Close()
	}()
	db.Order("created_at desc").Find(&[]Item{}).Count(&allitemscount)
	db.Order("created_at desc").Find(&allitems)
	var items []Item
	pagenum := 0
	citemcount := 0
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	for _, item := range allitems {
		ti, err := template.ParseFiles(getOrEnv("PHOTOS_HTML_DIRECTORY", htmlDir) + "/view.html")
		if err != nil {
			return err
		}
		var itemRandom Item
		for {
			itemRandom = allitems[rand.Intn(len(allitems))]
			if itemRandom.Href != item.Href {
				break
			}
		}

		itemM := map[string]interface{}{
			"item":       item,
			"itemRandom": itemRandom,
		}
		itempage := filepath.Join(outputdir, item.ImageUrl()+".html")
		err = os.MkdirAll(filepath.Dir(itempage), 0o755)
		if err != nil {
			return err
		}
		f, err := os.Create(itempage)
		if err != nil {
			return err
		}
		err = ti.Execute(f, itemM)
		if err != nil {
			return err
		}
		f.Close()

		items = append(items, item)
		citemcount += 1
		if citemcount < imagePerPage && item.Href != allitems[len(allitems)-1].Href {
			continue
		}
		pagePrevious := ""
		if pagenum > 1 {
			pagePrevious = fmt.Sprintf("page%d", pagenum-1)
		}

		pageNext := fmt.Sprintf("page%d", pagenum+1)
		if (pagenum*imagePerPage)+imagePerPage > int(allitemscount) {
			pageNext = ""
		}

		dico := map[string]interface{}{
			"pageNext":     pageNext,
			"pagePrevious": pagePrevious,
			"pageCurrent":  pagenum,
			"items":        items,
		}

		// generate template from file
		t, err := template.ParseFiles(getOrEnv("PHOTOS_HTML_DIRECTORY", htmlDir) + "/indexpp.html")
		if err != nil {
			return err
		}
		// execute template with data and write to individual file
		pagename := fmt.Sprintf("page%d.html", pagenum)
		if pagenum == 0 {
			pagename = "index.html"
		}
		f, err = os.Create(filepath.Join(outputdir, pagename))
		if err != nil {
			return err
		}
		// for _, v := range items {
		// 	fmt.Println(v.Href)
		// }
		// fmt.Println("---------")
		err = t.Execute(f, dico)
		if err != nil {
			return err
		}
		f.Close()
		citemcount = 0
		pagenum += 1
		fmt.Println(len(items))
		items = []Item{}
	}
	return nil
}

func page(c echo.Context) error {
	pageint, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Is this page a real page??.")
	}

	template, data := dopage(pageint)
	return c.Render(http.StatusOK, template, data)
}

func view(c echo.Context) error {
	var item Item
	var itemRandom Item

	db.Where("href=?", c.Param("href")).First(&item)
	db.Debug().Order("RAND()").Not(map[string]interface{}{"href": []string{c.Param("href")}}).First(&itemRandom)

	return c.Render(http.StatusOK, "view.html", map[string]interface{}{
		"item":       item,
		"itemRandom": itemRandom,
	})
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
	defer func() {
		dbConn, _ := db.DB()
		dbConn.Close()
	}()

	if os.Getenv("PHOTOS_DEBUG") != "" {
		db.Debug()
	}
	err = db.AutoMigrate(&Item{})
	if err != nil {
		return
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
	e.POST("/webhook", webhook)
	e.POST("/upload", upload)
	e.GET("/page/:page", page)
	e.GET("/:year/:month/:href", view)
	e.GET("/", index)

	return e.Start(
		fmt.Sprintf("%s:%s",
			getOrEnv("PHOTOS_HOST", host),
			getOrEnv("PHOTOS_PORT", port)))
}
