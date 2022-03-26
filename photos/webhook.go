package photos

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/google/go-github/v43/github"
	"github.com/labstack/echo/v4"
)

const blogRepository = "/home/www/chmouel/photos"

func run(dir, cmd string, args ...string) (string, error) {
	gitPath, err := exec.LookPath(cmd)
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

func process(event *github.PushEvent) {
	sha := event.Commits[0].ID
	blogDir := getOrEnv("BLOG_REPOSITORY", blogRepository)
	output, err := run(blogDir, "git", "fetch", "-a", "origin")
	if output != "" {
		log.Println(output)
	}
	if err != nil {
		log.Println(err.Error())
		return
	}

	output, err = run(blogDir, "git", "reset", "--hard", *sha)
	if output != "" {
		log.Println(output)
	}
	if err != nil {
		log.Println(err.Error())
		return
	}

	output, err = run(blogDir, "hugo", "--gc", "--minify")
	if output != "" {
		log.Println(output)
	}
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func webhook(c echo.Context) error {
	payload, err := github.ValidatePayload(c.Request(), []byte(os.Getenv("BLOG_WEBHOOK_SECRET")))
	if err != nil {
		return err
	}
	eventType, err := github.ParseWebHook(c.Request().Header.Get(github.EventTypeHeader), payload)
	if err != nil {
		return err
	}

	switch e := eventType.(type) {
	case *github.PushEvent:
		go process(e)
	default:
		return fmt.Errorf("event not supported")
	}
	return c.JSON(http.StatusAccepted, struct {
		Status string
	}{
		Status: "accepted",
	})
}
