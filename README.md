[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=metrumresearchgroup_gridengine_prometheus&metric=alert_status)](https://sonarcloud.io/dashboard?id=metrumresearchgroup_gridengine_prometheus)

[![Coverage Status](https://coveralls.io/repos/github/metrumresearchgroup/gridengine_prometheus/badge.svg?branch=master)](https://coveralls.io/github/metrumresearchgroup/gridengine_prometheus?branch=master)

# Prometheus Exporter for Sun Grid Engine

This is a Prometheus exporter for the Sun Grid Engine meant to be run on your master nodes. It utilizes Qstat on the command line and uses the gogridengine library to serialize its XML output into native objects and then format for prometheus consumption.

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