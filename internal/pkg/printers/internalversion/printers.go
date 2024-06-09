// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package internalversion

import (
	"fmt"
	"strings"

	coordinationv1 "k8s.io/api/coordination/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/apis/coordination"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/printers"

	printersutil "github.com/superproj/onex/internal/pkg/util/printers"
	"github.com/superproj/onex/pkg/apis/apps"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// AddHandlers adds print handlers for default OneX types dealing with internal versions.
// TODO: handle errors from Handler.
func AddHandlers(h printers.PrintHandler) {
	namespaceColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Status", Type: "string", Description: "The status of the namespace"},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}
	h.TableHandler(namespaceColumnDefinitions, printNamespace)
	h.TableHandler(namespaceColumnDefinitions, printNamespaceList)

	configMapColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Data", Type: "string", Description: apiv1.ConfigMap{}.SwaggerDoc()["data"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}
	h.TableHandler(configMapColumnDefinitions, printConfigMap)
	h.TableHandler(configMapColumnDefinitions, printConfigMapList)

	minerSetColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Ready", Type: "integer", Description: v1beta1.MinerSetStatus{}.SwaggerDoc()["readyReplicas"]},
		{Name: "Current", Type: "integer", Description: v1beta1.MinerSetStatus{}.SwaggerDoc()["replicas"]},
		{Name: "Available", Type: "string", Description: v1beta1.MinerSetStatus{}.SwaggerDoc()["availableReplicas"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "Selector", Type: "string", Priority: 1, Description: v1beta1.MinerSetSpec{}.SwaggerDoc()["selector"]},
	}
	h.TableHandler(minerSetColumnDefinitions, printMinerSet)
	h.TableHandler(minerSetColumnDefinitions, printMinerSetList)

	minerColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Status", Type: "string", Description: "The status of the miner"},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "Type", Type: "string", Priority: 1, Description: metav1.ObjectMeta{}.SwaggerDoc()["minerType"]},
	}

	h.TableHandler(minerColumnDefinitions, printMiner)
	h.TableHandler(minerColumnDefinitions, printMinerList)

	leaseColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Holder", Type: "string", Description: coordinationv1.LeaseSpec{}.SwaggerDoc()["holderIdentity"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}
	h.TableHandler(leaseColumnDefinitions, printLease)
	h.TableHandler(leaseColumnDefinitions, printLeaseList)

	statusColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Status", Type: "string", Description: metav1.Status{}.SwaggerDoc()["status"]},
		{Name: "Reason", Type: "string", Description: metav1.Status{}.SwaggerDoc()["reason"]},
		{Name: "Message", Type: "string", Description: metav1.Status{}.SwaggerDoc()["Message"]},
	}
	h.TableHandler(statusColumnDefinitions, printStatus)

	eventColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Last Seen", Type: "string", Description: apiv1.Event{}.SwaggerDoc()["lastTimestamp"]},
		{Name: "Type", Type: "string", Description: apiv1.Event{}.SwaggerDoc()["type"]},
		{Name: "Reason", Type: "string", Description: apiv1.Event{}.SwaggerDoc()["reason"]},
		{Name: "Object", Type: "string", Description: apiv1.Event{}.SwaggerDoc()["involvedObject"]},
		{Name: "Subobject", Type: "string", Priority: 1, Description: apiv1.Event{}.InvolvedObject.SwaggerDoc()["fieldPath"]},
		{Name: "Source", Type: "string", Priority: 1, Description: apiv1.Event{}.SwaggerDoc()["source"]},
		{Name: "Message", Type: "string", Description: apiv1.Event{}.SwaggerDoc()["message"]},
		{Name: "First Seen", Type: "string", Priority: 1, Description: apiv1.Event{}.SwaggerDoc()["firstTimestamp"]},
		{Name: "Count", Type: "string", Priority: 1, Description: apiv1.Event{}.SwaggerDoc()["count"]},
		{Name: "Name", Type: "string", Priority: 1, Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
	}
	h.TableHandler(eventColumnDefinitions, printEvent)
	h.TableHandler(eventColumnDefinitions, printEventList)

	chainColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		// {Name: "Status", Type: "string", Description: "The status of the miner"},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}

	h.TableHandler(chainColumnDefinitions, printChain)
	h.TableHandler(chainColumnDefinitions, printChainList)
}

func printNamespace(obj *api.Namespace, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, obj.Name, string(obj.Status.Phase), printersutil.TranslateTimestampSince(obj.CreationTimestamp))
	return []metav1.TableRow{row}, nil
}

func printNamespaceList(list *api.NamespaceList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printNamespace(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printConfigMap(obj *api.ConfigMap, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(
		row.Cells,
		obj.Name,
		int64(len(obj.Data)+len(obj.BinaryData)),
		printersutil.TranslateTimestampSince(obj.CreationTimestamp),
	)
	return []metav1.TableRow{row}, nil
}

func printConfigMapList(list *api.ConfigMapList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printConfigMap(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printMiner(obj *apps.Miner, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}

	phase := string(v1beta1.MinerPhasePending)
	if obj.Status.Phase != "" {
		phase = obj.Status.Phase
	}

	row.Cells = append(
		row.Cells,
		obj.Name,
		phase,
		printersutil.TranslateTimestampSince(obj.CreationTimestamp),
	)

	if options.Wide {
		row.Cells = append(row.Cells, obj.Spec.MinerType)
	}

	return []metav1.TableRow{row}, nil
}

func printMinerList(list *apps.MinerList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printMiner(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printMinerSetList(minerSetList *apps.MinerSetList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(minerSetList.Items))
	for i := range minerSetList.Items {
		r, err := printMinerSet(&minerSetList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printMinerSet(obj *apps.MinerSet, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}

	desiredReplicas := obj.Spec.Replicas
	currentReplicas := obj.Status.Replicas
	readyReplicas := obj.Status.ReadyReplicas
	availableReplicas := obj.Status.AvailableReplicas

	ready := fmt.Sprintf("%d/%d", int64(readyReplicas), int64(*desiredReplicas))
	row.Cells = append(
		row.Cells,
		obj.Name,
		ready,
		int64(currentReplicas),
		int64(availableReplicas),
		printersutil.TranslateTimestampSince(obj.CreationTimestamp),
	)
	if options.Wide {
		row.Cells = append(row.Cells, metav1.FormatLabelSelector(&obj.Spec.Selector))
	}

	return []metav1.TableRow{row}, nil
}

func printLease(obj *coordination.Lease, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}

	var holderIdentity string
	if obj.Spec.HolderIdentity != nil {
		holderIdentity = *obj.Spec.HolderIdentity
	}
	row.Cells = append(row.Cells, obj.Name, holderIdentity, printersutil.TranslateTimestampSince(obj.CreationTimestamp))
	return []metav1.TableRow{row}, nil
}

func printLeaseList(list *coordination.LeaseList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printLease(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printStatus(obj *metav1.Status, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, obj.Status, obj.Reason, obj.Message)

	return []metav1.TableRow{row}, nil
}

func printEvent(obj *api.Event, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}

	firstTimestamp := printersutil.TranslateTimestampSince(obj.FirstTimestamp)
	if obj.FirstTimestamp.IsZero() {
		firstTimestamp = printersutil.TranslateMicroTimestampSince(obj.EventTime)
	}

	lastTimestamp := printersutil.TranslateTimestampSince(obj.LastTimestamp)
	if obj.LastTimestamp.IsZero() {
		lastTimestamp = firstTimestamp
	}

	count := obj.Count
	if obj.Series != nil {
		lastTimestamp = printersutil.TranslateMicroTimestampSince(obj.Series.LastObservedTime)
		count = obj.Series.Count
	} else if count == 0 {
		// Singleton events don't have a count set in the new API.
		count = 1
	}

	var target string
	if len(obj.InvolvedObject.Name) > 0 {
		target = fmt.Sprintf("%s/%s", strings.ToLower(obj.InvolvedObject.Kind), obj.InvolvedObject.Name)
	} else {
		target = strings.ToLower(obj.InvolvedObject.Kind)
	}
	if options.Wide {
		row.Cells = append(row.Cells,
			lastTimestamp,
			obj.Type,
			obj.Reason,
			target,
			obj.InvolvedObject.FieldPath,
			formatEventSource(obj.Source, obj.ReportingController, obj.ReportingInstance),
			strings.TrimSpace(obj.Message),
			firstTimestamp,
			int64(count),
			obj.Name,
		)
	} else {
		row.Cells = append(row.Cells,
			lastTimestamp,
			obj.Type,
			obj.Reason,
			target,
			strings.TrimSpace(obj.Message),
		)
	}

	return []metav1.TableRow{row}, nil
}

// Sorts and prints the EventList in a human-friendly format.
func printEventList(list *api.EventList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printEvent(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printChain(obj *apps.Chain, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}

	row.Cells = append(
		row.Cells,
		obj.Name,
		printersutil.TranslateTimestampSince(obj.CreationTimestamp),
	)

	if options.Wide {
		// row.Cells = append(row.Cells, getMinerInternalIP(obj), getMinerExternalIP(obj), getInstanceID(obj))
	}

	return []metav1.TableRow{row}, nil
}

func printChainList(list *apps.ChainList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(list.Items))
	for i := range list.Items {
		r, err := printChain(&list.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// formatEventSource formats EventSource as a comma separated string excluding Host when empty.
// It uses reportingController when Source.Component is empty and reportingInstance when Source.Host is empty.
func formatEventSource(es api.EventSource, reportingController, reportingInstance string) string {
	return formatEventSourceComponentInstance(
		firstNonEmpty(es.Component, reportingController),
		firstNonEmpty(es.Host, reportingInstance),
	)
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if len(s) > 0 {
			return s
		}
	}
	return ""
}

func formatEventSourceComponentInstance(component, instance string) string {
	if len(instance) == 0 {
		return component
	}
	return component + ", " + instance
}
