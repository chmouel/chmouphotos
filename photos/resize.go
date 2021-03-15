package photos

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func Generate() error {
	htmlDir = getOrEnv("PHOTOS_HTML_DIRECTORY", htmlDir)

	items, err := readConfig()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "We could not read config???.")
	}

	for _, item := range items {
		orig := filepath.Join(htmlDir, "content", "images", item.Date.Format("2006/01"), item.Image)
		if _, err := os.Stat(orig); os.IsNotExist(err) {
			return errors.New(item.Href + " doesnt exist. clean your json.")
		}
		err = resize(filepath.Join(item.Date.DayMonthDir(), item.Image))
		if err != nil {
			return err
		}
	}

	return nil
}

func resize(filename string) error {
	htmlDir = getOrEnv("PHOTOS_HTML_DIRECTORY", htmlDir)
	fpath := filepath.Join(htmlDir, "content", "images", filename)

	sizes := []int{1000, 1200, 1600, 2000, 30, 300, 600}
	for _, size := range sizes {
		fsize := filepath.Join(htmlDir, "content", "images", "size", fmt.Sprintf("w%d", size), filename)
		if _, err := os.Stat(fsize); os.IsNotExist(err) {
			fmt.Printf("Sizing %s to dimension %d in %s\n", filename, size, fsize)
			err = os.MkdirAll(filepath.Dir(fsize), 0755)
			if err != nil {
				return (err)
			}
			fmt.Printf("Resize %d %s to %s\n", size, fpath, fsize)
			out, err := exec.Command("convert", "-resize", fmt.Sprintf("%dx", size), fpath, fsize).CombinedOutput()
			if err != nil {
				fmt.Println(out)
				return err
			}
		}
	}

	return nil
}
