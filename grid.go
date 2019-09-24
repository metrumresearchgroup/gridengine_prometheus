package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/metrumresearchgroup/gogridengine"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
}

func newGridEngine() *GridEngine {
	return &GridEngine{
		TotalSlots: prometheus.NewDesc(
			"total_slots_count",
			"Total Number of slots available to the host",
			[]string{"hostname"},
			nil),
		UsedSlots: prometheus.NewDesc(
			"used_slots_count",
			"Number of used slots on host",
			[]string{"hostname"},
			nil),
		ReservedSlots: prometheus.NewDesc(
			"reserved_slots_count",
			"Number of reserved slots on host",
			[]string{"hostname"},
			nil),
		LoadAverage: prometheus.NewDesc(
			"sge_load_average",
			"Load average of this specific SGE host",
			[]string{"hostname"},
			nil),
		FreeMemory: prometheus.NewDesc(
			"free_memory_bytes",
			"Number of bytes in free memory",
			[]string{"hostname"},
			nil),
		UsedMemory: prometheus.NewDesc(
			"sge_used_memory_bytes",
			"Number of bytes in used memory",
			[]string{"hostname"},
			nil),
		TotalMemory: prometheus.NewDesc(
			"sge_total_memory_bytes",
			"Number of bytes in total memory",
			[]string{"hostname"},
			nil),
		CPUUtilization: prometheus.NewDesc(
			"sge_cpu_utilization_percent",
			"Decimal representing total CPU utilization on host",
			[]string{"hostname"},
			nil),
		JobState: prometheus.NewDesc(
			"job_state_value",
			"Indicates whether job is running (1) or not (0)",
			[]string{"hostname", "name", "owner", "job_number"},
			nil),
		JobPriority: prometheus.NewDesc(
			"job_priority_value",
			"Qstat priority for given job",
			[]string{"hostname", "name", "owner", "job_number"},
			nil),
		JobSlots: prometheus.NewDesc(
			"job_slots_count",
			"Number of slots on the selected job",
			[]string{"hostname", "name", "owner", "job_number"},
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
	x, err := getQstatOutput()
	if err != nil {
		log.Error("There was an error processing the XML output", err)
		return
	}

	ji := gogridengine.JobInfo{}

	err = xml.Unmarshal([]byte(x), &ji)

	if err != nil {
		log.Error("Unable to marshal the XML cleanly into an object", err)
		return
	}

	//Now to begin iterating over the QueueList components
	for _, ql := range ji.QueueInfo.Queues {
		//Assumes all.q@ip-172-16-2-102.us-west-2.compute.internal structure
		hostname := strings.Split(ql.Name, "@")[1]

		ch <- prometheus.MustNewConstMetric(collector.UsedSlots, prometheus.GaugeValue, float64(ql.SlotsUsed), hostname)
		ch <- prometheus.MustNewConstMetric(collector.ReservedSlots, prometheus.GaugeValue, float64(ql.SlotsReserved), hostname)
		ch <- prometheus.MustNewConstMetric(collector.TotalSlots, prometheus.GaugeValue, float64(ql.SlotsTotal), hostname)
		ch <- prometheus.MustNewConstMetric(collector.LoadAverage, prometheus.GaugeValue, ql.LoadAverage, hostname)

		FreeMemory, err := ql.Resources.FreeMemory()

		if err != nil {
			log.Error("There was an error extracting Free Memory from the resource list", err)
			FreeMemory = gogridengine.StorageValue{
				Bytes: 0,
			}
		}

		ch <- prometheus.MustNewConstMetric(collector.FreeMemory, prometheus.GaugeValue, float64(FreeMemory.Bytes), hostname)

		UsedMemory, err := ql.Resources.MemoryUsed()

		if err != nil {
			log.Error("There was an error extracting Used Memory from the resource list", err)
			UsedMemory = gogridengine.StorageValue{
				Bytes: 0,
			}
		}

		ch <- prometheus.MustNewConstMetric(collector.UsedMemory, prometheus.GaugeValue, float64(UsedMemory.Bytes), hostname)

		TotalMemory, err := ql.Resources.TotalMemory()

		if err != nil {
			log.Error("There was an error extracting Total Memory from the resource list", err)
			TotalMemory = gogridengine.StorageValue{
				Bytes: 0,
			}
		}

		ch <- prometheus.MustNewConstMetric(collector.TotalMemory, prometheus.GaugeValue, float64(TotalMemory.Bytes), hostname)

		CPUUtilization, err := ql.Resources.CPU()

		if err != nil {
			log.Error("There was an error extracting CPU Utilization from the resource list", err)
			CPUUtilization = 0
		}

		ch <- prometheus.MustNewConstMetric(collector.CPUUtilization, prometheus.GaugeValue, CPUUtilization, hostname)

		//Iterate over Jobs
		for _, j := range ql.JobList {
			name := j.JobName
			owner := j.JobOwner
			number := strconv.FormatInt(j.JBJobNumber, 10)

			ch <- prometheus.MustNewConstMetric(collector.JobState, prometheus.GaugeValue, float64(isJobRunning(j)), hostname, name, owner, number)
			ch <- prometheus.MustNewConstMetric(collector.JobPriority, prometheus.GaugeValue, j.JATPriority, hostname, name, owner, number)
			ch <- prometheus.MustNewConstMetric(collector.JobSlots, prometheus.GaugeValue, float64(j.Slots), hostname, name, owner, number)

		}
	}

}

func getQstatOutput() (string, error) {
	if !isTest {
		return qStatFromExec()
	}

	//Fallthrough to generation by object randomly
	return generatedQstatOputput()
}

func qStatFromExec() (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	//Cowardly cancel on any other exit mode
	defer cancel()

	command := exec.CommandContext(ctx, "qstat -F xml")
	output := &bytes.Buffer{}
	command.Stdout = output
	err := command.Run()
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func generatedQstatOputput() (string, error) {

	ji := gogridengine.JobInfo{
		XMLName: xml.Name{
			Local: "job_info",
		},
		QueueInfo: gogridengine.QueueInfo{
			XMLName: xml.Name{
				Local: "queue_info",
			},
			Queues: []gogridengine.QueueList{
				{
					XMLName: xml.Name{
						Local: "Queue-List",
					},
					Name:          "all.q@testing.local", //Always needs the @ symbol
					SlotsTotal:    int32(random.Int()),
					SlotsUsed:     int32(random.Int()),
					SlotsReserved: int32(random.Int()),
					LoadAverage:   float64(random.Float64()),
					Resources: gogridengine.ResourceList{
						{
							Name:  "load_average",
							Type:  "hl",
							Value: fmt.Sprintf("%f", random.Float64()),
						},
						{
							Name:  "num_proc",
							Type:  "ag",
							Value: strconv.Itoa(random.Int()),
						},
						{
							Name:  "mem_free",
							Type:  "af",
							Value: fmt.Sprintf("%f", random.Float64()) + "M",
						},
						{
							Name:  "swap_free",
							Type:  "ae",
							Value: fmt.Sprintf("%f", random.Float64()) + "G",
						},
						{
							Name:  "virtual_free",
							Type:  "ad",
							Value: fmt.Sprintf("%f", random.Float64()) + "T",
						},
						{
							Name:  "mem_used",
							Type:  "ac",
							Value: fmt.Sprintf("%f", random.Float64()) + "G",
						},
						{
							Name:  "mem_total",
							Type:  "ab",
							Value: fmt.Sprintf("%f", random.Float64()) + "M",
						},
						{
							Name:  "cpu",
							Type:  "aa",
							Value: fmt.Sprintf("%f", random.Float64()),
						},
					},
					JobList: []gogridengine.JobList{
						{
							XMLName: xml.Name{
								Local: "job_list",
							},
							State:       "running",
							JBJobNumber: int64(random.Int()),
							JATPriority: random.Float64(),
							JobName:     "Job-" + strconv.Itoa(random.Int()),
							JobOwner:    "Owner-" + strconv.Itoa(random.Int()),
							Slots:       3,
						},
					},
				},
				{
					XMLName: xml.Name{
						Local: "Queue-List",
					},
					Name: "all.q@testing.second", //Always needs the @ symbol

					Resources: gogridengine.ResourceList{},
					JobList: []gogridengine.JobList{
						{
							XMLName: xml.Name{
								Local: "job_list",
							},
							State:       "running",
							JBJobNumber: 1,
							JATPriority: 1,
							JobName:     "Second-Host-Job",
							JobOwner:    "Owner",
							Slots:       14,
						},
					},
				},
			},
		},
	}

	return ji.GetXML()
}

func isJobRunning(job gogridengine.JobList) int {

	if job.State == "running" {
		return 1
	}

	return 0
}
