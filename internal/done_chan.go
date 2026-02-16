package internal

import "sync"

type DoneChan chan struct{}

type DoneFunc func() DoneChan

func wrapWaitGroup(wg *sync.WaitGroup) DoneChan {
	ret := make(DoneChan)
	go func() {
		wg.Wait()
		close(ret)
	}()
	return ret
}

func Combine(first DoneChan, second DoneChan) DoneChan {
	ret := make(DoneChan)
	go func() {
		<-first
		<-second
		close(ret)
	}()
	return ret
}
