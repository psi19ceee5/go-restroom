[![Go Reference](https://pkg.go.dev/badge/github.com/psi19ceee5/go-restroom.svg)](https://pkg.go.dev/github.com/psi19ceee5/go-restroom)
![Coverage](https://img.shields.io/badge/Coverage-95.5%25-brightgreen)

# go-restroom

A thread-order preserving access-queue mutex

## Why do I need a restroom?

Well, there are certain things you don't want to do in public. Hence, a little privacy may sometimes be appreciated 😉
Exactly for those purposes we have public restrooms: a person enters, locks the door from the inside and does what has to be done. In the
mean time, other people with private needs gather up outside. But they have to wait until the first person has finished, unlocked the door and left the restroom. Only then, the next person can enter and lock the door again. Since we are civilized people we do not fight over the order in which we enter, but build a nicely ordered queue in front of the door. To make things even more convenient, we nowadays deploy ticket systems in many queueing situations. Each person has to draw a ticket with a number and a display showing your number tells you this is your turn.

Now, why taking so much about sanitary rooms? Because this package implements exactly such a mechanic for concurrent go-routines queueing before a critical block of code which
 - must only ever be accessed by one go-routine at a time
 - must be accessed by the go-routines in a very well defined order

## Example

```go
import "github.com/psi19ceee5/go-restroom/pkg/restroom"

func myFun(lock *restroom.RoomLock) {
    // ...
    t := lock.WaitIfLocked(nil) // Draw a ticket and queue up if the room is locked
    lock.Lock(t)                // Once its your turn, lock the room by showing your ticket
    defer lock.Unlock(t)        // When you're done, don't forget to unlock the room, your ticket is discarded in the process
    // ...
}

func main() {
    // ...
    lock := restroom.NewRoomLock()

    go myFun(lock)

    // ...
}

```

Alternatively, the line `t := lock.WaitIfLocked(nil)` could be replaced by
```
t := lock.DrawNewTicket()
lock.WaitIfLocked(t)
```
to decouple the ticket grant from the idle waiting routine.
