package dispatcher

import (
	"github.com/TheEmfield/Architecture_of_software_systems/internal/buffer"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/calendar"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/device"
)

type FetchDispatcher struct {
	Buffer *buffer.Buffer
}

func NewFetchDispatcher(buffer *buffer.Buffer) *FetchDispatcher {
	return &FetchDispatcher{
		Buffer: buffer,
	}
}

func (f *FetchDispatcher) ProcessDeparture(freedDevice *device.Device, currentTime float64) *calendar.Event {
	app, slotIndex := f.Buffer.FindHighestPriorityApp()
	if app == nil {
		return nil
	}

	f.Buffer.Remove(slotIndex)
	serviceTime := freedDevice.Assign(app, currentTime)

	return &calendar.Event{
		Time:     currentTime + serviceTime,
		Type:     1,
		DeviceID: freedDevice.ID,
	}
}
