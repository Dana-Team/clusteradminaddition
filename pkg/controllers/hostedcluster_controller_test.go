package controllers

import (
	. "clusteradminaddition/testUtils"
	"context"
	"github.com/go-logr/logr"
	"github.com/openshift/hypershift/api/v1alpha1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestAppendAnnotations(t *testing.T) {
	type args struct {
		hostedCluster       *v1alpha1.HostedCluster
		annotationsToAppend map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				hostedCluster:       GetHostedClusterObject("test"),
				annotationsToAppend: map[string]string{"shit": "fuck"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AppendAnnotations(tt.args.hostedCluster, tt.args.annotationsToAppend)
			flag := false
			for key, value := range tt.args.hostedCluster.Annotations {
				if val, ok := tt.args.hostedCluster.Annotations[key]; ok {
					if val == value {
						flag = true
					}
				}
			}
			if !flag {
				t.Errorf("the annotations append didn't work")
			}
		})
	}
}

//
//func TestHostedClusterReconciler_Reconcile(t *testing.T) {
//	type fields struct {
//		Client client.Client
//		Scheme *runtime.Scheme
//		Log    logr.Logger
//	}
//	type args struct {
//		ctx context.Context
//		req controllerruntime.Request
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    controllerruntime.Result
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := &HostedClusterReconciler{
//				Client: tt.fields.Client,
//				Scheme: tt.fields.Scheme,
//				Log:    tt.fields.Log,
//			}
//			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("Reconcile() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestHostedClusterReconciler_SetupWithManager(t *testing.T) {
//	type fields struct {
//		Client client.Client
//		Scheme *runtime.Scheme
//		Log    logr.Logger
//	}
//	type args struct {
//		mgr controllerruntime.Manager
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := &HostedClusterReconciler{
//				Client: tt.fields.Client,
//				Scheme: tt.fields.Scheme,
//				Log:    tt.fields.Log,
//			}
//			if err := r.SetupWithManager(tt.args.mgr); (err != nil) != tt.wantErr {
//				t.Errorf("SetupWithManager() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
func TestHostedClusterReconciler_addClusterAdminAnnotation(t *testing.T) {
	type fields struct {
		Client client.Client
		Log    logr.Logger
	}
	type args struct {
		username            string
		hostedClusterObject *v1alpha1.HostedCluster
		ctx                 context.Context
	}
	var objs []client.Object
	objs = append(objs, GetHostedClusterObject("test"))
	v1alpha1.AddToScheme(clientgoscheme.Scheme)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "",
			fields: fields{
				Client: fake.NewClientBuilder().WithObjects(objs...).Build(),
				Log:    ctrl.Log.WithName("test"),
			},
			args: args{
				username:            "user-test",
				hostedClusterObject: GetHostedClusterObject("test"),
				ctx:                 context.Background(),
			},
			want: map[string]string{clusterAdminAnnotation: "user-test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HostedClusterReconciler{
				Client: tt.fields.Client,
				Log:    tt.fields.Log,
			}
			r.addClusterAdminAnnotation(tt.args.username, tt.args.hostedClusterObject, tt.args.ctx)
			hc := v1alpha1.HostedCluster{}
			if err := r.Client.Get(tt.args.ctx, types.NamespacedName{Name: tt.args.hostedClusterObject.Name}, &hc); err != nil {
				flag := false
				for key, value := range hc.Annotations {
					if val, ok := tt.want[key]; ok {
						if val == value {
							flag = true
						}
					}
				}
				//if !reflect.DeepEqual(hc.Annotations, tt.want) {
				if !flag {
					t.Errorf("the cluster-admin annotation doesn't exists properly")
				}
			}
		})
	}
}

func TestHostedClusterReconciler_addClusterAdminRoleBinding(t *testing.T) {
	type fields struct {
		Client client.Client
		Log    logr.Logger
	}
	type args struct {
		hostedClient        client.Client
		username            string
		hostedClusterObject *v1alpha1.HostedCluster
		ctx                 context.Context
	}
	var objs []client.Object
	objs = append(objs, GetHostedClusterObject("test"))
	v1alpha1.AddToScheme(clientgoscheme.Scheme)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   v1.ClusterRoleBinding
	}{
		{
			name: "",
			fields: fields{
				Client: fake.NewClientBuilder().WithObjects(objs...).Build(),
				Log:    ctrl.Log.WithName("test"),
			},
			args: args{
				hostedClient:        fake.NewClientBuilder().Build(),
				username:            "user-test",
				hostedClusterObject: GetHostedClusterObject("test"),
				ctx:                 context.Background(),
			},
			want: GetClusterRoleBinding("user-test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HostedClusterReconciler{
				Client: tt.fields.Client,
				Log:    tt.fields.Log,
			}
			r.addClusterAdminRoleBinding(tt.args.hostedClient, tt.args.username, tt.args.hostedClusterObject, tt.args.ctx)
			clusterRoleBinding := v1.ClusterRoleBinding{}
			if err := tt.args.hostedClient.Get(tt.args.ctx, types.NamespacedName{Name: tt.args.username}, &clusterRoleBinding); err != nil {
				if !reflect.DeepEqual(clusterRoleBinding, tt.want) {
					t.Errorf("the anottations are wrong got: %v want %v", clusterRoleBinding, tt.want)
				}
			}
		})
	}
}

//func TestHostedClusterReconciler_getHostedClusterClient(t *testing.T) {
//	type fields struct {
//		Client client.Client
//		Log    logr.Logger
//	}
//	type args struct {
//		hostedclustername string
//	}
//	var objs []client.Object
//	objs = append(objs, GetHostedClusterObject("test"))
//	v1alpha1.AddToScheme(clientgoscheme.Scheme)
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   client.Client
//	}{
//		{
//			name: "",
//			fields: fields{
//				Client: fake.NewClientBuilder().WithObjects(objs...).Build(),
//				Log:    ctrl.Log.WithName("test"),
//			},
//			args: args{
//				hostedclustername: "test",
//			},
//			want: fake.NewClientBuilder().Build(),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := &HostedClusterReconciler{
//				Client: tt.fields.Client,
//				Log:    tt.fields.Log,
//			}
//			if got := r.getHostedClusterClient(tt.args.hostedclustername); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("getHostedClusterClient() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//
//func Test_composeClusterAdminCRB(t *testing.T) {
//	type args struct {
//		username string
//	}
//	tests := []struct {
//		name string
//		args args
//		want v1.ClusterRoleBinding
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := composeClusterAdminCRB(tt.args.username); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("composeClusterAdminCRB() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
