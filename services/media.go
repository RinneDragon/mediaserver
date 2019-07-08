package services

import (
	"github.com/gorilla/websocket"
	"media-server/adapters"
	"media-server/schemas"
)

type MediaService struct {
	Adapter *adapters.KurentoMediaServer
}

var CandidatesQueue = make(map[string][]interface{})
var Users = make(map[string]*schemas.UserSession)
var WebRtc = make(map[string]string)

func (ms *MediaService) CreatePipeline(caller, callee string) (mediaPipelineId, sessionId, callerWebRtc, calleeWebRtc, recorderId *string, err error) {
	mediaPipelineId, sessionId, err = ms.Adapter.CreateMediaPipeline()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	callerWebRtc, err = ms.Adapter.CreateWebRtcEndpoint(*mediaPipelineId, *sessionId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	if candidates, ok := CandidatesQueue[caller]; ok {
		for _, candidate := range candidates {
			err := ms.Adapter.AddIceCandidate(*callerWebRtc, *sessionId, candidate)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
		}
	}

	recorderId, err = ms.Adapter.CreateRecorder(*mediaPipelineId, *sessionId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	calleeWebRtc, err = ms.Adapter.CreateWebRtcEndpoint(*mediaPipelineId, *sessionId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	if candidates, ok := CandidatesQueue[callee]; ok {
		for _, candidate := range candidates {
			err := ms.Adapter.AddIceCandidate(*calleeWebRtc, *sessionId, candidate)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
		}
	}

	err = ms.Adapter.Connect(*callerWebRtc, *calleeWebRtc, *sessionId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	err = ms.Adapter.Connect(*calleeWebRtc, *callerWebRtc, *sessionId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	err = ms.Adapter.Connect(*callerWebRtc, *recorderId, *sessionId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return
}

func (ms *MediaService) Stop(name string) error {
	to := Users[name].Peer
	err := Users[to].SendMessage(struct {
		Id      string `json:"id"`
		Message string `json:"message"`
	}{
		Id: "stopCommunication",
		Message: "remote user hanged out",
	})
	if err != nil {
		return err
	}
	err = ms.Adapter.Release(Users[name].MediaPipelineId, Users[name].SessionId)
	if err != nil {
		return err
	}
	return nil
}

func (ms *MediaService) IncomingCallResponse(from, to, callResponse, calleeSdp string) error {
	Users[to].SdpOffer = calleeSdp
	caller := Users[from]
	callee := Users[to]
	if callResponse == "accept" {
		mediaPipelineId, sessionId, callerWebRtc, calleeWebRtc, recorderId, err := ms.CreatePipeline(from, to)
		if err != nil {
			return err
		}
		WebRtc[from] = *callerWebRtc
		WebRtc[to] = *calleeWebRtc
		callerAnswer, err := ms.Adapter.GenerateSdpAnswer(*callerWebRtc, *sessionId, caller.SdpOffer)
		calleeAnswer, err := ms.Adapter.GenerateSdpAnswer(*calleeWebRtc, *sessionId, callee.SdpOffer)

		err = ms.Adapter.StartRecording(*recorderId, *callerWebRtc, *sessionId)
		if err != nil {
			return err
		}

		Users[from].MediaPipelineId = *mediaPipelineId
		Users[from].SessionId = *sessionId
		Users[to].MediaPipelineId = *mediaPipelineId
		Users[to].SessionId = *sessionId

		var messageCallee = struct{
			Id string `json:"id"`
			SdpAnswer string `json:"sdpAnswer"`
		}{
			Id: "startCommunication",
			SdpAnswer: *calleeAnswer,
		}
		err = callee.SendMessage(messageCallee)
		if err != nil {
			return err
		}

		var messageCaller = struct{
			Id string `json:"id"`
			Response string `json:"response"`
			SdpAnswer string `json:"sdpAnswer"`
		}{
			Id: "callResponse",
			Response : "accepted",
			SdpAnswer: *callerAnswer,
		}
		err = caller.SendMessage(messageCaller)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MediaService) Call(to, from, sdpOffer string) error {
	delete(CandidatesQueue, from)

	Users[from].SdpOffer = sdpOffer
	Users[to].Peer = from
	Users[from].Peer = to
	var message = struct{
		Id string `json:"id"`
		From string `json:"from"`
	}{
		Id: "incomingCall",
		From: from,
	}
	err := Users[to].SendMessage(message)
	return err
}

func (ms *MediaService) Register(name string, ws *websocket.Conn) {
	Users[name] = new(schemas.UserSession)
	Users[name].Ws = ws
}

func (ms *MediaService) OnIceCandidate(name string, candidate interface{}) error {
	if  WebRtc, ok := WebRtc[name]; ok {
		err := ms.Adapter.AddIceCandidate(WebRtc, Users[name].SessionId, candidate)
		if err != nil {
			return err
		}
	} else {
		if _, ok := CandidatesQueue[name]; !ok {
			CandidatesQueue[name] = make([]interface{}, 0)
		}
		CandidatesQueue[name] = append(CandidatesQueue[name], candidate)
	}
	return nil
}
