package live

import (
	"encoding/base64"
	"net/url"

	"github.com/autonomouskoi/core-tinygo"
)

func (p *Plugin) handleCommand() core.TypeRouter {
	return core.TypeRouter{
		int32(MessageTypeCommand_TOKEN_SET_REQ): p.handleCommandSetToken,
	}
}

func (p *Plugin) handleCommandSetToken(msg *core.BusMessage) *core.BusMessage {
	reply := core.DefaultReply(msg)
	var req TokenSetRequest
	if reply.Error = core.UnmarshalMessage(msg, &req); reply.Error != nil {
		return reply
	}
	if req.RawToken == "" {
		if req.Enabled == nil {
			delete(p.cfg.Tokens, req.GetLabel())
		} else {
			t, present := p.cfg.Tokens[req.Label]
			if !present {
				return reply
			}
			t.Enabled = req.GetEnabled()
		}
		p.writeCfg()
		return reply
	}

	b, err := base64.StdEncoding.DecodeString(req.RawToken)
	if err != nil {
		core.LogError("decoding token", "error", err.Error())
		reply.Error = core.InvalidTypeError(err)
		return reply
	}
	t := &Token{}
	if err := t.UnmarshalVT(b); err != nil {
		core.LogError("unmarshalling token", "length", len(b), "error", err.Error())
		reply.Error = core.InvalidTypeError(err)
		return reply
	}
	u, err := url.Parse(t.Issuer)
	if err != nil {
		core.LogError("parsing token issuer", "issuer", t.Issuer, "error", err.Error())
		reply.Error = core.InvalidTypeError(err)
		return reply
	}
	u.Path = "/u/" + t.GetSubject()
	p.cfg.Tokens[u.String()] = &TokenConfig{
		Token:   t,
		Enabled: true,
	}
	p.writeCfg()
	return reply
}
