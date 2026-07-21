package handler

import (
	"net/http"
	"strconv"

	logOption "github.com/digitalrealmforgestudios/d-logger/option"

	"idas-video/internal/adapter/inbound/http/middleware"
	"idas-video/internal/adapter/inbound/http/observability"
	"idas-video/internal/entity"
	"idas-video/internal/usecase/inbound"
)

type videoResponse struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Category    entity.Tier `json:"category"`
	VideoURL    string      `json:"videoUrl"`
	CreatedAt   int64       `json:"createdAt"`
	UpdatedAt   int64       `json:"updatedAt"`
}

type VideoHandler struct {
	usecase inbound.VideoAccessUsecase
}

func NewVideoHandler(usecase inbound.VideoAccessUsecase) *VideoHandler {
	return &VideoHandler{usecase: usecase}
}

func (handler *VideoHandler) ListVideos(writer http.ResponseWriter, request *http.Request) {
	log := observability.Child("http.handler.video")
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		writeError(writer, ErrHTTPUnauthorized)
		return
	}

	query := parseVideoListQuery(request)
	log.Info("list videos request received", observability.WithContext(request.Context()), logOption.Attribute("user.id", userID.String()), logOption.Attribute("page", query.Page), logOption.Attribute("offset", query.Offset), logOption.Attribute("sort", query.SortBy))
	result, err := handler.usecase.ListAccessibleVideos(request.Context(), userID, query)
	if err != nil {
		log.Warn("list videos failed", observability.WithContext(request.Context()), logOption.Error(err), logOption.Attribute("user.id", userID.String()))
		writeError(writer, err)
		return
	}

	responses := newVideoResponses(result.Items)
	log.Info("list videos succeeded", observability.WithContext(request.Context()), logOption.Attribute("user.id", userID.String()), logOption.Attribute("video.count", len(responses)), logOption.Attribute("video.total", result.Total))
	writeListSuccess(writer, responses, result.Total, result.Page, result.Offset, result.SortBy)
}

func (handler *VideoHandler) GetVideo(writer http.ResponseWriter, request *http.Request) {
	log := observability.Child("http.handler.video")
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		writeError(writer, ErrHTTPUnauthorized)
		return
	}

	videoID := request.PathValue("id")
	if !entity.IsUUID(videoID) {
		writeError(writer, entity.ErrInvalidRequest)
		return
	}

	log.Info("get video request received", observability.WithContext(request.Context()), logOption.Attribute("user.id", userID.String()), logOption.Attribute("video.id", videoID))
	video, err := handler.usecase.GetAccessibleVideoByID(request.Context(), userID, entity.UUID(videoID))
	if err != nil {
		log.Warn("get video failed", observability.WithContext(request.Context()), logOption.Error(err), logOption.Attribute("user.id", userID.String()), logOption.Attribute("video.id", videoID))
		writeError(writer, err)
		return
	}

	log.Info("get video succeeded", observability.WithContext(request.Context()), logOption.Attribute("user.id", userID.String()), logOption.Attribute("video.id", videoID))
	writeSuccess(writer, newVideoResponse(*video))
}

func newVideoResponses(videos []entity.Video) []videoResponse {
	responses := make([]videoResponse, 0, len(videos))
	for _, video := range videos {
		responses = append(responses, newVideoResponse(video))
	}

	return responses
}

func parseVideoListQuery(request *http.Request) inbound.VideoListQuery {
	values := request.URL.Query()
	return inbound.VideoListQuery{
		Page:   parsePositiveInt(values.Get("page"), defaultListPage),
		Offset: parseNonNegativeInt(values.Get("offset"), defaultListOffset),
		SortBy: parseVideoSort(values.Get("sort")),
	}
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseNonNegativeInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func parseVideoSort(value string) string {
	switch value {
	case "createdAtAsc", "createdAtDesc", "titleAsc", "titleDesc":
		return value
	default:
		return defaultListSortBy
	}
}

func newVideoResponse(video entity.Video) videoResponse {
	return videoResponse{
		ID:          video.ID.String(),
		Title:       video.Title,
		Description: video.Description,
		Category:    video.Category,
		VideoURL:    video.VideoURL,
		CreatedAt:   video.CreatedAt.UTC().Unix(),
		UpdatedAt:   video.UpdatedAt.UTC().Unix(),
	}
}
