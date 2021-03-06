// Copyright 2020 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	aggregator "github.com/vmware-tanzu/antrea/pkg/flowaggregator"
	"github.com/vmware-tanzu/antrea/pkg/signals"
)

func run(o *Options) error {
	klog.Infof("Flow aggregator starting...")
	// Set up signal capture: the first SIGTERM / SIGINT signal is handled gracefully and will
	// cause the stopCh channel to be closed; if another signal is received before the program
	// exits, we will force exit.
	stopCh := signals.RegisterSignalHandlers()

	k8sClient, err := createK8sClient()
	if err != nil {
		return fmt.Errorf("error when creating K8s client: %v", err)
	}
	flowAggregator := aggregator.NewFlowAggregator(o.externalFlowCollectorAddr, o.externalFlowCollectorProto, o.exportInterval, o.aggregatorTransportProtocol, o.flowAggregatorAddress, k8sClient)
	err = flowAggregator.InitCollectingProcess()
	if err != nil {
		return fmt.Errorf("error when creating collecting process: %v", err)
	}
	err = flowAggregator.InitAggregationProcess()
	if err != nil {
		return fmt.Errorf("error when creating aggregation process: %v", err)
	}
	go flowAggregator.Run(stopCh)
	<-stopCh
	klog.Infof("Stopping flow aggregator")
	return nil
}

func createK8sClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}
