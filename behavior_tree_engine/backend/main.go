package main

import (
	"behavior_tree_engine/backend/core"
	"behavior_tree_engine/backend/database"
	"behavior_tree_engine/backend/models"
	"encoding/json" // Required for Bind (indirectly) and potentially for logging context
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

// APIError represents a standard error response format
type APIError struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ExecutionRequest defines the structure for the behavior tree execution request payload
type ExecutionRequest struct {
	TargetState map[string]interface{} `json:"target_state"`
}

// ExecutionResponse defines the structure for the behavior tree execution response
type ExecutionResponse struct {
	RunID       string     `json:"run_id"`
	FinalStatus string     `json:"final_status"`
	Logs        []LogEntry `json:"logs"`
	DurationMs  int64      `json:"duration_ms"`
	Error       *APIError  `json:"error,omitempty"`
}

// LogEntry is a simplified structure for execution logs returned in API.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	NodeID    string `json:"node_id,omitempty"`
	NodeName  string `json:"node_name,omitempty"`
	NodeType  string `json:"node_type,omitempty"`
	Status    string `json:"status,omitempty"`
	Message   string `json:"message"`
}

func main() {
	// Initialize Database
	database.InitDatabase("behavior_trees.db")
	// core.InitDefaultRegistry() is called by its package init()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"}, // Frontend dev server
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
	}))

	// Routes
	api := e.Group("/api")

	api.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Behavior Tree Engine API")
	})

	// Behavior Tree CRUD
	btGroup := api.Group("/behavior-trees")
	btGroup.POST("", createBehaviorTree)
	btGroup.GET("", getAllBehaviorTrees)
	btGroup.GET("/:id", getBehaviorTree)
	// TODO: btGroup.PUT("/:id", updateBehaviorTree)
	// TODO: btGroup.DELETE("/:id", deleteBehaviorTree)

	// Behavior Tree Execution
	btGroup.POST("/:id/execute", executeBehaviorTree)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// --- Behavior Tree Handlers ---

func createBehaviorTree(c echo.Context) error {
	bt := new(models.BehaviorTree)
	if err := c.Bind(bt); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Message: "Invalid input", Details: err.Error()})
	}

	if bt.Name == "" {
		return c.JSON(http.StatusBadRequest, APIError{Message: "Behavior tree name is required"})
	}
	if bt.Definition == "" {
		// Ensure a default valid JSON if empty, e.g., {"nodes":[], "edges":[]}
		// For now, require it from client.
		return c.JSON(http.StatusBadRequest, APIError{Message: "Behavior tree definition is required"})
	}

	// Validate JSON structure of definition
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(bt.Definition), &js); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Message: "Behavior tree definition is not valid JSON", Details: err.Error()})
	}


	result := database.DB.Create(bt)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Message: "Failed to create behavior tree", Details: result.Error.Error()})
	}
	return c.JSON(http.StatusCreated, bt)
}

func getAllBehaviorTrees(c echo.Context) error {
	var trees []models.BehaviorTree
	result := database.DB.Order("created_at desc").Find(&trees)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Message: "Failed to retrieve behavior trees", Details: result.Error.Error()})
	}
	return c.JSON(http.StatusOK, trees)
}

func getBehaviorTree(c echo.Context) error {
	id := c.Param("id")
	var tree models.BehaviorTree
	result := database.DB.First(&tree, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, APIError{Message: "Behavior tree not found"})
		}
		return c.JSON(http.StatusInternalServerError, APIError{Message: "Failed to retrieve behavior tree", Details: result.Error.Error()})
	}
	return c.JSON(http.StatusOK, tree)
}

// --- Behavior Tree Execution Handler ---

func executeBehaviorTree(c echo.Context) error {
	startTime := time.Now()
	runID := fmt.Sprintf("run_%s", strconv.FormatInt(time.Now().UnixNano(), 10))

	treeID := c.Param("id")
	var dbTree models.BehaviorTree
	result := database.DB.First(&dbTree, treeID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, APIError{Message: "Behavior tree not found"})
		}
		return c.JSON(http.StatusInternalServerError, APIError{Message: "Failed to retrieve behavior tree", Details: result.Error.Error()})
	}

	var req ExecutionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Message: "Invalid request payload", Details: err.Error()})
	}

	if req.TargetState == nil {
		req.TargetState = make(map[string]interface{}) // Default to empty map
	}

	runnableTree, err := core.ParseTreeDefinition(dbTree.Definition)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Message: "Failed to parse behavior tree definition", Details: err.Error()})
	}
	runnableTree.ID = dbTree.Name // For logging purposes

	var apiLogs []LogEntry
	apiLogs = append(apiLogs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   fmt.Sprintf("Starting execution of tree '%s' (ID: %s) for run %s. Target: %+v", dbTree.Name, treeID, runID, req.TargetState),
	})

	deltaTime := 1.0 / 60.0 // Example: 60 FPS
	finalStatus := runnableTree.Tick(req.TargetState, deltaTime)

	apiLogs = append(apiLogs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   fmt.Sprintf("Tree '%s' execution finished. Final Status: %s. Final Target State: %+v", dbTree.Name, finalStatus.String(), req.TargetState),
		Status:    finalStatus.String(),
	})

	// Example of saving a more detailed log to the database
	// This is simplified; a real implementation would collect logs during Ticks.
	targetStateBytes, _ := json.Marshal(req.TargetState)
	dbLog := models.ExecutionLog{
		TreeID:    dbTree.ID,
		RunID:     runID,
		Timestamp: time.Now(),
		Status:    finalStatus.String(),
		Message:   fmt.Sprintf("Execution of tree %s completed with status %s.", dbTree.Name, finalStatus.String()),
		Context:   string(targetStateBytes),
	}
	if res := database.DB.Create(&dbLog); res.Error != nil {
		// Log this error but don't fail the entire request for it
		c.Logger().Errorf("Failed to save execution log for run %s: %v", runID, res.Error)
	}


	durationMs := time.Since(startTime).Milliseconds()

	return c.JSON(http.StatusOK, ExecutionResponse{
		RunID:       runID,
		FinalStatus: finalStatus.String(),
		Logs:        apiLogs,
		DurationMs:  durationMs,
	})
}

// TODO: Implement updateBehaviorTree and deleteBehaviorTree
