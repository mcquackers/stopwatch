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

func NewAlreadyStoppedErr(w *Stopwatch) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("stopwatch %s has already been stopped", w.Name),
	}
}

func NewNotStartedErr(w *Stopwatch) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("stopwatch %s has not been started", w.Name),
	}
}

func NewNotStoppedErr(w *Stopwatch) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("stopwatch %s has not been stopped", w.Name),
	}
}


func NewNotFoundErr() *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("no stopwatch found in ctx"),
	}
}

func NewNonExistentKeyErr(w *Stopwatch, missingKey key) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("stopwatch %s does not have key %s", w.Name, missingKey),
	}
}

func NewBadValueErr(be interface{}) *StopwatchErr {
	return &StopwatchErr{
		msg: fmt.Sprintf("found unexpected type in context: %T", be),
	}
}