package services

import (
	"acif-mediaserver/adapters"
	"acif-mediaserver/schemas"
	"github.com/gorilla/websocket"
)

type MediaService struct {
	Adapter *adapters.KurentoMediaServer
}

var CandidatesQueue = make(map[string][]interface{})
var Users = make(map[string]*schemas.UserSession)
var WebRtc = make(map[string]string)

func (ms *MediaService) CreatePipeline(caller, callee, userSession string) (mediaPipelineId, sessionId, callerWebRtc, calleeWebRtc, recorderId *string, err error) {
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

	var uri = new(string)
	recorderId, uri, err = ms.Adapter.CreateRecorder(*mediaPipelineId, *sessionId, userSession)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	Users[callee].RecordPath = *uri

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
	err = ms.Adapter.Connect(*calleeWebRtc, *recorderId, *sessionId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return
}

func (ms *MediaService) Stop(name string) (*string, error) {
	to := Users[name].Peer
	err := Users[to].SendMessage(struct {
		Id      string `json:"id"`
		Message string `json:"message"`
	}{
		Id:      "stopCommunication",
		Message: "remote user hanged out",
	})
	if err != nil {
		return nil, err
	}
	err = ms.Adapter.Release(Users[name].MediaPipelineId, Users[name].SessionId)
	if err != nil {
		return nil, err
	}
	Users[name].IsBusy = false
	Users[to].IsBusy = false
	delete(CandidatesQueue, name)
	delete(CandidatesQueue, to)
	delete(WebRtc, name)
	delete(WebRtc, to)

	var admin string
	if Users[to].Role == "admin" {
		admin = to
	} else {
		admin = name
	}
	return &admin, nil
}

func (ms *MediaService) IncomingCallResponse(from, to, callResponse, calleeSdp string) error {
	Users[to].SdpOffer = calleeSdp
	caller := Users[from]
	callee := Users[to]
	if callResponse == "accept" {
		mediaPipelineId, sessionId, callerWebRtc, calleeWebRtc, recorderId, err := ms.CreatePipeline(from, to, Users[to].SessionId)
		if err != nil {
			return err
		}
		WebRtc[from] = *callerWebRtc
		WebRtc[to] = *calleeWebRtc
		callerAnswer, err := ms.Adapter.GenerateSdpAnswer(*callerWebRtc, *sessionId, caller.SdpOffer)
		calleeAnswer, err := ms.Adapter.GenerateSdpAnswer(*calleeWebRtc, *sessionId, callee.SdpOffer)

		err = ms.Adapter.StartRecording(*recorderId, *calleeWebRtc, *sessionId)
		if err != nil {
			return err
		}

		Users[from].MediaPipelineId = *mediaPipelineId
		Users[from].SessionId = *sessionId
		Users[to].MediaPipelineId = *mediaPipelineId
		Users[to].SessionId = *sessionId

		var messageCallee = struct {
			Id        string `json:"id"`
			SdpAnswer string `json:"sdpAnswer"`
		}{
			Id:        "startCommunication",
			SdpAnswer: *calleeAnswer,
		}
		err = callee.SendMessage(messageCallee)
		if err != nil {
			return err
		}

		var messageCaller = struct {
			Id        string `json:"id"`
			Response  string `json:"response"`
			SdpAnswer string `json:"sdpAnswer"`
		}{
			Id:        "callResponse",
			Response:  "accepted",
			SdpAnswer: *callerAnswer,
		}
		err = caller.SendMessage(messageCaller)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MediaService) Call(from, sdpOffer string) error {
	//delete(CandidatesQueue, from)
	var to = ""
	for id, userSessionData := range Users {
		if userSessionData.Role == "admin" && !userSessionData.IsBusy {
			to = id
		}
	}

	if to == "" {
		err := Users[from].SendMessage(struct {
			Id string `json:"id"`
		}{
			Id: "allAdminsAreBusy",
		})
		return err
	}

	Users[to].IsBusy = true
	Users[from].IsBusy = true
	Users[from].SdpOffer = sdpOffer
	Users[to].Peer = from
	Users[from].Peer = to
	var message = struct {
		Id   string `json:"id"`
		From string `json:"from"`
		To   string `json:"to"`
	}{
		Id:   "incomingCall",
		From: from,
		To:   to,
	}
	err := Users[to].SendMessage(message)
	return err
}

func (ms *MediaService) Register(name string, role string, ws *websocket.Conn) error {
	Users[name] = new(schemas.UserSession)
	Users[name].Role = role
	Users[name].Ws = ws
	if err := Users[name].SendMessage(schemas.ResponseToClient{
		Id:       "registerResponse",
		Response: "accepted",
	}); err != nil {
		return err
	}
	return nil
}

func (ms *MediaService) OnIceCandidate(name string, candidate interface{}) error {
	if WebRtc, ok := WebRtc[name]; ok {
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
