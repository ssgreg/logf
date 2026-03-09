module github.com/ssgreg/logf/benchmarks

replace github.com/ssgreg/logf/v2 => ../

go 1.21

require (
	github.com/rs/zerolog v1.26.0
	github.com/ssgreg/logf/v2 v2.0.0-00010101000000-000000000000
	github.com/ssgreg/logrus v1.0.3
	go.uber.org/zap v1.19.1
)

require (
	github.com/sirupsen/logrus v1.8.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
