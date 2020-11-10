package poolmgr

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func (gp *GenericPool) startReadyPodController() {
	// create the pod watcher to filter by labels
	optionsModifier := func(options *metav1.ListOptions) {
		options.LabelSelector = labels.Set(
			gp.deployment.Spec.Selector.MatchLabels).AsSelector().String()
		options.FieldSelector = "status.phase=Running"
	}
	readyPodWatcher := cache.NewFilteredListWatchFromClient(gp.kubernetesClient.CoreV1().RESTClient(), "pods", gp.namespace, optionsModifier)

	gp.readyPodQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	gp.readyPodIndexer, gp.readyPodController = cache.NewIndexerInformer(readyPodWatcher, &apiv1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				gp.readyPodQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				gp.readyPodQueue.Forget(key)
				gp.readyPodQueue.Done(key)
			}
		},
	}, cache.Indexers{})
	go gp.readyPodController.Run(gp.stopReadyPodControllerCh)
}
