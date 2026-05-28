package publicpage

import (
	"net/http"

	apppublicpage "github.com/yorukot/netstamp/internal/controller/application/publicpage"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) listPages(r *http.Request) (*pageListOutput, error) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return nil, err
	}
	pages, err := h.service.ListPages(r.Context(), apppublicpage.ListPagesInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
	})
	if err != nil {
		return nil, mapPublicPageError(err, "list public pages failed")
	}

	return &pageListOutput{Body: pageListOutputBody{PublicPages: newPageBodies(pages)}}, nil
}

func (h *Handler) createPage(r *http.Request, body createPageInputBody) (*pageOutput, error) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return nil, err
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	page, err := h.service.CreatePage(r.Context(), apppublicpage.CreatePageInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
		Slug:          body.Slug,
		Title:         body.Title,
		Description:   body.Description,
		Enabled:       enabled,
	})
	if err != nil {
		return nil, mapPublicPageError(err, "create public page failed")
	}

	return &pageOutput{Body: pageOutputBody{PublicPage: newPageBody(page)}}, nil
}

func (h *Handler) getPage(r *http.Request) (*pageOutput, error) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return nil, err
	}
	page, err := h.service.GetPage(r.Context(), apppublicpage.GetPageInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
		PageID:        httpx.Path(r, "page_id"),
	})
	if err != nil {
		return nil, mapPublicPageError(err, "get public page failed")
	}

	return &pageOutput{Body: pageOutputBody{PublicPage: newPageBody(page)}}, nil
}

func (h *Handler) updatePage(r *http.Request, body updatePageInputBody) (*pageOutput, error) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return nil, err
	}
	page, err := h.service.UpdatePage(r.Context(), apppublicpage.UpdatePageInput{
		CurrentUserID:  userID,
		ProjectRef:     httpx.Path(r, "ref"),
		PageID:         httpx.Path(r, "page_id"),
		Slug:           body.Slug,
		Title:          body.Title,
		Description:    body.Description.Value,
		DescriptionSet: body.Description.Set,
		Enabled:        body.Enabled,
	})
	if err != nil {
		return nil, mapPublicPageError(err, "update public page failed")
	}

	return &pageOutput{Body: pageOutputBody{PublicPage: newPageBody(page)}}, nil
}

func (h *Handler) deletePage(r *http.Request) error {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return err
	}
	err = h.service.DeletePage(r.Context(), apppublicpage.DeletePageInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
		PageID:        httpx.Path(r, "page_id"),
	})
	if err != nil {
		return mapPublicPageError(err, "delete public page failed")
	}
	return nil
}

func (h *Handler) createFolder(r *http.Request, body createFolderInputBody) (*folderOutput, error) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return nil, err
	}
	sortOrder := int32(0)
	if body.SortOrder != nil {
		sortOrder = *body.SortOrder
	}
	folder, err := h.service.CreateFolder(r.Context(), apppublicpage.CreateFolderInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
		PageID:        httpx.Path(r, "page_id"),
		ParentID:      body.ParentID,
		Name:          body.Name,
		Description:   body.Description,
		SortOrder:     sortOrder,
	})
	if err != nil {
		return nil, mapPublicPageError(err, "create public page folder failed")
	}

	return &folderOutput{Body: folderOutputBody{Folder: newFolderBody(folder)}}, nil
}

func (h *Handler) updateFolder(r *http.Request, body updateFolderInputBody) (*folderOutput, error) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return nil, err
	}
	folder, err := h.service.UpdateFolder(r.Context(), apppublicpage.UpdateFolderInput{
		CurrentUserID:  userID,
		ProjectRef:     httpx.Path(r, "ref"),
		PageID:         httpx.Path(r, "page_id"),
		FolderID:       httpx.Path(r, "folder_id"),
		ParentID:       body.ParentID.Value,
		ParentIDSet:    body.ParentID.Set,
		Name:           body.Name,
		Description:    body.Description.Value,
		DescriptionSet: body.Description.Set,
		SortOrder:      body.SortOrder,
	})
	if err != nil {
		return nil, mapPublicPageError(err, "update public page folder failed")
	}

	return &folderOutput{Body: folderOutputBody{Folder: newFolderBody(folder)}}, nil
}

func (h *Handler) deleteFolder(r *http.Request) error {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return err
	}
	err = h.service.DeleteFolder(r.Context(), apppublicpage.DeleteFolderInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
		PageID:        httpx.Path(r, "page_id"),
		FolderID:      httpx.Path(r, "folder_id"),
	})
	if err != nil {
		return mapPublicPageError(err, "delete public page folder failed")
	}
	return nil
}

func (h *Handler) setFolderChecks(r *http.Request, body setFolderChecksInputBody) (*checksOutput, error) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		return nil, err
	}
	checks, err := h.service.SetFolderChecks(r.Context(), apppublicpage.SetFolderChecksInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
		PageID:        httpx.Path(r, "page_id"),
		FolderID:      httpx.Path(r, "folder_id"),
		CheckIDs:      body.CheckIDs,
	})
	if err != nil {
		return nil, mapPublicPageError(err, "set public page folder checks failed")
	}

	return &checksOutput{Body: checksOutputBody{Checks: newPublishedCheckBodies(checks)}}, nil
}

func (h *Handler) getPublicPage(r *http.Request) (*pageOutput, error) {
	page, err := h.service.GetPublicPage(r.Context(), apppublicpage.GetPublicPageInput{
		Slug: httpx.Path(r, "slug"),
	})
	if err != nil {
		return nil, mapPublicPageError(err, "get public page failed")
	}

	return &pageOutput{Body: pageOutputBody{PublicPage: newPageBody(page)}}, nil
}

func (h *Handler) queryPublicPingInsight(r *http.Request) (*pingInsightOutput, error) {
	from, err := httpx.QueryInt64(r, "from")
	if err != nil {
		return nil, err
	}
	to, err := httpx.QueryInt64(r, "to")
	if err != nil {
		return nil, err
	}
	maxDataPoints, err := httpx.QueryInt32(r, "maxDataPoints")
	if err != nil {
		return nil, err
	}
	output, err := h.service.QueryPublicPingInsight(r.Context(), apppublicpage.QueryPublicPingInsightInput{
		Slug:          httpx.Path(r, "slug"),
		ProbeID:       httpx.QueryString(r, "probeId"),
		CheckID:       httpx.QueryString(r, "checkId"),
		FromMs:        optionalInt64(from),
		ToMs:          optionalInt64(to),
		MaxDataPoints: optionalInt32(maxDataPoints),
	})
	if err != nil {
		return nil, mapPublicPageError(err, "query public ping insight failed")
	}

	return &pingInsightOutput{Body: newPingInsightBody(output)}, nil
}

type createPageInputBody struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

type updatePageInputBody struct {
	Slug        *string        `json:"slug,omitempty"`
	Title       *string        `json:"title,omitempty"`
	Description nullableString `json:"description,omitempty"`
	Enabled     *bool          `json:"enabled,omitempty"`
}

type createFolderInputBody struct {
	ParentID    *string `json:"parentId,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int32  `json:"sortOrder,omitempty"`
}

type updateFolderInputBody struct {
	ParentID    nullableString `json:"parentId,omitempty"`
	Name        *string        `json:"name,omitempty"`
	Description nullableString `json:"description,omitempty"`
	SortOrder   *int32         `json:"sortOrder,omitempty"`
}

type setFolderChecksInputBody struct {
	CheckIDs []string `json:"checkIds"`
}
