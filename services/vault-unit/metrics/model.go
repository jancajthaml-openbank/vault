// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/vault-unit/utils"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics represents metrics subroutine
type Metrics struct {
	utils.DaemonSupport
	output              string
	tenant              string
	refreshRate         time.Duration
	promisesAccepted    metrics.Counter
	commitsAccepted     metrics.Counter
	rollbacksAccepted   metrics.Counter
	createdAccounts     metrics.Counter
	updatedSnapshots    metrics.Meter
	snapshotCronLatency metrics.Timer
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, output string, refreshRate time.Duration) Metrics {
	return Metrics{
		DaemonSupport:       utils.NewDaemonSupport(ctx),
		output:              output,
		refreshRate:         refreshRate,
		promisesAccepted:    metrics.NewCounter(),
		commitsAccepted:     metrics.NewCounter(),
		rollbacksAccepted:   metrics.NewCounter(),
		createdAccounts:     metrics.NewCounter(),
		updatedSnapshots:    metrics.NewMeter(),
		snapshotCronLatency: metrics.NewTimer(),
	}
}

// MarshalJSON serialises Metrics as json bytes
func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return nil, fmt.Errorf("cannot marshall nil")
	}

	if metrics.promisesAccepted == nil || metrics.commitsAccepted == nil ||
		metrics.rollbacksAccepted == nil || metrics.createdAccounts == nil ||
		metrics.updatedSnapshots == nil || metrics.snapshotCronLatency == nil {
		return nil, fmt.Errorf("cannot marshall nil references")
	}

	var buffer bytes.Buffer

	buffer.WriteString("{\"snapshotCronLatency\":")
	buffer.WriteString(strconv.FormatFloat(metrics.snapshotCronLatency.Percentile(0.95), 'f', -1, 64))
	buffer.WriteString(",\"updatedSnapshots\":")
	buffer.WriteString(strconv.FormatInt(metrics.updatedSnapshots.Count(), 10))
	buffer.WriteString(",\"createdAccounts\":")
	buffer.WriteString(strconv.FormatInt(metrics.createdAccounts.Count(), 10))
	buffer.WriteString(",\"promisesAccepted\":")
	buffer.WriteString(strconv.FormatInt(metrics.promisesAccepted.Count(), 10))
	buffer.WriteString(",\"commitsAccepted\":")
	buffer.WriteString(strconv.FormatInt(metrics.commitsAccepted.Count(), 10))
	buffer.WriteString(",\"rollbacksAccepted\":")
	buffer.WriteString(strconv.FormatInt(metrics.rollbacksAccepted.Count(), 10))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON deserializes Metrics from json bytes
func (metrics *Metrics) UnmarshalJSON(data []byte) error {
	if metrics == nil {
		return fmt.Errorf("cannot unmarshall to nil")
	}

	if metrics.promisesAccepted == nil || metrics.commitsAccepted == nil ||
		metrics.rollbacksAccepted == nil || metrics.createdAccounts == nil ||
		metrics.updatedSnapshots == nil || metrics.snapshotCronLatency == nil {
		return fmt.Errorf("cannot unmarshall to nil references")
	}

	aux := &struct {
		SnapshotCronLatency float64 `json:"snapshotCronLatency"`
		UpdatedSnapshots    int64   `json:"updatedSnapshots"`
		CreatedAccounts     int64   `json:"createdAccounts"`
		PromisesAccepted    int64   `json:"promisesAccepted"`
		CommitsAccepted     int64   `json:"commitsAccepted"`
		RollbacksAccepted   int64   `json:"rollbacksAccepted"`
	}{}

	if err := utils.JSON.Unmarshal(data, &aux); err != nil {
		return err
	}

	metrics.promisesAccepted.Clear()
	metrics.promisesAccepted.Inc(aux.PromisesAccepted)
	metrics.commitsAccepted.Clear()
	metrics.commitsAccepted.Inc(aux.CommitsAccepted)
	metrics.rollbacksAccepted.Clear()
	metrics.rollbacksAccepted.Inc(aux.RollbacksAccepted)
	metrics.createdAccounts.Clear()
	metrics.createdAccounts.Inc(aux.CreatedAccounts)
	metrics.updatedSnapshots.Mark(aux.UpdatedSnapshots)
	metrics.snapshotCronLatency.Update(time.Duration(aux.SnapshotCronLatency))

	return nil
}
