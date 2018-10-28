package logf

import (
	"io/ioutil"
	"sync"
)

type ChannelWriter interface {
	Write(Entry)
}

type ChannelCloser interface {
	Close()
}

type Channel interface {
	ChannelWriter
	ChannelCloser
}

func NewBufferedChannel(capacity int, appenders ...Appender) Channel {
	ch := make(chan Entry, capacity)
	wg := sync.WaitGroup{}
	wg.Add(1)

	l := &bufferedChannel{
		channel: ch,
		wg:      &wg,
		buf:     NewBuffer(ioutil.Discard),
		formatter: NewJSONFormatter(&FormatterConfig{
			FieldKeyLevel:  "level",
			FieldKeyMsg:    "msg",
			FieldKeyTime:   "ts",
			FieldKeyName:   "logger",
			FieldKeyCaller: "caller",
			FormatTime:     RFC3339TimeFormatter,
		}),
	}
	go output(l)

	return l
}

type bufferedChannel struct {
	sync.Mutex
	channel chan Entry
	wg      *sync.WaitGroup

	formatter Formatter
	buf       *Buffer
	counter   int32
}

func (l *bufferedChannel) Close() {
	close(l.channel)
	l.wg.Wait()
}

// Log TODO
func (l *bufferedChannel) Write(e Entry) {
	l.channel <- e
	// l.Lock()
	// defer l.Unlock()
	// l.formatter.Format(l.buf, e)
}

func output(l *bufferedChannel) {
	defer l.wg.Done()

	var e Entry
	var ok bool
	for {
		select {
		case e, ok = <-l.channel:
		default:
			// l.appender.Flush()
			e, ok = <-l.channel
		}
		if !ok {
			break
		}
		_ = e
		l.formatter.Format(l.buf, e)

		// spew.Dump(e)
		// l.appender.Append(e)
	}

	// if err := l.appender.Flush(); err != nil {
	// 	fmt.Fprint(os.Stderr, err)
	// }
	// if err := l.appender.Close(); err != nil {
	// 	fmt.Fprint(os.Stderr, err)
	// }
}
