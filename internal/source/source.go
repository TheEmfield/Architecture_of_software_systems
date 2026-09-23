package source

import "math/rand/v2"

type Application struct {
	ID          int
	SourceID    int
	ArrivalTime float64
}

func NewApplication(id, sourceID int, arrivalTime float64) *Application {
	return &Application{
		ID:          id,
		SourceID:    sourceID,
		ArrivalTime: arrivalTime,
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

func NewSource(id int, minInterval, maxInterval, startTime float64) *Source {
	return &Source{
		id:             id,
		minInterval:    minInterval,
		maxInterval:    maxInterval,
		nextEventTime:  startTime,
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
