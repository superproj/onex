// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package job implements the controller that reconciles the lifecycle of a
// batch/v1beta1 Job (Pending -> Running -> Succeeded/Failed).
//
// It is a faithful port of the mechanism used by k8s.io/kubernetes/pkg/controller/job:
// ControllerExpectations gate reconciliation until the informer observes the
// controller's own writes, a backoffStore applies exponential backoff to
// consecutively failing dispatches, and a ControllerRefManager adopts/releases
// provider-managed child objects via controller references.
//
// The underlying workload is abstracted behind registry.Provider. Concrete
// providers (kubernetes / LLMTrain / aws-batch / spark-on-yarn) register
// themselves with the providers registry; dispatch resolves them by
// job.spec.type.
package job
