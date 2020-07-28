package stream

import (
	"errors"
	"strconv"
	"strings"

	"github.com/4406arthur/bello/utils/logger"
)

type Manager struct {
	prefix   string
	subjects chan string
	logger   logger.Logger
}

func NewManager(pre string, size int, log logger.Logger) *Manager {
	subs := make([]string, 0, size)
	for i := 0; i < size; i++ {
		subs = append(subs, pre+"-"+strconv.Itoa(i))
	}
	log.Debug("N", logger.Trace(), "all subjects: "+strings.Join(subs, ", "))
	m := Manager{
		prefix:   pre,
		subjects: make(chan string, size),
		logger:   log,
	}

	for i := range subs {
		m.subjects <- subs[i]
	}

	return &m
}

func (m *Manager) Checkout() (string, error) {
	select {
	case sub := <-m.subjects:
		m.logger.Debug("N", logger.Trace(), "get subject: "+sub)
		return sub, nil
	default:
		m.logger.Error("N", logger.Trace(), "cannot get any availabe subject")
		return "", errors.New("busy")
	}
}

func (m *Manager) Checkin(subject string) bool {
	select {
	case m.subjects <- subject:
		m.logger.Debug("N", logger.Trace(), "put back subject: "+subject)
		return true
	default:
		return false
	}
}
