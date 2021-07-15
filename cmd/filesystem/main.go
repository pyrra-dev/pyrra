package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/oklog/run"
	"github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/openapi"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
	"sigs.k8s.io/yaml"
)

var objectives = map[string]v1alpha1.ServiceLevelObjective{}

func main() {
	var gr run.Group
	{
		ctx, cancel := context.WithCancel(context.Background())
		gr.Add(func() error {
			files, err := filepath.Glob("/etc/pyrra/*.yaml")
			if err != nil {
				return err
			}
			for _, f := range files {
				bytes, err := ioutil.ReadFile(f)
				if err != nil {
					return err
				}

				var config v1alpha1.ServiceLevelObjective
				if err := yaml.UnmarshalStrict(bytes, &config); err != nil {
					return err
				}
				objectives[config.GetName()] = config
			}
			<-ctx.Done()
			return nil
		}, func(err error) {
			cancel()
		})
	}
	{
		router := openapiserver.NewRouter(
			openapiserver.NewObjectivesApiController(&FilesystemObjectiveServer{}),
		)
		server := http.Server{Addr: ":9444", Handler: router}

		gr.Add(func() error {
			log.Println("Starting up HTTP API on", server.Addr)
			return server.ListenAndServe()
		}, func(err error) {
			_ = server.Shutdown(context.Background())
		})
	}

	if err := gr.Run(); err != nil {
		log.Println(err)
		return
	}
}

type FilesystemObjectiveServer struct{}

func (f FilesystemObjectiveServer) ListObjectives(ctx context.Context) (openapiserver.ImplResponse, error) {
	list := make([]openapiserver.Objective, 0, len(objectives))
	for _, objective := range objectives {
		internal, err := objective.Internal()
		if err != nil {
			return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
		}
		list = append(list, openapi.ServerFromInternal(internal))
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: list,
	}, nil
}

func (f FilesystemObjectiveServer) GetObjective(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	slo, ok := objectives[name]
	if !ok {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}

	internal, err := slo.Internal()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapi.ServerFromInternal(internal),
	}, nil
}

func (f FilesystemObjectiveServer) GetMultiBurnrateAlerts(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetObjectiveErrorBudget(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetObjectiveStatus(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetREDRequests(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetREDErrors(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}
