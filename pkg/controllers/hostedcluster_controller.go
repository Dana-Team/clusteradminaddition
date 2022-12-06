/*
Copyright 2022.

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

package controllers

import (
	"context"
	utills "permission-granter-controller/pkg/utils"

	"github.com/go-logr/logr"
	"github.com/openshift/hypershift/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HostedClusterReconciler reconciles a HostedCluster object
type HostedClusterReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

type HostedClusterPredicate struct {
	predicate.Funcs
}

var (
	requesterAnnotation    = "dana.io/requester"
	clusterAdminAnnotation = "dana.io/addedclusteradmin"
)

//+kubebuilder:rbac:groups=hypershift.openshift.io.dana.io,resources=hostedclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hypershift.openshift.io.dana.io,resources=hostedclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hypershift.openshift.io.dana.io,resources=hostedclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HostedCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *HostedClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("hosted cluster", req.Name)
	hostedClusterObject := &v1alpha1.HostedCluster{}

	if err := r.Client.Get(ctx, req.NamespacedName, hostedClusterObject); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "could not decode object")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	hostedClient := r.getHostedClusterClient(hostedClusterObject.GetName())
	if val, ok := hostedClusterObject.GetAnnotations()[requesterAnnotation]; ok {
		r.addClusterAdminRoleBinding(hostedClient, val, hostedClusterObject, ctx)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HostedClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	hostedCluster := &unstructured.Unstructured{}
	hostedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "hypershift.openshift.io",
		Version: "v1alpha1",
		Kind:    "HostedCluster",
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(hostedCluster).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).WithEventFilter(HostedClusterPredicate{predicate.NewPredicateFuncs(func(object client.Object) bool {
		objAnnotaions := object.GetAnnotations()
		if _, ok := objAnnotaions[clusterAdminAnnotation]; ok {
			return false
		}
		return true
	})}).
		Complete(r)
}

//composeClusterAdminCRB the function gets username
//the function returns a ClusterRoleBinding giving the username the cluster-admin role
func composeClusterAdminCRB(username string) rbacv1.ClusterRoleBinding {
	return rbacv1.ClusterRoleBinding{
		ObjectMeta: v1api.ObjectMeta{
			Name: username,
		},
		Subjects: []rbacv1.Subject{
			{Kind: "User",
				Name: username},
		},
		RoleRef: rbacv1.RoleRef{Kind: "ClusterRole",
			Name: "cluster-admin"},
	}
}

//AppendAnnotations gets HostedCluster and Annotations to append
//The function appends the annotations to the HostedCluster
func AppendAnnotations(hostedCluster *v1alpha1.HostedCluster, annotationsToAppend map[string]string) {
	newAnnotations := hostedCluster.GetAnnotations()
	if len(newAnnotations) == 0 {
		newAnnotations = make(map[string]string)
	}
	for key, value := range annotationsToAppend {
		newAnnotations[key] = value
	}
	hostedCluster.SetAnnotations(newAnnotations)
}

//addClusterAdminAnnotation gets username of HostedCluster requester, HostedCluster and context
//The functions adds cluster-admin annotation with the username to the HostedCluster and updates it
func (r *HostedClusterReconciler) addClusterAdminAnnotation(username string, hostedClusterObject *v1alpha1.HostedCluster, ctx context.Context) {
	annotations := make(map[string]string)
	annotations[clusterAdminAnnotation] = username
	AppendAnnotations(hostedClusterObject, annotations)
	if err := r.Client.Update(ctx, hostedClusterObject); err != nil {
		r.Log.Error(err, "unable to update hosted cluster object")
	}
}

//getHostedClusterClient gets HostedCluster name and returns its client
func (r *HostedClusterReconciler) getHostedClusterClient(hostedclustername string) client.Client {
	//gets the HostedCluster client config
	hostedConfig, err := utills.GetHostedKubeRestConfig(r.Client, hostedclustername)
	if err != nil {
		r.Log.Error(err, "unable to get hosted cluster client")
	}
	//creates client from the client config
	hostedClusterClient, err := client.New(hostedConfig, client.Options{})
	if err != nil {
		r.Log.Error(err, "unable to get hosted cluster client")
	}
	return hostedClusterClient
}

//addClusterAdminRoleBinding gets HostedCluster client, HostedCluster requester username, the HostedCluster itself and context
//The function adds cluster-admin rolebinding to the username on the HostedCluster
func (r *HostedClusterReconciler) addClusterAdminRoleBinding(hostedClient client.Client, username string, hostedClusterObject *v1alpha1.HostedCluster, ctx context.Context) {
	clusterRoleBinding := composeClusterAdminCRB(username)
	err := hostedClient.Create(ctx, &clusterRoleBinding)
	if err != nil {
		r.Log.Error(err, "could not add cluster admin to the user")
	} else {
		r.addClusterAdminAnnotation(username, hostedClusterObject, ctx)
		r.Log.Info("user received cluster-admin role", "username", username)
	}
}
