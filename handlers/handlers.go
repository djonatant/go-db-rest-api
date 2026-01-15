package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/go-db-rest-api/database"
)

type Handler struct {
	manager *database.ConnectionManager
	config  database.DBConfig
}

func NewHandler(manager *database.ConnectionManager, config database.DBConfig) *Handler {
	return &Handler{
		manager: manager,
		config:  config,
	}
}

type QueryItem struct {
	Name string `json:"name"`
	Sql  string `json:"sql"`
}

type ConnectRequest struct {
	database.DBConfig
}

func (h *Handler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	if err := h.manager.Connect(id, req.DBConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"connection_id": id, "message": "connected successfully"})
}

func (h *Handler) Query(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := database.OpenConnection(h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
		return
	}
	defer db.Close()

	// Check if it's a multi-query request
	if queriesIntf, ok := req["queries"]; ok {
		queriesArr, ok := queriesIntf.([]interface{})
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "queries must be an array"})
			return
		}

		// Prepare parameters (exclude the "queries" field)
		params := make(map[string]interface{})
		for k, v := range req {
			if k != "queries" {
				params[k] = v
			}
		}

		results := make(map[string]interface{})
		for _, qObj := range queriesArr {
			qMap, ok := qObj.(map[string]interface{})
			if !ok {
				continue
			}

			name, _ := qMap["name"].(string)
			sqlStr, _ := qMap["sql"].(string)

			if name == "" || sqlStr == "" {
				continue
			}

			finalSQL := ProcessSQL(sqlStr, params)
			rows, err := db.Query(finalSQL)
			if err != nil {
				results[name] = gin.H{"error": err.Error()}
				continue
			}

			result, err := h.scanRows(rows)
			rows.Close()
			if err != nil {
				results[name] = gin.H{"error": err.Error()}
				continue
			}
			results[name] = result
		}

		c.JSON(http.StatusOK, results)
		return
	}

	// Fallback to single query mode
	queryIntf, ok := req["query"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query field is required"})
		return
	}
	query, ok := queryIntf.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query must be a string"})
		return
	}

	// Remove query from params map so it's not treated as a replacement param
	delete(req, "query")

	// Process the SQL based on params
	finalQuery := ProcessSQL(query, req)

	rows, err := db.Query(finalQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("query failed: %v", err)})
		return
	}
	defer rows.Close()

	result, err := h.scanRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		result = append(result, entry)
	}
	return result, nil
}

func (h *Handler) Disconnect(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.CloseConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disconnected successfully"})
}
