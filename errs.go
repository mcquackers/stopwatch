package stopwatch

import "fmt"

type StopwatchErr struct  {
	msg string
}

func (e *StopwatchErr) Error() string {
	return e.msg
}

func NewAlreadyStartedErr(w *Stopwatch) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("stopwatch %s has already been started", w.Name),
	}
}

func NewNotStartedErr(w *Stopwatch) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf(" stopwatch %s has not been started", w.Name),
	}
}

func NewAlreadyStoppedErr(w *Stopwatch) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("stopwatch %s has already been stopped", w.Name),
	}
}