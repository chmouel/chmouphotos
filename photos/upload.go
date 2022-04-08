package photos

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tjarratt/babble"
)

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

	err = os.MkdirAll(filepath.Dir(fpath), 0o755)
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

	newitem := Item{
		Image:       fname,
		Href:        href,
		Description: description,
		Title:       title,
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(&newitem).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	err = Generate()
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}
