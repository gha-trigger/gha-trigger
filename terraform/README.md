# Getting Started with Terraform

In the Getting Started, you can set up gha-dispatcher and experience CI with gha-dispatcher.

## Requirement

- Git
- Terraform
- GitHub Account
- AWS Account

## Procedure

### Create GitHub Repositories from template repositories

- Source Repository:
- CI Repository:

### Create Webhook Secret

- https://docs.github.com/en/developers/webhooks-and-events/webhooks/securing-your-webhooks

### Create GitHub App(s)

- https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app

You have to create GitHub App for two purposes.

1. Receive webhook and trigger GitHub Actions Workflow
2. Access Source Repository in CI of CI Repository

You can use one GitHub App for the above purposes or can create two GitHub Apps for each purpose.

You have to install GitHub App in Source Repository and CI Repository.
You can either use the same GitHub App or create GitHub Apps per repository.

#### 1. Receive webhook and trigger GitHub Actions Workflow

The minimum setting of GitHub App (1).

- Webhook: Active
- Permissions
  - Actions: Read and write
  - Issues: Read-only
  - Pull requests: Read and write

#### 2. Receive webhook and trigger GitHub Actions Workflow

The minimum setting of GitHub App (2).

- Webhook: Inactive
- Permissions
  - Commit statuses: Read and write
  - Contents: Read

### Set up Terraform Configuration

```console
$ git clone https://github.com/suzuki-shunsuke/gha-dispatcher
$ cd gha-dispatcher/terraform
```

[Download a zip file from Release page](https://github.com/suzuki-shunsuke/gha-dispatcher/releases) on this directory.

Create `config.yaml`, `secret.yaml`, and `terraform.tfvars` from templates.

```console
$ cp config.yaml.tmpl config.yaml
$ vi config.yaml

$ cp secret.yaml.tmpl secret.yaml
$ vi secret.yaml

$ cp terraform.tfvars.tmpl terraform.tfvars
$ vi terraform.tfvars
```

### Apply Terraform

Create resources.

```console
$ terraform apply [-refresh=false]
```

`-refresh=false` is useful to make terraform commands fast.

### Try

Create a pull request to source repository to test CI.

## Clean up

```
$ terraform destroy
```
