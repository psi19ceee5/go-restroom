package restroom

import (
	"errors"
	"sync"
)

// ErrAccessDenied is thrown upon an attempt to `Lock()` or `Unlock()` the restroom with an invalid `Ticket`
var ErrAccessDenied = errors.New("Access to restroom denied")

// A Ticket grants access to code blocks restricted by RoomLock.Lock() and RoomLock.Unlock() to the
// go-routine holding the Ticket. A Ticket is granted by calling RoomLock.
type Ticket struct {
	previous *Ticket
	next     *Ticket
}

type queue struct {
	tickets []*Ticket
	first   *Ticket
	last    *Ticket
}

func newQueue() queue {
	var tFirst Ticket
	tLast := Ticket{previous: &tFirst, next: nil}
	tFirst = Ticket{previous: nil, next: &tLast}

	return queue{
		tickets: []*Ticket{&tFirst, &tLast},
		first:   &tFirst,
		last:    &tLast,
	}
}

func (q *queue) getNewTicket() *Ticket {
	newTicket := &Ticket{
		previous: q.last.previous,
		next:     q.last,
	}

	q.pushBack(newTicket)

	return newTicket
}

func (q *queue) pushBack(ticket *Ticket) {
	q.tickets = append(q.tickets, ticket)

	q.last.previous.next = ticket
	q.last.previous = ticket
}

func (q *queue) popFront() {
	if len(q.tickets) <= 2 {
		return
	}

	q.first.next.next.previous = q.first
	q.first.next = q.first.next.next
}

func (q *queue) activeTicket() *Ticket {
	return q.first.next
}

// RoomLock is a type that helps synchronize the execution of handler functions for a single room
// using a queue-system it also makes sure, the go-routines access the restricted code in the same
// order as they arrive at WaitIfLocked()
type RoomLock struct {
	mutex sync.Mutex
	cond  *sync.Cond
	value bool
	queue queue
}

// NewRoomLock creates a new instance of RoomLock with `value“ set to false
func NewRoomLock() *RoomLock {
	lock := RoomLock{}
	lock.cond = sync.NewCond(&lock.mutex)
	lock.value = false
	lock.queue = newQueue()

	return &lock
}

// Lock locks the RoomLock (error must only be cached if unauthorized access attempts are of any interest)
func (lock *RoomLock) Lock(ticket *Ticket) error {
	if lock.queue.activeTicket() != ticket {
		return ErrAccessDenied
	}

	lock.set(true)

	return nil
}

// Unlock unlocks the RoomLock (error must only be cached if unauthorized access attempts are of any interest)
func (lock *RoomLock) Unlock(ticket *Ticket) error {
	if lock.queue.activeTicket() != ticket {
		return ErrAccessDenied
	}

	lock.queue.popFront() // trow your ticket in the bin
	lock.set(false)       // unlock the door

	return nil
}

// IsLocked tells if the room is currently locked (mainly used for testing)
func (lock *RoomLock) IsLocked() bool {
	return lock.value
}

// WaitIfLocked halts the program execution if the RoomLock is locked and does nothing if not.
// It implements the ticket system to grant access to go-routines in the same order in which
// they called this method. If called with `nil` as argument, a new Ticket is granted and returned.
func (lock *RoomLock) WaitIfLocked(ticket *Ticket) *Ticket {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	if ticket == nil {
		ticket = lock.queue.getNewTicket()
	}

	for lock.value || lock.queue.activeTicket() != ticket {
		lock.cond.Wait() // Wait temporarily releases the mutex until cond.Broadcast or cond.Signal
	}

	return ticket
}

// DrawNewTicket draws and returns a new Ticket without waiting or blocking. Can be used to call WaitIfLocked()
// later with an existing ticket.
func (lock *RoomLock) DrawNewTicket() *Ticket {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	return lock.queue.getNewTicket()
}

func (lock *RoomLock) set(val bool) {
	lock.mutex.Lock()
	lock.value = val
	lock.cond.Broadcast()
	lock.mutex.Unlock()
}
