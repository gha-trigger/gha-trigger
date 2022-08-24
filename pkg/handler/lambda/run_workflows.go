package lambda

import (
	"context"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/config"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) runWorkflows(ctx context.Context, logger *zap.Logger, body interface{}, workflows []*config.WorkflowConfig) (*Response, error) {
	if len(workflows) == 0 {
		return nil, nil //nolint:nilnil
	}

	numWorkflows := len(workflows)
	for i := 0; i < numWorkflows; i++ {
		workflow := workflows[i]
		// Run GitHub Actions Workflow
		inputs := make(map[string]interface{}, len(workflow.Inputs))
		for k, v := range workflow.Inputs {
			inputs[k] = v
			inputs["payload"] = body
		}
		_, err := handler.gh.RunWorkflow(ctx, workflow.RepoOwner, workflow.RepoName, workflow.WorkflowFileName, github.CreateWorkflowDispatchEventRequest{
			Ref:    workflow.Ref,
			Inputs: inputs,
		})
		if err != nil {
			logger.Error(
				"create a workflow dispatch event by file name",
				zap.Error(err),
				zap.String("workflow_repo_owner", workflow.RepoOwner),
				zap.String("workflow_repo_name", workflow.RepoName),
				zap.String("workflow_file_name", workflow.WorkflowFileName))
			continue
		}
	}
	return nil, nil //nolint:nilnil
}
