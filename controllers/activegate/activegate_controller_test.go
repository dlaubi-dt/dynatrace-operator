package activegate

import (
	"context"
	"testing"

	"github.com/Dynatrace/dynatrace-operator/api/v1alpha1"
	"github.com/Dynatrace/dynatrace-operator/controllers/kubemon"
	"github.com/Dynatrace/dynatrace-operator/controllers/kubesystem"
	"github.com/Dynatrace/dynatrace-operator/dtclient"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	testUID       = "test-uid"
	testPaasToken = "test-paas-token"
)

func init() {
	_ = v1alpha1.AddToScheme(scheme.Scheme) // Register OneAgent and Istio object schemas.
}

func TestReconcileActiveGate_Reconcile(t *testing.T) {
	t.Run(`Reconcile works with minimal setup`, func(t *testing.T) {
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
		r := &ReconcileActiveGate{
			client: fakeClient,
		}
		result, err := r.Reconcile(reconcile.Request{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
	t.Run(`Reconcile works with minimal setup and interface`, func(t *testing.T) {
		mockClient := &dtclient.MockDynatraceClient{}
		instance := &v1alpha1.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: testNamespace,
			}}
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, instance,
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testName,
					Namespace: testNamespace,
				}})
		r := &ReconcileActiveGate{
			client: fakeClient,
			dtcBuildFunc: func(_ client.Client, _ *v1alpha1.DynaKube, _ *corev1.Secret) (dtclient.Client, error) {
				return mockClient, nil
			},
		}
		result, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: testName},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
	t.Run(`Reconcile reconciles Kubernetes Monitoring if enabled`, func(t *testing.T) {
		mockClient := &dtclient.MockDynatraceClient{}
		instance := &v1alpha1.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: testNamespace,
			},
			Spec: v1alpha1.DynaKubeSpec{
				KubernetesMonitoringSpec: v1alpha1.KubernetesMonitoringSpec{
					Enabled: true,
				}}}
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, instance,
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testName,
					Namespace: testNamespace,
				},
				Data: map[string][]byte{dtclient.DynatracePaasToken: []byte(testPaasToken)}},
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: kubesystem.Namespace,
					UID:  testUID,
				}})
		r := &ReconcileActiveGate{
			client:    fakeClient,
			apiReader: fakeClient,
			scheme:    scheme.Scheme,
			dtcBuildFunc: func(_ client.Client, _ *v1alpha1.DynaKube, _ *corev1.Secret) (dtclient.Client, error) {
				return mockClient, nil
			},
		}

		mockClient.
			On("GetConnectionInfo").
			Return(dtclient.ConnectionInfo{
				TenantUUID: testName,
			}, nil)

		result, err := r.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: testName},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)

		var statefulSet appsv1.StatefulSet
		err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: instance.Name + kubemon.StatefulSetSuffix, Namespace: testNamespace}, &statefulSet)

		assert.NoError(t, err)
		assert.NotNil(t, statefulSet)
	})
}
