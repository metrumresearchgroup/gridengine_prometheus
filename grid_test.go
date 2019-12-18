package main

import (
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_newGridEngine(t *testing.T) {
	tests := []struct {
		name string
		want *GridEngine
	}{
		{
			name: "Normal operation Mode",
			want: &GridEngine{
				TotalSlots: prometheus.NewDesc(
					"total_slots_count",
					"Total Number of slots available to the host",
					[]string{"hostname", "queue"},
					nil),
				UsedSlots: prometheus.NewDesc(
					"used_slots_count",
					"Number of used slots on host",
					[]string{"hostname", "queue"},
					nil),
				ReservedSlots: prometheus.NewDesc(
					"reserved_slots_count",
					"Number of reserved slots on host",
					[]string{"hostname", "queue"},
					nil),
				LoadAverage: prometheus.NewDesc(
					"sge_load_average",
					"Load average of this specific SGE host",
					[]string{"hostname", "queue"},
					nil),
				FreeMemory: prometheus.NewDesc(
					"free_memory_bytes",
					"Number of bytes in free memory",
					[]string{"hostname", "queue"},
					nil),
				UsedMemory: prometheus.NewDesc(
					"sge_used_memory_bytes",
					"Number of bytes in used memory",
					[]string{"hostname", "queue"},
					nil),
				TotalMemory: prometheus.NewDesc(
					"sge_total_memory_bytes",
					"Number of bytes in total memory",
					[]string{"hostname", "queue"},
					nil),
				CPUUtilization: prometheus.NewDesc(
					"sge_cpu_utilization_percent",
					"Decimal representing total CPU utilization on host",
					[]string{"hostname", "queue"},
					nil),
				JobState: prometheus.NewDesc(
					"job_state_value",
					"Indicates whether job is running (1) or not (0)",
					[]string{"hostname", "queue", "name", "owner", "job_number", "task_id"},
					nil),
				JobPriority: prometheus.NewDesc(
					"job_priority_value",
					"Qstat priority for given job",
					[]string{"hostname", "queue", "name", "owner", "job_number", "task_id"},
					nil),
				JobSlots: prometheus.NewDesc(
					"job_slots_count",
					"Number of slots on the selected job",
					[]string{"hostname", "queue", "name", "owner", "job_number", "task_id"},
					nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newGridEngine(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newGridEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGridEngine_Describe(t *testing.T) {
	description := newGridEngine()
	channel := make(chan *prometheus.Desc, 11)
	type fields struct {
		TotalSlots     *prometheus.Desc
		UsedSlots      *prometheus.Desc
		ReservedSlots  *prometheus.Desc
		LoadAverage    *prometheus.Desc
		FreeMemory     *prometheus.Desc
		UsedMemory     *prometheus.Desc
		TotalMemory    *prometheus.Desc
		CPUUtilization *prometheus.Desc
		JobState       *prometheus.Desc
		JobPriority    *prometheus.Desc
		JobSlots       *prometheus.Desc
	}
	type args struct {
		ch chan<- *prometheus.Desc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Normal operation",
			fields: fields{
				TotalSlots:     description.TotalSlots,
				UsedSlots:      description.UsedSlots,
				ReservedSlots:  description.ReservedSlots,
				LoadAverage:    description.LoadAverage,
				FreeMemory:     description.FreeMemory,
				UsedMemory:     description.UsedMemory,
				TotalMemory:    description.TotalMemory,
				CPUUtilization: description.CPUUtilization,
				JobState:       description.JobState,
				JobPriority:    description.JobPriority,
				JobSlots:       description.JobSlots,
			},
			args: args{
				ch: channel,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &GridEngine{
				TotalSlots:     tt.fields.TotalSlots,
				UsedSlots:      tt.fields.UsedSlots,
				ReservedSlots:  tt.fields.ReservedSlots,
				LoadAverage:    tt.fields.LoadAverage,
				FreeMemory:     tt.fields.FreeMemory,
				UsedMemory:     tt.fields.UsedMemory,
				TotalMemory:    tt.fields.TotalMemory,
				CPUUtilization: tt.fields.CPUUtilization,
				JobState:       tt.fields.JobState,
				JobPriority:    tt.fields.JobPriority,
				JobSlots:       tt.fields.JobSlots,
			}
			collector.Describe(tt.args.ch)
		})
	}
}

func TestGridEngine_Collect(t *testing.T) {
	description := newGridEngine()
	channel := make(chan prometheus.Metric, 100)
	os.Setenv("GOGRIDENGINE_TEST", "true")
	os.Setenv("GOGRIDENGINE_TEST_SOURCE", "https://gist.githubusercontent.com/shairozan/03f6f6123b11483bb17fd2c6ee95c338/raw/d65e6d603e7563b7307f9b045171ea7693c95f40/small_sge.xml")

	//Prep the entropy components
	entropy = rand.NewSource(time.Now().UnixNano())
	random = rand.New(entropy)

	type fields struct {
		TotalSlots     *prometheus.Desc
		UsedSlots      *prometheus.Desc
		ReservedSlots  *prometheus.Desc
		LoadAverage    *prometheus.Desc
		FreeMemory     *prometheus.Desc
		UsedMemory     *prometheus.Desc
		TotalMemory    *prometheus.Desc
		CPUUtilization *prometheus.Desc
		JobState       *prometheus.Desc
		JobPriority    *prometheus.Desc
		JobSlots       *prometheus.Desc
	}
	type args struct {
		ch chan<- prometheus.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Normal Operation",
			fields: fields{
				TotalSlots:     description.TotalSlots,
				UsedSlots:      description.UsedSlots,
				ReservedSlots:  description.ReservedSlots,
				LoadAverage:    description.LoadAverage,
				FreeMemory:     description.FreeMemory,
				UsedMemory:     description.UsedMemory,
				TotalMemory:    description.TotalMemory,
				CPUUtilization: description.CPUUtilization,
				JobState:       description.JobState,
				JobPriority:    description.JobPriority,
				JobSlots:       description.JobSlots,
			},
			args: args{
				ch: channel,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &GridEngine{
				TotalSlots:     tt.fields.TotalSlots,
				UsedSlots:      tt.fields.UsedSlots,
				ReservedSlots:  tt.fields.ReservedSlots,
				LoadAverage:    tt.fields.LoadAverage,
				FreeMemory:     tt.fields.FreeMemory,
				UsedMemory:     tt.fields.UsedMemory,
				TotalMemory:    tt.fields.TotalMemory,
				CPUUtilization: tt.fields.CPUUtilization,
				JobState:       tt.fields.JobState,
				JobPriority:    tt.fields.JobPriority,
				JobSlots:       tt.fields.JobSlots,
			}
			collector.Collect(tt.args.ch)
		})
	}
}
