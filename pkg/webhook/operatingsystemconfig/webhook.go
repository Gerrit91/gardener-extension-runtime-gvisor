// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig

import (
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components/kubelet"
	oscutils "github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// Name is the webhook name
	Name = "gvisor-operatingsystemconfig-webhook"
)

var (
	logger = log.Log.WithName("gvisor-webhook")
)

// AddToManager creates a webhook and adds it to the manager.
func AddToManager(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	logger.Info("Adding webhook to manager")

	var (
		fciCodec = oscutils.NewFileContentInlineCodec()
		types    = []extensionswebhook.Type{
			{Obj: &extensionsv1alpha1.OperatingSystemConfig{}},
		}
		mutator = genericmutator.NewMutator(mgr, NewEnsurer(mgr, logger), oscutils.NewUnitSerializer(),
			kubelet.NewConfigCodec(fciCodec), fciCodec, logger)
	)

	// TODO: figure out from cluster resource if the workergroup needs to be targeted
	handler, err := extensionswebhook.NewBuilder(mgr, logger).WithMutator(mutator, types...).Build()
	if err != nil {
		return nil, err
	}

	webhook := &extensionswebhook.Webhook{
		Name:     Name,
		Provider: "",
		Types:    types,
		Target:   extensionswebhook.TargetSeed,
		Path:     "/webhooks/gvisor-osc",
		Webhook:  &admission.Webhook{Handler: handler},
	}

	return webhook, nil
}
