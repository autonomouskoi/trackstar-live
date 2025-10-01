package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	trackstar "github.com/autonomouskoi/trackstar/pb"
	"google.golang.org/protobuf/proto"
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
	if len(os.Args) != 6 {
		fatal("Usage: ", os.Args[0], "<token path> <session ID> <artist> <title> <idx>")
	}

	token, err := loadToken(os.Args[1])
	fatalIfError(err, "loading token")

	fmt.Printf(`
User:     %s
Issuer:   %s
Audience: %v
`,
		token.GetSubject(),
		token.GetIssuer(),
		token.GetAudience(),
	)

	now := time.Now().UnixMilli()
	idx, err := strconv.Atoi(os.Args[5])
	fatalIfError(err, "parsing index")

	tu := &trackstar.TrackUpdate{
		DeckId: "test-client",
		Track: &trackstar.Track{
			Artist: os.Args[3],
			Title:  os.Args[4],
		},
		When:  now,
		Index: int32(idx),
	}
	u, err := url.Parse(token.GetAudience()[0])
	fatalIfError(err, "parsing audience URL")
	u.Path = path.Join("_trackUpdate", token.GetSubject(), os.Args[2])

	b, err := proto.Marshal(tu)
	fatalIfError(err, "marshalling track update")

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(b))
	fatalIfError(err, "creating HTTP request")

	fmt.Println("POST -> ", u)
	req.Header.Set("Content-Type", "application/protobuf")
	req.Header.Set("x-extension-jwt", token.GetRawToken())

	resp, err := http.DefaultClient.Do(req)
	fatalIfError(err, "sending request")
	fmt.Println(resp.Status)
}
