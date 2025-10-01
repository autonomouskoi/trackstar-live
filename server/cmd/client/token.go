package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/trackstar-live/server"
)

func loadToken(path string) (*server.Token, error) {
	protoB64, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	protoB := make([]byte, base64.RawStdEncoding.DecodedLen(len(protoB64)))
	n, err := base64.StdEncoding.Decode(protoB, protoB64)
	if err != nil {
		return nil, fmt.Errorf("decoding base64: %w", err)
	}
	protoB = protoB[:n]
	var t server.Token
	if err := proto.Unmarshal(protoB, &t); err != nil {
		return nil, fmt.Errorf("unmarshalling token: %w", err)
	}
	return &t, nil
}
