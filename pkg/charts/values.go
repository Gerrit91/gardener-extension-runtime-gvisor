// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts

import (
	"github.com/gardener/gardener/pkg/chartrenderer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/gardener-extension-runtime-gvisor/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

// GVisorConfigKey is the key for the gVisor configuration.
const GVisorConfigKey = "config.yaml"

// RenderGVisorChart renders the gVisor chart
func RenderGVisorChart(renderer chartrenderer.Interface) ([]byte, error) {
	gvisorChartValues := map[string]interface{}{}

	release, err := renderer.RenderEmbeddedFS(charts.InternalChart, gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gvisorChartValues)
	if err != nil {
		return nil, err
	}
	return release.Manifest(), nil
}
