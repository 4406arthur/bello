package handler

import (
	"context"
	"encoding/binary"
	"net/http"
	"time"

	"github.com/4406arthur/bello/pkg/stream"
	"github.com/4406arthur/bello/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

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

func NewStreamHandler(p *stream.Pool, m *stream.Manager, log logger.Logger) *StreamHandler {
	return &StreamHandler{
		pool:    p,
		manager: m,
		logger:  log,
	}
}

func (s *StreamHandler) Flow(ctx *gin.Context) {

	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
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

	go s.streamFromWS(c, ws, streamIn, errChan)
	go s.streamRelay(c, nc, streamIn, subName, uniqueReplyTo, errChan)
	go s.streamFromMailbox(c, mailbox, streamOut, errChan)
	//stage := 1
	for {
		select {
		case messageFromSTT := <-streamOut:
			s.logger.Debug("N", logger.Trace(), "get message form STT: "+string(messageFromSTT))
			//TODO NLU logic
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

func (s *StreamHandler) streamFromWS(ctx context.Context, ws *websocket.Conn, ch chan<- []byte, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("N", logger.Trace(), "close goroutine")
			return
		default:
			_, message, err := ws.ReadMessage()
			if err != nil {
				s.logger.Error("N", logger.Trace(), err.Error())
				if isClosed(errCh) {
					return
				}
				errCh <- err
				return
			}
			if binary.Size(message) > 0 {
				ch <- message
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
			msg, err := mailbox.NextMsg(30 * time.Second)
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
