package restroom_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/psi19ceee5/go-restroom/pkg/restroom"
)

var _ = Describe("Restroom", func() {
	var roomLock *restroom.RoomLock
	var start, end []int
	var goRoutine func(int)

	BeforeEach(func() {
		roomLock = restroom.NewRoomLock()
		start, end = []int{}, []int{}
		goRoutine = func(number int) {
			fmt.Printf("routine %v (goRoutine) entered\n", number)
			roomLock.WaitIfLocked(number)
			fmt.Printf("routine %v (goRoutine) is done waiting\n", number)
			roomLock.Lock(number)
			defer roomLock.Unlock(number)
			fmt.Printf("routine %v (goRoutine) has locked the room\n", number)

			start = append(start, number)
			fmt.Printf("start = %v\n", start)
			time.Sleep(10 * time.Millisecond)
			end = append(end, number)

			fmt.Printf("routine %v (goRoutine) is leaving\n", number)
		}
	})

	Describe("Multiple go-routines", func() {
		Context("with varying runtime", func() {
			It("should be executed in order", func() {
				for i := 0; i < 10; i++ {
					go goRoutine(i)
					time.Sleep(time.Millisecond)
				}

				roomLock.WaitIfLocked(-1)

				Expect(start).To(Equal([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))
				Expect(end).To(Equal([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))
			})
		})
	})
})
