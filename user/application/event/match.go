package event

import (
	"context"
	"time"

	"github.com/weiqiangxu/micro_project/common-config/logger"
	"github.com/weiqiangxu/micro_project/net/transport"
)

var _ transport.Server = (*MatchEvent)(nil)

type MatchEventOption func(event *MatchEvent)

type MatchEvent struct {
	Ticker *time.Ticker
	Runner func() error
}

func WithTicker(t *time.Ticker) MatchEventOption {
	return func(event *MatchEvent) {
		event.Ticker = t
	}
}

func WithMatchCronAction(f func() error) MatchEventOption {
	return func(event *MatchEvent) {
		event.Runner = f
	}
}

func NewMatchEvent(options ...MatchEventOption) transport.Server {
	e := &MatchEvent{}
	for _, o := range options {
		o(e)
	}
	return e
}

func (e *MatchEvent) Start(ctx context.Context) error {
	go func() {
		for range e.Ticker.C {
			err := e.Runner()
			if err != nil {
				logger.Errorf("match event catch error %s", err.Error())
			}
		}
	}()
	return nil
}

func (e *MatchEvent) Stop(ctx context.Context) error {
	logger.Infow("stop match event !")
	e.Ticker.Stop()
	return nil
}
