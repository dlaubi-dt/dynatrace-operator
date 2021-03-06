package dtpullsecret

import (
	"context"
	"testing"

	dynatracev1alpha1 "github.com/Dynatrace/dynatrace-operator/api/v1alpha1"
	"github.com/Dynatrace/dynatrace-operator/dtclient"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	testEndpoint  = "http://test-endpoint.com/api"
	testPaasToken = "test-paas-token"
)

func init() {
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))

	utilruntime.Must(dynatracev1alpha1.AddToScheme(scheme.Scheme))
	// +kubebuilder:scaffold:scheme
}

func TestReconciler_Reconcile(t *testing.T) {
	t.Run(`Reconcile works with minimal setup`, func(t *testing.T) {
		mockDTC := &dtclient.MockDynatraceClient{}
		instance := &dynatracev1alpha1.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
				Name:      testName,
			}}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
				Name:      testName,
			},
			Data: map[string][]byte{dtclient.DynatracePaasToken: []byte(testPaasToken)},
		}
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
		r := NewReconciler(fakeClient, fakeClient, scheme.Scheme, instance, mockDTC, logf.Log, secret, "")

		mockDTC.
			On("GetConnectionInfo").
			Return(dtclient.ConnectionInfo{}, nil)

		err := r.Reconcile()

		assert.NoError(t, err)

		var pullSecret corev1.Secret
		err = fakeClient.Get(context.TODO(),
			client.ObjectKey{Name: testName + "-pull-secret", Namespace: testNamespace},
			&pullSecret)

		assert.NoError(t, err)
		assert.NotNil(t, pullSecret)
		assert.NotEmpty(t, pullSecret.Data)
		assert.Contains(t, pullSecret.Data, ".dockerconfigjson")
		assert.NotEmpty(t, pullSecret.Data[".dockerconfigjson"])
	})
	t.Run(`Reconcile does not reconcile with custom pull secret or custom image`, func(t *testing.T) {
		instance := &dynatracev1alpha1.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
				Name:      testName,
			},
			Spec: dynatracev1alpha1.DynaKubeSpec{
				CustomPullSecret: testValue,
			}}
		r := NewReconciler(nil, nil, nil, instance, nil, nil, nil, "")
		err := r.Reconcile()

		assert.NoError(t, err)

		instance = &dynatracev1alpha1.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
				Name:      testName,
			}}
		r = NewReconciler(nil, nil, nil, instance, nil, nil, nil, testValue)
		err = r.Reconcile()

		assert.NoError(t, err)
	})
	t.Run(`Reconcile creates correct docker config`, func(t *testing.T) {
		expectedJson := []byte(`{"Auths":{"test-endpoint.com":{"Username":"test-name","Password":"test-value","Auth":"dGVzdC1uYW1lOnRlc3QtdmFsdWU="}}}`)
		mockDTC := &dtclient.MockDynatraceClient{}
		instance := &dynatracev1alpha1.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
				Name:      testName,
			},
			Spec: dynatracev1alpha1.DynaKubeSpec{
				APIURL: testEndpoint,
			}}
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
		r := NewReconciler(fakeClient, fakeClient, scheme.Scheme, instance, mockDTC, logf.Log,
			&corev1.Secret{
				Data: map[string][]byte{
					dtclient.DynatracePaasToken: []byte(testValue),
				},
			}, "")

		mockDTC.
			On("GetConnectionInfo").
			Return(dtclient.ConnectionInfo{
				TenantUUID: testName,
			}, nil)

		err := r.Reconcile()

		assert.NoError(t, err)

		var pullSecret corev1.Secret
		err = fakeClient.Get(context.TODO(),
			client.ObjectKey{Name: testName + "-pull-secret", Namespace: testNamespace},
			&pullSecret)

		assert.NoError(t, err)
		assert.NotNil(t, pullSecret)
		assert.NotEmpty(t, pullSecret.Data)
		assert.Contains(t, pullSecret.Data, ".dockerconfigjson")
		assert.NotEmpty(t, pullSecret.Data[".dockerconfigjson"])
		assert.Equal(t, expectedJson, pullSecret.Data[".dockerconfigjson"])
	})
	t.Run(`Reconcile update secret if data changed`, func(t *testing.T) {
		expectedJson := []byte(`{"Auths":{"test-endpoint.com":{"Username":"test-name","Password":"test-value","Auth":"dGVzdC1uYW1lOnRlc3QtdmFsdWU="}}}`)
		mockDTC := &dtclient.MockDynatraceClient{}
		instance := &dynatracev1alpha1.DynaKube{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace,
				Name:      testName,
			},
			Spec: dynatracev1alpha1.DynaKubeSpec{
				APIURL: testEndpoint,
			}}
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
		r := NewReconciler(fakeClient, fakeClient, scheme.Scheme, instance, mockDTC, logf.Log,
			&corev1.Secret{
				Data: map[string][]byte{
					dtclient.DynatracePaasToken: []byte(testValue),
				},
			}, "")

		mockDTC.
			On("GetConnectionInfo").
			Return(dtclient.ConnectionInfo{
				TenantUUID: testName,
			}, nil)

		err := r.Reconcile()

		assert.NoError(t, err)

		var pullSecret corev1.Secret
		err = fakeClient.Get(context.TODO(),
			client.ObjectKey{Name: testName + "-pull-secret", Namespace: testNamespace},
			&pullSecret)

		assert.NoError(t, err)

		pullSecret.Data = nil
		err = fakeClient.Update(context.TODO(), &pullSecret)

		assert.NoError(t, err)

		err = r.Reconcile()

		assert.NoError(t, err)

		err = fakeClient.Get(context.TODO(),
			client.ObjectKey{Name: testName + "-pull-secret", Namespace: testNamespace},
			&pullSecret)

		assert.NoError(t, err)
		assert.NotNil(t, pullSecret)
		assert.NotEmpty(t, pullSecret.Data)
		assert.Contains(t, pullSecret.Data, ".dockerconfigjson")
		assert.NotEmpty(t, pullSecret.Data[".dockerconfigjson"])
		assert.Equal(t, expectedJson, pullSecret.Data[".dockerconfigjson"])
	})
}
