package simulator

import (
	"container/heap"
	"fmt"
	"sort"

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
	sources := make(map[int]*source.Source)
	for i := 1; i <= cfg.NumSources; i++ {
		src := source.NewSource(i, cfg.MinInterval, cfg.MaxInterval, 0.0) //пока 0.0, потом из конфига
		sources[i] = src
	}

	devices := make(map[int]*device.Device)
	deviceList := make([]*device.Device, 0, cfg.NumDevices)

	for i := 1; i <= cfg.NumDevices; i++ {
		dev := device.NewDevice(i, cfg.Lambda)
		devices[i] = dev
		deviceList = append(deviceList, dev)
	}

	buf := buffer.NewBuffer(cfg.BufferCapacity)
	staging := dispatcher.NewStagingDispatcher(buf, deviceList)
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
		EndTime:           cfg.MaxSimulationTime,
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

	fmt.Printf("[Time: %.2f] Event: ", s.CurrentTime)

	if event.Type == calendar.EventArrival {
		fmt.Printf("ARRIVAL from Source %d\n", event.SourceID)
		s.handleArrival(event)
	} else if event.Type == calendar.EventDeparture {
		fmt.Printf("DEPARTURE from Device %d\n", event.DeviceID)
		s.handleDeparture(event)
	}

	s.PrintState()
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
		fmt.Printf("  -> REFUSED (buffer full)\n")
	} else if depEvent != nil {
		heap.Push(&s.Calendar, depEvent)
		fmt.Printf("  -> Sent to Device %d\n", depEvent.DeviceID)
	} else {
		fmt.Printf("  -> Placed in buffer\n")
	}
}

func (s *Simulator) handleDeparture(event *calendar.Event) {
	dev := s.Devices[event.DeviceID]
	finishedApp := dev.Release()
	fmt.Printf("  -> Device %d finished Application %d\n", dev.ID, finishedApp.ID)

	depEvent := s.FetchDispatcher.ProcessDeparture(dev, s.CurrentTime)
	if depEvent != nil {
		heap.Push(&s.Calendar, depEvent)
		fmt.Printf("  -> Device %d took new application from buffer\n", dev.ID)
	} else {
		fmt.Printf("  -> Device %d is now idle\n", dev.ID)
	}
}

func (s *Simulator) PrintState() {
	fmt.Println("Event Calendar:")
	if s.Calendar.Len() == 0 {
		fmt.Println("  (empty)")
	} else {
		eventsCopy := make([]*calendar.Event, len(s.Calendar))
		copy(eventsCopy, s.Calendar)

		sort.Slice(eventsCopy, func(i, j int) bool {
			return eventsCopy[i].Time < eventsCopy[j].Time
		})

		for _, ev := range eventsCopy {
			if ev.Type == calendar.EventArrival {
				fmt.Printf("  Time: %.2f | Type: ARRIVAL   | Source: %d\n", ev.Time, ev.SourceID)
			} else {
				fmt.Printf("  Time: %.2f | Type: DEPARTURE | Device: %d\n", ev.Time, ev.DeviceID)
			}
		}
	}

	fmt.Println("Sources:")
	for i := 1; i <= len(s.Sources); i++ {
		src := s.Sources[i]
		gen, ref := src.GetStats()
		fmt.Printf("  Source %d: Next at %.2f, Generated: %d, Refused: %d\n", src.GetID(), src.GetNextEventTime(), gen, ref)
	}

	fmt.Println("Buffer:")
	for i, app := range s.Buffer.Slots {
		if app != nil {
			fmt.Printf("  Slot %d: App ID %d (Source %d, Arrival %.2f)\n", i+1, app.ID, app.SourceID, app.ArrivalTime)
		} else {
			fmt.Printf("  Slot %d: Empty\n", i+1)
		}
	}

	fmt.Println("Devices:")
	for i := 1; i <= len(s.Devices); i++ {
		dev := s.Devices[i]
		if dev.IsBusy {
			fmt.Printf("  Device %d: BUSY (App ID %d, Ends at %.2f)\n", dev.ID, dev.CurrentApp.ID, dev.ServiceEndTime)
		} else {
			fmt.Printf("  Device %d: IDLE\n", dev.ID)
		}
	}
	fmt.Println("--------------------")
}

func (s *Simulator) PrintFinalStats() {
	fmt.Println("\n=== FINAL STATISTICS ===")
	for i := 1; i <= len(s.Sources); i++ {
		src := s.Sources[i]
		gen, ref := src.GetStats()
		fmt.Printf("Source %d: Generated: %d, Refused: %d\n", src.GetID(), gen, ref)
	}
	for i := 1; i <= len(s.Devices); i++ {
		dev := s.Devices[i]
		fmt.Printf("Device %d: Served: %d, Total Busy Time: %.2f\n", dev.ID, dev.ServedCount, dev.TotalBusyTime)
	}
	fmt.Println("========================")
}
