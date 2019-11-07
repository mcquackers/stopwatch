package stopwatch

import (
	"sync"
	"time"
)

const (
	start key = "start"
	stop  key = "stop"
)

type key string
type record struct {
	ts int64
	comment string
}
type entries map[key]record

type Stopwatch struct {
	Name    string
	running bool
	keys    []key
	records entries

	rl *sync.Mutex
}

func New(name string) *Stopwatch {
	return &Stopwatch{
		Name:    name,
		keys:    make([]key, 0),
		records: make(map[key]record),

		rl: &sync.Mutex{},
	}
}

func (w *Stopwatch) Start() error {
	if w.stopped() {
		return NewAlreadyStoppedErr(w)
	}

	if w.started() {
		return NewAlreadyStartedErr(w)
	}

	w.setRunning(true)
	w.keys = append(w.keys, start)
	w.records[start] = newRecord("")
	return nil
}

func (w *Stopwatch) Lap(lapKey, lapComment string) error {
	if w.stopped() {
		return NewAlreadyStoppedErr(w)
	}
	if !w.started() {
		return NewNotStartedErr(w)
	}
	lk := newKey(lapKey)
	w.keys = append(w.keys, key(lk))
	w.records[lk] = newRecord(lapComment)
	return nil
}

func (w *Stopwatch) Running() bool {
	w.rl.Lock()
	defer w.rl.Unlock()
	return w.running
}

func (w *Stopwatch) Stop() error {
	if w.stopped() {
		return NewAlreadyStoppedErr(w)
	}

	if !w.started() {
		return NewNotStartedErr(w)
	}

	w.setRunning(false)
	w.keys = append(w.keys, stop)
	return nil
}

func newRecord(comment string) record {
	return record{
		ts: time.Now().Unix(),
		comment: comment,
	}
}

func newKey(keyName string) key {
	return key(keyName)
}

func (w *Stopwatch) setRunning(running bool) {
	w.rl.Lock()
	defer w.rl.Unlock()
	w.running = running
}

func (w *Stopwatch) started() bool {
	return len(w.keys) > 0 && w.keys[0] == start
}

func (w *Stopwatch) stopped() bool {
	lastIdx := len(w.keys) - 1
	return len(w.keys) > 0 && w.keys[lastIdx] == stop
}
