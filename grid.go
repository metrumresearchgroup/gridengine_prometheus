package gridengine_prometheus

import (
	"encoding/xml"
	"os"
	"strconv"
	"strings"

	"github.com/metrumresearchgroup/gogridengine"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

//GridEngine is the default struct we will use for collection
type GridEngine struct {
	TotalSlots    *prometheus.Desc
	UsedSlots     *prometheus.Desc
	ReservedSlots *prometheus.Desc
	LoadAverage   *prometheus.Desc
	//Resources
	FreeMemory     *prometheus.Desc
	UsedMemory     *prometheus.Desc
	TotalMemory    *prometheus.Desc
	CPUUtilization *prometheus.Desc
	//Job Details
	JobState    *prometheus.Desc
	JobPriority *prometheus.Desc
	JobSlots    *prometheus.Desc
	JobErrors   *prometheus.Desc
}

func NewGridEngine() *GridEngine {

	hostLabels := []string{
		"hostname",
		"queue",
		"qtype",
		"state",
	}

	jobLabels := []string{
		"hostname",
		"queue",
		"name",
		"owner",
		"job_number",
		"task_id",
		"state",
	}

	return &GridEngine{
		TotalSlots: prometheus.NewDesc(
			"total_slots_count",
			"Total Number of slots available to the host",
			hostLabels,
			nil),
		UsedSlots: prometheus.NewDesc(
			"used_slots_count",
			"Number of used slots on host",
			hostLabels,
			nil),
		ReservedSlots: prometheus.NewDesc(
			"reserved_slots_count",
			"Number of reserved slots on host",
			hostLabels,
			nil),
		LoadAverage: prometheus.NewDesc(
			"sge_load_average",
			"Load average of this specific SGE host",
			hostLabels,
			nil),
		FreeMemory: prometheus.NewDesc(
			"free_memory_bytes",
			"Number of bytes in free memory",
			hostLabels,
			nil),
		UsedMemory: prometheus.NewDesc(
			"sge_used_memory_bytes",
			"Number of bytes in used memory",
			hostLabels,
			nil),
		TotalMemory: prometheus.NewDesc(
			"sge_total_memory_bytes",
			"Number of bytes in total memory",
			hostLabels,
			nil),
		CPUUtilization: prometheus.NewDesc(
			"sge_cpu_utilization_percent",
			"Decimal representing total CPU utilization on host",
			hostLabels,
			nil),
		JobState: prometheus.NewDesc(
			"job_state_value",
			"Indicates whether job is running (1) or not (0)",
			jobLabels,
			nil),
		JobPriority: prometheus.NewDesc(
			"job_priority_value",
			"Qstat priority for given job",
			jobLabels,
			nil),
		JobSlots: prometheus.NewDesc(
			"job_slots_count",
			"Number of slots on the selected job",
			jobLabels,
			nil),
		JobErrors: prometheus.NewDesc(
			"job_errors",
			"Jobs that are reported in an errored or anomalous state",
			jobLabels,
			nil),
	}
}

//Describe provides prometheus with descriptions and details (not values) of each metric
func (collector *GridEngine) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.TotalSlots
	ch <- collector.UsedSlots
	ch <- collector.ReservedSlots
	//Resources
	ch <- collector.LoadAverage
	ch <- collector.FreeMemory
	ch <- collector.UsedMemory
	ch <- collector.TotalMemory
	ch <- collector.CPUUtilization
	//Job Components -> Additional Labels for identification
	ch <- collector.JobState
	ch <- collector.JobPriority
	ch <- collector.JobSlots
}

//Collect does all the work of actually generating and feeding metrics into the channel
func (collector *GridEngine) Collect(ch chan<- prometheus.Metric) {

	//How to get the XML String
	x, err := gogridengine.GetQstatOutput(make(map[string]string))
	if err != nil {
		log.WithError(err).Error("There was an error processing the XML output")
		return
	}

	ji := gogridengine.JobInfo{}

	err = xml.Unmarshal([]byte(x), &ji)

	if err != nil {
		log.WithError(err).Error("Unable to marshal the XML cleanly into an object")
		return
	}

	//Now to begin iterating over the QueueList components
	for _, ql := range ji.QueueInfo.Queues {
		//Assumes all.q@ip-172-16-2-102.us-west-2.compute.internal structure
		pieces := strings.Split(ql.Name, "@")
		queue := pieces[0]
		hostname := pieces[1]
		qtype := ql.QType
		state := ql.State
		ch <- prometheus.MustNewConstMetric(collector.UsedSlots, prometheus.GaugeValue, float64(ql.SlotsUsed), hostname, queue, qtype, state)
		ch <- prometheus.MustNewConstMetric(collector.ReservedSlots, prometheus.GaugeValue, float64(ql.SlotsReserved), hostname, queue, qtype, state)
		ch <- prometheus.MustNewConstMetric(collector.TotalSlots, prometheus.GaugeValue, float64(ql.SlotsTotal), hostname, queue, qtype, state)
		ch <- prometheus.MustNewConstMetric(collector.LoadAverage, prometheus.GaugeValue, ql.LoadAverage, hostname, queue, qtype, state)

		FreeMemory, err := ql.Resources.FreeMemory()

		if err != nil {
			log.WithError(err).Error("There was an error extracting Free Memory from the resource list")
			FreeMemory = gogridengine.StorageValue{
				Bytes: 0,
			}
		}

		ch <- prometheus.MustNewConstMetric(collector.FreeMemory, prometheus.GaugeValue, float64(FreeMemory.Bytes), hostname, queue)

		UsedMemory, err := ql.Resources.MemoryUsed()

		if err != nil {
			log.WithError(err).Error("There was an error extracting Used Memory from the resource list")
			UsedMemory = gogridengine.StorageValue{
				Bytes: 0,
			}
		}

		ch <- prometheus.MustNewConstMetric(collector.UsedMemory, prometheus.GaugeValue, float64(UsedMemory.Bytes), hostname, queue)

		TotalMemory, err := ql.Resources.TotalMemory()

		if err != nil {
			log.WithError(err).Error("There was an error extracting Total Memory from the resource list")
			TotalMemory = gogridengine.StorageValue{
				Bytes: 0,
			}
		}

		ch <- prometheus.MustNewConstMetric(collector.TotalMemory, prometheus.GaugeValue, float64(TotalMemory.Bytes), hostname, queue)

		CPUUtilization, err := ql.Resources.CPU()

		if err != nil {
			log.WithError(err).Error("There was an error extracting CPU Utilization from the resource list")
			CPUUtilization = 0
		}

		ch <- prometheus.MustNewConstMetric(collector.CPUUtilization, prometheus.GaugeValue, CPUUtilization, hostname, queue)

		//Iterate over Running Jobs
		for _, j := range ql.JobList {
			processJob(j, ch, collector, hostname, queue)
		}
	}

	for _, j := range ji.PendingJobs.JobList {
		//Process the hostname as the master
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "localhost"
		}
		processJob(j, ch, collector, hostname, "pending")
	}

}

func processJob(j gogridengine.Job, ch chan<- prometheus.Metric, collector *GridEngine, hostname string, queue string) {
	name := j.JobName
	owner := j.JobOwner
	number := strconv.FormatInt(j.JBJobNumber, 10)
	taskID := strconv.Itoa(int(j.Tasks.TaskID))

	ch <- prometheus.MustNewConstMetric(collector.JobState, prometheus.GaugeValue, float64(gogridengine.IsJobRunning(j)), hostname, queue, name, owner, number, taskID, j.State)
	ch <- prometheus.MustNewConstMetric(collector.JobPriority, prometheus.GaugeValue, j.JATPriority, hostname, queue, name, owner, number, taskID, j.State)
	ch <- prometheus.MustNewConstMetric(collector.JobSlots, prometheus.GaugeValue, float64(j.Slots), hostname, queue, name, owner, number, taskID, j.State)
	ch <- prometheus.MustNewConstMetric(collector.JobErrors, prometheus.GaugeValue, float64(gogridengine.IsJobInErrorState(j)), hostname, queue, name, owner, number, taskID, j.State)
}
