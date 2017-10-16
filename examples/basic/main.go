package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/ssgreg/bottleneck"
	// "github.com/ssgreg/logf"
)

type test interface {
	Test()
}

type Greg struct {
	B int
}

func (g *Greg) TakeSnapshot() interface{} {
	return g
}

type TimeSnapshotter struct {
	Time time.Time
}

func (g TimeSnapshotter) TakeSnapshot() interface{} {
	return g.Time.Unix()
}

type GregR int

// type MyFields struct {
// 	fields []int

// 	base *MyFields
// }

// func (f *MyFields) Get() []int {
// 	if f.base != nil {
// 		return append(f.base.Get(), f.fields...)
// 	}
// 	return f.fields
// }

// type NewFields struct {
// 	fields []int
// 	base   *NewFields
// }

// func (f *NewFields) Get() ([]int, *NewFields) {
// 	return f.fields, f.base
// }

func main() {
	// lgr := &log.Logger{
	// 	Handler: json.New(ioutil.Discard),
	// 	Level:   log.ErrorLevel,
	// }
	// lgr.Info("test")

	// logger := logf.New(logf.LoggerParams{
	// 	Level:    logf.InfoLevel,
	// 	Capacity: 100,
	// 	Appender: logf.NewFileAppender("dat",
	// 		&logf.JSONFormatter{
	// 			TimestampFormat: time.RFC3339Nano,
	// 		}),
	// }) //.WithField("Greg", 5.4)
	// defer logger.Close()

	// logger.Info().Msgf("%v", errors.New("Test"))

	w, _ := os.Create("dat")
	defer w.Close()

	// logger1 := logrus.New()
	// logger1.SetLevel(logrus.DebugLevel)
	// logger1.Out = w
	// logger1.Formatter = &logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano}
	// logger := logger1.WithFields(logrus.Fields{"Greg": 5.4, "struct": Greg{1}})

	// logger.Debug("debug")
	// logger.Info("info")
	// logger.Warnf("test %d", 1)
	// logger.Error("err")
	// //	logger.Fatal("crit")
	// logger.Panic("test")

	// logger := log.NewLogger3(w, "models", log.NewJSONFormatter("bench"))
	// logger.SetLevel(log.LevelInfo)
	// logger.Info(fmt.Sprintf("%s \"%d", "Greeg", 1), "Greg", 5.4, "struct", Greg{1}, "Field", &Greg{3})

	//	logger, _ := zap.NewProduction()
	// cfg := zap.NewProductionConfig()
	// cfg.OutputPaths = []string{"dat1"}
	// logger, _ := cfg.Build()
	// defer logger.Sync()

	// logger := zerolog.New(w).With().Timestamp().Logger().Level(zerolog.ErrorLevel)
	logger := zerolog.New(w).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	lg := logger.With().Str("string", "123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,123456789,").Logger()

	lg.Info().Msg("1")
	lg.Info().Msg("2")

	wgs := sync.WaitGroup{}
	wgs.Add(1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	bc := bottleneck.NewCalculator()

	// t := time.Now()
	for j := 0; j < 1; j++ {
		go func(j1 int) {
			wgs.Wait()
			for i := 0; i < 0; i++ {
				//				logger.WithField("Field", &Greg{j1*100000 + i}).Infof(fmt.Sprint(j1*100000 + i))
				// logger.WithFields([]logf.Field{{"Uint", j1*100000 + i}, {"Float", 34875.3459}}...).Info("Test")

				// logger.Info("Test")

				// logger.WithFields(
				// 	logf.FieldInt("Uint", j1*100000+i),
				// 	logf.FieldFloat("Float", 34875.3459),
				// // logf.FieldAny("Greg", Greg{j1*100000 + i}),
				// ).Msg("Test")

				// logger.Info().WithFields(
				// 	logf.FieldInt("Uint", j1*100000+i),
				// 	logf.FieldFloat("Float", 34875.3459),
				// ).Msg("test")

				// logger.Info().Int("Uint", j1*100000+i).Float64("Float", 34875.3459).Msg("Test")

				// logger.Pinfo().WithInt("Uint", i).WithFloat64("Float", 34875.3485).Msg("Test")
				// bc.TimeSlice(1)
				// logger.Pinfo().Msg("Test")
				// bc.TimeSlice(0)

				// logger.WithFields1(func() []logf.Field {
				// 	return []logf.Field{logf.FieldInt("Uint", j1*100000+i), logf.FieldFloat("Float", 34875.3459)}
				// })

				// logger.WithField("Uint", j1*100000+i).Info(strconv.Itoa(j1*100000 + i))

				// "Greg", TimeSnapshotter{time.Now()}
				// logger.Info(fmt.Sprint(j1*100000+i),
				// 	//"failed to fetch URL",
				// 	// Structured context as strongly typed Field values.
				// 	zap.Float64("Greg", 5.4),
				// 	// zap.Any("struct", Greg{1}),
				// 	zap.Any("Field", &Greg{j1*100000 + i}),
				// )

				// if i == 10000 {
				// 	time.Sleep(time.Second * 10)
				// }
				// if i == 20000 {
				// 	time.Sleep(time.Second * 10)
				// }

				// logger.Info(fmt.Sprintf("%s \"%d", "Greeg", 1), "Greg", 5.4, "struct", Greg{1}, "Field", &Greg{3})
				// logger.WithFields(logrus.Fields{"Field": "greg"}).Infof("%s \"%d", "Greeg", 1)

				// logger.WithField("Field", &Greg{3}).Infof("%s \n\"%d", "Greeg", 1)
				// logger.WithFields(logf.Fields{"Field": 3}).Infof("%s \"%d", "Greeg", 1)

				// logger.WithPlainFields("Field", "greg").Infof("%s \"%d", "Greeg", 1)
				// logger.Info("aa", "bb")
			}
			wg.Done()
		}(j)
	}
	bt := time.Now()
	wgs.Done()

	wg.Wait()
	fmt.Fprint(os.Stderr, "final:", time.Now().Sub(bt), "\n")
	fmt.Println(bc.Stats())

}
