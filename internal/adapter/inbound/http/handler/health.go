package handler

import (
	"net/http"
	"time"
)

type healthResponse struct {
	BuildID   string `json:"buildId"`
	BuildTime string `json:"buildTime"`
	Uptime    string `json:"uptime"`
}

type HealthHandler struct {
	buildID   string
	buildTime string
	startedAt time.Time
}

func NewHealthHandler(buildID string, buildTime string) *HealthHandler {
	return &HealthHandler{
		buildID:   buildID,
		buildTime: buildTime,
		startedAt: time.Now().UTC(),
	}
}

func (handler *HealthHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writeSuccess(writer, healthResponse{
		BuildID:   handler.buildID,
		BuildTime: handler.buildTime,
		Uptime:    time.Since(handler.startedAt).Round(time.Second).String(),
	})
}
