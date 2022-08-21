package cloudfunction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/domain"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/github"
)

func init() { //nolint:gochecknoinits
	ctx := context.Background()
	if err := initWithError(ctx); err != nil {
		panic(err)
	}
}

type Handler struct {
	Actions domain.ActionsService
}

var handler *Handler //nolint:gochecknoglobals

func initWithError(ctx context.Context) error {
	// read config
	// read env
	// read secret
	// initialize handler
	handler = &Handler{}
	secretClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Println(err)
	}
	defer secretClient.Close()
	return nil
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func Main(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	headerKeys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		headerKeys = append(headerKeys, k)
	}
	log.Println("header: ", strings.Join(headerKeys, " "))
	if _, err := io.Copy(w, r.Body); err != nil {
		log.Println(err)
	}
	// parse request
	// validate request
	// route and filter request
	// list labels and changed files
	// Run GitHub Actions Workflow

	if resp, err := handler.Actions.CreateWorkflowDispatchEventByFileName(ctx, "", "", "", github.CreateWorkflowDispatchEventRequest{
		Ref:    "main",
		Inputs: map[string]interface{}{},
	}); err != nil {
		m := map[string]interface{}{
			"error":       fmt.Sprintf("create a workflow dispatch event by file name: %s", err.Error()),
			"status_code": resp.StatusCode,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(m); err != nil {
			log.Println(err)
		}
	}
}
