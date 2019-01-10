package benchmarks

import (
	"errors"
	"fmt"
	"time"

	"github.com/ssgreg/logf"
	"go.uber.org/zap/zapcore"
)

type user struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
type users []*user

func (u *user) EncodeLogfObject(enc logf.FieldEncoder) error {
	enc.EncodeFieldString("name", u.Name)
	enc.EncodeFieldString("email", u.Email)
	enc.EncodeFieldInt64("createdAt", u.CreatedAt.UnixNano())

	return nil
}

// func (uu users) MarshalLogArray(arr zapcore.ArrayEncoder) error {
// 	var err error
// 	for i := range uu {
// 		err = multierr.Append(err, arr.AppendObject(uu[i]))
// 	}
// 	return err
// }

func (u *user) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", u.Name)
	enc.AddString("email", u.Email)
	enc.AddInt64("createdAt", u.CreatedAt.UnixNano())
	return nil
}

var (
	errExample = errors.New("example")

	messages = makePseudoMessages(1000)

	tenInts    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	tenStrings = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	tenTimes = []time.Time{
		time.Unix(0, 0), time.Unix(1, 0), time.Unix(2, 0), time.Unix(3, 0), time.Unix(4, 0),
		time.Unix(5, 0), time.Unix(6, 0), time.Unix(7, 0), time.Unix(8, 0), time.Unix(9, 0),
	}
	oneUser = &user{
		Name:      "Grigory Zubankov",
		Email:     "greg@acronis.com",
		CreatedAt: time.Date(1980, 3, 2, 12, 0, 0, 0, time.UTC),
	}
	tenUsers = users{oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser}
)

func makePseudoMessages(n int) []string {
	messages := make([]string, n)
	for i := range messages {
		messages[i] = fmt.Sprintf("A text that pretend to be a real message in case of length %d", i)
	}
	return messages
}

func getMessage(n int) string {
	return messages[n%len(messages)]
}
