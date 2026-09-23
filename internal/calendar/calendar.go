package calendar

type EventType int

const (
	EventArrival EventType = iota
	EventDeparture
)

type Event struct {
	Time     float64
	Type     EventType
	SourceID int
	DeviceID int
}

type EventCalendar []*Event

func (h EventCalendar) Len() int            { return len(h) }
func (h EventCalendar) Less(i, j int) bool  { return h[i].Time < h[j].Time }
func (h EventCalendar) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *EventCalendar) Push(x interface{}) { *h = append(*h, x.(*Event)) }
func (h *EventCalendar) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
