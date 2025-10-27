package application

import (
	"errors"
	"net/http"

	"k8s-monitoring-app/internal/core"
	serverModel "k8s-monitoring-app/internal/server/model"
	model "k8s-monitoring-app/pkg/application/model"

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
		log.Error(ctx, errors.New("id is empty")).Msg("error getting application")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	application, err := serverModel.ServerRepos.Application.Get(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error getting application")
		return sc.String(http.StatusNotFound, "application not found")
	}

	return sc.JSON(http.StatusOK, application)
}

func (s *service) List(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applications, err := serverModel.ServerRepos.Application.List(ctx)
	if err != nil {
		log.Error(ctx, err).Msg("error listing applications")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusOK, applications)
}

func (s *service) ListByProject(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	projectID := sc.Param("project_id")

	if len(projectID) == 0 {
		log.Error(ctx, errors.New("project_id is empty")).Msg("error listing applications by project")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	applications, err := serverModel.ServerRepos.Application.ListByProject(ctx, projectID)
	if err != nil {
		log.Error(ctx, err).Msg("error listing applications by project")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusOK, applications)
}

func (s *service) Add(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	application := model.Application{}
	if err := sc.Bind(&application); err != nil {
		log.Error(ctx, err).Msg("error binding application")
		return sc.String(http.StatusBadRequest, "invalid request body")
	}

	// Validate that the project exists
	_, err := serverModel.ServerRepos.Project.Get(ctx, application.ProjectID)
	if err != nil {
		log.Error(ctx, err).Msg("error getting project")
		return sc.String(http.StatusBadRequest, "project not found")
	}

	if err := serverModel.ServerRepos.Application.Add(ctx, &application); err != nil {
		log.Error(ctx, err).Msg("error add application")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusCreated, application)
}

func (s *service) Update(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	// First get the existing application to check it exists
	_, err := serverModel.ServerRepos.Application.Get(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error getting application")
		return sc.String(http.StatusNotFound, "application not found")
	}

	application := model.Application{}
	if err := sc.Bind(&application); err != nil {
		log.Error(ctx, err).Msg("error binding application")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}
	application.ID = id

	// Validate that the project exists if it's being changed
	if application.ProjectID != "" {
		_, err := serverModel.ServerRepos.Project.Get(ctx, application.ProjectID)
		if err != nil {
			log.Error(ctx, err).Msg("error getting project")
			return sc.String(http.StatusBadRequest, "project not found")
		}
	}

	if err := serverModel.ServerRepos.Application.Update(ctx, &application); err != nil {
		log.Error(ctx, err).Msg("error updating application")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return sc.JSON(http.StatusOK, application)
}

func (s *service) Delete(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	if len(id) == 0 {
		log.Error(ctx, errors.New("id is empty")).Msg("error deleting application")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}

	err := serverModel.ServerRepos.Application.Delete(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error deleting application")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return sc.JSON(http.StatusOK, map[string]bool{"success": true})
}
