// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts_test

import (
	"fmt"

	"github.com/gardener/gardener/pkg/chartrenderer"
	mockchartrenderer "github.com/gardener/gardener/pkg/chartrenderer/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"helm.sh/helm/v3/pkg/releaseutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	internalcharts "github.com/gardener/gardener-extension-runtime-gvisor/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

var _ = Describe("Chart package test", func() {
	Describe("#RenderGvisorChart", func() {
		var (
			ctrl              *gomock.Controller
			mockChartRenderer *mockchartrenderer.MockInterface

			testManifestContent = "test-content"
			mkManifest          = func(name string) releaseutil.Manifest {
				return releaseutil.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
			}
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockChartRenderer = mockchartrenderer.NewMockInterface(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("Render Gvisor chart correctly", func() {
			renderedValues := map[string]interface{}{}

			mockChartRenderer.EXPECT().RenderEmbeddedFS(internalcharts.InternalChart, gvisor.ChartPath, gvisor.ReleaseName, metav1.NamespaceSystem, gomock.Eq(renderedValues)).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []releaseutil.Manifest{
					mkManifest(charts.GVisorConfigKey),
				},
			}, nil)

			_, err := charts.RenderGVisorChart(mockChartRenderer)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
