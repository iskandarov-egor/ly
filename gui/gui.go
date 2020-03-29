package gui

import (
	"fmt"
	"os"
	"bytes"
	"net/http"
	"github.com/vmihailenco/msgpack"
	"github.com/gorilla/websocket"
)

type Gui struct {}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func DecodeMessage(blob []byte) (msg Message, err error) {
	fmt.Println("got blob", blob)

	var msgType int
	decoder := msgpack.NewDecoder(bytes.NewReader(blob))
	msgType, err = decoder.DecodeInt()
	if err != nil {
		return nil, fmt.Errorf("decode message type: %v", err)
	}
	switch msgType {
		case MessageTypePing:
			var ping PingMessage
			err = decoder.Decode(&ping)
			if err != nil {
				return nil, fmt.Errorf("decode ping message: %v", err)
			}
			return ping, nil
		case MessageTypeRender:
			var msg RenderMessage
			err = decoder.Decode(&msg)
			if err != nil {
				return nil, fmt.Errorf("decode render message: %v", err)
			}
			return msg, nil
		default:
			return nil, fmt.Errorf("unk message type %v", msgType)
	}
}

func Type2Code(m Message) int {
	switch m.(type) {
		case ImageMessage: return MessageTypeImage
		case PingMessage: return MessageTypePing
		case CanvasSizeMessage: return MessageTypeCanvasSize
		case LineMessage: return MessageTypeLine
	}
	panic("wot")
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
	fmt.Println("client connected")
	go func() {
		for {
			msg := <- s.QueueOut
			var buf bytes.Buffer
			encoder := msgpack.NewEncoder(&buf)
			err := encoder.Encode(Type2Code(msg))
			if err != nil {
				fmt.Println("msgpack marshal: %v", err)
				continue
			}
			err = encoder.Encode(msg)
			if err != nil {
				fmt.Println("msgpack marshal: %v", err)
				continue
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buf.Bytes()); err != nil {
				fmt.Println(err)
			}
			//println("msg went")
		}
	}()
	go func() {
		for {
			mtype, blob, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			if mtype != websocket.BinaryMessage {
				fmt.Println("got non binary msg")
				continue
			}
			message, err := DecodeMessage(blob)
			if err != nil {
				fmt.Println("unmarshal msg: ", err)
				continue
			}
			s.QueueIn <- message
		}
	}()
}

const (
	// server commands
	MessageTypePing = 44001
	MessageTypeImage = 44002
	MessageTypeCanvasSize = 44003
	MessageTypeLine = 44005
	// client commands
	MessageTypeRender = 44004
)

type CanvasSizeMessage struct {
	W, H int
}

type ImageMessage struct {
	RGBA []byte
	X, Y int
	W, H int
}

type PingMessage struct {
	Hello string
}

type Selection struct {
	Left, Top, Right, Bottom int
}

type RenderMessage struct {
	Area *Selection
}

type LineMessage struct {
	X1, Y1 int
	X2, Y2 int
	R, G, B float32
}

type Message interface {}

type Server struct {
	QueueIn chan Message
	QueueOut chan Message
}

func NewServer() *Server {
	return &Server{
		QueueIn: make(chan Message),
		QueueOut: make(chan Message),
	}
}

func (s *Server) Serve() {
	http.HandleFunc("/", s.handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("start gui: %v", err)
		os.Exit(1)
	}
}

func (s *Server) NewCanvas() *Canvas {
	return &Canvas{Server: s}
}

func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true };
}

func main() {
	s := NewServer()
	go s.Serve()

	for {
		msgI := <-s.QueueIn
		switch msg := msgI.(type) {
			case PingMessage:
				fmt.Println("got ping message:", msg.Hello)
			default:
				fmt.Println("unk message")
		}
	}
}
