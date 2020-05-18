package together

import (
    "fmt"
    "sync"
    "sync/atomic"
)

type rail struct {
    at int
    queue chan train
    open int32
}

type train struct {
    delta int
}

type RailSwitch struct {
    rails map[int] rail
    queue chan rail

    at int
    value int
    control sync.Mutex
    register sync.Mutex
}

/*

Queue 
    // queue has to create new queue when there is no queue for `at`
    // queue has to just pass when current is at `at`
    // queue has to stop when current is not at `at`
    // queue should not use goroutine for sending as it should stop
    // current has to receive in order to let pass

Proceed
    // proceed has to panic when it is not at `at`
    // proceed has to deduct 1 from current queue

*/

func NewRailSwitch() *RailSwitch {

    rs       := new(RailSwitch)
    rs.rails  = make(map[int] rail)
    rs.queue  = make(chan rail)
    
    rs.control.Lock()

    go func() {
        for r := range rs.queue {

            at      := r.at
            queue   := r.queue
            rs.at    = at
            rs.value = 0

            rs.control.Unlock()

            print(at); print(" = "); print(queue == nil); println("")
            if queue == nil {
                rs.control.Lock()
                continue
            }

            for t := range queue {
                rs.value += t.delta
                print(rs.at); print(" - "); print(rs.value); println("")
                if rs.value == 0 {
                    rs.control.Lock()
                    atomic.CompareAndSwapInt32(&r.open, 1, 0)
                    break
                }
            }

        }
    }()

    return rs

}

func(rs *RailSwitch) Queue(at, delta int) {

    rs.register.Lock()
    r, ok := rs.rails[at]

    // Check for rail
    if !ok {
        q := make(chan train)
        r := rail{at, q, 0}
        rs.rails[at] = r
    }

    rs.register.Unlock()

    // Check for at
    rs.control.Lock()
    if closed := atomic.CompareAndSwapInt32(&r.open, 0, 1); closed {
        go func() {
            rs.queue <- r
        }()
    }
    rs.control.Unlock()

    r.queue <- train{delta}

}

func(rs *RailSwitch) Proceed(at int) {
    if rs.at != at {
        panic(fmt.Sprintf("together: invalid proceed call for %d while it is at %d", at, rs.at))
    }
    rs.rails[at].queue <- train{-1}
}

func(rs *RailSwitch) Wait() {
    rs.queue <- rail{}
}
