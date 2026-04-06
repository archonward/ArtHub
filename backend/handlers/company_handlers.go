package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
)

func CompaniesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetCompanies(w, r)
	case http.MethodPost:
		CreateCompany(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func CompanyResource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetCompanyByID(w, r)
	case http.MethodPut:
		UpdateCompany(w, r)
	case http.MethodDelete:
		DeleteCompany(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func GetCompanies(w http.ResponseWriter, r *http.Request) {
	rows, err := db().Query(`
		SELECT id, ticker, name, description, created_by, created_at, updated_at
		FROM companies
		ORDER BY ticker ASC, id ASC
	`)
	if err != nil {
		log.Printf("GetCompanies query failed: %v", err)
		writeError(w, http.StatusInternalServerError, "companies_query_failed", "failed to load companies")
		return
	}
	defer rows.Close()

	companies := make([]Company, 0)
	for rows.Next() {
		var company Company
		if err := rows.Scan(&company.ID, &company.Ticker, &company.Name, &company.Description, &company.CreatedBy, &company.CreatedAt, &company.UpdatedAt); err != nil {
			log.Printf("GetCompanies scan failed: %v", err)
			writeError(w, http.StatusInternalServerError, "companies_parse_failed", "failed to parse companies")
			return
		}
		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetCompanies rows failed: %v", err)
		writeError(w, http.StatusInternalServerError, "companies_read_failed", "failed to read companies")
		return
	}

	writeJSON(w, http.StatusOK, companies)
}

func GetCompanyByID(w http.ResponseWriter, r *http.Request) {
	companyID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_company_id", "company id must be a positive integer")
		return
	}

	var company Company
	err = db().QueryRow(`
		SELECT id, ticker, name, description, created_by, created_at, updated_at
		FROM companies
		WHERE id = ?
	`, companyID).Scan(&company.ID, &company.Ticker, &company.Name, &company.Description, &company.CreatedBy, &company.CreatedAt, &company.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "company_not_found", "company not found")
		return
	}
	if err != nil {
		log.Printf("GetCompanyByID failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_query_failed", "failed to load company")
		return
	}

	writeJSON(w, http.StatusOK, company)
}

func CreateCompany(w http.ResponseWriter, r *http.Request) {
	user, err := requireAuthenticatedUser(r)
	if err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	var input struct {
		Ticker      string `json:"ticker"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedBy   int    `json:"created_by"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	input.Ticker = normalizeTicker(input.Ticker)
	input.Name = trimRequired(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Ticker == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "ticker is required")
		return
	}
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "name is required")
		return
	}

	result, err := db().Exec(`
		INSERT INTO companies (ticker, name, description, created_by)
		VALUES (?, ?, ?, ?)
	`, input.Ticker, input.Name, input.Description, user.ID)
	if err != nil {
		log.Printf("CreateCompany insert failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_create_failed", "failed to create company")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("CreateCompany lastInsertId failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_create_failed", "failed to retrieve company")
		return
	}

	var company Company
	err = db().QueryRow(`
		SELECT id, ticker, name, description, created_by, created_at, updated_at
		FROM companies
		WHERE id = ?
	`, id).Scan(&company.ID, &company.Ticker, &company.Name, &company.Description, &company.CreatedBy, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		log.Printf("CreateCompany reload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_query_failed", "failed to retrieve company")
		return
	}

	writeJSON(w, http.StatusCreated, company)
}

func UpdateCompany(w http.ResponseWriter, r *http.Request) {
	user, err := requireAuthenticatedUser(r)
	if err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	companyID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_company_id", "company id must be a positive integer")
		return
	}

	var existing Company
	err = db().QueryRow(`
		SELECT id, ticker, name, description, created_by, created_at, updated_at
		FROM companies
		WHERE id = ?
	`, companyID).Scan(&existing.ID, &existing.Ticker, &existing.Name, &existing.Description, &existing.CreatedBy, &existing.CreatedAt, &existing.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "company_not_found", "company not found")
		return
	}
	if err != nil {
		log.Printf("UpdateCompany lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_query_failed", "failed to load company")
		return
	}

	if err := authorizeOwnership(user, existing.CreatedBy); err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	var input struct {
		Ticker      string `json:"ticker"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &input); err != nil {
		code, message := malformedJSONError(err)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	input.Ticker = normalizeTicker(input.Ticker)
	input.Name = trimRequired(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Ticker == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "ticker is required")
		return
	}
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "name is required")
		return
	}

	if _, err := db().Exec(`
		UPDATE companies
		SET ticker = ?, name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, input.Ticker, input.Name, input.Description, companyID); err != nil {
		log.Printf("UpdateCompany update failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_update_failed", "failed to update company")
		return
	}

	existing.Ticker = input.Ticker
	existing.Name = input.Name
	existing.Description = input.Description
	writeJSON(w, http.StatusOK, existing)
}

func DeleteCompany(w http.ResponseWriter, r *http.Request) {
	user, err := requireAuthenticatedUser(r)
	if err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	companyID, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_company_id", "company id must be a positive integer")
		return
	}

	var ownerID int
	if err := db().QueryRow(`SELECT created_by FROM companies WHERE id = ?`, companyID).Scan(&ownerID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "company_not_found", "company not found")
		return
	} else if err != nil {
		log.Printf("DeleteCompany lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_query_failed", "failed to load company")
		return
	}

	if err := authorizeOwnership(user, ownerID); err != nil {
		status, code, message := authError(err)
		writeError(w, status, code, message)
		return
	}

	tx, err := db().Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "company_delete_failed", "failed to delete company")
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`DELETE FROM companies WHERE id = ?`, companyID)
	if err != nil {
		log.Printf("DeleteCompany company failed: %v", err)
		writeError(w, http.StatusInternalServerError, "company_delete_failed", "failed to delete company")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "company_not_found", "company not found")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "company_delete_failed", "failed to finalize company deletion")
		return
	}

	writeNoContent(w)
}

func normalizeTicker(input string) string {
	return strings.ToUpper(strings.TrimSpace(input))
}
