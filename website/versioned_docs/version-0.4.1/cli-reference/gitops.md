## gitops

Weave GitOps

### Synopsis

Command line utility for managing Kubernetes applications via GitOps.

### Examples

```

  # Get verbose output for any gitops command
  gitops [command] -v, --verbose

  # Get gitops app help
  gitops help app

  # Add application to gitops control from a local git repository
  gitops add app . --name <myapp>
  OR
  gitops add app <myapp-directory>

  # Add application to gitops control from a github repository
  gitops add app \
    --name <myapp> \
    --url git@github.com:myorg/<myapp> \
    --branch prod-<myapp>

  # Get status of application under gitops control
  gitops get app podinfo

  # Get help for gitops add app command
  gitops add app -h
  gitops help add app

  # Show manifests that would be installed by the gitops install command
  gitops install --dry-run

  # Install gitops in the wego-system namespace
  gitops install

  # Get the version of gitops along with commit, branch, and flux version
  gitops version

  To learn more, you can find our documentation at https://docs.gitops.weave.works/

```

### Options

```
  -e, --endpoint string    The Weave GitOps Enterprise HTTP API endpoint
  -h, --help               help for gitops
      --namespace string   Weave GitOps runtime namespace (default "wego-system")
  -v, --verbose            Enable verbose output
```

### SEE ALSO

* [gitops add](gitops_add.md)	 - Add a new Weave GitOps resource
* [gitops beta](gitops_beta.md)	 - Experimental commands
* [gitops delete](gitops_delete.md)	 - Delete one or many Weave GitOps resources
* [gitops flux](gitops_flux.md)	 - Use flux commands
* [gitops get](gitops_get.md)	 - Display one or many Weave GitOps resources
* [gitops install](gitops_install.md)	 - Install or upgrade GitOps
* [gitops resume](gitops_resume.md)	 - Resume your GitOps automations
* [gitops suspend](gitops_suspend.md)	 - Suspend your GitOps automations
* [gitops ui](gitops_ui.md)	 - Manages Gitops UI
* [gitops uninstall](gitops_uninstall.md)	 - Uninstall GitOps
* [gitops upgrade](gitops_upgrade.md)	 - Upgrade to Weave GitOps Enterprise
* [gitops version](gitops_version.md)	 - Display gitops version

###### Auto generated by spf13/cobra on 3-Nov-2021