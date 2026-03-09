module github.com/ssgreg/logf/benchmarks

replace github.com/ssgreg/logf/v2 => ../

go 1.24.0

require (
	github.com/rs/zerolog v1.34.0
	github.com/ssgreg/logf/v2 v2.0.0-00010101000000-000000000000
	github.com/ssgreg/logrus v1.0.3
	go.uber.org/zap v1.27.1
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/term v0.40.0 // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
)
