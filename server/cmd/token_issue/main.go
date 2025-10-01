package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/autonomouskoi/trackstar-live/server"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(-1)
}

func fatalIfError(err error, msg string) {
	if err != nil {
		fatal("error: ", msg, ": ", err)
	}
}

func main() {
	if len(os.Args) != 3 {
		fatal("usage: ", os.Args[0], "<config path>", "<user id>")
	}

	cfg, err := server.LoadConfig(os.Args[1])
	fatalIfError(err, "loading config")

	u, err := url.Parse(cfg.MyURL)
	fatalIfError(err, "parsing server URL")
	u.Path = path.Join(u.Path, "_issue")

	form := url.Values{}
	form.Set("user_id", os.Args[2])
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	fatalIfError(err, "creating request")

	req.Header.Set("x-extension-jwt", cfg.MyKeyInput)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	fatalIfError(err, "sending request")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, resp.Status)
		io.Copy(os.Stderr, resp.Body)
		fatal("request failed")
	}
	w := base64.NewEncoder(base64.StdEncoding, os.Stdout)
	io.Copy(w, resp.Body)
	w.Close()
	fmt.Println()
}
