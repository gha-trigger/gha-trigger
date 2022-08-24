# gha-trigger

GitHub App for Secure GitHub Actions

## Status

Work in progress. This isn't production ready yet.

## Goal

Run GitHub Actions Workflow securely.
Prevent GitHub Actions Workflow from being modifying and running malicious commands.

## Background

GitHub Actions is very powerful CI Platform, but also has a security risk that someone modifies workflow and CI scripts and run malicious commands.
For example, secrets with strong permission may be abused and stolen.

You can use other CI Platform to prevent workflows from being modifying, but we would like to use GitHub Actions because GitHub Actions is very powerful.

So we design the architecture and develop GitHub App to achieve the above goal.

## Architecture

![gha-trigger architecture](https://user-images.githubusercontent.com/13323303/186283702-cb3d7de1-6bb0-45dc-8387-d251068484a1.png)

You create two GitHub repositories.

- Main Repository
  - Users develop this repository
  - Disable GitHub Actions
- CI Repository
  - Manage GitHub Actions Workflows and CI scripts
  - Only CI maintainers have write permissiono and other developers have only read permission

You install GitHub App `gha-trigger` in these repositories.
When events such as `push` and `pull_request` occur in Main Repository, the webhook is sent to `gha-trigger`.
`gha-trigger` validates and filters webhooks and triggers GitHub Actions Workflows of CI Repository via GitHub API.
Workflows of CI Repository update commit statuses of Main Repository and send pull request comments so that developers can refer CI results from Main Repository's pull request pages.

The important thing is that workflows and CI scripts are managed at the repository other than `Main Repository` and only restricted people have the write permission of `CI Repository`.
This prevents developers from modifying workflows and CI scripts and makes GitHub Actions secure.

## Supported runtime

gha-trigger supports only AWS Lambda at the moment,
but we're considering to support other platform such as Google Cloud Function too.

## Stateless

`gha-trigger` doesn't use any Databases such as RDB at the moment, which makes the management of `gha-trigger` easy.

## How to rerun and cancel CI

Developers don't have the write permission of CI Repository, so they can't rerun and cancel workflows directly.
But they can rerun and cancel workflows via pull request comments.

- Rerun workflows: `/rerun-workflow <workflow id> [<workflow id> ...]`
- Rerun failed jobs: `/rerun-failed-jobs <workflow id> [<workflow id> ...]`
- Rerun jobs: `/rerun-job <job id> [<job id> ...]`
- Cancel workflows: `/cancel <workflow id> [<workflow id> ...]`

## How to trigger workflows manually

If you would like to add workflows that developers run manually, you have to create a repository for those workflows.
Let's call the repository `Manual Trigger Repository`.
You have to install GitHub App in `Main Repository` and `Manual Trigger Repository` so that workflows can access `Main Repository`.
You also have to give developers the write permission of `Manual Trigger Repository`, so you have to be careful the treat of this repository.

One of the usecase of this repository we assume is that developers scaffold pull requests of Main Repository.

For example, [tfaction](https://github.com/suzuki-shunsuke/tfaction) provides the feature.

[Scaffold working directory by GitHub Actions workflow_dispatch event](https://suzuki-shunsuke.github.io/tfaction/docs/feature/scaffold-working-dir)

In that case, you can give GitHub App only permission to push commits to `Main Repository`.
If GitHub App can create pull requests to `Main Repository`, a developer can approve and merge them himself. This is risky so workflows should create only feature branches in `Main Repository` and let developers open pull requests themselves.

## Pros and Cons

### Pros

The pros of `gha-trigger` is that you can run GitHub Actions securely.
You can prevent GitHub Actions Workflow from being modifying and running malicious commands.

### Cons

Compared with normal GitHub Actions usage, `gha-trigger` has some drawbacks.

- `github.token` of `CI Repository` can't be used to access `Main Repository`
- You have to fix workfows to migrate existing workflows to `gha-trigger`
- `gha-trigger` uses not Checks API but Commit Status API
- `gha-trigger` calls GitHub API so it has a risk of GitHub API rate limit issue
- The experience for rerunning and canceling CI isn't good
- It spends money
- You have to set up and maintain `gha-trigger`
  - Continous update
  - Monitoring
  - Trouble shooting and user support

## Getting Started

Coming soon.

## How to set up

Coming soon.

## Configuration

`gha-trigger` supports only environment variables as source of configuration,
but we are considering other sources such as S3, DynamoDB, AWS AppConfig, and so on.

e.g.

```yaml
---
aws:
  region: us-east-1
  secretsmanager:
    region: us-east-1
    secret_id: test-gha-trigger
github_app:
  app_id: 123456789
events:
  - matches:
      - repo_owner: suzuki-shunsuke
        repo_name: example-terraform-monorepo-2
        events:
          - pull_request
        branches:
          - main
    workflows:
      - repo_owner: suzuki-shunsuke
        repo_name: example-terraform-monorepo-2-ci
        workflow_file_name: test_pull_request.yaml
        ref: pull_request
  - matches:
      - repo_owner: suzuki-shunsuke
        repo_name: example-terraform-monorepo-2
        events:
          - push
        branches:
          - main
    workflows:
      - repo_owner: suzuki-shunsuke
        repo_name: example-terraform-monorepo-2-ci
        workflow_file_name: test.yaml
        ref: main
```

### Secrets

`gha-trigger` requires the following secrets.

- webhook_secret: GitHub App's Webhook Secret
- github_app_private_key: GitHub App's private key

`gha-trigger` supports only AWS SecretsManager at the moment.

## LICENSE

[MIT](LICENSE)
