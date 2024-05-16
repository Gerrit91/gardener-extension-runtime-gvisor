// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"slices"

	"github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-runtime-gvisor/imagevector"
)

// NewEnsurer creates a new controlplane ensurer.
func NewEnsurer(mgr manager.Manager, logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		client: mgr.GetClient(),
		logger: logger.WithName("gvisor-controlplane-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	client client.Client
	logger logr.Logger
}

// EnsureAdditionalFiles ensures additional systemd files
// "old" might be "nil" and must always be checked.
func (e *ensurer) EnsureAdditionalFiles(_ context.Context, _ gcontext.GardenContext, new, _ *[]extensionsv1alpha1.File) error {
	e.logger.Info("ensuring files")

	ociImage := imagevector.FindImage("runtime-gvisor-installation")

	*new = webhook.EnsureFileWithPath(*new, extensionsv1alpha1.File{
		Path:        path.Join(extensionsv1alpha1.ContainerDRuntimeContainersBinFolder, "containerd-shim-runsc-v1"),
		Permissions: ptr.To(int32(0644)),
		Content: extensionsv1alpha1.FileContent{
			ImageRef: &extensionsv1alpha1.FileContentImageRef{
				Image:           ociImage,
				FilePathInImage: "/var/content/containerd-shim-runsc-v1",
			},
		},
	})

	*new = webhook.EnsureFileWithPath(*new, extensionsv1alpha1.File{
		Path:        path.Join(extensionsv1alpha1.ContainerDRuntimeContainersBinFolder, "runsc"),
		Permissions: ptr.To(int32(0644)),
		Content: extensionsv1alpha1.FileContent{
			ImageRef: &extensionsv1alpha1.FileContentImageRef{
				Image:           ociImage,
				FilePathInImage: "/var/content/runsc",
			},
		},
	})

	return nil
}

// EnsureContainerdConfig ensures the containerd config.
// "old" might be "nil" and must always be checked.
func (e *ensurer) EnsureContainerdConfig(_ context.Context, _ gcontext.GardenContext, new, _ *extensionsv1alpha1.CRIConfig) error {
	e.logger.Info("ensuring containerd config")

	if new.Containerd == nil {
		new.Containerd = &extensionsv1alpha1.ContainerdConfig{}
	}

	raw, err := json.Marshal(struct {
		RuntimeType string `json:"runtime_type"`
	}{
		RuntimeType: "io.containerd.runsc.v1",
	})
	if err != nil {
		return fmt.Errorf("unable to marshal containerd config: %w", err)
	}

	new.Containerd.Plugins = ensurePluginConfiguration(new.Containerd.Plugins, extensionsv1alpha1.PluginConfig{
		Path: []string{"io.containerd.grpc.v1.cri", "containerd", "runtimes", "runsc"},
		Values: &apiextensionsv1.JSON{
			Raw: raw,
		},
	})

	raw, err = json.Marshal(struct {
		RuntimeType string `json:"runtime_type"`
	}{
		RuntimeType: "io.containerd.runc.v2",
	})
	if err != nil {
		return fmt.Errorf("unable to marshal containerd config: %w", err)
	}

	new.Containerd.Plugins = ensurePluginConfiguration(new.Containerd.Plugins, extensionsv1alpha1.PluginConfig{
		Path: []string{"io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc"},
		Values: &apiextensionsv1.JSON{
			Raw: raw,
		},
	})

	return nil
}

func ensurePluginConfiguration(plugins []extensionsv1alpha1.PluginConfig, plugin extensionsv1alpha1.PluginConfig) []extensionsv1alpha1.PluginConfig {
	var res []extensionsv1alpha1.PluginConfig

	found := false
	for _, p := range plugins {
		if slices.Equal(p.Path, plugin.Path) {
			found = true
			p.Values = plugin.Values
		}

		res = append(res, p)
	}

	if !found {
		res = append(res, plugin)
	}

	return res
}
