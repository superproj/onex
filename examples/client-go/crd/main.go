package main

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	namespace, _, err := configLoader.Namespace()
	if err != nil {
		panic(err)
	}

	cfg, err := configLoader.ClientConfig()
	if err != nil {
		panic(err)
	}

	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	// identify out custom resource
	gvr := schema.GroupVersionResource{
		Group:    "webapp.my.domain",
		Version:  "v1",
		Resource: "guestbooks",
	}
	// retrieve the resource of kind Pizza named 'margherita'
	res, err := dc.Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), "guestbook-sample", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return
		}
		panic(err)
	}
	data, _ := json.Marshal(res)
	fmt.Println(string(data))

	// grab the status if exists
	status, ok := res.Object["status"]
	if !ok {
		// otherwise create it
		status = make(map[string]interface{})
	}

	// change the 'margherita' price
	status.(map[string]interface{})["cost"] = 6.50
	res.Object["status"] = status

	// update the 'margherita' custom resource with the new price
	_, err = dc.Resource(gvr).Namespace(namespace).Update(context.TODO(), res, metav1.UpdateOptions{})
	if err != nil {
		panic(err)
	}
}
