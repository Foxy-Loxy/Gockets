package connection

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gockets/models"
	"gockets/src/callback"
	"gockets/src/channel"
	"gockets/src/tickerHelper"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if subscriberChannel, ok := channel.SubscriberChannels[vars["subscriberToken"]]; ok {
		socketConnection, _ := upgrader.Upgrade(w, r, nil)
		socketConnection.SetCloseHandler(func(code int, text string,) error {
			switch code {
			case 1001:
				log.Print("Client going away")
				break
			case 1000:
				log.Print("Regular shutdown")
				break
			default:
				log.Print("Shutdown of connection with code " + strconv.Itoa(code))
			}
			subscriberChannel.Listeners--

			return nil
		})
		socketConnection.SetPongHandler(func(appData string) error {
			log.Print("Pong handler started")
			_ = socketConnection.SetReadDeadline(tickerHelper.GetPingDeadline())
			log.Print("Set read deadline")
			return nil
		})
		subscriberChannel.Listeners++

		go pushDataToConnection(socketConnection, subscriberChannel, tickerHelper.RunTicker())
		go readDataFromConnection(socketConnection, subscriberChannel)
		go callback.HandleSentData(subscriberChannel)
	} else {
		preparedJson, _ := json.Marshal(models.Response{
			Message: "Subscriber token not found",
			Type:    "ERR",
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(preparedJson)
	}
}

func readDataFromConnection(socket *websocket.Conn, channel *models.Channel) {
	for {
		messageType, p, err := socket.ReadMessage()
		if err != nil {
			log.Println(err)
			log.Println(messageType)
			log.Println("Error caught. Removing a listener")
			channel.Listeners--
			_ = socket.Close()
			return
		}

		switch messageType {
		case websocket.TextMessage:
			log.Println("Got a text message from listener")
			channel.SubscriberMessagesChannel <- string(p)
			break
		case websocket.CloseMessage:
			log.Println("Listener closed connection")
			channel.Listeners--
			_ = socket.Close()
		}
	}
}

func pushDataToConnection(socket *websocket.Conn, channel *models.Channel, ticker *time.Ticker) {
	for {
		select {
		case data := <-channel.SubscriberChannel:
			_ = socket.WriteMessage(websocket.TextMessage, []byte(data))
			break
		case signal := <-channel.PublisherChannel:
			switch signal {
			case models.ChannelCloseSignal:
				_ = socket.Close()
				channel.ResponseChannel <- models.ChannelSignalOk
			}
			break
		case <- ticker.C:
			log.Print("Writing ping message")
			log.Print(time.Now().String())
			log.Print(tickerHelper.GetPingDeadline().String())
			_ = socket.WriteControl(websocket.PingMessage, []byte{}, tickerHelper.GetPingDeadline())
		}
		for {
			_, _, err := socket.ReadMessage()
			if err != nil {
				return
			}
		}
	}
}

func PushToConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response
	if publisherChannel, ok := channel.PublisherChannels[vars["publisherToken"]]; ok {
		if publisherChannel.SubscriberChannel == nil {
			response = models.Response{
				Message: "Subscriber has not subscribed yet",
				Type:    "INF",
			}
		} else {
			for i := 0; i < publisherChannel.Listeners; i++ {
				body, _ := ioutil.ReadAll(r.Body)
				publisherChannel.SubscriberChannel <- string(body)
			}
			response = models.Response{
				Message: "Successfully pushed data to subscriber",
				Type:    "INF",
			}
		}
	} else {
		response = models.Response{
			Message: "Publisher token not found",
			Type:    "ERR",
		}
	}

	preparedJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}

func CloseConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response

	if publisherChannel, ok := channel.PublisherChannels[vars["publisherToken"]]; ok {

		for i := 0; i < publisherChannel.Listeners; i++ {
			publisherChannel.PublisherChannel <- models.ChannelCloseSignal
		}

		delete(channel.PublisherChannels, publisherChannel.PublisherToken)
		delete(channel.SubscriberChannels, publisherChannel.SubscriberToken)

		response = models.Response{
			Message: "Successfully closed connection",
			Type:    "OK",
		}
	} else {
		response = models.Response{
			Message: "Publisher token not found",
			Type:    "ERR",
		}
	}

	preparedJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}
