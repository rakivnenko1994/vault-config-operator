/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vaultresourcecontroller

import (
	"context"

	"github.com/redhat-cop/operator-utils/pkg/util"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type VaultResource struct {
	vaultEndpoint  *vaultutils.VaultEndpoint
	reconcilerBase *util.ReconcilerBase
}

func NewVaultResource(reconcilerBase *util.ReconcilerBase, obj client.Object) *VaultResource {
	return &VaultResource{
		reconcilerBase: reconcilerBase,
		vaultEndpoint:  vaultutils.NewVaultEndpoint(obj),
	}
}

func (r *VaultResource) Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if util.IsBeingDeleted(instance) {
		if !util.HasFinalizer(instance, redhatcopv1alpha1.GetFinalizer(instance)) {
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(ctx, instance)
		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return r.reconcilerBase.ManageError(ctx, instance, err)
		}
		util.RemoveFinalizer(instance, redhatcopv1alpha1.GetFinalizer(instance))
		err = r.reconcilerBase.GetClient().Update(ctx, instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return r.reconcilerBase.ManageError(ctx, instance, err)
		}
		return reconcile.Result{}, nil
	}

	err := r.manageReconcileLogic(ctx, instance)
	if err != nil {
		log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return r.reconcilerBase.ManageError(ctx, instance, err)
	}

	return r.reconcilerBase.ManageSuccess(ctx, instance)
}

func (r *VaultResource) manageCleanUpLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)
	err := r.vaultEndpoint.DeleteIfExists(context)
	if err != nil {
		log.Error(err, "unable to delete vault resource", "instance", instance)
		return err
	}
	return nil
}

func (r *VaultResource) manageReconcileLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)
	// prepare internal values
	err := instance.(vaultutils.VaultObject).PrepareInternalValues(context, instance)
	if err != nil {
		log.Error(err, "unable to prepare internal values", "instance", instance)
		return err
	}
	err = r.vaultEndpoint.CreateOrUpdate(context)
	if err != nil {
		log.Error(err, "unable to create/update vault resource", "instance", instance)
		return err
	}
	return nil
}
