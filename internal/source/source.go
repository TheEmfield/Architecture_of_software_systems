package source

import (
	"math/rand/v2"

	"github.com/TheEmfield/Architecture_of_software_systems/internal/config"
)

type Application struct {
	ID               int
	SourceID         int
	ArrivalTime      float64
	BufferEntryTime  float64
	ServiceStartTime float64
}

func NewApplication(id, sourceID int, arrivalTime float64) *Application {
	return &Application{
		ID:              id,
		SourceID:        sourceID,
		ArrivalTime:     arrivalTime,
		BufferEntryTime: -1.0,
	}
}

type Source struct {
	id             int
	minInterval    float64
	maxInterval    float64
	nextEventTime  float64
	generatedCount int
	refusedCount   int
}

func NewSource(cfg config.Source) *Source {
	return &Source{
		id:             cfg.Id,
		minInterval:    cfg.MinInterval,
		maxInterval:    cfg.MaxInterval,
		nextEventTime:  cfg.StartTime,
		generatedCount: 0,
		refusedCount:   0,
	}
}

func (s *Source) GenerateNextInterval() float64 {
	return s.minInterval + rand.Float64()*(s.maxInterval-s.minInterval)
}

func (s *Source) GetNextApplication() *Application {
	app := NewApplication(s.generatedCount, s.id, s.nextEventTime)
	s.generatedCount++

	s.nextEventTime += s.GenerateNextInterval()

	return app
}

func (s *Source) GetNextEventTime() float64 {
	return s.nextEventTime
}

func (s *Source) GetID() int {
	return s.id
}

func (s *Source) RecordRefusal() {
	s.refusedCount++
}

func (s *Source) GetStats() (generated int, refused int) {
	return s.generatedCount, s.refusedCount
}
