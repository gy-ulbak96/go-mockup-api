package helmclient

import (
	"context"
	"fmt"
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

var storage = repo.File{}

// const (
// 	defaultCachePath            = "/tmp/.helmcache"
// 	defaultRepositoryConfigPath = "/tmp/.helmrepo"
// )

func (c *HelmClient) getChart(chartName string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, string, error) {
	chartPath, err := chartPathOptions.LocateChart(chartName, c.Settings)
	if err != nil {
		return nil, "", err
	}

	helmChart, err := loader.Load(chartPath)
	if err != nil {
		return nil, "", err
	}

	return helmChart, chartPath, err
}

func (c *HelmClient) AddOrUpdateChartRepo(entry repo.Entry) error {
	return c.addOrUpdateChartRepo(entry)
}

func (c *HelmClient) addOrUpdateChartRepo(entry repo.Entry) error {
	chartRepo, err := repo.NewChartRepository(&entry, c.Providers)
	if err != nil {
		return err
	}

	chartRepo.CachePath = c.Settings.RepositoryCache

	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return err
	}

	if c.storage.Has(entry.Name) {
		//c.DebugLog("WARNING: repository name %q already exists", entry.Name)
		return nil
	}

	c.storage.Update(&entry)
	err = c.storage.WriteFile(c.Settings.RepositoryConfig, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func (c *HelmClient) UpdateChartRepos() error {
	return c.updateChartRepos()
}

func (c *HelmClient) updateChartRepos() error {
	for _, entry := range c.storage.Repositories {
		chartRepo, err := repo.NewChartRepository(entry, c.Providers)
		if err != nil {
			return err
		}

		chartRepo.CachePath = c.Settings.RepositoryCache
		_, err = chartRepo.DownloadIndexFile()
		if err != nil {
			return err
		}

		c.storage.Update(entry)
	}

	return c.storage.WriteFile(c.Settings.RepositoryConfig, 0o644)
}

func (c *HelmClient) GetReleases() ([]*release.Release, error) {
	return c.getReleases()
}

func (c *HelmClient) getReleases() ([]*release.Release, error) {
	client := action.NewList(c.ActionConfig)
	client.Deployed = true
	rel, err := client.Run()
	if err != nil {
		return rel, err
	}

	return rel, nil
}

func (c *HelmClient) GetRelease(releaseName string) (*release.Release, error) {
	return c.getRelease(releaseName)
}

func (c *HelmClient) getRelease(releaseName string) (*release.Release, error) {
	client := action.NewGet(c.ActionConfig)
	rel, err := client.Run(releaseName)
	if err != nil {
		return rel, err
	}

	return rel, nil
}

func (c *HelmClient) Install(ctx context.Context, spec *ChartSpec) (*release.Release, error) {
	return c.install(ctx, spec)
}

func (c *HelmClient) install(ctx context.Context, spec *ChartSpec) (*release.Release, error) {
	client := action.NewInstall(c.ActionConfig)
	mergeInstallOptions(spec, client)

	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}

	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return nil, err
	}

	helmChart, err = updateDependencies(helmChart, &client.ChartPathOptions, chartPath, c, client.DependencyUpdate, spec)
	if err != nil {
		return nil, err
	}

	values, err := getValuesMap(spec)
	if err != nil {
		return nil, err
	}

	if err := c.lint(chartPath, values); err != nil {
		return nil, err
	}

	rel, err := client.RunWithContext(ctx, helmChart, values)
	if err != nil {
		return rel, err
	}

	return rel, nil
}

func (c *HelmClient) Upgrade(ctx context.Context, spec *ChartSpec, opts *GenericHelmOptions) (*release.Release, error) {
	return c.upgrade(ctx, spec, opts)
}

func (c *HelmClient) upgrade(ctx context.Context, spec *ChartSpec, opts *GenericHelmOptions) (*release.Release, error) {
	client := action.NewUpgrade(c.ActionConfig)
	mergeUpgradeOptions(spec, client)
	// if client.Install = true
	

	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}

	helmChart, chartPath, err := c.getChart(spec.ChartName, &client.ChartPathOptions)
	if err != nil {
		return nil, err
	}

	helmChart, err = updateDependencies(helmChart, &client.ChartPathOptions, chartPath, c, client.DependencyUpdate, spec)
	if err != nil {
		return nil, err
	}

	values, err := getValuesMap(spec)
	if err != nil {
		return nil, err
	}

	if err := c.lint(chartPath, values); err != nil {
		return nil, err
	}

	if client.Install == true {
		exists, err := c.getRelease(spec.ReleaseName)
		fmt.Printf(exists)
		if err == nil{
			//릴리즈가 있다. 즉, 업그레이드
			upgradedRelease, upgradeErr := client.RunWithContext(ctx, spec.ReleaseName, helmChart, values)
			if upgradeErr != nil {
				if upgradedRelease != nil && opts != nil && opts.RollBack != nil {
					return nil, opts.RollBack.Rollback(spec)
				}
				return nil, upgradeErr
			}
		
			return upgradedRelease, nil

		} else {
			//릴리즈가 없다. 즉, 신규 설치
			rel, err := client.install(ctx, spec)
			if err != nil{
				return rel, nil
			} else {
				return nil, err
			}
			}
		}
	}




func (c *HelmClient) lint(chartPath string, values map[string]interface{}) error {
	client := action.NewLint()
	result := client.Run([]string{chartPath}, values)
	if len(result.Errors) > 0 {
		return fmt.Errorf("linting for chartpath %q failed", chartPath)
	}

	return nil
}

func (c *HelmClient) UnInstall(releaseName string) (*release.UninstallReleaseResponse, error) {
	return c.unInstall(releaseName)
}

func (c *HelmClient) unInstall(releaseName string) (*release.UninstallReleaseResponse, error) {
	client := action.NewUninstall(c.ActionConfig)
	rel, err := client.Run(releaseName)
	if err != nil {
		return rel, err
	}

	return rel, nil
}

func (c *HelmClient) Rollback(spec *ChartSpec) error {
	return c.rollback(spec)
}

func (c *HelmClient) rollback(spec *ChartSpec) error {
	client := action.NewRollback(c.ActionConfig)

	mergeRollbackOptions(spec, client)

	return client.Run(spec.ReleaseName)
}

func CreateClient(namespace string, kubeConfig []byte) (Client, error) {
	settings := cli.New()
	// settings.RepositoryCache = defaultCachePath
	// settings.RepositoryConfig = defaultRepositoryConfigPath
	actionConfig := new(action.Configuration)
	clientGetter := NewRESTClientGetter(namespace, kubeConfig, nil)
	if err := actionConfig.Init(clientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err.Error())
		return nil, err
	}

	return &HelmClient{
		Settings:     settings,
		Providers:    getter.All(settings),
		storage:      &storage,
		ActionConfig: actionConfig,
	}, nil
}

func getValuesMap(chartSpec *ChartSpec) (map[string]interface{}, error) {
	var values map[string]interface{}
	err := yaml.Unmarshal([]byte(chartSpec.ValuesYaml), &values)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func updateDependencies(helmChart *chart.Chart, chartPathOptions *action.ChartPathOptions, chartPath string, c *HelmClient, dependencyUpdate bool, spec *ChartSpec) (*chart.Chart, error) {
	if req := helmChart.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(helmChart, req); err != nil {
			if dependencyUpdate {
				man := &downloader.Manager{
					ChartPath:        chartPath,
					Keyring:          chartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          c.Providers,
					RepositoryConfig: c.Settings.RepositoryConfig,
					RepositoryCache:  c.Settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}

				helmChart, _, err = c.getChart(spec.ChartName, chartPathOptions)
				if err != nil {
					return nil, err
				}

			} else {
				return nil, err
			}
		}
	}
	return helmChart, nil
}

func mergeInstallOptions(chartSpec *ChartSpec, installOptions *action.Install) {
	installOptions.CreateNamespace = chartSpec.CreateNamespace
	installOptions.DisableHooks = chartSpec.DisableHooks
	installOptions.Replace = chartSpec.Replace
	installOptions.Wait = chartSpec.Wait
	installOptions.DependencyUpdate = chartSpec.DependencyUpdate
	installOptions.Timeout = chartSpec.Timeout
	installOptions.Namespace = chartSpec.Namespace
	installOptions.ReleaseName = chartSpec.ReleaseName
	installOptions.Version = chartSpec.Version
	installOptions.GenerateName = chartSpec.GenerateName
	installOptions.NameTemplate = chartSpec.NameTemplate
	installOptions.Atomic = chartSpec.Atomic
	installOptions.SkipCRDs = chartSpec.SkipCRDs
	installOptions.DryRun = chartSpec.DryRun
	installOptions.SubNotes = chartSpec.SubNotes
}

func mergeUpgradeOptions(chartSpec *ChartSpec, upgradeOptions *action.Upgrade) {
	upgradeOptions.Version = chartSpec.Version
	upgradeOptions.Namespace = chartSpec.Namespace
	upgradeOptions.Timeout = chartSpec.Timeout
	upgradeOptions.Wait = chartSpec.Wait
	upgradeOptions.DependencyUpdate = chartSpec.DependencyUpdate
	upgradeOptions.DisableHooks = chartSpec.DisableHooks
	upgradeOptions.Force = chartSpec.Force
	upgradeOptions.ResetValues = chartSpec.ResetValues
	upgradeOptions.ReuseValues = chartSpec.ReuseValues
	upgradeOptions.Recreate = chartSpec.Recreate
	upgradeOptions.MaxHistory = chartSpec.MaxHistory
	upgradeOptions.Atomic = chartSpec.Atomic
	upgradeOptions.CleanupOnFail = chartSpec.CleanupOnFail
	upgradeOptions.DryRun = chartSpec.DryRun
	upgradeOptions.SubNotes = chartSpec.SubNotes
	upgradeOptions.Install = chartSpec.Install
}

func mergeRollbackOptions(chartSpec *ChartSpec, rollbackOptions *action.Rollback) {
	rollbackOptions.DisableHooks = chartSpec.DisableHooks
	rollbackOptions.DryRun = chartSpec.DryRun
	rollbackOptions.Timeout = chartSpec.Timeout
	rollbackOptions.CleanupOnFail = chartSpec.CleanupOnFail
	rollbackOptions.Force = chartSpec.Force
	rollbackOptions.MaxHistory = chartSpec.MaxHistory
	rollbackOptions.Recreate = chartSpec.Recreate
	rollbackOptions.Wait = chartSpec.Wait
}
