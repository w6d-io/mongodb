/*
Copyright 2021 WILDCARD SA.

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

package main

import (
	"flag"
	"os"

	zapraw "go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/w6d-io/mongodb/internal/config"
	"github.com/w6d-io/mongodb/internal/util"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dbv1alpha1 "github.com/w6d-io/mongodb/api/v1alpha1"
	"github.com/w6d-io/mongodb/controllers"
	//+kubebuilder:scaffold:imports
)

// Version microservice version
var Version = ""

// Revision git commit
var Revision = ""

// GoVersion ...
var GoVersion = ""

// Built Date built
var Built = ""

// OsArch ...
var OsArch = ""

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(dbv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", util.LookupEnvOrString("METRICS_ADDRESS", ":8080"), "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", util.LookupEnvOrString("PROBE_ADDRESS", ":8081"), "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", util.LookupEnvOrBool("ENABLE_LEADER", false),
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development:     os.Getenv("RELEASE") != "prod",
		StacktraceLevel: zapcore.PanicLevel,
		Encoder:         zapcore.NewConsoleEncoder(util.TextEncoderConfig()),
	}
	//opts.BindFlags(flag.CommandLine)
	config.BindFlag(flag.CommandLine)
	util.BindFlags(&opts, flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts), zap.RawZapOpts(zapraw.AddCaller(), zapraw.AddCallerSkip(-1))))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "644757cd.w6d.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.MongoDBReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MongoDB"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MongoDB")
		os.Exit(1)
	}
	if err = (&controllers.MongoDBUserReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MongoDBUser"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MongoDBUser")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&dbv1alpha1.MongoDB{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "MongoDB")
			os.Exit(1)
		}
		if err = (&dbv1alpha1.MongoDBUser{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "MongoDBUser")
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager", "Version", Version, "Built",
		Built, "Revision", Revision, "Arch", OsArch, "GoVersion", GoVersion)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
