package live

import (
	"fmt"
	"net/url"
	"path"

	"github.com/autonomouskoi/core-tinygo"
	"github.com/autonomouskoi/core-tinygo/svc"
	trackstar "github.com/autonomouskoi/trackstar-tinygo"
)

func (p *Plugin) handleTrackstarEvent() core.TypeRouter {
	return core.TypeRouter{
		int32(trackstar.MessageTypeEvent_TRACK_UPDATE): p.handleTrackstarEventTrackUpdate,
	}
}

func (p *Plugin) handleTrackstarEventTrackUpdate(msg *core.BusMessage) *core.BusMessage {
	var tu trackstar.TrackUpdate
	if err := core.UnmarshalMessage(msg, &tu); err != nil {
		return nil
	}

	for _, tCfg := range p.cfg.GetTokens() {
		if !tCfg.Enabled {
			continue
		}
		token := tCfg.GetToken()
		u, err := url.Parse(token.GetIssuer())
		if err != nil {
			sendTSLEvent(nil, fmt.Errorf("parsing issuer URL: %w", err))
			core.LogError("parsing issuer URL",
				"issuer", token.GetIssuer(),
				"error", err.Error(),
			)
			continue
		}
		u.Path = path.Join(u.Path,
			"_trackUpdate",
			token.GetSubject(),
			p.session,
		)
		core.LogDebug("sending POST", "url", u.String())
		httpReq := &svc.WebclientHTTPRequest{
			Request: &svc.HTTPRequest{
				Method: "POST",
				Url:    u.String(),
				Header: map[string]*svc.StringValues{
					"x-extension-jwt": {Values: []string{token.GetRawToken()}},
					"Content-Type":    {Values: []string{"application/protobuf"}},
				},
			},
			RequestBody: &svc.BodyDisposition{
				BodyAs: &svc.BodyDisposition_Inline{Inline: msg.GetMessage()},
			},
		}
		resp, err := svc.WebclientRequest(httpReq, 5000)
		if err != nil {
			sendTSLEvent(nil, fmt.Errorf("sending HTTP request: %w", err))
			core.LogError("sending HTTP request", "error", err.Error())
			if be, ok := err.(*core.Error); ok {
				core.LogBusError("sending HTTP request", be)
			}
			continue
		}
		if resp.StatusCode != 200 {
			sendTSLEvent(nil, fmt.Errorf("non-200 status: %s", resp.GetStatus()))
			core.LogError("non-200 status sending HTTP request", "status", resp.GetStatus())
			continue
		}
		sendTSLEvent(&tu, nil)
	}
	return nil
}

func sendTSLEvent(tu *trackstar.TrackUpdate, err error) {
	msg := &core.BusMessage{
		Topic: BusTopic_TRACKSTAR_LIVE_EVENT.String(),
		Type:  int32(MessageTypeEvent_TRACK_SEND_EVENT),
	}
	if err != nil {
		msg.Error = &core.Error{
			Code:        int32(core.CommonErrorCode_UNKNOWN),
			UserMessage: core.String(err.Error()),
		}
	} else {
		core.MarshalMessage(msg, tu)
	}
	if err := core.Send(msg); err != nil {
		core.LogError("sending live event", "error", err.Error())
	}
}
