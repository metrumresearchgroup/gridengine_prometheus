module github.com/metrumresearchgroup/gridengine_prometheus

go 1.13

require (
	github.com/go-kit/log v0.2.0
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/common v0.35.0
	github.com/prometheus/exporter-toolkit v0.7.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.2
	github.com/yuriykis/gogridengine v0.0.2-0.20220907073204-aa644c6d14d6
)

replace github.com/yuriykis/gogridengine => /home/yuriy/GridEngine/gogridengine
