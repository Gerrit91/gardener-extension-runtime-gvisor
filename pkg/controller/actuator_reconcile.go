// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/charts"
)

const (
	// GVisorInstallationManagedResourceName is the name of the managed resource installation.
	GVisorInstallationManagedResourceName = "extension-runtime-gvisor-installation"
	// GVisorManagedResourceName is the name of the managed resource.
	GVisorManagedResourceName = "extension-runtime-gvisor"
)

// Reconcile implements ContainerRuntime.Actuator.
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, cr *extensionsv1alpha1.ContainerRuntime, cluster *extensionscontroller.Cluster) error {
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("could not create chart renderer for shoot '%s', %w", cr.Namespace, err)
	}

	log.Info("Preparing gVisor installation", "shoot", cluster.Shoot.Name, "shootNamespace", cluster.Shoot.Namespace)
	// create MR containing the prerequisites for the installation DaemonSet
	gVisorChart, err := charts.RenderGVisorChart(chartRenderer)
	if err != nil {
		return err
	}

	if err := managedresources.CreateForShoot(ctx, a.client, cr.Namespace, GVisorManagedResourceName, "extension-runtime-gvisor", false, map[string][]byte{charts.GVisorConfigKey: gVisorChart}); err != nil {
		return err
	}

	installMRName := fmt.Sprintf("%s-%s", GVisorInstallationManagedResourceName, cr.Spec.WorkerPool.Name)
	if err := managedresources.DeleteForShoot(ctx, a.client, cr.Namespace, installMRName); err != nil {
		return err
	}

	return nil
}
