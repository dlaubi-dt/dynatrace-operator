package activegate

/**
The following functions have been copied from dynatrace-oneagent-operator
and are candidates to be made into a library:

* BuildDynatraceClient
* verifySecret
* getTokensName

*/

import (
	"context"
	"fmt"

	dynatracev1alpha1 "github.com/Dynatrace/dynatrace-operator/api/v1alpha1"
	"github.com/Dynatrace/dynatrace-operator/dtclient"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type options struct {
	Opts []dtclient.Option
}

// BuildDynatraceClient creates a new Dynatrace client using the settings configured on the given instance.
func BuildDynatraceClient(rtc client.Client, instance *dynatracev1alpha1.DynaKube, secret *corev1.Secret) (dtclient.Client, error) {
	if instance == nil {
		return nil, fmt.Errorf("could not build dynatrace client: instance is nil")
	}
	namespace := instance.GetNamespace()
	spec := instance.Spec
	tokens, err := NewTokens(secret)
	if err != nil {
		return nil, err
	}

	opts := newOptions()
	opts.appendCertCheck(&spec)
	opts.appendNetworkZone(&spec)

	err = opts.appendProxySettings(rtc, &spec, namespace)
	if err != nil {
		return nil, err
	}

	err = opts.appendTrustedCerts(rtc, &spec, namespace)
	if err != nil {
		return nil, err
	}

	return dtclient.NewClient(spec.APIURL, tokens.ApiToken, tokens.PaasToken, opts.Opts...)
}

func newOptions() *options {
	return &options{
		Opts: []dtclient.Option{},
	}
}

func (opts *options) appendNetworkZone(spec *dynatracev1alpha1.DynaKubeSpec) {
	if spec.NetworkZone != "" {
		opts.Opts = append(opts.Opts, dtclient.NetworkZone(spec.NetworkZone))
	}
}

func (opts *options) appendCertCheck(spec *dynatracev1alpha1.DynaKubeSpec) {
	opts.Opts = append(opts.Opts, dtclient.SkipCertificateValidation(spec.SkipCertCheck))
}

func (opts *options) appendProxySettings(rtc client.Client, spec *dynatracev1alpha1.DynaKubeSpec, namespace string) error {
	if p := spec.Proxy; p != nil {
		if p.ValueFrom != "" {
			proxySecret := &corev1.Secret{}
			err := rtc.Get(context.TODO(), client.ObjectKey{Name: p.ValueFrom, Namespace: namespace}, proxySecret)
			if err != nil {
				return fmt.Errorf("failed to get proxy secret: %w", err)
			}

			proxyURL, err := ExtractToken(proxySecret, Proxy)
			if err != nil {
				return fmt.Errorf("failed to extract proxy secret field: %w", err)
			}
			opts.Opts = append(opts.Opts, dtclient.Proxy(proxyURL))
		} else if p.Value != "" {
			opts.Opts = append(opts.Opts, dtclient.Proxy(p.Value))
		}
	}
	return nil
}

func (opts *options) appendTrustedCerts(rtc client.Client, spec *dynatracev1alpha1.DynaKubeSpec, namespace string) error {
	if spec.TrustedCAs != "" {
		certs := &corev1.ConfigMap{}
		if err := rtc.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: spec.TrustedCAs}, certs); err != nil {
			return fmt.Errorf("failed to get certificate configmap: %w", err)
		}
		if certs.Data[Certificates] == "" {
			return fmt.Errorf("failed to extract certificate configmap field: missing field certs")
		}
		opts.Opts = append(opts.Opts, dtclient.Certs([]byte(certs.Data[Certificates])))
	}
	return nil
}

const (
	Proxy        = "proxy"
	Certificates = "certs"
)
