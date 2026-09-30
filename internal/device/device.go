package device

import (
	"math"
	"math/rand/v2"

	"github.com/TheEmfield/Architecture_of_software_systems/internal/config"
	"github.com/TheEmfield/Architecture_of_software_systems/internal/source"
)

type Device struct {
	ID             int
	Lambda         float64
	IsBusy         bool
	CurrentApp     *source.Application
	ServiceEndTime float64
	TotalBusyTime  float64
	ServedCount    int
}

func NewDevice(cfg config.Device) *Device {
	return &Device{
		ID:             cfg.Id,
		Lambda:         cfg.Lambda,
		IsBusy:         false,
		CurrentApp:     nil,
		ServiceEndTime: 0,
		TotalBusyTime:  0,
		ServedCount:    0,
	}
}

func (d *Device) IsFree() bool {
	return !d.IsBusy
}

func (d *Device) Assign(app *source.Application, currentTime float64) float64 {
	d.IsBusy = true
	d.CurrentApp = app
	d.ServedCount++
	app.ServiceStartTime = currentTime
	serviceTime := -math.Log(rand.Float64()) / d.Lambda
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
