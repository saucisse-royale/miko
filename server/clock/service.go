// Provides functions to manage time, relative to the in-game world.
package clock

import (
	"git.emersion.fr/saucisse-royale/miko.git/server/message"
	"time"
)

// The duration of a single tick.
const TickDuration = time.Millisecond * 20

type Service struct {
	ticks message.AbsoluteTick
}

func (s *Service) Tick() {
	s.ticks++
}

func (s *Service) GetAbsoluteTick() message.AbsoluteTick {
	return s.ticks
}

func (s *Service) GetRelativeTick() message.Tick {
	return s.ToRelativeTick(s.ticks)
}

func (s *Service) ToRelativeTick(at message.AbsoluteTick) message.Tick {
	return message.Tick(at)
}

func (s *Service) ToAbsoluteTick(rt message.Tick) message.AbsoluteTick {
	current := s.GetRelativeTick()

	at := message.AbsoluteTick(rt) + message.AbsoluteTick(s.ticks-s.ticks%message.MaxTick)
	if current < rt {
		at -= message.AbsoluteTick(message.MaxTick)
	}

	return at
}

func (s *Service) Sync(t message.Tick) {
	s.ticks = 0
	s.ticks = s.ToAbsoluteTick(t)
}

func NewService() *Service {
	return &Service{}
}
