package live

import "github.com/autonomouskoi/core-tinygo"

func (p *Plugin) handleRequest() core.TypeRouter {
	return core.TypeRouter{
		int32(MessageTypeRequest_CONFIG_GET_REQ): p.handleRequestGetConfig,
	}
}

func (p *Plugin) handleRequestGetConfig(msg *core.BusMessage) *core.BusMessage {
	reply := core.DefaultReply(msg)
	cfg := p.cfg.CloneVT()
	for _, tCfg := range cfg.Tokens {
		if tCfg.Token != nil {
			tCfg.Token.RawToken = ""
		}
	}
	core.MarshalMessage(reply, &GetConfigResponse{
		Config: cfg,
	})
	return reply
}
