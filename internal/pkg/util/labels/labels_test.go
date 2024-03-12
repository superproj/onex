// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package labels

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNameLabelValue(t *testing.T) {
	g := gomega.NewWithT(t)
	tests := []struct {
		name           string
		machineSetName string
		want           string
	}{
		{
			name:           "return the name if it's less than 63 characters",
			machineSetName: "machineSetName",
			want:           "machineSetName",
		},
		{
			name:           "return  for a name with more than 63 characters",
			machineSetName: "machineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetName",
			want:           "hash_FR_ghQ_z",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustFormatValue(tt.machineSetName)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestMustMatchLabelValueForName(t *testing.T) {
	g := gomega.NewWithT(t)
	tests := []struct {
		name           string
		machineSetName string
		labelValue     string
		want           bool
	}{
		{
			name:           "match labels when MachineSet name is short",
			machineSetName: "ms1",
			labelValue:     "ms1",
			want:           true,
		},
		{
			name:           "don't match different labels when MachineSet name is short",
			machineSetName: "ms1",
			labelValue:     "notMS1",
			want:           false,
		},
		{
			name:           "don't match labels when MachineSet name is long",
			machineSetName: "machineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetName",
			labelValue:     "hash_Nx4RdE_z",
			want:           false,
		},
		{
			name:           "match labels when MachineSet name is long",
			machineSetName: "machineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetNamemachineSetName",
			labelValue:     "hash_FR_ghQ_z",
			want:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustEqualValue(tt.machineSetName, tt.labelValue)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
