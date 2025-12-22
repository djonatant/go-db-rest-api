package handlers

import (
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

type QueryRequest struct {
	Query string `json:"query" binding:"required"`
}

func (h *Handler) Query(c *gin.Context) {
	var req QueryRequest
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

	rows, err := db.Query(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("query failed: %v", err)})
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
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

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Disconnect(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.CloseConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disconnected successfully"})
}
