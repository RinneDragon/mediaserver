package adapters

import (
	"acif-mediaserver/schemas"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
)

const (
	RECORDING_PATH = "file:///tmp/record-%v.webm"
)

type KurentoMediaServer struct {
	Ws *websocket.Conn
}

func (kms *KurentoMediaServer) GenerateSdpAnswer(webRtcEndpointId, sessionId, sdpOffer string) (*string, error) {
	answer, err := kms.CreateOffer(webRtcEndpointId, sessionId, sdpOffer)
	if err != nil {
		return nil, err
	}
	err = kms.GatherCandidates(webRtcEndpointId, sessionId)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (kms *KurentoMediaServer) Release(objectId, sessionId string) error {
	var id = rand.Int()
	err := kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "release",
		Params: struct {
			Object    string `json:"object"`
			SessionId string `json:"sessionId"`
		}{
			Object:    objectId,
			SessionId: sessionId,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return nil
}

func (kms *KurentoMediaServer) GatherCandidates(webRtcEndpointId, sessionId string) error {
	var id = rand.Int()
	err := kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "invoke",
		Params: struct {
			Object          string `json:"object"`
			Operation       string `json:"operation"`
			OperationParams struct {
				Sink interface{} `json:"sink"`
			} `json:"operationParams"`
			SessionId string `json:"sessionId"`
		}{
			Object:    webRtcEndpointId,
			Operation: "gatherCandidates",
			SessionId: sessionId,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return nil
}

func (kms *KurentoMediaServer) AddIceCandidate(webRtcEndpointId, sessionId string, candidate interface{}) error {
	var id = rand.Int()
	err := kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "invoke",
		Params: struct {
			Object          string `json:"object"`
			Operation       string `json:"operation"`
			OperationParams struct {
				Sink      interface{} `json:"sink"`
				Candidate interface{} `json:"candidate"`
			} `json:"operationParams"`
			SessionId string `json:"sessionId"`
		}{
			Object:    webRtcEndpointId,
			Operation: "addIceCandidate",
			OperationParams: struct {
				Sink      interface{} `json:"sink"`
				Candidate interface{} `json:"candidate"`
			}{
				Sink:      webRtcEndpointId,
				Candidate: candidate,
			},
			SessionId: sessionId,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return nil
}

func (kms *KurentoMediaServer) CreateMediaPipeline() (mediaPipelineId, sessionId *string, err error) {
	var id = rand.Int()
	err = kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "create",
		Params: struct {
			Type              string      `json:"type"`
			ConstructorParams interface{} `json:"constructorParams"`
			Properties        interface{} `json:"properties"`
		}{
			Type:              "MediaPipeline",
			ConstructorParams: nil,
			Properties:        nil,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return nil, nil, err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return nil, nil, err
	}
	if response.Error != nil {
		return nil, nil,
			errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return &response.Result.Value, &response.Result.SessionId, nil
}

func (kms *KurentoMediaServer) CreateWebRtcEndpoint(mediaPipelineId, sessionId string) (webRtcEndpointId *string, err error) {
	var id = rand.Int()
	err = kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "create",
		Params: struct {
			Type              string      `json:"type"`
			ConstructorParams interface{} `json:"constructorParams"`
			Properties        interface{} `json:"properties"`
			SessionId         string      `json:"sessionId"`
		}{
			Type: "WebRtcEndpoint",
			ConstructorParams: struct {
				MediaPipeline string `json:"mediaPipeline"`
			}{
				MediaPipeline: mediaPipelineId,
			},
			Properties: nil,
			SessionId:  sessionId,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return nil, err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return &response.Result.Value, nil
}

func (kms *KurentoMediaServer) Connect(firstMediaElement, secondMediaElement, sessionId string) (err error) {
	var id = rand.Int()
	err = kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "invoke",
		Params: struct {
			Object          string `json:"object"`
			Operation       string `json:"operation"`
			OperationParams struct {
				Sink string `json:"sink"`
			} `json:"operationParams"`
			SessionId string `json:"sessionId"`
		}{
			Object:    firstMediaElement,
			Operation: "connect",
			OperationParams: struct {
				Sink string `json:"sink"`
			}{
				Sink: secondMediaElement,
			},
			SessionId: sessionId,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return nil
}

func (kms *KurentoMediaServer) CreateOffer(webRtcEndpointId, sessionId, SDP string) (SDPAnswer *string, err error) {
	var id = rand.Int()
	err = kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "invoke",
		Params: struct {
			Object          string `json:"object"`
			Operation       string `json:"operation"`
			OperationParams struct {
				Offer string `json:"offer"` //SDP Offer
			} `json:"operationParams"`
			SessionId string `json:"sessionId"`
		}{
			Object:    webRtcEndpointId,
			Operation: "processOffer",
			OperationParams: struct {
				Offer string `json:"offer"`
			}{Offer: SDP},
			SessionId: sessionId,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return nil, err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return &response.Result.Value, nil
}

func (kms *KurentoMediaServer) CreateRecorder(mediaPipelineId, sessionId string) (recorderEndpoint *string, err error) {
	var id = rand.Int()
	err = kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "create",
		Params: struct {
			Type              string      `json:"type"`
			ConstructorParams interface{} `json:"constructorParams"`
			Properties        interface{} `json:"properties"`
			SessionId         string      `json:"sessionId"`
		}{
			Type: "RecorderEndpoint",
			ConstructorParams: struct {
				MediaPipeline string `json:"mediaPipeline"`
				URI           string `json:"uri"`
			}{
				MediaPipeline: mediaPipelineId,
				URI:           fmt.Sprintf(RECORDING_PATH, id),
			},
			Properties: nil,
			SessionId:  sessionId,
		},
		Jsonrpc: "2.0",
	})
	if err != nil {
		return nil, err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return &response.Result.Value, nil
}

func (kms *KurentoMediaServer) StartRecording(recordEndpointId, webRtcEndpointId, sessionId string) (err error) {
	var id = rand.Int()
	err = kms.Ws.WriteJSON(schemas.Request{
		Id:     id,
		Method: "invoke",
		Params: struct {
			Object          string `json:"object"`
			Operation       string `json:"operation"`
			OperationParams struct {
				Sink string `json:"sink"`
			} `json:"operationParams"`
			SessionId string `json:"sessionId"`
		}{
			Object:    recordEndpointId,
			Operation: "record",
			OperationParams: struct {
				Sink string `json:"sink"`
			}{
				Sink: webRtcEndpointId,
			},
			SessionId: sessionId,
		},
		Jsonrpc: "2.0",
	})

	if err != nil {
		return err
	}

	var response schemas.Response
	err = kms.Ws.ReadJSON(&response)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return errors.New(fmt.Sprintf("%s: %s", response.Error.Message, response.Error.Data))
	}

	return nil
}
