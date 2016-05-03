package kubelego

import (
	"errors"
	"reflect"

	k8sApi "k8s.io/kubernetes/pkg/api"
	k8sExtensions "k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util"
)

func (kl *KubeLego) InitKube() error {

	// Try in cluster client first
	kubeClient, err := client.NewInCluster()
	if err != nil {
		kl.Log().Warnf("failed to create in-cluster client: %v.", err)

		// fall back to 127.0.0.1:8080 for dev
		kubeClient, err = client.New(
			&client.Config{
				Host: "127.0.0.1:8080",
			},
		)
		if err != nil {
			kl.Log().Warnf("failed to create test cluster client: %v.", err)
			return errors.New("kube init failed as both in-cluster and dev connection unavailable")
		}
	}

	version, err := kubeClient.ServerVersion()
	if err != nil {
		return err
	}
	kl.Log().Infof("connected to kubernetes api %s", version.String())

	kl.kubeClient = kubeClient
	return nil
}

func (kl *KubeLego) Namespace() string {
	return kl.LegoNamespace
}

func (kl *KubeLego) WatchConfig() {

	oldList := &k8sExtensions.IngressList{}

	rateLimiter := util.NewTokenBucketRateLimiter(0.1, 1)

	ingClient := kl.kubeClient.Extensions().Ingress(k8sApi.NamespaceAll)

	for {
		rateLimiter.Accept()

		list, err := ingClient.List(k8sApi.ListOptions{})
		if err != nil {
			kl.Log().Warn("Error while retrieving ingress list: ", err)
			continue
		}

		if reflect.DeepEqual(oldList, list) {
			continue
		}
		oldList = list

		kl.Reconfigure()

	}

}
