package buffer

import "github.com/TheEmfield/Architecture_of_software_systems/internal/source"

type Buffer struct {
	Capacity int
	Slots    []*source.Application
}

func NewBuffer(capacity int) *Buffer {
	return &Buffer{
		Capacity: capacity,
		Slots:    make([]*source.Application, capacity),
	}
}

func (b *Buffer) IsFull() bool {
	for _, app := range b.Slots {
		if app == nil {
			return false
		}
	}

	return true
}

func (b *Buffer) IsEmpty() bool {
	for _, app := range b.Slots {
		if app != nil {
			return false
		}
	}

	return true
}

func (b *Buffer) FindFirstFreeSlot() int {
	for id, app := range b.Slots {
		if app == nil {
			return id
		}
	}

	return -1
}

func (b *Buffer) Add(app *source.Application, slotIndex int) {
	b.Slots[slotIndex] = app
}

func (b *Buffer) Remove(slotIndex int) *source.Application {
	app := b.Slots[slotIndex]
	b.Slots[slotIndex] = nil

	return app
}

func (b *Buffer) FindHighestPriorityApp() (*source.Application, int) {
	resId := -1
	var a *source.Application
	for id, app := range b.Slots {
		if app != nil {
			if a == nil {
				a = app
			} else {
				if app.SourceID < a.SourceID {
					a = app
					resId = id
				} else if app.SourceID == a.SourceID && app.ArrivalTime < a.ArrivalTime {
					a = app
					resId = id
				}
			}
		}
	}

	return a, resId
}
