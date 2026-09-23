package simulator

import (
	"container/heap"
	"fmt"

	"github.com/TheEmfield/Architecture_of_software_systems/internal/buffer"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/calendar"
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

func NewSimulator() *Simulator {
	src1 := source.NewSource(1, 2.0, 5.0, 0.0)
	src2 := source.NewSource(2, 3.0, 6.0, 0.0)

	dev1 := device.NewDevice(1, 4.0)
	dev2 := device.NewDevice(2, 4.0)

	buf := buffer.NewBuffer(5)

	staging := dispatcher.NewStagingDispatcher(buf, []*device.Device{dev1, dev2})
	fetch := dispatcher.NewFetchDispatcher(buf)

	cal := make(calendar.EventCalendar, 0)
	heap.Init(&cal)

	heap.Push(&cal, &calendar.Event{
		Time: src1.GetNextEventTime(), Type: calendar.EventArrival, SourceID: src1.GetID(),
	})
	heap.Push(&cal, &calendar.Event{
		Time: src2.GetNextEventTime(), Type: calendar.EventArrival, SourceID: src2.GetID(),
	})

	return &Simulator{
		CurrentTime:       0.0,
		EndTime:           100.0,
		Calendar:          cal,
		Sources:           map[int]*source.Source{1: src1, 2: src2},
		Devices:           map[int]*device.Device{1: dev1, 2: dev2},
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
		src.RecordRefusal() // Д10О5: Отказ вновь пришедшей
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
