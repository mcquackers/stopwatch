package stopwatch

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

func TestStopwatchCreation(t *testing.T) {
	scs := new(creationSuite)
	suite.Run(t, scs)
}

type creationSuite struct {
	stopwatchName string
	suite.Suite
}

func (cs *creationSuite) SetupTest() {
	cs.stopwatchName = "testWatch"
}

func (cs *creationSuite) TestNew() {
	w := New(cs.stopwatchName)
	assert.NotNil(cs.T(), w)
	assert.IsType(cs.T(), &Stopwatch{}, w)

	assert.Equal(cs.T(), cs.stopwatchName, w.Name)
	assert.NotNil(cs.T(), w.keys)
	assert.NotNil(cs.T(), w.records)
	assert.False(cs.T(), w.running)
	assert.NotNil(cs.T(), w.rl)
}

//Start
//DoubleStart
//StartAfterStop
func TestStopwatchStart(t *testing.T) {
	ss := new(startSuite)
	suite.Run(t, ss)
}

type startSuite struct {
	w *Stopwatch
	suite.Suite
}

func (ss *startSuite) SetupTest() {
	ss.w = New("test")
}

func (ss *startSuite) TestStart_Success() {
	expectedStartTime := time.Now().Unix()
	err := ss.w.Start()
	assert.Nil(ss.T(), err)
	assert.True(ss.T(), ss.w.Running())
	assert.Equal(ss.T(), 1, len(ss.w.keys))
	assert.Equal(ss.T(), start, ss.w.keys[0])
	assert.Equal(ss.T(), 1, len(ss.w.records))
	startRecord, exists := ss.w.records[start]
	assert.True(ss.T(), exists)
	assert.NotZero(ss.T(), startRecord)
	assert.Equal(ss.T(), expectedStartTime, startRecord.ts)
	assert.Zero(ss.T(), startRecord.comment)
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
	w *Stopwatch
	suite.Suite
}

func (st *stopSuite) SetupTest() {
	st.w = New("test")
}

func (st *stopSuite) TestStop_Success() {
	err := st.w.Start()
	assert.Nil(st.T(), err)
	err = st.w.Stop()
	assert.Nil(st.T(), err)
	assert.Equal(st.T(), 2, len(st.w.keys))
	assert.Equal(st.T(), stop, st.w.keys[1])
}

func (st *stopSuite) TestStop_Error_NotStarted() {
	err := st.w.Stop()
	assert.NotNil(st.T(), err)
	assert.Equal(st.T(), NewNotStartedErr(st.w).Error(), err.Error())
}

func (st *stopSuite) TestStop_Error_AlreadyStopped() {
	err := st.w.Start()
	assert.Nil(st.T(), err)
	err = st.w.Stop()
	assert.Nil(st.T(), err)
	err = st.w.Stop()
	assert.NotNil(st.T(), err)
	assert.Equal(st.T(), NewAlreadyStoppedErr(st.w).Error(), err.Error())
}

//Success
//Error_NotStarted
//Error_AlreadyStopped
func TestLap(t *testing.T) {
	ls := new(lapSuite)
	suite.Run(t, ls)
}

type lapSuite struct {
	w *Stopwatch

	key     string
	comment string
	suite.Suite
}

func (ls *lapSuite) SetupTest() {
	ls.w = New("test")
	ls.key = "key"
	ls.comment = "lap-comment"
}

func (ls *lapSuite) TestLap_Success() {
	err := ls.w.Start()
	assert.Nil(ls.T(), err)
	//basic return check
	expectedTs := time.Now().Unix()
	err = ls.w.Lap(ls.key, ls.comment)
	assert.Nil(ls.T(), err)

	//internal function check
	assert.Equal(ls.T(), 2, len(ls.w.keys))
	assert.Equal(ls.T(), ls.w.keys[1], newKey(ls.key))
	assert.Equal(ls.T(), 2, len(ls.w.records))

	//integrity check
	lapRecord, exists := ls.w.records[newKey(ls.key)]
	assert.True(ls.T(), exists)
	assert.NotZero(ls.T(), lapRecord)
	assert.Equal(ls.T(), expectedTs, lapRecord.ts)
	assert.Equal(ls.T(), ls.comment, lapRecord.comment)
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
	assert.NotZero(t, r)
	assert.Equal(t, time.Now().Unix(), r.ts)
	assert.Equal(t, comment, r.comment)
}
