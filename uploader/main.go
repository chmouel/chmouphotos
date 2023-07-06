package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	// / HOST where to bind the upload
	chmouPhotosHost = "localhost"
	chmouPhotosPort = "1322"
	redirectURL     = "https://github.com/chmouel/chmouphotos/actions"
)

//go:embed html/upload.html
var indexTmpl []byte

func getOrEnv(env string, def string) string {
	if os.Getenv(env) != "" {
		def = os.Getenv(env)
	}
	return def
}

func RunGit(dir string, args ...string) (string, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		// nolint: nilerr
		return "", nil
	}
	c := exec.Command(gitPath, args...)
	var output bytes.Buffer
	c.Stderr = &output
	c.Stdout = &output
	// This is the optional working directory. If not set, it defaults to the current
	// working directory of the process.
	if dir != "" {
		c.Dir = dir
	}
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("error running, %s, output: %s error: %w", args, output.String(), err)
	}
	return output.String(), nil
}

func getDir() string {
	if env := os.Getenv("CHMOUPHOTOS_DIR"); env != "" {
		return env
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	defpath := filepath.Dir(ex)
	return defpath
}

func main() {
	e := echo.New()
	e.GET("/photos/upload", func(c echo.Context) error {
		return c.HTML(http.StatusOK, string(indexTmpl))
	})
	e.POST("/photos/upload", upload)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{}))
	if err := e.Start(chmouPhotosHost + ":" + chmouPhotosPort); err != nil {
		log.Fatal(err)
	}
}

func upload(c echo.Context) error {
	rootDir := getDir()
	alphanum, err := regexp.Compile("[^a-zA-Z0-9-]+")
	description := c.FormValue("description")
	title := c.FormValue("title")

	if title == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "You are missing a title")
	}
	href := strings.Trim(
		alphanum.ReplaceAllString(
			strings.ReplaceAll(
				strings.ToLower(
					strings.TrimSpace(title)),
				" ",
				"-"),
			""),
		"-")

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "You are missing a file")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	fdir := filepath.Join(rootDir, "content", href)
	savepath := filepath.Join(fdir, href+filepath.Ext(file.Filename))
	if _, err := os.Stat(savepath); err == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "post already exist")
	}
	fmt.Println(description)
	c.Logger().Debugf("Saving %s to %s", file.Filename, savepath)

	err = os.MkdirAll(fdir, 0o755)
	if err != nil {
		return err
	}
	dst, err := os.Create(savepath)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	dt := time.Now()
	date := dt.Format("2006-01-02T15:04:05")
	md := filepath.Join(fdir, "index.md")
	mdcontent := fmt.Sprintf("---\ntitle: %s\ndate: %s+02:00\n---\n%s", title, date, description)
	err = ioutil.WriteFile(md, []byte(mdcontent), 0o644)
	if err != nil {
		return err
	}

	if output, err := RunGit(rootDir, "add", filepath.Join("content", href)); err != nil {
		return fmt.Errorf("cannot add content: %s err: %w", output, err)
	}

	if output, err := RunGit(rootDir, "commit", "-m", fmt.Sprintf("add post %s", href), filepath.Join("content", href)); err != nil {
		return fmt.Errorf("cannot commit : %s err: %w", output, err)
	}

	if output, err := RunGit(rootDir, "pull", "--rebase", "origin"); err != nil {
		return fmt.Errorf("cannot pull ff only with output: %s err: %w", output, err)
	}

	if output, err := RunGit(rootDir, "push", "origin", "main"); err != nil {
		return fmt.Errorf("cannot push origin main: %s err: %w", output, err)
	}

	return c.Redirect(http.StatusMovedPermanently, redirectURL)
}
