// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package nextid

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	"github.com/superproj/onex/internal/pkg/config"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
)

const (
	IDCounterAnnotation = "onex.io/id-counter"
)

var idGeneratorConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      config.IDGeneraterName.String(),
		Namespace: metav1.NamespaceSystem,
		Annotations: map[string]string{
			IDCounterAnnotation: "0",
		},
	},
}

func GetNextID(client clientset.Interface) (uint64, error) {
	var seqID uint64
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		idgenerator, err := config.IDGeneraterName.GetConfig(client)
		if err != nil {
			if apierrors.IsNotFound(err) {
				if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(
					context.Background(),
					idGeneratorConfigMap,
					metav1.CreateOptions{},
				); err != nil {
					klog.ErrorS(err, "Failed to create id generator")
					return err
				}

				seqID = 0
				return nil
			}

			klog.ErrorS(err, "Failed to get id generator")
			return err
		}

		annotations := idgenerator.GetAnnotations()
		if annotations == nil || annotations[IDCounterAnnotation] == "" {
			return fmt.Errorf("counter's annotation illegal")
		}
		rv, err := strconv.ParseUint(annotations[IDCounterAnnotation], 10, 64)
		if err != nil {
			return err
		}
		rv++
		idgenerator.Annotations[IDCounterAnnotation] = fmt.Sprintf("%d", rv)
		if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Update(
			context.Background(),
			idgenerator,
			metav1.UpdateOptions{},
		); err != nil {
			return err
		}

		seqID = rv
		return nil
	}); err != nil {
		return 0, err
	}

	return seqID, nil
}
