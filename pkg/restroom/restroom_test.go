package restroom_test

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/psi19ceee5/go-restroom/pkg/restroom"
)

var _ = Describe("Restroom", func() {
	var roomLock *restroom.RoomLock
	var control []string
	var controlExp []string
	var goRoutineImplicit func(int)
	var goRoutineExplicit func(int)

	BeforeEach(func() {
		roomLock = restroom.NewRoomLock()
		control = []string{}
		controlExp = []string{
			"0-start",
			"0-end",
			"1-start",
			"1-end",
			"2-start",
			"2-end",
			"3-start",
			"3-end",
			"4-start",
			"4-end",
			"5-start",
			"5-end",
			"6-start",
			"6-end",
			"7-start",
			"7-end",
			"8-start",
			"8-end",
			"9-start",
			"9-end",
		}
		goRoutineImplicit = func(number int) {
			t := roomLock.WaitIfLocked(nil)
			roomLock.Lock(t)
			defer roomLock.Unlock(t)

			control = append(control, fmt.Sprintf("%d-start", number))
			randTime := time.Duration(rand.Intn(100))
			time.Sleep(randTime * time.Millisecond)
			control = append(control, fmt.Sprintf("%d-end", number))
		}
		goRoutineExplicit = func(number int) {
			t := roomLock.DrawNewTicket()

			// doing some stuff in the mean time - waiting in the queue is boring
			randTime := time.Duration(rand.Intn(100))
			time.Sleep(randTime * time.Millisecond)

			roomLock.WaitIfLocked(t)
			roomLock.Lock(t)
			defer roomLock.Unlock(t)

			control = append(control, fmt.Sprintf("%d-start", number))
			randTime = time.Duration(rand.Intn(100))
			time.Sleep(randTime * time.Millisecond)
			control = append(control, fmt.Sprintf("%d-end", number))
		}
	})

	Describe("Multiple go-routines", func() {
		Context("with varying runtime", func() {
			Context("using implicit Ticket grant", func() {
				It("should be executed in order", func() {
					for i := range 10 {
						go goRoutineImplicit(i)
						time.Sleep(time.Millisecond) // needed to ensure correct dispatch order
					}

					roomLock.WaitIfLocked(nil)

					Ω(control).To(Equal(controlExp))
				})
			})
			Context("using explicit Ticket grant", func() {
				It("should be executed in order", func() {
					for i := range 10 {
						go goRoutineExplicit(i)
						time.Sleep(time.Millisecond) // needed to ensure correct dispatch order
					}

					roomLock.WaitIfLocked(nil)

					Ω(control).To(Equal(controlExp))
				})
			})
		})
	})
	Describe("An outside routine", func() {
		It("should not be able to unlock a room from the outside", func() {
			t := roomLock.DrawNewTicket()
			roomLock.Lock(t)

			t2 := roomLock.DrawNewTicket()
			err := roomLock.Unlock(t2)

			Ω(roomLock.IsLocked()).To(Equal(true))
			Ω(err).To(Equal(restroom.ErrAccessDenied))
		})
	})
})
