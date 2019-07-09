package controllers

import (
	"acif-mediaserver/adapters"
	"acif-mediaserver/schemas"
	"acif-mediaserver/services"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"net/http"
	"os"
)

func Call(c echo.Context) error {
	dialer := websocket.DefaultDialer
	ws, response, err := dialer.Dial(os.Getenv("KURENTO_HOST"), nil)
	if err != nil {
		fmt.Println(response, err)
	}
	defer ws.Close()
	ms := &services.MediaService{
		Adapter: &adapters.KurentoMediaServer{
			Ws: ws,
		},
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	defer conn.Close()

	for {
		var message schemas.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		switch message.Id {
		case "register":
			ms.Register(message.Name, conn)
		case "call":
			//инициирует сеанс связи только оператор
			err = ms.Call(message.To, message.From, message.SdpOffer)
			if err != nil {
				c.Logger().Error(err)
				return err
			}
		case "incomingCallResponse":
			err = ms.IncomingCallResponse(message.From, message.To, message.CallResponse, message.SdpOffer)
			if err != nil {
				c.Logger().Error(err)
				return err
			}
		case "stop":
			//инициализирует остановку сессии только оператор
			err = ms.Stop(message.Name)
			if err != nil {
				c.Logger().Error(err)
				return err
			}
			//отправляем запрос на привязку к данным сессии ссылки на видеозапись
			err = adapters.BindVideoLinkToSession(message.SessionId, services.Users[services.Users[message.Name].Peer].RecordPath)
			if err != nil {
				c.Logger().Error(err)
				return err
			}
		case "onIceCandidate":
			err = ms.OnIceCandidate(message.From, message.Candidate)
			if err != nil {
				c.Logger().Error(err)
				return err
			}
		default:
			fmt.Println(message.Id)
		}
	}
}

