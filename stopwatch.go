package stopwatch

import (
	"context"
	"math"
	"sync"
	"time"
)

const (
	start key = "start"
	stop  key = "stop"
)

const ctxStopwatch = "stopwatch"

type key string

func (k key) String() string {
	return string(k)
}

type record struct {
	ts      int64
	comment string
}
type entries map[key]record

type Split struct {
	Name     string
	Comment  string
	Duration time.Duration
}

type Stopwatch struct {
	Name    string
	Logger  Logger
	running bool
	keys    []key
	records entries

	rl *sync.Mutex
}
type Logger interface {
	Log(timestamp int64, key string, comment string)
}

type nopLogger struct{}

func (l *nopLogger) Log(_ int64, _, _ string) {}

func New(name string, logger Logger) *Stopwatch {
	if logger == nil {
		logger = &nopLogger{}
	}

	return &Stopwatch{
		Name:    name,
		Logger:  logger,
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
	startComment := ""
	startRecord := newRecord(startComment)
	w.records[start] = startRecord
	w.Logger.Log(startRecord.ts, start.String(), startComment)
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
	w.Logger.Log(time.Now().UTC().UnixNano(), lapKey, lapComment)
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

	stopComment := ""
	stopRecord := newRecord(stopComment)
	w.records[stop] = stopRecord
	w.Logger.Log(stopRecord.ts, stop.String(), stopComment)
	return nil
}

func (w *Stopwatch) Report() (Report, error) {
	if !w.started() {
		return Report{}, NewNotStartedErr(w)
	}

	if !w.stopped() {
		return Report{}, NewNotStoppedErr(w)
	}

	duration, err := w.calculateDuration(start, stop)
	if err != nil {
		return Report{}, err
	}

	splits := w.calculateSplits()
	rpt := Report{
		Duration: duration,
		Splits:   splits,
	}
	return rpt, nil
}

func (w *Stopwatch) calculateSplits() []Split {
	splits := make([]Split, len(w.keys)-1)
	var splitStart, splitEnd key
	for i := range splits {
		splitStart, splitEnd = w.keys[i], w.keys[i+1]
		splits[i] = w.calculateSplit(splitStart, splitEnd)
	}

	return splits
}

func (w *Stopwatch) calculateSplit(begin, end key) Split {
	rec := w.records[begin]
	dur, _ := w.calculateDuration(begin, end)
	return newSplit(begin.String(), rec.comment, dur)
}

func newSplit(splitName string, splitComment string, dur time.Duration) Split {
	return Split{
		Name:     splitName,
		Comment:  splitComment,
		Duration: dur,
	}
}
func (w *Stopwatch) calculateDuration(from, to key) (time.Duration, error) {
	fromRecord, exists := w.records[from]
	if !exists {
		return time.Duration(0), NewNonExistentKeyErr(w, from)
	}

	toRecord, exists := w.records[to]
	if !exists {
		return time.Duration(0), NewNonExistentKeyErr(w, to)
	}

	dur := math.Abs(float64(fromRecord.ts - toRecord.ts))
	return time.Duration(dur), nil
}

type Report struct {
	Duration time.Duration
	Splits   []Split
}

func newRecord(comment string) record {
	return record{
		ts:      time.Now().UTC().UnixNano(),
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


//Context Stopwatch Handling
func CtxNew(ctx context.Context, name string, logger Logger) context.Context {
	return context.WithValue(ctx, ctxStopwatch, New(name, logger))
}

func CtxStart(ctx context.Context) error {
	w, err := getStopwatchFromCtx(ctx)
	if err != nil {
		return  err
	}

	return w.Start()
}

func CtxStop(ctx context.Context) error {
	w, err := getStopwatchFromCtx(ctx)
	if err != nil {
		return err
	}

	return w.Stop()
}

func CtxLap(ctx context.Context, lapKey, lapComment string) error {
	w, err := getStopwatchFromCtx(ctx)
	if err != nil {
		return err
	}

	return w.Lap(lapKey, lapComment)
}

func CtxReport(ctx context.Context) (Report, error) {
	w, err := getStopwatchFromCtx(ctx)
	if err != nil {
		return Report{}, err
	}

	return w.Report()
}

func getStopwatchFromCtx(ctx context.Context) (*Stopwatch, error) {
	wi := ctx.Value(ctxStopwatch)
	if wi == nil {
		return nil, NewNotFoundErr()
	}

	w, ok := wi.(*Stopwatch)
	if !ok {
		return nil, NewBadValueErr(wi)
	}

	return w, nil
}