//
// Copyright 2020 IBM Corporation
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

package authentication

import (
	"context"
	"math/rand"
	"time"

	certmgr "github.com/IBM/ibm-iam-operator/pkg/apis/certmanager/v1alpha1"
	operatorv1alpha1 "github.com/IBM/ibm-iam-operator/pkg/apis/operator/v1alpha1"
	res "github.com/IBM/ibm-iam-operator/pkg/resources"
	userv1 "github.com/openshift/api/user/v1"
	regen "github.com/zach-klippenstein/goregen"
	reg "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_authentication")
var fullAccess int32 = 0777
var trueVar bool = true
var falseVar bool = false
var seconds60 int64 = 60
var partialAccess int32 = 420
var authServicePort int32 = 9443
var identityProviderPort int32 = 4300
var identityManagerPort int32 = 4500
var serviceAccountName string = "ibm-iam-operand-restricted"

var cpu10 = resource.NewMilliQuantity(10, resource.DecimalSI)            // 10m
var cpu50 = resource.NewMilliQuantity(50, resource.DecimalSI)            // 50m
var cpu100 = resource.NewMilliQuantity(100, resource.DecimalSI)          // 100m
var cpu1000 = resource.NewMilliQuantity(1000, resource.DecimalSI)        // 1000m
var memory32 = resource.NewQuantity(100*1024*1024, resource.BinarySI)    // 32Mi
var memory128 = resource.NewQuantity(128*1024*1024, resource.BinarySI)   // 128Mi
var memory150 = resource.NewQuantity(150*1024*1024, resource.BinarySI)   // 128Mi
var memory350 = resource.NewQuantity(350*1024*1024, resource.BinarySI)   // 350Mi
var memory1024 = resource.NewQuantity(1024*1024*1024, resource.BinarySI) // 1024Mi

var rule = `^([a-z0-9]){32,}$`
var wlpClientID = generateRandomString(rule)
var wlpClientSecret = generateRandomString(rule)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Authentication Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAuthentication{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("authentication-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Authentication
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.Authentication{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Certificate and requeue the owner Authentication
	err = c.Watch(&source.Kind{Type: &certmgr.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Authentication{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Certificate and requeue the owner Authentication
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Authentication{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Certificate and requeue the owner Authentication
	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Authentication{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Service and requeue the owner Authentication
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Authentication{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource ConfigMap and requeue the owner Authentication
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Authentication{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Ingress and requeue the owner Authentication
	err = c.Watch(&source.Kind{Type: &net.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Authentication{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Deployment and requeue the owner Authentication
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Authentication{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAuthentication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAuthentication{}

// ReconcileAuthentication reconciles a Authentication object
type ReconcileAuthentication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Authentication object and makes changes based on the state read
// and what is in the Authentication.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAuthentication) Reconcile(contect context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Authentication")
	var requeueResult bool = false

	// Fetch the Authentication instance
	instance := &operatorv1alpha1.Authentication{}
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

	// Credit: kubebuilder book
	finalizerName := "authentication.operator.ibm.com"
	// Determine if the Authentication CR  is going to be deleted
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted, but add our finalizer so we know to remove this object later when it is going to be deleted
		if !containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				log.Error(err, "Error adding the finalizer to the CR")
				return reconcile.Result{}, err
			}
		}
	} else {
		// Object scheduled to be deleted
		if containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			if err := r.deleteExternalResources(instance); err != nil {
				log.Error(err, "Error deleting resources created by this operator")

				return reconcile.Result{}, err
			}

			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				log.Error(err, "Error updating the CR to remove the finalizer")
				return reconcile.Result{}, err
			}

		}
		return reconcile.Result{}, nil
	}

	// Check if this Certificate already exists and create it if it doesn't
	currentCertificate := &certmgr.Certificate{}
	err = r.handleCertificate(instance, currentCertificate, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists and create it if it doesn't
	currentService := &corev1.Service{}
	err = r.handleService(instance, currentService, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Secret already exists and create it if it doesn't
	currentSecret := &corev1.Secret{}
	err = r.handleSecret(instance, wlpClientID, wlpClientSecret, currentSecret, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Job already exists and create it if it doesn't
	currentJob := &batchv1.Job{}
	err = r.handleJob(instance, currentJob, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	//Check if this ConfigMap already exists and create it if it doesn't
	currentConfigMap := &corev1.ConfigMap{}
	err = r.handleConfigMap(instance, wlpClientID, wlpClientSecret, currentConfigMap, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Ingress already exists and create it if it doesn't
	currentIngress := &net.Ingress{}
	err = r.handleIngress(instance, currentIngress, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ClusterRole already exists and create it if it doesn't
	currentClusterRole := &rbacv1.ClusterRole{}
	err = r.handleClusterRole(instance, currentClusterRole, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ClusterRole already exists and create it if it doesn't
	currentClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = r.handleClusterRoleBinding(instance, currentClusterRoleBinding, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	currentWebhook := &reg.MutatingWebhookConfiguration{}
	err = r.handleWebhook(instance, currentWebhook, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists and create it if it doesn't
	currentDeployment := &appsv1.Deployment{}
	err = r.handleDeployment(instance, currentDeployment, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this User already exists and create it if it doesn't
	currentUser := &userv1.User{}
	err = r.handleUser(instance, currentUser, &requeueResult)
	if err != nil {
		return reconcile.Result{}, err
	}

	if requeueResult {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

// Removes some of the resources created by this controller for the CR including
// The clusterrole, clusterrolebinding and User resources created by Authentication
func (r *ReconcileAuthentication) deleteExternalResources(instance *operatorv1alpha1.Authentication) error {

	crMap := generateCRData()
	crbMap := generateCRBData("dummy", "dummy")
	userName := instance.Spec.Config.DefaultAdminUser
	csCfgAnnotationName := res.GetCsConfigAnnotation(instance.Namespace)

	// These code changes handles all use cases:
	// - fresh install in saas or on-prem mode and
	// - upgrade on older releases in on-prem mode
	webhook := "namespace-admission-config"
	if instance.Spec.Config.IBMCloudSaas {
		// in saas mode
		webhook = webhook + "-" + instance.Namespace
	} else if instance.Spec.Config.OnPremMultipleDeploy {
		// multiple deployment in on-prem mode
		webhook = webhook + "-" + instance.Namespace
	}

	// Remove multiple deployment common-service/config annotation
	for crName := range crMap {
		if err := removeCsAnnotationFromCR(r.client, crName, csCfgAnnotationName); err != nil {
			return err
		}
	}
	for crbName := range crbMap {
		if err := removeCsAnnotationFromCRB(r.client, crbName, csCfgAnnotationName); err != nil {
			return err
		}
	}

	log.V(0).Info("Wait for 2 seconds.")
	time.Sleep(time.Second * 2)

	// Finally check and remove Cluster Role
	for crName := range crMap {
		if err := removeCR(r.client, crName); err != nil {
			return err
		}
	}

	// Remove Cluster Role Binding
	for crbName := range crbMap {
		if err := removeCRB(r.client, crbName); err != nil {
			return err
		}
	}

	log.V(0).Info("Wait for 2 seconds.")
	time.Sleep(time.Second * 2)

	// Remove User

	if err := removeUser(r.client, userName); err != nil {
		return err
	}

	// Remove Webhook

	if err := removeWebhook(r.client, webhook); err != nil {
		return err
	}

	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// Functions to remove cluster scoped resources

func isOidcAdminBindingCRBExists(client client.Client) bool {
	// checks if oidc-admin-binding ClusterRoleBinding exists or not
	crbName := "oidc-admin-binding"
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: crbName, Namespace: ""}, clusterRoleBinding); err != nil && errors.IsNotFound(err) {
		return false
	}
	return true
}

func removeCsAnnotationFromCR(client client.Client, crName string, csCfgAnnotationName string) error {
	// Remove common-service/config annotation
	clusterRole := &rbacv1.ClusterRole{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: crName, Namespace: ""}, clusterRole); err != nil && errors.IsNotFound(err) {
		log.V(1).Info("Error getting cluster role", crName, err)
		return nil
	} else if err == nil {
		if len(clusterRole.ObjectMeta.Annotations) > 0 {
			if _, ok := clusterRole.ObjectMeta.Annotations[csCfgAnnotationName]; ok {
				delete(clusterRole.ObjectMeta.Annotations, csCfgAnnotationName)
				if err = client.Update(context.Background(), clusterRole); err != nil {
					// if error, retry second time to avoid manual deletion after uninstall
					if err2 := client.Update(context.Background(), clusterRole); err2 != nil {
						log.V(1).Info("Error removing common-service/config from cluster role", "name", crName, "error message", err2)
						return err2
					}
				}
			}
		}
	} else {
		return err
	}
	return nil
}

func removeCsAnnotationFromCRB(client client.Client, crbName string, csCfgAnnotationName string) error {
	// Remove common-service/config annotation
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: crbName, Namespace: ""}, clusterRoleBinding); err != nil && errors.IsNotFound(err) {
		log.V(1).Info("Error getting cluster role binding", crbName, err)
		return nil
	} else if err == nil {
		if len(clusterRoleBinding.ObjectMeta.Annotations) > 0 {
			if _, ok := clusterRoleBinding.ObjectMeta.Annotations[csCfgAnnotationName]; ok {
				delete(clusterRoleBinding.ObjectMeta.Annotations, csCfgAnnotationName)
				if err = client.Update(context.Background(), clusterRoleBinding); err != nil {
					// if error, retry second time to avoid manual deletion after uninstall
					if err2 := client.Update(context.Background(), clusterRoleBinding); err2 != nil {
						log.V(1).Info("Error removing common-service/config annotation from cluster role binding", "name", crbName, "error message", err2)
						return err2
					}
				}
			}
		}
	} else {
		return err
	}
	return nil
}

func removeCR(client client.Client, crName string) error {
	// Delete Clusterrole
	clusterRole := &rbacv1.ClusterRole{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: crName, Namespace: ""}, clusterRole); err != nil && errors.IsNotFound(err) {
		log.V(1).Info("Error getting cluster role", crName, err)
		return nil
	} else if err == nil {
		if !res.IsCsConfigAnnotationExists(clusterRole.ObjectMeta.Annotations) {
			if err = client.Delete(context.Background(), clusterRole); err != nil {
				log.V(1).Info("Error deleting cluster role", "name", crName, "error message", err)
				return err
			}
		}
	} else {
		return err
	}
	return nil
}

func removeCRB(client client.Client, crbName string) error {
	// Delete ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: crbName, Namespace: ""}, clusterRoleBinding); err != nil && errors.IsNotFound(err) {
		log.V(1).Info("Error getting cluster role binding", crbName, err)
		return nil
	} else if err == nil {
		if crbName == "oidc-admin-binding" {
			clusterRoleBinding.ObjectMeta.Finalizers = []string{}
			if err = client.Update(context.Background(), clusterRoleBinding); err != nil {
				log.V(1).Info("Error updating cluster role binding", "name", crbName, "error message", err)
				return err
			}
		}
		if !res.IsCsConfigAnnotationExists(clusterRoleBinding.ObjectMeta.Annotations) {
			if err = client.Delete(context.Background(), clusterRoleBinding); err != nil {
				log.V(1).Info("Error deleting cluster role binding", "name", crbName, "error message", err)
				return err
			}
		}
	} else {
		return err
	}
	return nil
}

func removeUser(client client.Client, userName string) error {
	// Delete User
	user := &userv1.User{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: userName, Namespace: ""}, user); err != nil && errors.IsNotFound(err) {
		log.V(1).Info("Error getting user", userName, err)
		return nil
	} else if err == nil {
		// check if oidc-admin-binding CRB exists or not
		if !isOidcAdminBindingCRBExists(client) {
			if err = client.Delete(context.Background(), user); err != nil {
				log.V(1).Info("Error deleting user", "name", userName, "error message", err)
				return err
			}
		}
	} else {
		return err
	}
	return nil
}

func removeWebhook(client client.Client, webhookName string) error {
	// Delete Webhook
	webhook := &reg.MutatingWebhookConfiguration{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: webhookName, Namespace: ""}, webhook); err != nil && errors.IsNotFound(err) {
		log.V(1).Info("Error getting webhook", webhookName, err)
		return nil
	} else if err == nil {
		if err = client.Delete(context.Background(), webhook); err != nil {
			log.V(1).Info("Error deleting webhook", "name", webhookName, "error message", err)
			return err
		}
	} else {
		return err
	}
	return nil
}

func generateRandomString(rule string) string {

	generator, _ := regen.NewGenerator(rule, &regen.GeneratorArgs{
		RngSource:               rand.NewSource(time.Now().UnixNano()),
		MaxUnboundedRepeatCount: 1})
	randomString := generator.Generate()
	return randomString
}
