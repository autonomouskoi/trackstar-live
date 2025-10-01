package main

import (
	"github.com/extism/go-pdk"

	"github.com/autonomouskoi/core-tinygo"
	plugin "github.com/autonomouskoi/trackstar-live/go"
)

var (
	p *plugin.Plugin
)

//go:export start
func Initialize() int32 {
	core.LogDebug("starting up")

	var err error
	p, err = plugin.New()
	if err != nil {
		core.LogError("creating rosco", "error", err.Error())
		core.Exit(err.Error())
		return -1
	}

	core.LogInfo("ready")

	return 0
}

//go:export recv
func Recv() int32 {
	msg := &core.BusMessage{}
	if err := msg.UnmarshalVT(pdk.Input()); err != nil {
		core.LogError("unmarshalling message", "error", err.Error())
		return 0
	}
	p.Handle(msg)
	return 0
}

func main() {}
