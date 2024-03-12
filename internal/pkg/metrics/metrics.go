// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

const (
	DefaultMetricsAddress           = ":8081"
	DefaultMinerSetMetricsAddress   = ":8082"
	DefaultMinerMetricsAddress      = ":8083"
	DefaultNodeServerMetricsAddress = ":8084"
)

var (
	// MinerCountDesc is a metric about miner object count in the cluster.
	MinerCountDesc = prometheus.NewDesc("napi_miner_items", "Count of miner objects currently at the apiserver", nil, nil)
	// MinerSetCountDesc Count of minerset object count at the apiserver.
	MinerSetCountDesc = prometheus.NewDesc("napi_minerset_items", "Count of minersets at the apiserver", nil, nil)
	// MinerInfoDesc is a metric about miner object info in the cluster.
	MinerInfoDesc = prometheus.NewDesc("napi_miner_created_timestamp_seconds", "Timestamp of the napi managed Miner creation time", []string{"name", "namespace", "spec_provider_id", "node", "api_version", "phase"}, nil)
	// MinerSetInfoDesc is a metric about miner object info in the cluster.
	MinerSetInfoDesc = prometheus.NewDesc("napi_minerset_created_timestamp_seconds", "Timestamp of the napi managed Minerset creation time", []string{"name", "namespace", "api_version"}, nil)

	// MinerSetStatusAvailableReplicasDesc is the information of the Minerset's status for available replicas.
	MinerSetStatusAvailableReplicasDesc = prometheus.NewDesc("napi_miner_set_status_replicas_available", "Information of the napi managed Minerset's status for available replicas", []string{"name", "namespace"}, nil)

	// MinerSetStatusReadyReplicasDesc is the information of the Minerset's status for ready replicas.
	MinerSetStatusReadyReplicasDesc = prometheus.NewDesc("napi_miner_set_status_replicas_ready", "Information of the napi managed Minerset's status for ready replicas", []string{"name", "namespace"}, nil)

	// MinerSetStatusReplicasDesc is the information of the Minerset's status for replicas.
	MinerSetStatusReplicasDesc = prometheus.NewDesc("napi_miner_set_status_replicas", "Information of the napi managed Minerset's status for replicas", []string{"name", "namespace"}, nil)

	// MinerCollectorUp is a Prometheus metric, which reports reflects successful collection and reporting of all the metrics.
	MinerCollectorUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "napi_mao_collector_up",
		Help: "Node API Controller metrics are being collected and reported successfully",
	}, []string{"kind"})

	failedInstanceCreateCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "napi_instance_create_failed",
			Help: "Number of times provider instance create has failed.",
		}, []string{"name", "namespace", "reason"},
	)

	failedInstanceUpdateCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "napi_instance_update_failed",
			Help: "Number of times provider instance update has failed.",
		}, []string{"name", "namespace", "reason"},
	)

	failedInstanceDeleteCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "napi_instance_delete_failed",
			Help: "Number of times provider instance delete has failed.",
		}, []string{"name", "namespace", "reason"},
	)
)

// Metrics for use in the Miner controller.
var (
	// MinerPhaseTransitionSeconds is a metric to capute the time between a Miner being created and entering a particular phase.
	MinerPhaseTransitionSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "napi_miner_phase_transition_seconds",
			Help:    "Number of seconds between Miner creation and Miner transition to a phase.",
			Buckets: []float64{5, 10, 20, 30, 60, 90, 120, 180, 240, 300, 360, 480, 600},
		}, []string{"phase"},
	)
)

// MinerCollector is implementing prometheus.Collector interface.
type MinerCollector struct {
	client    client.Client
	namespace string
}

// MinerLabels is the group of labels that are applied to the miner metrics.
type MinerLabels struct {
	Name      string
	Namespace string
	Reason    string
}

func NewMinerCollector(client client.Client, namespace string) *MinerCollector {
	return &MinerCollector{
		client:    client,
		namespace: namespace,
	}
}

// Collect is method required to implement the prometheus.Collector(prometheus/client_golang/prometheus/collector.go) interface.
func (mc *MinerCollector) Collect(ch chan<- prometheus.Metric) {
	mc.collectMinerMetrics(ch)
	mc.collectMinerSetMetrics(ch)
}

// Describe implements the prometheus.Collector interface.
func (mc MinerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- MinerCountDesc
	ch <- MinerSetCountDesc
}

// Collect implements the prometheus.Collector interface.
func (mc MinerCollector) collectMinerMetrics(ch chan<- prometheus.Metric) {
	minerList, err := mc.listMiners()
	if err != nil {
		MinerCollectorUp.With(prometheus.Labels{"kind": "napi_miner_items"}).Set(float64(0))
		return
	}
	MinerCollectorUp.With(prometheus.Labels{"kind": "napi_miner_items"}).Set(float64(1))

	for _, miner := range minerList {
		podName := ""
		if miner.Status.PodRef != nil {
			podName = miner.Status.PodRef.Name
		}
		// Only gather metrics for miners with a phase.  This indicates
		// That the miner-controller is running on this cluster.
		phase := miner.Status.Phase
		if phase != "" {
			ch <- prometheus.MustNewConstMetric(
				MinerInfoDesc,
				prometheus.GaugeValue,
				float64(miner.ObjectMeta.GetCreationTimestamp().Time.Unix()),
				miner.ObjectMeta.Name,
				miner.ObjectMeta.Namespace,
				podName,
				miner.TypeMeta.APIVersion,
				phase,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(MinerCountDesc, prometheus.GaugeValue, float64(len(minerList)))
	klog.V(4).InfoS("collectminerMetrics exit")
}

// collectMinerSetMetrics is method to collect minerSet related metrics.
func (mc MinerCollector) collectMinerSetMetrics(ch chan<- prometheus.Metric) {
	minerSetList, err := mc.listMinerSets()
	if err != nil {
		MinerCollectorUp.With(prometheus.Labels{"kind": "napi_minerset_items"}).Set(float64(0))
		return
	}
	MinerCollectorUp.With(prometheus.Labels{"kind": "napi_minerset_items"}).Set(float64(1))
	ch <- prometheus.MustNewConstMetric(MinerSetCountDesc, prometheus.GaugeValue, float64(len(minerSetList)))

	for _, minerSet := range minerSetList {
		ch <- prometheus.MustNewConstMetric(
			MinerSetInfoDesc,
			prometheus.GaugeValue,
			float64(minerSet.GetCreationTimestamp().Time.Unix()),
			minerSet.Name, minerSet.Namespace, minerSet.TypeMeta.APIVersion,
		)
		ch <- prometheus.MustNewConstMetric(
			MinerSetStatusAvailableReplicasDesc,
			prometheus.GaugeValue,
			float64(minerSet.Status.AvailableReplicas),
			minerSet.Name, minerSet.Namespace,
		)
		ch <- prometheus.MustNewConstMetric(
			MinerSetStatusReadyReplicasDesc,
			prometheus.GaugeValue,
			float64(minerSet.Status.ReadyReplicas),
			minerSet.Name, minerSet.Namespace,
		)
		ch <- prometheus.MustNewConstMetric(
			MinerSetStatusReplicasDesc,
			prometheus.GaugeValue,
			float64(minerSet.Status.Replicas),
			minerSet.Name, minerSet.Namespace,
		)
	}
}

func (mc MinerCollector) listMiners() ([]v1beta1.Miner, error) {
	miners := &v1beta1.MinerList{}
	if err := mc.client.List(context.Background(), miners, client.InNamespace(mc.namespace)); err != nil {
		klog.ErrorS(err, "Failed to list miners")
		return nil, err
	}

	return miners.Items, nil
}

func (mc MinerCollector) listMinerSets() ([]v1beta1.MinerSet, error) {
	minersets := &v1beta1.MinerSetList{}
	if err := mc.client.List(context.Background(), minersets, client.InNamespace(mc.namespace)); err != nil {
		klog.ErrorS(err, "Failed to list miner sets")
		return nil, err
	}

	return minersets.Items, nil
}

func RegisterFailedInstanceCreate(labels *MinerLabels) {
	failedInstanceCreateCount.With(prometheus.Labels{
		"name":      labels.Name,
		"namespace": labels.Namespace,
		"reason":    labels.Reason,
	}).Inc()
}

func RegisterFailedInstanceUpdate(labels *MinerLabels) {
	failedInstanceUpdateCount.With(prometheus.Labels{
		"name":      labels.Name,
		"namespace": labels.Namespace,
		"reason":    labels.Reason,
	}).Inc()
}

func RegisterFailedInstanceDelete(labels *MinerLabels) {
	failedInstanceDeleteCount.With(prometheus.Labels{
		"name":      labels.Name,
		"namespace": labels.Namespace,
		"reason":    labels.Reason,
	}).Inc()
}

func init() {
	prometheus.MustRegister(KratosMetricSeconds, KratosServerMetricRequests, MinerCollectorUp)
	metrics.Registry.MustRegister(MinerPhaseTransitionSeconds)
	metrics.Registry.MustRegister(
		failedInstanceCreateCount,
		failedInstanceUpdateCount,
		failedInstanceDeleteCount,
	)
}
