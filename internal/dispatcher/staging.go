package dispatcher

import (
	"github.com/TheEmfield/Architecture_of_software_systems/internal/buffer"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/calendar"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/device"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/source"
)

type StagingDispatcher struct {
	Buffer  *buffer.Buffer
	Devices map[int]*device.Device
}

func NewStagingDispatcher(buffer *buffer.Buffer, devices map[int]*device.Device) *StagingDispatcher {
	return &StagingDispatcher{
		Buffer:  buffer,
		Devices: devices,
	}
}

func (s *StagingDispatcher) FindFreeDevice() *device.Device {
	for _, d := range s.Devices {
		if !d.IsBusy {
			return d
		}
	}

	return nil
}

func (s *StagingDispatcher) ProcessArrival(app *source.Application, currentTime float64) (*calendar.Event, bool) {
	d := s.FindFreeDevice()
	if d != nil {
		serviceTime := d.Assign(app, currentTime)

		return &calendar.Event{
			Time:     currentTime + serviceTime,
			Type:     calendar.EventDeparture,
			DeviceID: d.ID,
		}, false
	}

	slot := s.Buffer.FindFirstFreeSlot()
	if slot != -1 {
		s.Buffer.Slots[slot] = app
		return nil, false
	}

	return nil, true
}
