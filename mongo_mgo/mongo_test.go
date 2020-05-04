package mongo

// This example code runs OK with mongo v3.6.3

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	mgo "github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	hasSessionSleep bool

	session *mgo.Session

	Collection = "test"
	Database   = "test"
	URI        = "localhost:27017"
)

const (
	returnContextKey = "want_return"
	early            = "early"
)

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	Collection string
	Database   string
	URI        string
}

//type TestModel struct {
//	State  string `bson:"state"`
//	NewKey int    `bson:"new_key,omitempty"`
//}

type ungraceful struct{}

func (t ungraceful) shutdown(ctx context.Context, session *mgo.Session, closedChannel chan bool) {
	fmt.Printf("Doing Test Harness: shutdown() ...\n")
	time.Sleep(timeLeft + (100 * time.Millisecond))
	if ctx.Value(returnContextKey) == early {
		fmt.Printf("Got key: early\n")
		return
	}
	if ctx.Err() != nil {
		fmt.Printf("ctx Err:%v\n", ctx.Err())
		return
	}

	fmt.Printf("Test mock - reached session.Close() ...should NEVER get here !!\n")
	session.Close()

	closedChannel <- true
}

func TestSuccessfulCloseMongoSession(t *testing.T) {
	_, err := setupSession()
	if err != nil {
		fmt.Printf("mongo instance not available, skip close tests: %v\n", err)
		return
	}

	if err = cleanupTestData(session.Copy()); err != nil {
		fmt.Printf("Failed to delete test data: %v\n", err)
	}
	fmt.Printf("'start' points to ==> %T, %v\n", start, start)

	Convey("Safely close mongo session", t, func() {
		if !hasSessionSleep {
			Convey("with no context deadline", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				err := Close(ctx, session.Copy())

				So(err, ShouldBeNil)
			})
		}

		Convey("within context timeout (deadline)", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err := Close(ctx, session.Copy())

			So(err, ShouldBeNil)
		})

		Convey("within context deadline", func() {
			time := time.Now().Local().Add(time.Second * time.Duration(2))
			ctx, cancel := context.WithDeadline(context.Background(), time)
			defer cancel()
			err := Close(ctx, session.Copy())

			So(err, ShouldBeNil)
		})
	})

	if err = setUpTestData(session.Copy()); err != nil {
		fmt.Printf("Failed to insert test data, skipping tests: %v\n", err)
		os.Exit(1)
	}

	Convey("Timed out from safely closing mongo session", t, func() {
		Convey("with no context deadline", func() {
			fmt.Printf("NOTE: The test code will now switch to using the 'mock' function to handle simulated database issues ...\n")
			start = ungraceful{}
			fmt.Printf("'start' points to ==> %T, %v\n", start, start)
			copiedSession := session.Copy()
			go func() {
				_ = slowQueryMongo(copiedSession)
			}()
			// Sleep for half a second for mongo query to begin
			time.Sleep(500 * time.Millisecond)

			// Force context exit with a 'key'
			ctx := context.WithValue(context.Background(), returnContextKey, early)
			err := Close(ctx, copiedSession) // NOTE: the 'session.Close()' in the call to this Close() is not reached due to 'ctx' exit clause.

			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("closing mongo timed out"))
			time.Sleep(500 * time.Millisecond)
		})

		Convey("with context deadline", func() {
			copiedSession := session.Copy()
			go func() {
				_ = slowQueryMongo(copiedSession)
			}()
			// Sleep for half a second for mongo query to begin
			time.Sleep(500 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			err := Close(ctx, copiedSession)

			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, context.DeadlineExceeded)
		})
	})

	if err = cleanupTestData(session.Copy()); err != nil {
		fmt.Printf("Failed to delete test data: %v\n", err)
	}
	fmt.Printf("Doing 2.5 second extra delay to allow session close's ...\n")
	time.Sleep(2500 * time.Millisecond)
}

func cleanupTestData(session *mgo.Session) error {
	defer func() {
		fmt.Printf("Exiting: cleanupTestData\n")
		session.Close()
	}()

	err := session.DB(Database).DropDatabase()
	if err != nil {
		return err
	}

	return nil
}

func slowQueryMongo(session *mgo.Session) error {
	defer func() {
		fmt.Printf("Exiting: slowQueryMongo\n")
		session.Close()
	}()

	time.Sleep(2 * time.Second)

	_, err := session.DB(Database).C(Collection).Find(bson.M{"$where": "sleep(2000) || true"}).Count()
	if err != nil {
		return err
	}

	return nil
}

func getTestData() []bson.M {
	return []bson.M{
		bson.M{
			"_id":   "1",
			"state": "first",
		},
		bson.M{
			"_id":   "2",
			"state": "second",
		},
	}
}

func setUpTestData(session *mgo.Session) error {
	defer func() {
		fmt.Printf("Exiting: setUpTestData\n")
		session.Close()
	}()

	if _, err := session.DB(Database).C(Collection).Upsert(bson.M{"_id": "1"}, getTestData()[0]); err != nil {
		return err
	}

	if _, err := session.DB(Database).C(Collection).Upsert(bson.M{"_id": "2"}, getTestData()[1]); err != nil {
		return err
	}

	return nil
}

func setupSession() (*Mongo, error) {
	mongo := &Mongo{
		Collection: Collection,
		Database:   Database,
		URI:        URI,
	}

	if session != nil {
		return nil, errors.New("Failed to initialise mongo")
	}
	var err error

	if session, err = mgo.Dial(URI); err != nil {
		return nil, err
	}

	session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	session.SetMode(mgo.Strong, true)
	return mongo, nil
}
