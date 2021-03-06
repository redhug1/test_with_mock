package mongo

// To allow this example code to run without having mongo installed, all code that uses mongo
// has been commented out.

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// The 'Shutdown' interface ...
// ( achieves function overloading / re-assignment in mongo_test.go
//   that is ... allowing one to 'mock' out the original shutdown() below
//   with a different one in mongo_test.go - that has the same function signature )
//

// Shutdown represents an interface to the shutdown method
type Shutdown interface {
	shutdown(ctx context.Context, ses int /*session *mgo.Session*/, closedChannel chan bool)
}

type graceful struct{}

func (t graceful) shutdown(ctx context.Context, ses int /*session *mgo.Session*/, closedChannel chan bool) {
	fmt.Printf("Doing Original: shutdown() ...\n")

	//	session.Close()

	closedChannel <- true
}

var (
	start    Shutdown = graceful{}
	timeLeft          = 1000 * time.Millisecond
)

// Close represents mongo session closing within the context deadline
func Close(ctx context.Context, ses int /*session *mgo.Session*/) error {
	closedChannel := make(chan bool)
	defer close(closedChannel)

	if deadline, ok := ctx.Deadline(); ok {
		// Add some time to timeLeft so case where ctx.Done in select
		// statement below gets called before time.After(timeLeft) gets called.
		// This is so the context error is returned over hardcoded error.
		timeLeft = deadline.Sub(time.Now()) + (10 * time.Millisecond)
	}

	go func() {
		start.shutdown(ctx, ses /*session*/, closedChannel)
		return
	}()

	select {
	case <-time.After(timeLeft):
		return errors.New("closing mongo timed out")
	case <-closedChannel:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
