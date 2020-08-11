package handler

import (
	"context"
	"encoding/binary"
	"net/http"
	"time"

	"github.com/4406arthur/bello/pkg/entity"
	"github.com/4406arthur/bello/pkg/stream"
	"github.com/4406arthur/bello/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/looplab/fsm"
	"github.com/nats-io/nats.go"
	"github.com/pquerna/ffjson/ffjson"
)

//StreamHandler ...
type StreamHandler struct {
	pool    *stream.Pool
	manager *stream.Manager
	logger  logger.Logger
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//NewStreamHandler ...
func NewStreamHandler(p *stream.Pool, m *stream.Manager, log logger.Logger) *StreamHandler {
	return &StreamHandler{
		pool:    p,
		manager: m,
		logger:  log,
	}
}

//Flow ...
func (s *StreamHandler) Flow(ctx *gin.Context) {

	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	//define mrcp websocket fsm
	FSM := fsm.NewFSM(
		"open",
		fsm.Events{
			{Name: "start", Src: []string{"open"}, Dst: "listen"},
			{Name: "stop", Src: []string{"listen"}, Dst: "open"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				s.logger.Info("N", logger.Trace(), "enter: "+e.Dst+" from: "+e.Src)
			},
		},
	)

	subName, err := s.manager.Checkout()
	if err != nil {
		ctx.AbortWithStatus(429)
		return
	}
	nc, _ := s.pool.Get()
	// Create a unique subject name for replies.
	uniqueReplyTo := nats.NewInbox()
	// Listen for response
	mailbox, err := nc.SubscribeSync(uniqueReplyTo)
	if err != nil {
		s.logger.Error("N", logger.Trace(), err.Error())
	}
	c, cancel := context.WithCancel(ctx)
	streamIn := make(chan []byte, 30)
	streamOut := make(chan []byte, 30)
	errChan := make(chan error, 3)

	//resource release
	defer func() {
		// time.Sleep(100 * time.Millisecond)
		cancel()
		ws.Close()
		s.pool.Put(nc)
		s.manager.Checkin(subName)
		mailbox.Unsubscribe()
		close(streamIn)
		close(streamOut)
		close(errChan)
	}()

	//send a listening cmd to MRCP
	ListenAction := entity.Response{
		ErrCode: 0,
		State:   "listening",
	}
	ws.WriteJSON(ListenAction)

	go s.streamFromWS(c, FSM, ws, streamIn, errChan)
	go s.streamRelay(c, nc, streamIn, subName, uniqueReplyTo, errChan)
	go s.streamFromMailbox(c, mailbox, streamOut, errChan)

	for {
		select {
		case messageFromSTT := <-streamOut:
			s.logger.Debug("N", logger.Trace(), "get message form STT: "+string(messageFromSTT))
			//TODO NLU logic
			ws.WriteMessage(1, messageFromSTT)
		case err := <-errChan:
			s.logger.Error("N", logger.Trace(), "catch error"+err.Error())
			return
		case <-ctx.Done():
			s.logger.Error("N", logger.Trace(), "catch client request done event")
			return
		case <-time.After(60 * time.Second):
			s.logger.Error("N", logger.Trace(), "timeout")
			return
		}
	}
}

func (s *StreamHandler) streamFromWS(ctx context.Context, FSM *fsm.FSM, ws *websocket.Conn, ch chan<- []byte, errCh chan error) {
	action := &entity.Action{}
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("N", logger.Trace(), "close goroutine")
			return
		default:
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				s.logger.Error("N", logger.Trace(), err.Error())
				if isClosed(errCh) {
					return
				}
				errCh <- err
				return
			}
			if binary.Size(msg) == 0 {
				break
			}

			switch FSM.Current() {
			case "open":
				if msgType == 1 {
					//json decode msg
					ffjson.Unmarshal(msg, &action)
					s.logger.Debug("N", logger.Trace(), "ASHD:"+action.Action)
					if action.Action == entity.ActionStart {
						//ws.WriteMessage(1, []byte("lets rock"))
						FSM.Event("start")
					}
				}
			case "listen":
				if msgType == 1 {
					//json decode msg
					ffjson.Unmarshal(msg, &action)
					if action.Action == entity.ActionStop {
						ws.WriteMessage(1, []byte("bye"))
						FSM.Event("stop")
					}
				}
				if msgType == 2 {
					ch <- msg
				}
			}
		}
	}
}

func (s *StreamHandler) streamRelay(ctx context.Context, nc *nats.Conn, ch <-chan []byte, subject string, replyTo string, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("N", logger.Trace(), "close goroutine")
			return
		default:
			message, ok := <-ch
			//if channel aleady close
			if !ok {
				return
			}
			if err := nc.PublishRequest(subject, replyTo, message); err != nil {
				s.logger.Error("N", logger.Trace(), err.Error())
				if isClosed(errCh) {
					return
				}
				errCh <- err
				return
			}
		}
	}
}

func (s *StreamHandler) streamFromMailbox(ctx context.Context, mailbox *nats.Subscription, ch chan<- []byte, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("N", logger.Trace(), "close goroutine")
			return
		default:
			msg, err := mailbox.NextMsg(60 * time.Second)
			if err != nil {
				s.logger.Error("N", logger.Trace(), err.Error())
				if isClosed(errCh) {
					return
				}
				errCh <- err
				return
			}
			ch <- msg.Data
		}
	}
}

func isClosed(ch chan error) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
