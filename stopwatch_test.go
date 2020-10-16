package stopwatch

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

const nanoTsDelta = 5000

func EqualLog(t *testing.T, expected testLog, actual testLog) {
	assert.Equal(t, expected.key, actual.key)
	assert.Equal(t, expected.comment, actual.comment)
	assert.InDelta(t, expected.ts, actual.ts, nanoTsDelta)
}

func AssertEqualRecord(t *testing.T, expected record, actual record) {
	assert.Equal(t, expected.comment, actual.comment)
	assert.InDelta(t, expected.ts, actual.ts, nanoTsDelta)
}

type testLogger struct {
	logs []testLog
}

type testLog struct {
	ts      int64
	key     string
	comment string
}

func (l *testLogger) Log(ts int64, key string, comment string) {
	l.logs = append(l.logs, testLog{
		ts:      ts,
		key:     key,
		comment: comment,
	})
}

func TestStopwatchCreation(t *testing.T) {
	scs := new(creationSuite)
	suite.Run(t, scs)
}

type creationSuite struct {
	stopwatchName string
	testLogger    *testLogger
	suite.Suite
}

func (cs *creationSuite) SetupTest() {
	cs.stopwatchName = "testWatch"
	cs.testLogger = &testLogger{}
}

func (cs *creationSuite) TestNew_WithNilLogger() {
	w := New(cs.stopwatchName, nil)
	assert.NotNil(cs.T(), w)
	assert.IsType(cs.T(), &Stopwatch{}, w)

	assert.Equal(cs.T(), cs.stopwatchName, w.Name)
	assert.Equal(cs.T(), &nopLogger{}, w.Logger)
	assert.NotNil(cs.T(), w.keys)
	assert.NotNil(cs.T(), w.records)
	assert.False(cs.T(), w.running)
	assert.NotNil(cs.T(), w.rl)
}

func (cs *creationSuite) TestNew_WithTestLogger() {
	w := New(cs.stopwatchName, cs.testLogger)
	assert.NotNil(cs.T(), w)
	assert.IsType(cs.T(), &Stopwatch{}, w)

	assert.IsType(cs.T(), &testLogger{}, w.Logger)
}

//Start
//DoubleStart
//StartAfterStop
func TestStopwatchStart(t *testing.T) {
	ss := new(startSuite)
	suite.Run(t, ss)
}

type startSuite struct {
	w      *Stopwatch
	logger *testLogger
	suite.Suite
}

func (ss *startSuite) SetupTest() {
	ss.logger = &testLogger{}
	ss.w = New("test", ss.logger)
}

func (ss *startSuite) TestStart_Success() {
	expectedStartTime := time.Now().UnixNano()
	err := ss.w.Start()
	assert.Nil(ss.T(), err)
	assert.True(ss.T(), ss.w.Running())

	//confirm start adds appropriate key
	assert.Equal(ss.T(), 1, len(ss.w.keys))
	assert.Equal(ss.T(), start, ss.w.keys[0])
	assert.Equal(ss.T(), 1, len(ss.w.records))

	//confirm start adds appropriate record
	startRecord, exists := ss.w.records[start]
	assert.True(ss.T(), exists)
	assert.NotZero(ss.T(), startRecord)
	expectedRecord := record{
		ts:      expectedStartTime,
		comment: "",
	}
	AssertEqualRecord(ss.T(), expectedRecord, startRecord)

	//confirm log
	expectedNumOfLogs := 1
	assert.Equal(ss.T(), expectedNumOfLogs, len(ss.logger.logs))
	expectedLog := testLog{
		ts:      expectedStartTime,
		key:     string(start),
		comment: "",
	}
	logToCheck := ss.logger.logs[0]
	EqualLog(ss.T(), expectedLog, logToCheck)
}

func (ss *startSuite) TestStart_Error_AlreadyStarted() {
	err := ss.w.Start()
	assert.Nil(ss.T(), err)
	err = ss.w.Start()
	assert.NotNil(ss.T(), err)
	assert.IsType(ss.T(), NewAlreadyStartedErr(ss.w).Error(), err.Error())
}

func (ss *startSuite) TestStart_Error_AlreadyStopped() {
	err := ss.w.Start()
	assert.Nil(ss.T(), err)
	err = ss.w.Stop()
	assert.Nil(ss.T(), err)
	err = ss.w.Start()
	assert.NotNil(ss.T(), err)
	assert.Equal(ss.T(), NewAlreadyStoppedErr(ss.w).Error(), err.Error())
}

//Success
//Error_NotStarted
//Error_AlreadyStopped
func TestStopwatchStop(t *testing.T) {
	st := new(stopSuite)
	suite.Run(t, st)
}

type stopSuite struct {
	w          *Stopwatch
	testLogger *testLogger
	suite.Suite
}

func (st *stopSuite) SetupTest() {
	st.testLogger = &testLogger{}
	st.w = New("test", st.testLogger)
}

func (st *stopSuite) TestStop_Success() {
	err := st.w.Start()
	assert.Nil(st.T(), err)
	expectedStopTs := time.Now().UTC().UnixNano()
	err = st.w.Stop()
	assert.Nil(st.T(), err)
	expectedLenOfKeys := 2
	assert.Equal(st.T(), expectedLenOfKeys, len(st.w.keys))
	assert.Equal(st.T(), stop, st.w.keys[1])
	expectedLenOfRecords := 2
	assert.Equal(st.T(), expectedLenOfRecords, len(st.w.records))
	expectedRecord := record{
		ts:      expectedStopTs,
		comment: "",
	}
	stopRecord := st.w.records[stop]
	AssertEqualRecord(st.T(), expectedRecord, stopRecord)

	expectedNumOfLogs := 2
	assert.Equal(st.T(), expectedNumOfLogs, len(st.testLogger.logs))
	expectedLog := testLog{
		ts:      expectedStopTs,
		key:     string(stop),
		comment: "",
	}
	EqualLog(st.T(), expectedLog, st.testLogger.logs[1])
}

func (st *stopSuite) TestStop_Error_NotStarted() {
	err := st.w.Stop()
	assert.NotNil(st.T(), err)
	assert.Equal(st.T(), NewNotStartedErr(st.w).Error(), err.Error())

	expectedNumOfLogs := 0
	assert.Equal(st.T(), expectedNumOfLogs, len(st.testLogger.logs))
}

func (st *stopSuite) TestStop_Error_AlreadyStopped() {
	err := st.w.Start()
	assert.Nil(st.T(), err)
	err = st.w.Stop()
	assert.Nil(st.T(), err)
	err = st.w.Stop()
	assert.NotNil(st.T(), err)
	assert.Equal(st.T(), NewAlreadyStoppedErr(st.w).Error(), err.Error())

	expectedNumOfLogs := 2
	assert.Equal(st.T(), expectedNumOfLogs, len(st.testLogger.logs))
}

//Success
//Error_NotStarted
//Error_AlreadyStopped
func TestLap(t *testing.T) {
	ls := new(lapSuite)
	suite.Run(t, ls)
}

type lapSuite struct {
	w      *Stopwatch
	logger *testLogger

	key     string
	comment string
	suite.Suite
}

func (ls *lapSuite) SetupTest() {
	ls.logger = &testLogger{}
	ls.w = New("test", ls.logger)
	ls.key = "key"
	ls.comment = "lap-comment"
}

func (ls *lapSuite) TestLap_Success() {
	err := ls.w.Start()
	assert.Nil(ls.T(), err)
	//basic return check
	expectedTs := time.Now().UnixNano()
	err = ls.w.Lap(ls.key, ls.comment)
	assert.Nil(ls.T(), err)

	//internal function check
	expectedNumOfKeys := 2
	assert.Equal(ls.T(), expectedNumOfKeys, len(ls.w.keys))
	assert.Equal(ls.T(), ls.w.keys[1], newKey(ls.key))
	assert.Equal(ls.T(), expectedNumOfKeys, len(ls.w.records))

	//integrity check
	lapRecord, exists := ls.w.records[newKey(ls.key)]
	assert.True(ls.T(), exists)
	assert.NotZero(ls.T(), lapRecord)
	assert.InDelta(ls.T(), expectedTs, lapRecord.ts, nanoTsDelta)
	assert.Equal(ls.T(), ls.comment, lapRecord.comment)

	//log check
	expectedNumOfLogs := 2
	assert.Equal(ls.T(), expectedNumOfLogs, len(ls.logger.logs))
	expectedLapLog := testLog{
		ts:      lapRecord.ts,
		key:     ls.key,
		comment: lapRecord.comment,
	}
	assert.Equal(ls.T(), expectedLapLog, ls.logger.logs[1])
}

func (ls *lapSuite) TestLap_Error_NotStarted() {
	err := ls.w.Lap(ls.key, ls.comment)
	assert.NotNil(ls.T(), err)
	assert.IsType(ls.T(), &StopwatchErr{}, err)
	assert.Equal(ls.T(), NewNotStartedErr(ls.w).Error(), err.Error())
}

func (ls *lapSuite) TestLap_Error_AlreadyStopped() {
	_ = ls.w.Start()
	err := ls.w.Stop()
	assert.Nil(ls.T(), err)
	err = ls.w.Lap(ls.key, ls.comment)
	assert.NotNil(ls.T(), err)
	assert.IsType(ls.T(), &StopwatchErr{}, err)
	assert.Equal(ls.T(), NewAlreadyStoppedErr(ls.w).Error(), err.Error())
}

func TestNewRecord(t *testing.T) {
	comment := "test-comment"
	r := newRecord(comment)
	now := time.Now().UnixNano()
	assert.NotZero(t, r)
	assert.InDelta(t, now, r.ts, nanoTsDelta)
	assert.Equal(t, comment, r.comment)
}

//Success
// - No records
// - Some Records
//Error
// - Not Started
// - Not Stopped

func TestStopwatch_FullReport(t *testing.T) {
	frs := new(fullReportSuite)
	suite.Run(t, frs)
}

type fullReportSuite struct {
	logger *testLogger
	w      *Stopwatch
	suite.Suite
}

func (frs *fullReportSuite) SetupTest() {
	frs.logger = &testLogger{}
	frs.w = New("test-stopwatch", frs.logger)
}

func (frs *fullReportSuite) TestFullReport_Success() {
	_ = frs.w.Start()
	keys := []string{"first", "second", "third"}
	comments := []string{"woof", "meow", "chirp"}
	assert.Equal(frs.T(), len(keys), len(comments))
	for i := range keys {
		_ = frs.w.Lap(keys[i], comments[i])
	}
	_ = frs.w.Stop()

	rpt, err := frs.w.Report()
	assert.NotNil(frs.T(), rpt)
	assert.IsType(frs.T(), Report{}, rpt)
	assert.Nil(frs.T(), err)

	expectedDuration, err := frs.w.calculateDuration(start, stop)
	assert.Nil(frs.T(), err)
	assert.Equal(frs.T(), expectedDuration, rpt.Duration)
	assert.NotNil(frs.T(), rpt.Splits)
	expectedNumberOfSplits := len(frs.w.keys) - 1
	assert.Equal(frs.T(), expectedNumberOfSplits, len(rpt.Splits))

	expectedSplits := frs.w.calculateSplits()
	assert.Equal(frs.T(), expectedSplits, rpt.Splits)
}

func (frs *fullReportSuite) TestReport_Error_NotStarted() {
	rpt, err := frs.w.Report()
	assert.Zero(frs.T(), rpt)
	assert.NotNil(frs.T(), err)

	expectedErr := NewNotStartedErr(frs.w)
	assert.Equal(frs.T(), expectedErr.Error(), err.Error())
}

func (frs *fullReportSuite) TestFullReport_Error_NotStopped() {
	_ = frs.w.Start()
	rpt, err := frs.w.Report()
	assert.NotNil(frs.T(), err)
	assert.Zero(frs.T(), rpt)

	assert.IsType(frs.T(), &StopwatchErr{}, err)
	expectedErr := NewNotStoppedErr(frs.w)
	assert.Equal(frs.T(), expectedErr.Error(), err.Error())
}

//Success
//Error
// -- NonExistentKey

func TestStopwatch_CalculateDuration(t *testing.T) {
	cds := new(calculateDurationSuite)
	suite.Run(t, cds)
}

type calculateDurationSuite struct {
	w *Stopwatch
	suite.Suite
}

func (cds *calculateDurationSuite) SetupTest() {
	cds.w = New("test", nil)
}

func (cds *calculateDurationSuite) TestCalculateDuration_Success() {
	startTime := time.Now()
	startToA := 100 * time.Millisecond
	aTime := startTime.Add(startToA)
	aToB := 3 * time.Second
	bTime := aTime.Add(aToB)
	bToStop := 200 * time.Millisecond
	stopTime := bTime.Add(bToStop)

	cds.w.records = map[key]record{
		start: record{
			ts: startTime.UnixNano(),
		},
		newKey("a"): record{
			ts: aTime.UnixNano(),
		},
		newKey("b"): record{
			ts: bTime.UnixNano(),
		},
		stop: record{
			ts: stopTime.UnixNano(),
		},
	}

	duration, err := cds.w.calculateDuration(start, stop)
	assert.Nil(cds.T(), err)
	assert.NotZero(cds.T(), duration)
	assert.IsType(cds.T(), time.Duration(0), duration)
	assert.Equal(cds.T(), stopTime.Sub(startTime), duration)

	duration, err = cds.w.calculateDuration(newKey("a"), newKey("b"))
	assert.Nil(cds.T(), err)
	assert.Equal(cds.T(), bTime.Sub(aTime), duration)

	//ensure absolute value of duration
	duration, err = cds.w.calculateDuration(newKey("b"), newKey("a"))
	assert.Nil(cds.T(), err)
	assert.Equal(cds.T(), bTime.Sub(aTime), duration)
}

func (cds *calculateDurationSuite) TestCalculateDuration_Error_NonexistentKey() {
	cds.w.Start()
	cds.w.Lap("a", "")
	cds.w.Lap("b", "")
	cds.w.Stop()

	dur, err := cds.w.calculateDuration(start, newKey("c"))
	assert.Zero(cds.T(), dur)
	assert.NotNil(cds.T(), err)
	assert.IsType(cds.T(), &StopwatchErr{}, err)
}

func TestCalculateSplits(t *testing.T) {
	keys := []key{start, newKey("one"), newKey("two"), newKey("three"), stop}
	keymap := map[key]string{
		keys[0]: "",
		keys[1]: "woof",
		keys[2]: "meow",
		keys[3]: "chirp",
		keys[4]: "",
	}
	w := New("test", &testLogger{})
	_ = w.Start()
	time.Sleep(500 * time.Nanosecond)
	expectedDiff := 5000 * time.Nanosecond
	for i, v := range keys {
		if i == 0 || i == len(keymap)-1 {
			continue
		}
		_ = w.Lap(keys[i].String(), keymap[v])
		time.Sleep(expectedDiff)
	}
	_ = w.Stop()

	splits := w.calculateSplits()
	assert.Equal(t, len(w.keys)-1, len(splits))

	for i, k := range keys {
		if i == len(keys)-1 {
			break
		}
		expectedDur, err := w.calculateDuration(keys[i], keys[i+1])
		assert.Nil(t, err)
		expectedSplit := Split{
			Name:     k.String(),
			Comment:  keymap[k],
			Duration: expectedDur,
		}
		assert.Equal(t, expectedSplit, splits[i])
	}
}

type ctxSuite struct {
	ctx    context.Context
	logger *testLogger
	suite.Suite
}

func (cs *ctxSuite) SetupTest() {
	cs.ctx = context.Background()
	cs.logger = &testLogger{}
}

type ctxNewSuite struct {
	name string
	ctxSuite
}

func TestStopwatchContextCreationSuite(t *testing.T) {
	cns := new(ctxNewSuite)
	suite.Run(t, cns)
}

func (cns *ctxNewSuite) SetupTest() {
	cns.ctxSuite.SetupTest()
	cns.name = "test"
}

func (cns *ctxNewSuite) TestCtxNew_Success() {
	cns.ctx = CtxNew(cns.ctx, cns.name, cns.logger)

	w := cns.ctx.Value(ctxStopwatch)
	assert.NotNil(cns.T(), w)
	assert.IsType(cns.T(), &Stopwatch{}, w)

	sw := w.(*Stopwatch)
	assert.Equal(cns.T(), cns.name, sw.Name)

	sw, err := getStopwatchFromCtx(cns.ctx)
	assert.Nil(cns.T(), err)
	assert.NotNil(cns.T(), sw)
	assert.IsType(cns.T(), &Stopwatch{}, sw)
	assert.Equal(cns.T(), cns.name, sw.Name)
}

func (cns *ctxNewSuite) TestCtxNew_NilCtx() {
	newCtx := CtxNew(nil, cns.name, cns.logger)
	assert.NotNil(cns.T(), newCtx)

	w, err := getStopwatchFromCtx(newCtx)
	assert.Nil(cns.T(), err)
	assert.NotNil(cns.T(), w)
}

//Success
//Error
// - AlreadyStarted
// - AlreadyStopped
// - NotFoundInCtx
func TestStopwatchStartContext(t *testing.T) {
	css := new(ctxStartSuite)
	suite.Run(t, css)
}

type ctxStartSuite struct {
	name string
	ctxSuite
}

func (css *ctxStartSuite) SetupTest() {
	css.ctxSuite.SetupTest()
	css.name = "test"
	css.ctx = CtxNew(css.ctx, css.name, css.logger)
}

func (css *ctxStartSuite) TestCtxStart_Success() {
	var err error
	err = CtxStart(css.ctx)
	assert.Nil(css.T(), err)

	w, err := getStopwatchFromCtx(css.ctx)
	assert.Nil(css.T(), err)
	assert.True(css.T(), w.Running())
}

func (css *ctxStartSuite) TestCtxStart_Error_AlreadyStarted() {
	err := CtxStart(css.ctx)
	assert.Nil(css.T(), err)
	err = CtxStart(css.ctx)
	assert.NotNil(css.T(), err)
	assert.IsType(css.T(), &StopwatchErr{}, err)

	w, e := getStopwatchFromCtx(css.ctx)
	assert.Nil(css.T(), e)
	expectedErr := NewAlreadyStartedErr(w)
	assert.Equal(css.T(), expectedErr.Error(), err.Error())
}
func (css *ctxStartSuite) TestCtxStart_Error_AlreadyStopped() {
	_ = CtxStart(css.ctx)
	err := CtxStop(css.ctx)
	assert.Nil(css.T(), err)

	err = CtxStart(css.ctx)
	assert.NotNil(css.T(), err)
	assert.IsType(css.T(), &StopwatchErr{}, err)

	w, e := getStopwatchFromCtx(css.ctx)
	assert.Nil(css.T(), e)
	expectedErr := NewAlreadyStoppedErr(w)
	assert.Equal(css.T(), expectedErr.Error(), err.Error())
}

func (css *ctxStartSuite) TestCtxStart_Error_NotFoundInCtx() {
	ctx := context.Background()
	err := CtxStart(ctx)
	assert.NotNil(css.T(), err)
	assert.IsType(css.T(), &StopwatchErr{}, err)

	expectedErr := NewNotFoundErr()
	assert.Equal(css.T(), expectedErr.Error(), err.Error())
}

//Success
//Error
// - NotStarted
// - AlreadyStopped
// - NotFoundInCtx
func TestStopwatchContextStop(t *testing.T) {
	cst := new(ctxStopSuite)
	suite.Run(t, cst)
}

type ctxStopSuite struct {
	name string
	ctxSuite
}

func (cst *ctxStopSuite) SetupTest() {
	cst.ctxSuite.SetupTest()
	cst.name = "test"
	cst.ctx = CtxNew(cst.ctx, cst.name, cst.logger)
}

func (cst *ctxStopSuite) TestCtxStop_Success() {
	err := CtxStart(cst.ctx)
	assert.Nil(cst.T(), err)
	err = CtxStop(cst.ctx)
	assert.Nil(cst.T(), err)

	w, err := getStopwatchFromCtx(cst.ctx)
	assert.Nil(cst.T(), err)

	expectedLenOfKeys := 2
	assert.Equal(cst.T(), expectedLenOfKeys, len(w.keys))
	assert.Equal(cst.T(), w.keys[0], start)
	assert.Equal(cst.T(), w.keys[1], stop)
}

func (cst *ctxStopSuite) TestCtxStop_Error_NotStarted() {
	err := CtxStop(cst.ctx)
	assert.NotNil(cst.T(), err)

	w, e := getStopwatchFromCtx(cst.ctx)
	assert.Nil(cst.T(), e)
	expectedErr := NewNotStartedErr(w)
	assert.Equal(cst.T(), expectedErr.Error(), err.Error())
}

func (cst *ctxStopSuite) TestCtxStop_Error_AlreadyStopped() {
	_ = CtxStart(cst.ctx)
	_ = CtxStop(cst.ctx)
	err := CtxStop(cst.ctx)
	assert.NotNil(cst.T(), err)

	w, _ := getStopwatchFromCtx(cst.ctx)
	expectedErr := NewAlreadyStoppedErr(w)
	assert.Equal(cst.T(), expectedErr.Error(), err.Error())
}

func (cst *ctxStopSuite) TestCtxStop_Error_NotFoundInCtx() {
	ctx := context.Background()
	err := CtxStop(ctx)
	assert.NotNil(cst.T(), err)

	expectedErr := NewNotFoundErr()
	assert.Equal(cst.T(), expectedErr.Error(), err.Error())
}

//Success
//Error
// - NotStarted
// - AlreadyStopped
// - NotFoundInCtx

func TestCtxLap(t *testing.T) {
	cls := new(ctxLapSuite)
	suite.Run(t, cls)
}

type ctxLapSuite struct {
	name    string
	key     string
	comment string
	ctxSuite
}

func (cls *ctxLapSuite) SetupTest() {
	cls.name = "test"
	cls.key = "key"
	cls.comment = "comment"
	cls.ctxSuite.SetupTest()
}

func (cls *ctxLapSuite) TestCtxLap_Success() {
	cls.ctx = CtxNew(cls.ctx, cls.name, cls.logger)
	_ = CtxStart(cls.ctx)
	expectedTs := time.Now().UnixNano()
	err := CtxLap(cls.ctx, cls.key, cls.comment)
	assert.Nil(cls.T(), err)

	w, err := getStopwatchFromCtx(cls.ctx)
	assert.Nil(cls.T(), err)

	expectedLenOfKeys := 2
	assert.Equal(cls.T(), expectedLenOfKeys, len(w.keys))
	assert.Equal(cls.T(), w.keys[0], start)
	assert.Equal(cls.T(), w.keys[1], newKey(cls.key))

	assert.Equal(cls.T(), expectedLenOfKeys, len(w.records))
	l, exists := w.records[newKey(cls.key)]
	assert.True(cls.T(), exists)
	expectedEntry := record{
		ts: expectedTs,
		comment: cls.comment,
	}
	AssertEqualRecord(cls.T(), expectedEntry, l)
}

func (cls *ctxLapSuite) TestCtxLap_Error_NotStarted() {
	cls.ctx = CtxNew(cls.ctx, cls.name, cls.logger)
	err := CtxLap(cls.ctx, cls.key, cls.comment)
	assert.NotNil(cls.T(), err)

	w, e := getStopwatchFromCtx(cls.ctx)
	assert.Nil(cls.T(), e)

	expectedErr := NewNotStartedErr(w)
	assert.Equal(cls.T(), expectedErr.Error(), err.Error())

	expectedLenOfKeys := 0
	assert.Equal(cls.T(), expectedLenOfKeys, len(w.keys))
	assert.Equal(cls.T(), expectedLenOfKeys, len(w.records))
}

func (cls *ctxLapSuite) TestCtxLap_Error_AlreadyStopped() {
	cls.ctx = CtxNew(cls.ctx, cls.name, cls.logger)
	_ = CtxStart(cls.ctx)
	_ = CtxStop(cls.ctx)
	err := CtxLap(cls.ctx, cls.key, cls.comment)
	assert.NotNil(cls.T(), err)

	w, e := getStopwatchFromCtx(cls.ctx)
	assert.Nil(cls.T(), e)
	expectedErr := NewAlreadyStoppedErr(w)
	assert.Equal(cls.T(), expectedErr.Error(), err.Error())

	expectedLenOfKeys := 2
	assert.Equal(cls.T(), expectedLenOfKeys, len(w.keys))
	assert.Equal(cls.T(), expectedLenOfKeys, len(w.records))
}

func (cls *ctxLapSuite) TestCtxLap_Error_NotFound() {
	ctx := context.Background()
	err := CtxLap(ctx, cls.name, cls.comment)
	assert.NotNil(cls.T(), err)

	expectedErr := NewNotFoundErr()
	assert.Equal(cls.T(), expectedErr.Error(), err.Error())
}

//Success
//Error
// - NotStarted
// - NotStopped
// - NotFound

func TestCtxReport(t *testing.T) {
	crs := new(ctxReportSuite)
	suite.Run(t, crs)
}

type ctxReportSuite struct {
	name string
	ctxSuite
}

func (crs *ctxReportSuite) SetupTest() {
	crs.name = "test"
	crs.ctxSuite.SetupTest()
}

func (crs *ctxReportSuite) TestCtxReport_Success() {
	crs.ctx = CtxNew(crs.ctx, crs.name, crs.logger)

	_ = CtxStart(crs.ctx)
	_ = CtxStop(crs.ctx)

	rpt, err := CtxReport(crs.ctx)
	assert.Nil(crs.T(), err)
	assert.NotZero(crs.T(), rpt)
	assert.IsType(crs.T(), Report{}, rpt)
}

func (crs *ctxReportSuite) TestCtxReport_Error_NotStarted() {
	crs.ctx = CtxNew(crs.ctx, crs.name, crs.logger)
	rpt, err := CtxReport(crs.ctx)
	assert.Zero(crs.T(), rpt)
	assert.NotNil(crs.T(), err)

	w, e := getStopwatchFromCtx(crs.ctx)
	assert.Nil(crs.T(), e)

	expectedErr := NewNotStartedErr(w)
	assert.Equal(crs.T(), expectedErr.Error(), err.Error())
}

func (crs *ctxReportSuite) TestCtxReport_Error_NotStopped() {
	crs.ctx = CtxNew(crs.ctx, crs.name, crs.logger)
	_ = CtxStart(crs.ctx)

	rpt, err := CtxReport(crs.ctx)
	assert.Zero(crs.T(), rpt)
	assert.NotNil(crs.T(), err)

	w, e := getStopwatchFromCtx(crs.ctx)
	assert.Nil(crs.T(), e)

	expectedErr := NewNotStoppedErr(w)
	assert.Equal(crs.T(), expectedErr.Error(), err.Error())
}

func (crs *ctxReportSuite) TestCtxReport_Error_NotFound() {
	ctx := context.Background()
	rpt, err := CtxReport(ctx)
	assert.Zero(crs.T(), rpt)
	assert.NotNil(crs.T(), err)

	expectedErr := NewNotFoundErr()
	assert.Equal(crs.T(), expectedErr.Error(), err.Error())
}
