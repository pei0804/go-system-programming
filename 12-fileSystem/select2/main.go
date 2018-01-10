package main

import (
	"log"
	"syscall"
)

func main() {
	// create kqueue
	kq, err := syscall.Kqueue()
	if err != nil {
		log.Println("Error creating Kqueue descriptor!")
		return
	}
	// open folder
	fd, err := syscall.Open("./test", syscall.O_RDONLY, 0)
	if err != nil {
		log.Println("Error opening folder descriptor!")
		return
	}
	// build kevent
	ev1 := syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_VNODE,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE | syscall.EV_ONESHOT,
		Fflags: syscall.NOTE_DELETE | syscall.NOTE_WRITE,
		Data:   0,
		Udata:  nil,
	}
	// configure timeout
	timeout := syscall.Timespec{
		Sec:  0,
		Nsec: 0,
	}
	// wait for events
	for {
		// create kevent
		events := make([]syscall.Kevent_t, 10)
		nev, err := syscall.Kevent(kq, []syscall.Kevent_t{ev1}, events, &timeout)
		if err != nil {
			log.Println("Error creating kevent")
		}
		// check if there was an event
		for i := 0; i < nev; i++ {
			// log
			log.Printf("Event [%d] -> %+v", i, events[i])
		}
	}
}
