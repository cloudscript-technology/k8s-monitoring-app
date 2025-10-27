package project

import (
	"errors"
	"net/http"

	"k8s-monitoring-app/internal/core"
	serverModel "k8s-monitoring-app/internal/server/model"
	model "k8s-monitoring-app/pkg/project/model"

	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
)

type service struct{}

func NewService() model.Service {
	return &service{}
}

func (s *service) Get(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	id := sc.Param("id")

	if len(id) == 0 {
		log.Error(ctx, errors.New("id is empty")).Msg("error getting project")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	project, err := serverModel.ServerRepos.Project.Get(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error getting project")
		return sc.String(http.StatusNotFound, "project not found")
	}

	return sc.JSON(http.StatusOK, project)
}

func (s *service) List(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	projects, err := serverModel.ServerRepos.Project.List(ctx)
	if err != nil {
		log.Error(ctx, err).Msg("error listing projects")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusOK, projects)
}

func (s *service) Add(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	project := model.Project{}
	if err := sc.Bind(&project); err != nil {
		log.Error(ctx, err).Msg("error binding project")
		return sc.String(http.StatusBadRequest, "invalid request body")
	}
	if err := serverModel.ServerRepos.Project.Add(ctx, &project); err != nil {
		log.Error(ctx, err).Msg("error add project")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusCreated, project)
}

func (s *service) Update(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	// First get the existing project to check it exists
	_, err := serverModel.ServerRepos.Project.Get(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error getting project")
		return sc.String(http.StatusNotFound, "project not found")
	}

	project := model.Project{}
	if err := sc.Bind(&project); err != nil {
		log.Error(ctx, err).Msg("error binding project")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}
	project.ID = id

	if err := serverModel.ServerRepos.Project.Update(ctx, &project); err != nil {
		log.Error(ctx, err).Msg("error updating project")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return sc.JSON(http.StatusOK, project)
}

func (s *service) Delete(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	if len(id) == 0 {
		log.Error(ctx, errors.New("id is empty")).Msg("error deleting project")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}

	err := serverModel.ServerRepos.Project.Delete(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error deleting project")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return sc.JSON(http.StatusOK, map[string]bool{"success": true})
}
