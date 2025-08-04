package restroom

import (
	"fmt"
	"sync"
)

type ticket struct {
	previous *ticket
	next     *ticket
}

type queue struct {
	tickets []*ticket
	first   *ticket
	last    *ticket
}

func NewQueue() queue {
	var tFirst ticket
	tLast := ticket{previous: &tFirst, next: nil}
	tFirst = ticket{previous: nil, next: &tLast}

	return queue{
		tickets: []*ticket{&tFirst, &tLast},
		first:   &tFirst,
		last:    &tLast,
	}
}

func (q *queue) getNewTicket() *ticket {
	newTicket := &ticket{
		previous: q.last.previous,
		next:     q.last,
	}

	q.pushBack(newTicket)

	return newTicket
}

func (q *queue) pushBack(ticket *ticket) {
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

func (q *queue) activeTicket() *ticket {
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
	lock.queue = NewQueue()

	return &lock
}

// Lock locks the RoomLock
func (lock *RoomLock) Lock(number int) {
	fmt.Printf("routine %v (Lock) is locking the room \n", number)
	lock.set(true)
}

// Unlock unlocks the RoomLock
func (lock *RoomLock) Unlock(number int) {
	fmt.Printf("routine %v (Unlock) discards the ticket \n", number)
	lock.queue.popFront() // trow your ticket in the bin
	fmt.Printf("routine %v (Unlock) unlocks the room \n", number)
	lock.set(false) // unlock the door
}

// WaitIfLocked halts the program execution if the RoomLock is locked and does nothing if not
// it implements the ticket system to grant access to go-routines in the same order in which
// they called this method
func (lock *RoomLock) WaitIfLocked(number int) {
	fmt.Printf("routine %v (WaitIfLocked) entered\n", number)
	lock.mutex.Lock()
	fmt.Printf("routine %v (WaitIfLocked) is drawing the next ticket\n", number)
	ticket := lock.queue.getNewTicket()
	fmt.Printf("routine %v (WaitIfLocked) got a Ticket \n", number)

	for lock.value || lock.queue.activeTicket() != ticket {
		fmt.Printf("routine %v (WaitIfLocked) is waiting \n", number)
		lock.cond.Wait() // Wait temporarily releases the mutex until cond.Broadcast or cond.Signal
	}

	fmt.Printf("routine %v (WaitIfLocked) left the waiting loop \n", number)

	lock.mutex.Unlock()

	fmt.Printf("routine %v (WaitIfLocked) is leaving \n", number)
}

func (lock *RoomLock) set(val bool) {
	lock.mutex.Lock()
	lock.value = val
	lock.cond.Broadcast()
	lock.mutex.Unlock()
}
