package device

import (
	"math"
	"math/rand/v2"

	"github.com/TheEmfield/Architecture_of_software_systems/internal/source"
)

type Device struct {
	ID              int
	MeanServiceTime float64
	IsBusy          bool
	CurrentApp      *source.Application
	ServiceEndTime  float64
	TotalBusyTime   float64
	ServedCount     int
}

func NewDevice(id int, meanServiceTime float64) *Device {
	return &Device{
		ID:              id,
		MeanServiceTime: meanServiceTime,
		IsBusy:          false,
		CurrentApp:      nil,
		ServiceEndTime:  0,
		TotalBusyTime:   0,
		ServedCount:     0,
	}
}

func (d *Device) IsFree() bool {
	return !d.IsBusy
}

func (d *Device) Assign(app *source.Application, currentTime float64) float64 {
	d.IsBusy = true
	d.CurrentApp = app
	d.ServedCount++
	serviceTime := -d.MeanServiceTime * math.Log(rand.Float64())
	d.ServiceEndTime = currentTime + serviceTime
	d.TotalBusyTime += serviceTime

	return serviceTime
}

func (d *Device) Release() *source.Application {
	app := d.CurrentApp
	d.IsBusy = false
	d.CurrentApp = nil

	return app
}
