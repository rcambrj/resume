package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
)

type GithubMarkdownBody struct {
	Text    string `json:"text"`
	Mode    string `json:"mode"`
	Context string `json:"context"`
}

func getMarkdown() ([]byte, error) {
	md, err := os.ReadFile("README.md")
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	return md, nil
}
func getMarkup(md []byte) (*string, error) {
	reqBody := GithubMarkdownBody{
		Text:    string(md),
		Mode:    "gfm",
		Context: "rcambrj/resume",
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.github.com/markdown", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("unable to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch markdown from github: %w", err)
	}
	defer res.Body.Close()
	resBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body of http response: %w", err)
	}
	resBodyString := string(resBodyBytes)
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("unexpected status code fetching from github: %d %s", res.StatusCode, resBodyString))
	}
	return &resBodyString, nil
}

func getGithubCSS() (*string, error) {
	res, err := http.Get("https://raw.githubusercontent.com/sindresorhus/github-markdown-css/main/github-markdown.css")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch github css: %w", err)
	}
	defer res.Body.Close()
	resBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body of http response: %w", err)
	}
	resBodyString := string(resBodyBytes)
	return &resBodyString, nil
}

func wrapHTML(markup string, css string) string {
	return fmt.Sprintf(
		`<!DOCTYPE html><html><head><meta charset="utf-8">
		<style>
		  %s
		  @media print {
			.markdown-body {
			  width: 980px;
			}
		  }
		  .markdown-body {
			box-sizing: border-box;
			min-width: 200px;
			max-width: 980px;
			margin: 0 auto;
			padding: 15px;
		  }
		</style>
		</head><body>
		<div class="markdown-body">%s</div>
		</body></html>`,
		css,
		markup)
}

func writeHTML(html string) error {
	err := os.WriteFile("README.html", []byte(html), 0644)
	if err != nil {
		return fmt.Errorf("unable to write html file: %w", err)
	}
	return nil
}

func writePDF(html string) error {
	page := rod.New().MustConnect().MustPage().MustSetDocumentContent(html).MustWaitLoad()
	pdf, err := page.PDF(&proto.PagePrintToPDF{})
	if err != nil {
		return err
	}
	err = utils.OutputFile("README.pdf", pdf)
	return nil
}

func main() {
	log.Print("Opening file...")
	markdown, err := getMarkdown()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Fetching Github CSS...")
	css, err := getGithubCSS()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Converting markdown to HTML...")
	markup, err := getMarkup(markdown)
	if err != nil {
		log.Fatal(err)
	}
	html := wrapHTML(*markup, *css)
	log.Print("Writing HTML to file...")
	err = writeHTML(html)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Writing PDF to file...")
	err = writePDF(html)
	if err != nil {
		log.Fatal(err)
	}
}
