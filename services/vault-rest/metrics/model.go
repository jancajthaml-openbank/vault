// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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

	"github.com/jancajthaml-openbank/vault-rest/utils"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics represents metrics subroutine
type Metrics struct {
	utils.DaemonSupport
	output               string
	refreshRate          time.Duration
	getAccountLatency    metrics.Timer
	createAccountLatency metrics.Timer
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, output string, refreshRate time.Duration) Metrics {
	return Metrics{
		DaemonSupport:        utils.NewDaemonSupport(ctx),
		output:               output,
		refreshRate:          refreshRate,
		getAccountLatency:    metrics.NewTimer(),
		createAccountLatency: metrics.NewTimer(),
	}
}

// MarshalJSON serialises Metrics as json bytes
func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return nil, fmt.Errorf("cannot marshall nil")
	}

	if metrics.getAccountLatency == nil || metrics.createAccountLatency == nil {
		return nil, fmt.Errorf("cannot marshall nil references")
	}

	var buffer bytes.Buffer

	buffer.WriteString("{\"getAccountLatency\":")
	buffer.WriteString(strconv.FormatFloat(metrics.getAccountLatency.Percentile(0.95), 'f', -1, 64))
	buffer.WriteString(",\"createAccountLatency\":")
	buffer.WriteString(strconv.FormatFloat(metrics.createAccountLatency.Percentile(0.95), 'f', -1, 64))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON deserializes Metrics from json bytes
func (metrics *Metrics) UnmarshalJSON(data []byte) error {
	if metrics == nil {
		return fmt.Errorf("cannot unmarshall to nil")
	}

	if metrics.getAccountLatency == nil || metrics.createAccountLatency == nil {
		return fmt.Errorf("cannot unmarshall to nil references")
	}

	aux := &struct {
		GetAccountLatency    float64 `json:"getAccountLatency"`
		CreateAccountLatency float64 `json:"createAccountLatency"`
	}{}

	if err := utils.JSON.Unmarshal(data, &aux); err != nil {
		return err
	}

	metrics.getAccountLatency.Update(time.Duration(aux.GetAccountLatency))
	metrics.createAccountLatency.Update(time.Duration(aux.CreateAccountLatency))

	return nil
}
