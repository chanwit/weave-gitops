package app

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// type DeploymentType string
// type SourceType string

// const (
//  DeployTypeKustomize DeploymentType = "kustomize"
//  DeployTypeHelm      DeploymentType = "helm"

//  SourceTypeGit  SourceType = "git"
//  SourceTypeHelm SourceType = "helm"
// )

// AppService entity that manages applications
type AppService interface {
	// Add adds a new application to the cluster
	Add(params AddParams) error
	// Get returns a given applicaiton
	Get(name types.NamespacedName) (*wego.Application, error)
	// GetCommits returns a list of commits for an application
	GetCommits(params CommitParams) ([]gitprovider.Commit, error)
	// Remove removes an application from the cluster
	Remove(params RemoveParams) error
	// Status returns flux resources status and the last successful reconciliation time
	Status(params StatusParams) (string, string, error)
	// Pause pauses the gitops automation for an app
	Pause(params PauseParams) error
	// Unpause resumes the gitops automation for an app
	Unpause(params UnpauseParams) error
}

type App struct {
	Osys               osys.Osys
	Git                git.Git
	Flux               flux.Flux
	Kube               kube.Kube
	Logger             logger.Logger
	GitProviderFactory func(token string) (gitproviders.GitProvider, error)
	// TODO: @jpellizzari adding this as a temporary stop-gap to maintain the current behavior for external config repos.
	// As of https://github.com/weaveworks/weave-gitops/pull/587,
	// we are not addressing this case yet. Many of the unit tests check for exact function call
	// arguments, which will get skipped when the AuthService is used (and therefore tests will fail).
	// Follow up issue where this will be addressed: https://github.com/weaveworks/weave-gitops/issues/592
	temporaryGitClientFactory func(osysClient osys.Osys, privKeypath string) (git.Git, error)
}

func New(logger logger.Logger, git git.Git, flux flux.Flux, kube kube.Kube, osys osys.Osys) *App {
	return &App{
		Git:                       git,
		Flux:                      flux,
		Kube:                      kube,
		Logger:                    logger,
		Osys:                      osys,
		GitProviderFactory:        createGitProvider,
		temporaryGitClientFactory: temporaryCreateGitClient,
	}
}

// Make sure App implements all the required methods.
var _ AppService = &App{}

func createGitProvider(token string) (gitproviders.GitProvider, error) {
	provider, err := gitproviders.New(gitproviders.Config{
		Provider: gitproviders.GitProviderGitHub,
		Token:    token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed initializing git provider: %w", err)
	}

	return provider, nil
}

func temporaryCreateGitClient(osysClient osys.Osys, privKeypath string) (git.Git, error) {
	auth, err := osysClient.SelectAuthMethod(privKeypath)
	if err != nil {
		return nil, fmt.Errorf("error selecting auth method for external config repo: %w", err)
	}
	return git.New(auth, wrapper.NewGoGit()), nil
}

func (a *App) getDeploymentType(ctx context.Context, name string, namespace string) (wego.DeploymentType, error) {
	app, err := a.Kube.GetApplication(ctx, types.NamespacedName{Name: name, Namespace: namespace})
	if err != nil {
		return wego.DeploymentTypeKustomize, err
	}

	return wego.DeploymentType(app.Spec.DeploymentType), nil
}

func (a *App) getSuspendedStatus(ctx context.Context, name, namespace string, deploymentType wego.DeploymentType) (bool, error) {
	var automation client.Object

	switch deploymentType {
	case wego.DeploymentTypeKustomize:
		automation = &kustomizev1.Kustomization{}
	case wego.DeploymentTypeHelm:
		automation = &helmv2.HelmRelease{}
	default:
		return false, fmt.Errorf("invalid deployment type: %v", deploymentType)
	}

	if err := a.Kube.GetResource(ctx, types.NamespacedName{Namespace: namespace, Name: name}, automation); err != nil {
		return false, err
	}

	suspendStatus := false

	switch at := automation.(type) {
	case *kustomizev1.Kustomization:
		suspendStatus = at.Spec.Suspend
	case *helmv2.HelmRelease:
		suspendStatus = at.Spec.Suspend
	}
	return suspendStatus, nil
}

func (a *App) pauseOrUnpause(suspendAction wego.SuspendActionType, name, namespace string) error {
	ctx := context.Background()
	deploymentType, err := a.getDeploymentType(ctx, name, namespace)
	if err != nil {
		return fmt.Errorf("unable to determine deployment type for %s: %s", name, err)
	}

	suspendStatus, err := a.getSuspendedStatus(ctx, name, namespace, deploymentType)
	if err != nil {
		return fmt.Errorf("failed to get suspended status: %s", err)
	}

	switch deploymentType {
	case wego.DeploymentTypeKustomize:
		deploymentType = "kustomization"
	case wego.DeploymentTypeHelm:
		deploymentType = "helmrelease"
	default:
		return fmt.Errorf("invalid deployment type: %v", deploymentType)
	}

	switch suspendAction {
	case wego.SuspendAction:
		if suspendStatus {
			a.Logger.Printf("app %s is already paused\n", name)
			return nil
		}
		out, err := a.Flux.SuspendOrResumeApp(suspendAction, name, namespace, string(deploymentType))
		if err != nil {
			return fmt.Errorf("unable to pause %s err: %s", name, err)
		}
		a.Logger.Printf("%s\n gitops automation paused for %s\n", string(out), name)
		return nil
	case wego.ResumeAction:
		if !suspendStatus {
			a.Logger.Printf("app %s is already reconciling\n", name)
			return nil
		}
		out, err := a.Flux.SuspendOrResumeApp(suspendAction, name, namespace, string(deploymentType))
		if err != nil {
			return fmt.Errorf("unable to unpause %s err: %s", name, err)
		}
		a.Logger.Printf("%s\n gitops automation unpaused for %s\n", string(out), name)
		return nil
	}
	return fmt.Errorf("invalid suspend action")
}

func IsClusterReady(l logger.Logger, k kube.Kube) error {
	l.Waitingf("Checking cluster status")
	clusterStatus := k.GetClusterStatus(context.Background())

	switch clusterStatus {
	case kube.Unmodified:
		return fmt.Errorf("Wego not installed... exiting")
	case kube.Unknown:
		return fmt.Errorf("Wego can not determine cluster status... exiting")
	}
	l.Successf(clusterStatus.String())

	return nil
}