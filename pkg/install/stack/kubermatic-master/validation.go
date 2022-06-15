/*
Copyright 2020 The Kubermatic Kubernetes Platform contributors.

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

package kubermaticmaster

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	semverlib "github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"

	kubermaticv1 "k8c.io/kubermatic/v2/pkg/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/pkg/controller/operator/defaults"
	"k8c.io/kubermatic/v2/pkg/features"
	"k8c.io/kubermatic/v2/pkg/install/stack"
	"k8c.io/kubermatic/v2/pkg/install/util"
	k8csemver "k8c.io/kubermatic/v2/pkg/semver"
	"k8c.io/kubermatic/v2/pkg/serviceaccount"
	"k8c.io/kubermatic/v2/pkg/util/yamled"
)

func (m *MasterStack) ValidateState(ctx context.Context, opt stack.DeployOptions) []error {
	var errs []error

	// validation can only happen if KKP was already installed, otherwise the resource types
	// won't even be known by the kube-apiserver
	crdsExists, err := util.HasAllReadyCRDs(ctx, opt.KubeClient, []string{
		"clusters.kubermatic.k8c.io",
		"seeds.kubermatic.k8c.io",
	})
	if err != nil {
		return append(errs, fmt.Errorf("failed to check for CRDs: %w", err))
	}

	if !crdsExists {
		return nil // nothing to do
	}

	// we need the actual, effective versioning configuration, which most users will
	// probably not override
	defaulted, err := defaults.DefaultConfiguration(opt.KubermaticConfiguration, zap.NewNop().Sugar())
	if err != nil {
		return append(errs, fmt.Errorf("failed to apply default values to the KubermaticConfiguration: %w", err))
	}

	allSeeds, err := opt.SeedsGetter()
	if err != nil {
		return append(errs, fmt.Errorf("failed to list Seeds: %w", err))
	}

	upgradeConstraints, contraintErrs := getAutoUpdateConstraints(defaulted)
	if len(contraintErrs) > 0 {
		return contraintErrs
	}

	for seedName, seed := range allSeeds {
		opt.Logger.WithField("seed", seedName).Info("Checking seed cluster…")

		// create client into seed
		seedClient, err := opt.SeedClientGetter(seed)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create client for Seed cluster %q: %w", seedName, err))
			continue
		}

		// list all userclusters
		clusters := kubermaticv1.ClusterList{}
		if err := seedClient.List(ctx, &clusters); err != nil {
			errs = append(errs, fmt.Errorf("failed to list user clusters on Seed %q: %w", seedName, err))
			continue
		}

		// check that each cluster still matches the configured versions
		for _, cluster := range clusters.Items {
			clusterVersion := cluster.Spec.Version

			if !clusterVersionIsConfigured(clusterVersion, defaulted, upgradeConstraints) {
				errs = append(errs, fmt.Errorf("cluster %s (version %s) on Seed %s would not be supported anymore", cluster.Name, clusterVersion, seedName))
			}
		}
	}

	return errs
}

func getAutoUpdateConstraints(defaultedConfig *kubermaticv1.KubermaticConfiguration) ([]*semverlib.Constraints, []error) {
	var errs []error

	upgradeConstraints := []*semverlib.Constraints{}

	for i, update := range defaultedConfig.Spec.Versions.Updates {
		// only consider automated updates, otherwise we might accept an unsupported
		// cluster that is never manually updated
		if update.Automatic == nil || !*update.Automatic {
			continue
		}

		from, err := semverlib.NewConstraint(update.From)
		if err != nil {
			errs = append(errs, fmt.Errorf("`from` constraint %q for update rule %d is invalid: %w", update.From, i, err))
			continue
		}

		upgradeConstraints = append(upgradeConstraints, from)
	}

	return upgradeConstraints, errs
}

func clusterVersionIsConfigured(version k8csemver.Semver, defaultedConfig *kubermaticv1.KubermaticConfiguration, constraints []*semverlib.Constraints) bool {
	// is this version still straight up supported?
	for _, configured := range defaultedConfig.Spec.Versions.Versions {
		if configured.Equal(&version) {
			return true
		}
	}

	sversion := version.Semver()

	// is an upgrade path defined from the current version to something else?
	for _, update := range constraints {
		if update.Check(sversion) {
			return true
		}
	}

	return false
}

func (*MasterStack) ValidateConfiguration(config *kubermaticv1.KubermaticConfiguration, helmValues *yamled.Document, opt stack.DeployOptions, logger logrus.FieldLogger) (*kubermaticv1.KubermaticConfiguration, *yamled.Document, []error) {
	kubermaticFailures := validateKubermaticConfiguration(config)
	for idx, e := range kubermaticFailures {
		kubermaticFailures[idx] = prefixError("KubermaticConfiguration: ", e)
	}

	helmFailures := validateHelmValues(config, helmValues, opt, logger)
	for idx, e := range helmFailures {
		helmFailures[idx] = prefixError("Helm values: ", e)
	}

	return config, helmValues, append(kubermaticFailures, helmFailures...)
}

func validateKubermaticConfiguration(config *kubermaticv1.KubermaticConfiguration) []error {
	failures := []error{}

	if config.Namespace != KubermaticOperatorNamespace {
		failures = append(failures, errors.New("the namespace must be \"kubermatic\""))
	}

	if config.Spec.Ingress.Domain == "" {
		failures = append(failures, errors.New("spec.ingress.domain cannot be left empty"))
	}

	// only validate auth-related keys if we are not setting up a headless system
	if !config.Spec.FeatureGates[features.HeadlessInstallation] {
		failures = validateRandomSecret(config, config.Spec.Auth.ServiceAccountKey, "spec.auth.serviceAccountKey", failures)

		if err := serviceaccount.ValidateKey([]byte(config.Spec.Auth.ServiceAccountKey)); err != nil {
			failures = append(failures, fmt.Errorf("spec.auth.serviceAccountKey is invalid: %w", err))
		}

		if config.Spec.FeatureGates[features.OIDCKubeCfgEndpoint] {
			failures = validateRandomSecret(config, config.Spec.Auth.IssuerClientSecret, "spec.auth.issuerClientSecret", failures)
			failures = validateRandomSecret(config, config.Spec.Auth.IssuerCookieKey, "spec.auth.issuerCookieKey", failures)
		}
	}

	return failures
}

func validateRandomSecret(config *kubermaticv1.KubermaticConfiguration, value string, path string, failures []error) []error {
	if value == "" {
		secret, err := randomString()
		if err == nil {
			failures = append(failures, fmt.Errorf("%s must be a non-empty secret, for example: %s", path, secret))
		} else {
			failures = append(failures, fmt.Errorf("%s must be a non-empty secret", path))
		}
	}

	return failures
}

type dexClient struct {
	ID string `yaml:"id"`
}

func validateHelmValues(config *kubermaticv1.KubermaticConfiguration, helmValues *yamled.Document, opt stack.DeployOptions, logger logrus.FieldLogger) []error {
	failures := []error{}

	path := yamled.Path{"kubermaticOperator", "imagePullSecret"}
	if value, _ := helmValues.GetString(path); value == "" {
		logger.Warnf("Helm values: %s is empty, setting to spec.imagePullSecret from KubermaticConfiguration", path.String())
		helmValues.Set(path, config.Spec.ImagePullSecret)
	}

	if !config.Spec.FeatureGates[features.HeadlessInstallation] {
		path := yamled.Path{"dex", "ingress", "host"}
		if domain, _ := helmValues.GetString(path); domain == "" {
			logger.WithField("domain", config.Spec.Ingress.Domain).Warnf("Helm values: %s is empty, setting to spec.ingress.domain from KubermaticConfiguration", path.String())
			helmValues.Set(path, config.Spec.Ingress.Domain)
		}
	}

	defaultedConfig, err := defaults.DefaultConfiguration(config, zap.NewNop().Sugar())
	if err != nil {
		failures = append(failures, fmt.Errorf("failed to process KubermaticConfiguration: %w", err))
		return failures // must stop here, without defaulting the clientID check can be misleading
	}

	if !config.Spec.FeatureGates[features.HeadlessInstallation] {
		clientID := defaultedConfig.Spec.Auth.ClientID
		hasDexIssues := false
		clients := []dexClient{}

		if err := helmValues.DecodeAtPath(yamled.Path{"dex", "clients"}, &clients); err != nil {
			hasDexIssues = true
			logger.Warn("Helm values: There are no Dex/OAuth clients configured.")
		} else {
			hasMatchingClient := false

			for _, client := range clients {
				if client.ID == clientID {
					hasMatchingClient = true
					break
				}
			}

			if !hasMatchingClient {
				hasDexIssues = true
				logger.Warnf("Helm values: The Dex configuration does not contain a `%s` client to allow logins to the Kubermatic dashboard.", clientID)
			}
		}

		connectors, _ := helmValues.GetArray(yamled.Path{"dex", "connectors"})
		staticPasswords, _ := helmValues.GetArray(yamled.Path{"dex", "staticPasswords"})

		if len(connectors) == 0 && len(staticPasswords) == 0 {
			hasDexIssues = true
			logger.Warn("Helm values: There are no connectors or static passwords configured for Dex.")
		}

		if hasDexIssues {
			logger.Warnf("If you intend to use Dex, please refer to the example configuration to define a `%s` client and connectors.", clientID)
		}
	}

	return failures
}

func prefixError(prefix string, e error) error {
	return fmt.Errorf("%s%w", prefix, e)
}

func randomString() (string, error) {
	c := 32
	b := make([]byte, c)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
