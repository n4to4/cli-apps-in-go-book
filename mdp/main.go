package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	header = `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
	  <title>Markdown Preview Tool</title>
    </meta>
  </head>
  <body>
`
	footer = `
  </body>
</html>
`
)

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, os.Stderr, *skipPreview); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func run(filename string, out io.Writer, skipPreview bool) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData := parseContent(input)

	temp, err := os.CreateTemp("", "mdp*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	outName := temp.Name()
	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if skipPreview {
		return nil
	}

	return preview(outName)
}

func parseContent(input []byte) []byte {
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	var buffer bytes.Buffer
	buffer.WriteString(header)
	buffer.Write(body)
	buffer.WriteString(footer)

	return buffer.Bytes()
}

func saveHTML(outFname string, data []byte) error {
	return os.WriteFile(outFname, data, 0644)
}

func preview(fname string) error {
	cName := "open"
	cParams := []string{fname}

	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	return exec.Command(cPath, cParams...).Run()
}
