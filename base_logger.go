package logf

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Entry struct {
	time   time.Time
	level  Level
	format string
	args   []interface{}
	fields FieldGetter
}

// LoggerParams TODO
type LoggerParams struct {
	Level     Level
	Capacity  int
	Appender  Appender
	Appenders []Appender
}

type FieldGetter interface {
	Fields() ([]Field, Logger)
}

type baseLogger interface {
	Log(lv Level, format string, args []interface{}, fields FieldGetter)

	Level() Level

	Close()
}

func newBaseLogger(params LoggerParams) baseLogger {
	ch := make(chan *Entry, params.Capacity)
	wg := sync.WaitGroup{}
	wg.Add(1)

	l := &baseLoggerImpl{
		level:    params.Level,
		channel:  ch,
		wg:       &wg,
		appender: &MultiAppender{append(params.Appenders, params.Appender)},
	}
	go output(l)

	return l
}

type baseLoggerImpl struct {
	level      Level
	channel    chan *Entry
	wg         *sync.WaitGroup
	appender   Appender
	closeMutex sync.Mutex
	closed     bool
}

// Wait TODO
func (l *baseLoggerImpl) Close() {
	l.closeMutex.Lock()
	defer l.closeMutex.Unlock()
	if !l.closed {
		close(l.channel)
		l.wg.Wait()
		l.closed = true
	}
}

// Log TODO
func (l *baseLoggerImpl) Log(lv Level, format string, args []interface{}, fields FieldGetter) {
	data := &Entry{format: format, args: args, fields: fields}
	_ = data
	for i, a := range data.args {
		data.args[i] = TakeSnapshot(a)
	}
	data.time = time.Now()
	data.level = lv
	l.channel <- data
}

// Level TODO
func (l *baseLoggerImpl) Level() Level {
	return l.level
}

// add bottlenecks
func output(l *baseLoggerImpl) {
	defer l.wg.Done()

	// time.Sleep(time.Second * 5)
	// bt := time.Now()

	// for entry := range l.channel {
	// 	l.appender.Append(entry)
	// }

	// bc := NewBottleneckCalculator(2)
	for {
		// bc.NTM(0)
		var entry *Entry
		select {
		case entry = <-l.channel:
		default:
			select {
			case entry = <-l.channel:
			case <-time.After(time.Second):
				// bc.TimeSlice(0)
				// bc.NTM(1)
				l.appender.Flush()
				// bc.TimeSlice(1)
				// bc.NTM(0)
				entry = <-l.channel
			}

		}
		if entry == nil {
			break
		}

		// TODO: подумать над оптимизацией контекста (форматировать одни и те же поля только один раз)

		// bc.NTM(1)
		// bc.TimeSlice(0)
		l.appender.Append(entry)
		// entry.fields.(*extLogger).fields = entry.fields.(*extLogger).fields[:0]
		// loggerPool.Put(entry.fields)
		// bc.TimeSlice(1)
	}

	// fmt.Println(CounterPool)

	// fmt.Fprint(os.Stderr, time.Now().Sub(bt), "\n")
	// bs := bc.Stats()
	// fmt.Fprint(os.Stderr, bs[0], bs[1], "\n")
	if err := l.appender.Flush(); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	if err := l.appender.Close(); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
