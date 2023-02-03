package fakes

import "sync"

type ExitHandlerInterface struct {
	ErrorCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Err error
		}
		Stub func(error)
	}
}

func (f *ExitHandlerInterface) Error(param1 error) {
	f.ErrorCall.mutex.Lock()
	defer f.ErrorCall.mutex.Unlock()
	f.ErrorCall.CallCount++
	f.ErrorCall.Receives.Err = param1
	if f.ErrorCall.Stub != nil {
		f.ErrorCall.Stub(param1)
	}
}
