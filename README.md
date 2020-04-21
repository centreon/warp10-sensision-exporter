# Warp10 Sensision exporter

## Build the exporter

Download the modules dependencies.

```
go mod download
```

Build the sensision metrics files. The tools command take the version of Warp10 in argument.

```
cd tools
go run generate_sensision_metrics.go 2.4.0
cp sensision.go ../collector
cd ..
go get -d github.com/centreon/warp10-sensision-exporter
```

Build the binary

```
go build -a -o warp10_sensision_exporter
```

## Configure Warp10 to expose Sensision

To expose Sensision, we neeed to add the following option in starting command of Warp10.

```
-D sesision.server.port=8082
```
