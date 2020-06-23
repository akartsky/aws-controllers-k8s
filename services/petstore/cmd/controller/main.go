// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"os"

	ackrt "github.com/aws/aws-service-operator-k8s/pkg/runtime"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlrt "sigs.k8s.io/controller-runtime"

	svctypes "github.com/aws/aws-service-operator-k8s/services/petstore/apis/v1alpha1"
	svcresource "github.com/aws/aws-service-operator-k8s/services/petstore/pkg/resource"
)

var (
	awsServiceAPIGroup = "petstore.services.k8s.aws"
	awsServiceAlias    = "petstore"
	scheme             = runtime.NewScheme()
	setupLog           = ctrlrt.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = svctypes.AddToScheme(scheme)
}

func main() {
	var ackCfg ackrt.Config
	ackCfg.BindFlags()
	flag.Parse()

	mgr, err := ctrlrt.NewManager(ctrlrt.GetConfigOrDie(), ctrlrt.Options{
		Scheme:             scheme,
		Port:               ackCfg.BindPort,
		MetricsBindAddress: ackCfg.MetricsAddr,
		LeaderElection:     ackCfg.EnableLeaderElection,
		LeaderElectionID:   awsServiceAPIGroup,
	})
	if err != nil {
		setupLog.Error(
			err, "unable to create controller manager",
			"aws.service", awsServiceAlias,
		)
		os.Exit(1)
	}

	stopChan := ctrlrt.SetupSignalHandler()

	setupLog.Info(
		"initializing service controller",
		"aws.service", awsServiceAlias,
	)
	sc := ackrt.NewServiceController(
		awsServiceAlias, awsServiceAPIGroup,
	).WithLogger(
		ctrlrt.Log,
	).WithResourceManagerFactories(
		svcresource.GetManagerFactories(),
	)
	if err = sc.BindControllerManager(mgr); err != nil {
		setupLog.Error(
			err, "unable bind to controller manager to service controller",
			"aws.service", awsServiceAlias,
		)
		os.Exit(1)
	}

	setupLog.Info(
		"starting manager",
		"aws.service", awsServiceAlias,
	)
	if err := mgr.Start(stopChan); err != nil {
		setupLog.Error(
			err, "unable to start controller manager",
			"aws.service", awsServiceAlias,
		)
		os.Exit(1)
	}
}