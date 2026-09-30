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

type SourceStats struct {
	Generated int
	Refused   int
	Tbp       []float64
	Tobsl     []float64
}

type Simulator struct {
	CurrentTime       float64
	EndTime           float64
	Calendar          calendar.EventCalendar
	Sources           map[int]*source.Source
	Devices           map[int]*device.Device
	Buffer            *buffer.Buffer
	StagingDispatcher *dispatcher.StagingDispatcher
	FetchDispatcher   *dispatcher.FetchDispatcher
	Stats             map[int]*SourceStats
}

func NewSimulator(cfg *config.Simulator) *Simulator {
	sources := make(map[int]*source.Source)
	for i := 0; i < cfg.NumSources; i++ {
		src := source.NewSource(cfg.Sources[i])
		sources[cfg.Sources[i].Id] = src
	}

	devices := make(map[int]*device.Device)
	deviceList := make([]*device.Device, 0, cfg.NumDevices)

	for i := 0; i < cfg.NumDevices; i++ {
		dev := device.NewDevice(cfg.Devices[i])
		devices[cfg.Devices[i].Id] = dev
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

	stats := make(map[int]*SourceStats)

	return &Simulator{
		CurrentTime:       0.0,
		EndTime:           cfg.MaxSimulationTime,
		Calendar:          cal,
		Sources:           sources,
		Devices:           devices,
		Buffer:            buf,
		StagingDispatcher: staging,
		FetchDispatcher:   fetch,
		Stats:             stats,
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

	if finishedApp != nil {
		tobsl := s.CurrentTime - finishedApp.ServiceStartTime
		tbp := 0.0
		if finishedApp.BufferEntryTime >= 0 {
			tbp = finishedApp.ServiceStartTime - finishedApp.BufferEntryTime
		}

		if s.Stats[finishedApp.SourceID] == nil {
			s.Stats[finishedApp.SourceID] = &SourceStats{}
		}
		s.Stats[finishedApp.SourceID].Tobsl = append(s.Stats[finishedApp.SourceID].Tobsl, tobsl)
		s.Stats[finishedApp.SourceID].Tbp = append(s.Stats[finishedApp.SourceID].Tbp, tbp)
	}

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
	fmt.Println("\nТАБЛИЦА 1: Характеристики источников ВС")
	fmt.Printf("%-3s | %-8s | %-8s | %-8s | %-10s | %-10s | %-10s | %-8s | %-8s\n",
		"№", "Заявок", "Отказов", "P_отк", "T_преб", "T_БП", "T_обсл", "D_БП", "D_обсл")

	for i := 1; i <= len(s.Sources); i++ {
		src := s.Sources[i]
		gen, ref := src.GetStats()

		pOtk := 0.0
		if gen > 0 {
			pOtk = float64(ref) / float64(gen)
		}

		tbpAvg, tbpVar := 0.0, 0.0
		tobslAvg, tobslVar := 0.0, 0.0
		tprebAvg := 0.0

		if stats, ok := s.Stats[i]; ok && stats != nil {
			tbpAvg, tbpVar = calcMeanVar(stats.Tbp)
			tobslAvg, tobslVar = calcMeanVar(stats.Tobsl)
			tprebAvg = tbpAvg + tobslAvg
		}

		fmt.Printf("%-3d | %-8d | %-8d | %-8.4f | %-10.4f | %-10.4f | %-10.4f | %-8.4f | %-8.4f\n",
			i, gen, ref, pOtk, tprebAvg, tbpAvg, tobslAvg, tbpVar, tobslVar)
	}

	fmt.Println("\nТАБЛИЦА 2: Характеристики приборов ВС")
	fmt.Printf("%-3s | %-12s | %-15s\n", "№", "Обслужено", "K_исп")

	for i := 1; i <= len(s.Devices); i++ {
		dev := s.Devices[i]
		kIsp := 0.0
		if s.CurrentTime > 0 {
			kIsp = dev.TotalBusyTime / s.CurrentTime
		}
		if kIsp > 1.0 {
			kIsp = 1.0
		}
		fmt.Printf("%-3d | %-12d | %-15.4f\n", i, dev.ServedCount, kIsp)
	}
}

func calcMeanVar(data []float64) (mean, variance float64) {
	if len(data) == 0 {
		return 0, 0
	}
	var sum, sumSq float64
	for _, v := range data {
		sum += v
		sumSq += v * v
	}
	mean = sum / float64(len(data))
	variance = (sumSq / float64(len(data))) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	return mean, variance
}
