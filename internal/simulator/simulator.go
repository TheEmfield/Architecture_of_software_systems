package simulator

import (
	"container/heap"
	"fmt"

	"github.com/TheEmfield/Architecture_of_software_systems/internal/buffer"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/calendar"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/config"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/device"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/dispatcher"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/source"
)

type Simulator struct {
	CurrentTime       float64
	EndTime           float64
	Calendar          calendar.EventCalendar
	Sources           map[int]*source.Source
	Devices           map[int]*device.Device
	Buffer            *buffer.Buffer
	StagingDispatcher *dispatcher.StagingDispatcher
	FetchDispatcher   *dispatcher.FetchDispatcher
}

func NewSimulator(cfg *config.Simulator) *Simulator {
	var sources map[int]*source.Source
	for i := range cfg.NumSources {
		sources[i] = source.NewSource(i, cfg.MinInterval, cfg.MaxInterval, 0.0) //пока 0.0, дальше с определенного периода времени
	}

	var devices map[int]*device.Device
	for i := range cfg.NumSources {
		devices[i] = device.NewDevice(i, cfg.MeanServiceTime)
	}

	buf := buffer.NewBuffer(cfg.BufferCapacity)

	staging := dispatcher.NewStagingDispatcher(buf, devices)
	fetch := dispatcher.NewFetchDispatcher(buf)

	cal := make(calendar.EventCalendar, 0)
	heap.Init(&cal)

	for _, src := range sources {
		heap.Push(&cal, &calendar.Event{
			Time: src.GetNextEventTime(), Type: calendar.EventArrival, SourceID: src.GetID(),
		})
	}

	return &Simulator{
		CurrentTime:       0.0,
		EndTime:           100.0,
		Calendar:          cal,
		Sources:           sources,
		Devices:           devices,
		Buffer:            buf,
		StagingDispatcher: staging,
		FetchDispatcher:   fetch,
	}
}

func (s *Simulator) Step() bool {
	if s.Calendar.Len() == 0 || s.CurrentTime >= s.EndTime {
		return false
	}

	event := heap.Pop(&s.Calendar).(*calendar.Event)
	s.CurrentTime = event.Time

	fmt.Printf("\n[Время: %.2f] Событие: ", s.CurrentTime)

	if event.Type == calendar.EventArrival {
		fmt.Printf("ПРИХОД заявки от источника %d\n", event.SourceID)
		s.handleArrival(event)
	} else if event.Type == calendar.EventDeparture {
		fmt.Printf("ОСВОБОЖДЕНИЕ прибора %d\n", event.DeviceID)
		s.handleDeparture(event)
	}

	return true
}

func (s *Simulator) handleArrival(event *calendar.Event) {
	src := s.Sources[event.SourceID]

	app := src.GetNextApplication()
	heap.Push(&s.Calendar, &calendar.Event{
		Time: src.GetNextEventTime(), Type: calendar.EventArrival, SourceID: src.GetID(),
	})

	depEvent, isRefused := s.StagingDispatcher.ProcessArrival(app, s.CurrentTime)

	if isRefused {
		src.RecordRefusal()
		fmt.Printf("  -> ОТКАЗ заявке %d (буфер полон)\n", app.ID)
	} else if depEvent != nil {
		heap.Push(&s.Calendar, depEvent)
		fmt.Printf("  -> Заявка %d пошла сразу на прибор %d\n", app.ID, depEvent.DeviceID)
	} else {
		fmt.Printf("  -> Заявка %d помещена в буфер\n", app.ID)
	}
}

func (s *Simulator) handleDeparture(event *calendar.Event) {
	dev := s.Devices[event.DeviceID]

	finishedApp := dev.Release()
	fmt.Printf("  -> Прибор %d завершил работу над заявкой %d\n", dev.ID, finishedApp.ID)

	depEvent := s.FetchDispatcher.ProcessDeparture(dev, s.CurrentTime)

	if depEvent != nil {
		heap.Push(&s.Calendar, depEvent)
		fmt.Printf("  -> Прибор %d сразу взял новую заявку из буфера\n", dev.ID)
	} else {
		fmt.Printf("  -> Прибор %d перешел в режим простоя (буфер пуст)\n", dev.ID)
	}
}

func (s *Simulator) PrintStats() {
	fmt.Println("\n=== СТАТИСТИКА ===")
	for _, src := range s.Sources {
		gen, ref := src.GetStats()
		fmt.Printf("Источник %d: Сгенерировано: %d, Отказов: %d\n", src.GetID(), gen, ref)
	}
	for _, dev := range s.Devices {
		fmt.Printf("Прибор %d: Обслужено заявок: %d, Общее время работы: %.2f\n", dev.ID, dev.ServedCount, dev.TotalBusyTime)
	}
}
