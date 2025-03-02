//
// Copyright 2022 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package issuer

import (
	"context"

	"golang.org/x/mod/semver"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metaerrors "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1"
	certmanagerv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1alpha1"
	"github.com/ibm/ibm-cert-manager-operator/pkg/controller/certmanager"
)

var log = logf.Log.WithName("controller_issuer")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Issuer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIssuer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("issuer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Issuer
	err = c.Watch(&source.Kind{Type: &certmanagerv1alpha1.Issuer{}}, &handler.EnqueueRequestForObject{}, ignoreStatusPredicate{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Issuer
	err = c.Watch(&source.Kind{Type: &certmanagerv1.Issuer{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &certmanagerv1alpha1.Issuer{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileIssuer implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIssuer{}

// ReconcileIssuer reconciles a Issuer object
type ReconcileIssuer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Issuer object and makes changes based on the state read
// and what is in the Issuer.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIssuer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Issuer")

	// Fetch the Issuer instance
	instance := &certmanagerv1alpha1.Issuer{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//Check RHACM
	rhacmVersion, _, rhacmErr := certmanager.CheckRhacm(r.client)
	if rhacmErr != nil {
		// missing RHACM CR or CRD means RHACM does not exist
		if errors.IsNotFound(rhacmErr) || metaerrors.IsNoMatchError(rhacmErr) {
			log.Error(rhacmErr, "Could not find RHACM")
		} else {
			return reconcile.Result{}, rhacmErr
		}
	}
	if rhacmVersion != "" {
		rhacmVersion = "v" + rhacmVersion
		deployOperand := semver.Compare(rhacmVersion, "v2.3")

		if deployOperand < 0 {
			log.Info("RHACM version is less than 2.3, so not reconciling Issuer")
			return reconcile.Result{}, nil
		}
	}

	reqLogger.Info("### DEBUG ### v1alpha1 Issuer created", "Issuer.Namespace", instance.Namespace, "Issuer.Name", instance.Name)

	reqLogger.Info("### DEBUG ### Creating v1 Issuer", "Issuer.Namespace", instance.Namespace, "Issuer.Name", instance.Name)

	annotations := instance.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["ibm-cert-manager-operator-generated"] = "true"

	v1Issuer := &certmanagerv1.Issuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Issuer",
			APIVersion: "cert-manager.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: annotations,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				ACME:       convertACME(instance.Spec.ACME),
				CA:         convertCA(instance.Spec.CA),
				Vault:      convertVault(instance.Spec.Vault),
				SelfSigned: convertSelfSigned(instance.Spec.SelfSigned),
				Venafi:     convertVenafi(instance.Spec.Venafi),
			},
		},
	}

	// Set the issuer v1alpha1 as the controller of the issuer v1
	if err := controllerutil.SetControllerReference(instance, v1Issuer, r.scheme); err != nil {
		reqLogger.Error(err, "### DEBUG ### failed to set owner reference for %s", v1Issuer)
		return reconcile.Result{}, err
	}

	if err := r.client.Create(context.TODO(), v1Issuer); err != nil {
		if errors.IsAlreadyExists(err) {
			existingIssuer := &certmanagerv1.Issuer{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: v1Issuer.Namespace, Name: v1Issuer.Name}, existingIssuer); err != nil {
				reqLogger.Error(err, "### DEBUG ### Failed to get v1 Issuer")
				return reconcile.Result{}, err
			}
			if !equality.Semantic.DeepEqual(v1Issuer.Labels, existingIssuer.Labels) || !equality.Semantic.DeepEqual(v1Issuer.Spec, existingIssuer.Spec) {
				v1Issuer.SetResourceVersion(existingIssuer.GetResourceVersion())
				v1Issuer.SetAnnotations(existingIssuer.GetAnnotations())
				if err := r.client.Update(context.TODO(), v1Issuer); err != nil {
					reqLogger.Error(err, "### DEBUG ### Failed to update v1 Issuer")
					return reconcile.Result{}, err
				}
				reqLogger.Info("### DEBUG #### Updated v1 Issuer")
			}

			reqLogger.Info("### DEBUG ### Converting Issuer status")
			status := convertStatus(existingIssuer.Status)
			instance.Status = status
			reqLogger.Info("### DEBUG ### Updating v1alpha1 Issuer status")
			if err := r.client.Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "### DEBUG ### error updating")
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "### DEBUG ### Failed to create v1 Issuer")
		return reconcile.Result{}, err
	}

	reqLogger.Info("### DEBUG #### Created v1 Issuer")

	return reconcile.Result{}, nil
}

type ignoreStatusPredicate struct{}

func (i ignoreStatusPredicate) Create(e event.CreateEvent) bool {
	return true
}

func (i ignoreStatusPredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (i ignoreStatusPredicate) Update(e event.UpdateEvent) bool {
	return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
}

func (i ignoreStatusPredicate) Generic(e event.GenericEvent) bool {
	return false
}
