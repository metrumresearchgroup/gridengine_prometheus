# Prometheus Exporter for Sun Grid Engine

This is a prometheus exporter for the sun grid engine meant to be run on your master nodes. It utilizes Qstat on the command line and uses the gogridengine library to serialize its XML output into native objects and then format for prometheus consumption.

## Opinions

This exporter has various opinions about how data is reported, primarily based on how it's reported into the XML structures from Qstat:

* All metrics have a "hostname" hey
    * This hostname is derivative of the Qlist Name (split by @ symbol)
    * This facilitates PromQL statements for locating / isolating queries by host
* Hostname is the primary identifier for host level metrics. This includes:
    * Load Averages
    * Resource Values:
        * mem_free
        * swap_used
        * cpu
* Job Details are recorded with the following labels:
    * hostname
    * Job Jumber
    * Job Name
    * Owner
* Values that fit into this category are:
    * State (Running = 1 , Not = 0)
    * Priority
    * Slots

## Testing
For the sake of testing, there's an environment variable called `TEST` that if set to `"true"`, will cause the collector to bypass trying to run the command line output and generate XML (2 instances) with randomly generated and some invalid information. The invalid information is for unit testing purposes, but also to ensure that base values are still reported by the collector. This is exceptionally beneficial if you're looking to write cusotm grafana dashboards, as you can setup prometheus, the collector, and grafana in a local compose instance to basically consume generated data. 