[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=metrumresearchgroup_gridengine_prometheus&metric=alert_status)](https://sonarcloud.io/dashboard?id=metrumresearchgroup_gridengine_prometheus)

[![Coverage Status](https://coveralls.io/repos/github/metrumresearchgroup/gridengine_prometheus/badge.svg?branch=master)](https://coveralls.io/github/metrumresearchgroup/gridengine_prometheus?branch=master)

# Prometheus Exporter for Sun Grid Engine

This is a Prometheus exporter for the Sun Grid Engine meant to be run on your master nodes. It utilizes Qstat on the command line and uses the gogridengine library to serialize its XML output into native objects and then format for prometheus consumption. As long as the path for the executing user contains qstat, everything should work as the command execution inherits everything from the user. 

# Environment Variables
`TEST`: `true` for test mode which will not attempt to reach out to the command line but will rather generate data. 
`LISTEN_PORT` : Defines what port the application should listen on

# Running
There is one optional flag available to the binary, which is `-pidfile`. This should indicate where the pidfile for the application should be placed, and primarily services to facilitate service managers such as uptstart or systemd.

## Default
`./gridengine_prometheus`
This will run the application on the default port (9081)

`./gridengine_prometheus -pidfile /tmp/pid.pid`

Will run the application on the default port and write it's PID into a file located at `/tmp/pid.pid`

## Opinions

This exporter has various opinions about how data is reported, primarily based on the XML structures from Qstat:

* All metrics have a "hostname" hey
    * This hostname is derivative of the Qlist Name (split by @ symbol)
    * This facilitates PromQL statements for locating / isolating queries by host
* Hostname is the primary identifier for host level metrics. This includes:
    * Load Averages
    * Resource Values:
        * mem_free
        * swap_used
        * cpu ...
* Job Details are recorded with the following labels:
    * hostname
    * Job Jumber
    * Job Name
    * Owner
* Values that fit into this category are:
    * State (Running = 1 , Not = 0)
    * Priority
    * Slots

With these labels, it should be easy to create variable driven dashboards to allow scientists to drill down to their specific jobs across any or all hosts at a time. 

## Testing
For the sake of testing, there's an environment variable called `TEST` that if set to `"true"`, will cause the collector to bypass trying to run the command line output and generate XML (2 instances) with some static and some invalid information. The invalid information is for unit testing purposes, but also to ensure that base values are still reported by the collector. This is exceptionally beneficial if you're looking to write custom grafana dashboards, as you can setup prometheus, the collector, and grafana in a local compose instance to basically consume generated data. 

## Grafana

If you want to work with grafana or try the existing dashboards, the docker-compose file in this directory will setup :

* Prometheus
* Grafana
* Exporter

The exporter will be running in test mode and will generate a mix of static / non-static content. Prometheus is auto configured to scrape the exporter by name, and grafana is set with the magical username / pass of "admin / admin" although it'll make you change it on first setup. 