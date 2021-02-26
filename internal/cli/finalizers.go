package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const timeout = 5 * time.Second

type Finalizers struct {
	notify func(c chan<- os.Signal, sig ...os.Signal)
	exit   func(int)
	funcs  []func()
}

func NewFinalizer() *Finalizers {
	fin := &Finalizers{
		notify: signal.Notify,
		exit:   os.Exit,
	}
	fin.setupCloseHandler()

	return fin
}

func (f *Finalizers) Add(function func()) {
	f.funcs = append(f.funcs, function)
}

func (f *Finalizers) setupCloseHandler() {
	c := make(chan os.Signal, 1)
	f.notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-c
		fmt.Printf("\r- Signal '%v' received from Terminal. Exiting...\n ", sig)
		lastChan := make(chan struct{})

		go func() {
			for _, f := range f.funcs {
				if f != nil {
					f()
				}
			}
			lastChan <- struct{}{}
		}()

		go func() {
			time.Sleep(timeout)
			lastChan <- struct{}{}
		}()

		<-lastChan
		f.exit(0)
	}()
}
