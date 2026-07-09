package admin

import (
	"encoding/json"
	"net/http"
	"time"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

type dataExportResponseBody struct {
	Format     string                               `json:"format"`
	ExportedAt time.Time                            `json:"exportedAt"`
	Tables     map[string][]domainsystem.RawDataRow `json:"tables"`
}

type dataImportResponseBody struct {
	Result dataImportResultBody `json:"result"`
}

type dataImportResultBody struct {
	Format         string `json:"format"`
	ImportedTables int    `json:"importedTables"`
	ImportedRows   int    `json:"importedRows"`
}

func (h *Handler) handleExportData(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	export, err := h.service.ExportData(r.Context(), appadmin.ExportDataInput{CurrentUserID: userID})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "export admin data failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, dataExportResponse(export))
}

func (h *Handler) handleImportData(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body dataExportResponseBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}

	result, err := h.service.ImportData(r.Context(), appadmin.ImportDataInput{
		CurrentUserID: userID,
		Export: appadmin.DataExport{
			Format:     body.Format,
			ExportedAt: body.ExportedAt,
			Tables:     cloneRawTables(body.Tables),
		},
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "import admin data failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, dataImportResponse(result))
}

func dataExportResponse(export appadmin.DataExport) dataExportResponseBody {
	return dataExportResponseBody{
		Format:     export.Format,
		ExportedAt: export.ExportedAt,
		Tables:     cloneRawTables(export.Tables),
	}
}

func dataImportResponse(result appadmin.DataImportResult) dataImportResponseBody {
	return dataImportResponseBody{
		Result: dataImportResultBody{
			Format:         result.Format,
			ImportedTables: result.ImportedTables,
			ImportedRows:   result.ImportedRows,
		},
	}
}

func cloneRawTables(input map[string][]domainsystem.RawDataRow) map[string][]domainsystem.RawDataRow {
	if input == nil {
		return nil
	}
	output := make(map[string][]domainsystem.RawDataRow, len(input))
	for table, rows := range input {
		output[table] = make([]domainsystem.RawDataRow, 0, len(rows))
		for _, row := range rows {
			output[table] = append(output[table], domainsystem.RawDataRow(append(json.RawMessage(nil), row...)))
		}
	}
	return output
}
