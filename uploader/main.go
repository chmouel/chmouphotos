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
)

var (
	// / HOST where to bind the upload
	chmouPhotosHost = "localhost"
	chmouPhotosPort = "1314"
	chmouPhotosDir  = "/tmp/photos"
)

//go:embed html/upload.html
var uploadPage []byte

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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.ParseMultipartForm(32 << 20)
		alphanum, err := regexp.Compile("[^a-zA-Z0-9-]+")
		description := r.FormValue("description")
		title := r.FormValue("title")
		href := strings.Trim(
			alphanum.ReplaceAllString(
				strings.ReplaceAll(
					strings.ToLower(
						strings.TrimSpace(title)),
					" ",
					"-"),
				""),
			"-")

		file, handler, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			fmt.Println(err)
			return
		}
		defer file.Close()
		postDir := filepath.Join(getOrEnv("CHMOUPHOTOS_DIR", chmouPhotosDir), "content", href)
		err = os.MkdirAll(postDir, 0o755)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		f, err := os.OpenFile(filepath.Join(postDir, handler.Filename), os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
		defer f.Close()

		dt := time.Now()
		date := dt.Format("2006-01-02T15:04:05")
		md := filepath.Join(postDir, "index.md")
		mdcontent := fmt.Sprintf(`---
title: %s
date: %s+02:00
image: %s
---
%s
`, title, date, handler.Filename, description)
		err = ioutil.WriteFile(md, []byte(mdcontent), 0o644)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
		io.Copy(f, file)

		output, err := RunGit(getOrEnv("CHMOUPHOTOS_DIR", chmouPhotosDir), "add", filepath.Join("content", href))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			fmt.Println(err)
			return
		}
		if output != "" {
			fmt.Println(output)
		}
		output, err = RunGit(getOrEnv("CHMOUPHOTOS_DIR", chmouPhotosDir), "commit", "-m", fmt.Sprintf("add post %s", href), filepath.Join("content", href))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if output != "" {
			fmt.Println(output)
		}

		output, err = RunGit(getOrEnv("CHMOUPHOTOS_DIR", chmouPhotosDir), "push", "origin", "main")
		if err != nil {
			fmt.Println(err)
			return
		}
		if output != "" {
			fmt.Println(output)
		}

		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(uploadPage)
	})
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", getOrEnv("CHMOUPHOTOS_HOST", chmouPhotosHost), getOrEnv("CHMOUPHOTOS_PORT", chmouPhotosPort)), mux)
	if err != nil {
		log.Fatal(err)
	}
}
