package live

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/core-tinygo"
	"github.com/autonomouskoi/core-tinygo/svc"
	trackstar "github.com/autonomouskoi/trackstar-tinygo"
)

var (
	cfgKVKey = []byte("config")
)

type Plugin struct {
	cfg     *Config
	router  core.TopicRouter
	session string
}

func New() (*Plugin, error) {
	p := &Plugin{}

	if err := p.loadConfig(); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	session, _, err := svc.CurrentTimeMillis()
	if err != nil {
		return nil, fmt.Errorf("getting current time: %w", err)
	}
	p.session = strconv.FormatInt(session, 10)

	p.router = core.TopicRouter{
		trackstar.BusTopic_TRACKSTAR_EVENT.String(): p.handleTrackstarEvent(),
		BusTopic_TRACKSTAR_LIVE_REQUEST.String():    p.handleRequest(),
		BusTopic_TRACKSTAR_LIVE_COMMAND.String():    p.handleCommand(),
		"352fad0f027de97e":                          p.handleDirect(),
	}

	for topic := range p.router {
		core.LogDebug("subscribing", "topic", topic)
		if err := core.Subscribe(topic); err != nil {
			return nil, fmt.Errorf("subscribing to topic %s: %w", topic, err)
		}
	}

	return p, nil
}

func (p *Plugin) Handle(msg *core.BusMessage) {
	p.router.Handle(msg)
}

func (p *Plugin) handleDirect() core.TypeRouter {
	return core.TypeRouter{}
}

func (p *Plugin) loadConfig() error {
	p.cfg = &Config{
		Tokens: map[string]*TokenConfig{},
	}
	core.LogError("DELETING CONFIG")
	return nil
	if err := core.KVGetProto(cfgKVKey, p.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}
	return nil
}

func (p *Plugin) writeCfg() {
	if err := core.KVSetProto(cfgKVKey, p.cfg); err != nil {
		core.LogError("writing config", "error", err.Error())
	}
}
