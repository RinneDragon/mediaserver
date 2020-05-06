package schemas

import "github.com/gorilla/websocket"

type UserSession struct {
	MediaPipelineId string
	SessionId       string
	RecordPath      string
	SdpOffer        string
	Peer            string
	Ws              *websocket.Conn
}

type Message struct {
	Id           string      `json:"id"`
	From         string      `json:"from"`
	To           string      `json:"to"`
	SdpOffer     string      `json:"sdpOffer"`
	Candidate    interface{} `json:"candidate"`
	Name         string      `json:"name"`
	CallResponse string      `json:"callResponse"`
	SessionId    string      `json:"sessionId"`
}

type ResponseToClient struct {
	Id       string `json:"id"`
	Response string `json:"response"`
}

func (u *UserSession) SendMessage(message interface{}) (err error) {
	err = u.Ws.WriteJSON(message)
	return
}
