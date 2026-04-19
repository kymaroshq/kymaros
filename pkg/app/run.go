/*
Copyright 2026.

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

// Package app provides the main entry point for the Kymaros operator.
// External wrappers can import this package and call Run() to start
// the operator with any plugins registered via pkg/plugin.
package app

import (
	"crypto/tls"
	"flag"
	"log/slog"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	kymarosapi "github.com/kymaroshq/kymaros/internal/api"
	"github.com/kymaroshq/kymaros/internal/controller"
	"github.com/kymaroshq/kymaros/internal/sandbox"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(restorev1alpha1.AddToScheme(scheme))
}

// Run starts the Kymaros operator. It parses flags, sets up the controller
// manager, API server, and health checks, then blocks until the context is
// cancelled. Any adapters, notifiers, or routes registered via pkg/plugin
// init() functions are available when the server starts.
//
// Run calls os.Exit on fatal errors and should be called from main().
func Run() {
	var metricsAddr string
	var metricsCertPath, metricsCertName, metricsCertKey string
	var webhookCertPath, webhookCertName, webhookCertKey string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var apiPort int
	var staticDir string
	var tlsOpts []func(*tls.Config)
	var notifSlackEnabled bool
	var notifSlackWebhookSecret string
	var notifSlackDefaultChannel string
	var notifWebhookEnabled bool
	var notifWebhookSecret string
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.StringVar(&webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	flag.StringVar(&webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	flag.StringVar(&webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.IntVar(&apiPort, "api-port", 8080, "Port for the API server and dashboard")
	flag.StringVar(&staticDir, "static-dir", "/static", "Path to frontend static files")
	flag.BoolVar(&notifSlackEnabled, "notification-slack-enabled", false,
		"Enable global Slack notifications when a RestoreTest has no per-test notification config.")
	flag.StringVar(&notifSlackWebhookSecret, "notification-slack-webhook-secret", "",
		"Name of the Secret (in operator namespace) containing the Slack webhook URL (key: url).")
	flag.StringVar(&notifSlackDefaultChannel, "notification-slack-default-channel", "#alerts",
		"Default Slack channel used for global notifications.")
	flag.BoolVar(&notifWebhookEnabled, "notification-webhook-enabled", false,
		"Enable global webhook notifications when a RestoreTest has no per-test notification config.")
	flag.StringVar(&notifWebhookSecret, "notification-webhook-secret", "",
		"Name of the Secret (in operator namespace) containing the generic webhook URL (key: url).")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	if !flag.Parsed() {
		flag.Parse()
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("Disabling HTTP/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookTLSOpts := tlsOpts
	webhookServerOptions := webhook.Options{
		TLSOpts: webhookTLSOpts,
	}

	if len(webhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", webhookCertPath, "webhook-cert-name", webhookCertName, "webhook-cert-key", webhookCertKey)

		webhookServerOptions.CertDir = webhookCertPath
		webhookServerOptions.CertName = webhookCertName
		webhookServerOptions.KeyName = webhookCertKey
	}

	webhookServer := webhook.NewServer(webhookServerOptions)

	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		metricsServerOptions.CertDir = metricsCertPath
		metricsServerOptions.CertName = metricsCertName
		metricsServerOptions.KeyName = metricsCertKey
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "08ce2b2c.kymaros.io",
	})
	if err != nil {
		setupLog.Error(err, "Failed to start manager")
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "Failed to create kubernetes clientset")
		os.Exit(1)
	}

	if err := (&controller.RestoreTestReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Sandbox:   sandbox.NewManager(mgr.GetClient(), slog.Default()),
		Clientset: clientset,
		GlobalNotifications: controller.GlobalNotificationDefaults{
			SlackEnabled:        notifSlackEnabled,
			SlackWebhookSecret:  notifSlackWebhookSecret,
			SlackDefaultChannel: notifSlackDefaultChannel,
			WebhookEnabled:      notifWebhookEnabled,
			WebhookSecret:       notifWebhookSecret,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Failed to create controller", "controller", "RestoreTest")
		os.Exit(1)
	}

	apiServer := kymarosapi.NewServer(mgr.GetClient(), apiPort, staticDir, mgr.GetConfig())
	if err := mgr.Add(apiServer); err != nil {
		setupLog.Error(err, "Failed to add API server")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Failed to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Failed to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Failed to run manager")
		os.Exit(1)
	}
}
