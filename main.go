package main

import (
	"context"
	"fmt"
	"net/http"
	"bytes"
	"encoding/json"
	"os"
	"encoding/base64"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Demo 🚀",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	var tool mcp.Tool

	// Initialize Kanboard API client
	apiEndpoint := os.Getenv("KANBOARD_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://your-kanboard-url/jsonrpc.php"
	}

	apiKey := os.Getenv("KANBOARD_API_KEY")
	if apiKey == "" {
		apiKey = "your-kanboard-api-key"
	}

	kbUsername := os.Getenv("KANBOARD_USERNAME")
	if kbUsername == "" {
		kbUsername = "your-kanboard-username" // Default or placeholder
	}

	kbPassword := os.Getenv("KANBOARD_PASSWORD")
	if kbPassword == "" {
		kbPassword = "your-kanboard-password" // Default or placeholder
	}

	kbClient := newKanboardClient(apiEndpoint, apiKey, kbUsername, kbPassword)

	tool = mcp.NewTool("get_projects",
		mcp.WithDescription("List all projects"),
	)
	s.AddTool(tool, kbClient.getProjectsHandler)

	tool = mcp.NewTool("create_project",
		mcp.WithDescription("Create new projects"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the project to create"),
		),
	)
	s.AddTool(tool, kbClient.createProjectHandler)

	tool = mcp.NewTool("get_tasks",
		mcp.WithDescription("Get project tasks"),
		mcp.WithString("project_name",
			mcp.Required(),
			mcp.Description("Name of the project to get tasks from"),
		),
	)
	s.AddTool(tool, kbClient.getTasksHandler)

	tool = mcp.NewTool("create_task",
		mcp.WithDescription("Create new tasks"),
		mcp.WithString("project_name",
			mcp.Required(),
			mcp.Description("Name of the project to add the task to"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Title of the task to create"),
		),
	)
	s.AddTool(tool, kbClient.createTaskHandler)

	tool = mcp.NewTool("update_task",
		mcp.WithDescription("Modify existing tasks"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to update"),
		),
		mcp.WithString("description",
			mcp.Description("New description for the task"),
		),
		mcp.WithString("title",
			mcp.Description("New title for the task"),
		),
	)
	s.AddTool(tool, kbClient.updateTaskHandler)

	tool = mcp.NewTool("delete_task",
		mcp.WithDescription("Remove tasks"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to delete"),
		),
	)
	s.AddTool(tool, kbClient.deleteTaskHandler)

	tool = mcp.NewTool("get_task_details",
		mcp.WithDescription("Get detailed task info"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get details for"),
		),
	)
	s.AddTool(tool, kbClient.getTaskDetailsHandler)

	tool = mcp.NewTool("move_task",
		mcp.WithDescription("Move tasks between columns"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to move"),
		),
		mcp.WithString("column_name",
			mcp.Required(),
			mcp.Description("Name of the column to move the task to"),
		),
	)
	s.AddTool(tool, kbClient.moveTaskHandler)

	tool = mcp.NewTool("get_users",
		mcp.WithDescription("List all system users"),
	)
	s.AddTool(tool, kbClient.getUsersHandler)

	tool = mcp.NewTool("assign_task",
		mcp.WithDescription("Assign tasks to users"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to assign"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to assign the task to"),
		),
	)
	s.AddTool(tool, kbClient.assignTaskHandler)

	tool = mcp.NewTool("set_task_due_date",
		mcp.WithDescription("Set task deadlines"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to set the due date for"),
		),
		mcp.WithString("due_date",
			mcp.Required(),
			mcp.Description("Due date in YYYY-MM-DD format"),
		),
	)
	s.AddTool(tool, kbClient.setTaskDueDateHandler)

	tool = mcp.NewTool("add_task_comment",
		mcp.WithDescription("Add task comments"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to add a comment to"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user adding the comment"),
		),
		mcp.WithString("comment",
			mcp.Required(),
			mcp.Description("Content of the comment"),
		),
	)
	s.AddTool(tool, kbClient.addTaskCommentHandler)

	tool = mcp.NewTool("get_task_comments",
		mcp.WithDescription("Get task comments"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get comments for"),
		),
	)
	s.AddTool(tool, kbClient.getTaskCommentsHandler)

	tool = mcp.NewTool("assign_user_to_project",
		mcp.WithDescription("Assign a user to a project with a specific role"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to assign the user to"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to assign"),
		),
		mcp.WithString("role",
			mcp.Description("Role to assign (e.g., project-member, project-manager)"),
		),
	)
	s.AddTool(tool, kbClient.assignUserToProjectHandler)

	tool = mcp.NewTool("get_user_by_name",
		mcp.WithDescription("Get user by name"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("Username of the user to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getUserByNameHandler)

	tool = mcp.NewTool("create_user",
		mcp.WithDescription("Create a new user"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("Username for the new user"),
		),
		mcp.WithString("password",
			mcp.Required(),
			mcp.Description("Password for the new user"),
		),
		mcp.WithString("name",
			mcp.Description("Full name of the new user"),
		),
		mcp.WithString("email",
			mcp.Description("Email address of the new user"),
		),
	)
	s.AddTool(tool, kbClient.createUserHandler)

	tool = mcp.NewTool("update_user",
		mcp.WithDescription("Modify an existing user"),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to update"),
		),
		mcp.WithString("username",
			mcp.Description("New username"),
		),
		mcp.WithString("password",
			mcp.Description("New password"),
		),
		mcp.WithString("name",
			mcp.Description("New full name"),
		),
		mcp.WithString("email",
			mcp.Description("New email address"),
		),
	)
	s.AddTool(tool, kbClient.updateUserHandler)

	tool = mcp.NewTool("remove_user",
		mcp.WithDescription("Remove a user"),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeUserHandler)

	tool = mcp.NewTool("get_columns",
		mcp.WithDescription("List project columns"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get columns from"),
		),
	)
	s.AddTool(tool, kbClient.getColumnsHandler)

	tool = mcp.NewTool("create_column",
		mcp.WithDescription("Add new columns"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to add the column to"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the column to create"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Task limit for the new column"),
		),
	)
	s.AddTool(tool, kbClient.createColumnHandler)

	tool = mcp.NewTool("update_column",
		mcp.WithDescription("Modify column settings"),
		mcp.WithNumber("column_id",
			mcp.Required(),
			mcp.Description("ID of the column to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the column"),
		),
		mcp.WithNumber("limit",
			mcp.Description("New task limit for the column"),
		),
	)
	s.AddTool(tool, kbClient.updateColumnHandler)

	tool = mcp.NewTool("delete_column",
		mcp.WithDescription("Remove columns"),
		mcp.WithNumber("column_id",
			mcp.Required(),
			mcp.Description("ID of the column to delete"),
		),
	)
	s.AddTool(tool, kbClient.deleteColumnHandler)

	tool = mcp.NewTool("reorder_columns",
		mcp.WithDescription("Change column positions"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("ID of the project containing the columns"),
		),
		mcp.WithNumber("column_id",
			mcp.Required(),
			mcp.Description("ID of the column to reorder"),
		),
		mcp.WithNumber("new_position",
			mcp.Required(),
			mcp.Description("New position for the column"),
		),
	)
	s.AddTool(tool, kbClient.reorderColumnsHandler)

	tool = mcp.NewTool("get_categories",
		mcp.WithDescription("List project categories"),
		mcp.WithString("project_id",
			mcp.Description("ID of the project to get categories from (optional)"),
		),
	)
	s.AddTool(tool, kbClient.getCategoriesHandler)

	tool = mcp.NewTool("create_category",
		mcp.WithDescription("Add task categories"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the category to create"),
		),
		mcp.WithString("project_id",
			mcp.Description("ID of the project to add the category to (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createCategoryHandler)

	tool = mcp.NewTool("update_category",
		mcp.WithDescription("Modify categories"),
		mcp.WithNumber("category_id",
			mcp.Required(),
			mcp.Description("ID of the category to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the category"),
		),
	)
	s.AddTool(tool, kbClient.updateCategoryHandler)

	tool = mcp.NewTool("delete_category",
		mcp.WithDescription("Remove categories"),
		mcp.WithNumber("category_id",
			mcp.Required(),
			mcp.Description("ID of the category to delete"),
		),
	)
	s.AddTool(tool, kbClient.deleteCategoryHandler)

	tool = mcp.NewTool("get_swimlanes",
		mcp.WithDescription("List project swimlanes"),
		mcp.WithString("project_id",
			mcp.Description("ID of the project to get swimlanes from (optional)"),
		),
	)
	s.AddTool(tool, kbClient.getSwimlanesHandler)

	tool = mcp.NewTool("create_swimlane",
		mcp.WithDescription("Add team swimlanes"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to add the swimlane to"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the swimlane to create"),
		),
	)
	s.AddTool(tool, kbClient.createSwimlaneHandler)

	tool = mcp.NewTool("update_swimlane",
		mcp.WithDescription("Modify swimlanes"),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the swimlane"),
		),
	)
	s.AddTool(tool, kbClient.updateSwimlaneHandler)

	tool = mcp.NewTool("delete_swimlane",
		mcp.WithDescription("Remove swimlanes"),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to delete"),
		),
	)
	s.AddTool(tool, kbClient.deleteSwimlaneHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

type kanboardClient struct {
	apiEndpoint string
	apiKey      string
	username    string
	password    string
}

func newKanboardClient(apiEndpoint, apiKey, username, password string) *kanboardClient {
	return &kanboardClient{
		apiEndpoint: apiEndpoint,
		apiKey:      apiKey,
		username:    username,
		password:    password,
	}
}

func (kc *kanboardClient) callKanboardAPI(ctx context.Context, method string, params interface{}) (interface{}, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      1,
		"params":  params,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", kc.apiEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if kc.username != "" && kc.password != "" {
		auth := kc.username + ":" + kc.password
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", basicAuth)
	} else {
		req.Header.Set("X-API-Auth", kc.apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Jsonrpc string      `json:"jsonrpc"`
		ID      int         `json:"id"`
		Result  interface{} `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("Failed to decode API response: %v", err)
	}

	if apiResponse.Error != nil {
		return nil, fmt.Errorf("Kanboard API error: %s (Code: %d)", apiResponse.Error.Message, apiResponse.Error.Code)
	}

	return apiResponse.Result, nil
}

func (kc *kanboardClient) getProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getAllProjects", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"name": name}
	result, err := kc.callKanboardAPI(ctx, "createProject", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_name, err := request.RequireString("project_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// First, get the project ID from the project name
	result, err := kc.callKanboardAPI(ctx, "getProjectByName", map[string]string{"name": project_name})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project ID: %v", err)), nil
	}

	// Assuming the result is an object, parse it to extract the ID
	var projectInfo struct {
		ID string `json:"id"`
	}
	// Marshal and unmarshal to ensure correct type conversion from interface{} to struct
	tempBytes, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal project info for parsing: %v", err)), nil
	}
	if err := json.Unmarshal(tempBytes, &projectInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse project info: %v", err)), nil
	}

	if projectInfo.ID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found", project_name)), nil
	}

	params := map[string]interface{}{"project_id": projectInfo.ID}
	result, err = kc.callKanboardAPI(ctx, "getAllTasks", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get tasks: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_name, err := request.RequireString("project_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// First, get the project ID from the project name
	result, err := kc.callKanboardAPI(ctx, "getProjectByName", map[string]string{"name": project_name})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project ID: %v", err)), nil
	}

	var projectInfo struct {
		ID string `json:"id"`
	}
	tempBytes, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal project info for parsing: %v", err)), nil
	}
	if err := json.Unmarshal(tempBytes, &projectInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse project info: %v", err)), nil
	}

	if projectInfo.ID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found", project_name)), nil
	}

	params := map[string]interface{}{
		"project_id": projectInfo.ID,
		"title":      title,
	}
	result, err = kc.callKanboardAPI(ctx, "createTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": task_id}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}

	title := request.GetString("title", "")
	if title != "" {
		params["title"] = title
	}

	result, err := kc.callKanboardAPI(ctx, "updateTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) deleteTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id}
	result, err := kc.callKanboardAPI(ctx, "removeTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTaskDetailsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id}
	result, err := kc.callKanboardAPI(ctx, "getTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get task details: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) moveTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	column_name, err := request.RequireString("column_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// To move a task, we need the column_id. First, get the project ID for the task.
	// This involves a couple of steps: get the task, then get its project ID.
	result, err := kc.callKanboardAPI(ctx, "getTask", map[string]int{"task_id": task_id})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get task details: %v", err)), nil
	}

	var taskInfo struct {
		ProjectID string `json:"project_id"`
	}
	tempBytes, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal task info for parsing: %v", err)), nil
	}
	if err := json.Unmarshal(tempBytes, &taskInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse task info: %v", err)), nil
	}

	if taskInfo.ProjectID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Could not determine project for task #%d", task_id)), nil
	}

	// Now, get the columns for that project to find the column_id by name.
	result, err = kc.callKanboardAPI(ctx, "getColumns", map[string]string{"project_id": taskInfo.ProjectID})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get columns for project: %v", err)), nil
	}

	var columns []struct {
		ID   string `json:"id"`
		Name string `json:"title"` // Kanboard API returns 'title' for column name
	}

	tempBytes, err = json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal columns info for parsing: %v", err)), nil
	}
	if err := json.Unmarshal(tempBytes, &columns); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse columns info: %v", err)), nil
	}

	var targetColumnID string
	for _, col := range columns {
		if col.Name == column_name {
			targetColumnID = col.ID
			break
		}
	}

	if targetColumnID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Column '%s' not found in project for task #%d", column_name, task_id)), nil
	}

	// Finally, move the task.
	moveParams := map[string]interface{}{
		"task_id":    task_id,
		"column_id":  targetColumnID,
		"position":   1, // Assuming position 1 for simplicity
		"swimlane_id": 0, // Assuming default swimlane
		"project_id": taskInfo.ProjectID, // Required for moveTaskPosition
	}
	result, err = kc.callKanboardAPI(ctx, "moveTaskPosition", moveParams)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getAllUsers", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getUserByNameHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, err := request.RequireString("username")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"username": username}
	result, err := kc.callKanboardAPI(ctx, "getUserByName", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get user by name: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, err := request.RequireString("username")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	password, err := request.RequireString("password")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"username": username,
		"password": password,
	}

	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	email := request.GetString("email", "")
	if email != "" {
		params["email"] = email
	}

	result, err := kc.callKanboardAPI(ctx, "createUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create user: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": user_id}

	username := request.GetString("username", "")
	if username != "" {
		params["username"] = username
	}
	password := request.GetString("password", "")
	if password != "" {
		params["password"] = password
	}
	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}
	email := request.GetString("email", "")
	if email != "" {
		params["email"] = email
	}

	result, err := kc.callKanboardAPI(ctx, "updateUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update user: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": user_id}
	result, err := kc.callKanboardAPI(ctx, "removeUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove user: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getColumnsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id_str, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}

	params := map[string]int{"project_id": project_id}
	result, err := kc.callKanboardAPI(ctx, "getColumns", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get columns: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createColumnHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id_str, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id": project_id,
		"title":      name, // Kanboard API expects 'title'
	}

	limit := request.GetInt("limit", 0)
	if limit != 0 {
		params["task_limit"] = limit
	}

	result, err := kc.callKanboardAPI(ctx, "addColumn", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create column: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateColumnHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	column_id, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": column_id}

	name := request.GetString("name", "")
	if name != "" {
		params["title"] = name // Kanboard API expects 'title'
	} else {
		// If name is not provided, fetch the existing column to get its current title
		result, err := kc.callKanboardAPI(ctx, "getColumn", map[string]int{"column_id": column_id})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get column details to infer title: %v", err)), nil
		}
		var columnInfo struct {
			Title string `json:"title"`
		}
		tempBytes, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal column info for parsing: %v", err)), nil
		}
		if err := json.Unmarshal(tempBytes, &columnInfo); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse column info to infer title: %v", err)), nil
		}
		params["title"] = columnInfo.Title // Use the existing title
	}

	limit := request.GetInt("limit", 0)
	if limit != 0 {
		params["task_limit"] = limit
	}

	result, err := kc.callKanboardAPI(ctx, "updateColumn", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update column: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) deleteColumnHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	column_id, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"column_id": column_id}
	result, err := kc.callKanboardAPI(ctx, "removeColumn", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete column: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) reorderColumnsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id_str, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}
	new_position, err := request.RequireInt("new_position")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id":  project_id,
		"column_id":   request.GetInt("column_id", 0), // column_id is needed but not exposed directly in NLP, might need to get it via column name
		"position":    new_position,
	}

	// Kanboard's moveColumnPosition requires a column_id. If we only have name, we need to resolve it.
	// For simplicity, assuming column_id is provided, or a mechanism to resolve it exists for the LLM.

	result, err := kc.callKanboardAPI(ctx, "moveColumnPosition", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reorder columns: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getCategoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id := request.GetString("project_id", "") // Categories can be global or project-specific

	var result interface{}
	var err error
	if project_id != "" {
		params := map[string]string{"project_id": project_id}
		result, err = kc.callKanboardAPI(ctx, "getAllCategories", params)
	} else {
		result, err = kc.callKanboardAPI(ctx, "getAllCategories", nil)
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get categories: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id := request.GetString("project_id", "")
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"name": name}
	if project_id != "" {
		params["project_id"] = project_id
	}

	result, err := kc.callKanboardAPI(ctx, "createCategory", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create category: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category_id, err := request.RequireInt("category_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": category_id}
	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	result, err := kc.callKanboardAPI(ctx, "updateCategory", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update category: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) deleteCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category_id, err := request.RequireInt("category_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"category_id": category_id}
	result, err := kc.callKanboardAPI(ctx, "removeCategory", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete category: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getSwimlanesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id := request.GetString("project_id", "") // Swimlanes can be global or project-specific

	var result interface{}
	var err error
	if project_id != "" {
		params := map[string]string{"project_id": project_id}
		result, err = kc.callKanboardAPI(ctx, "getAllSwimlanes", params)
	} else {
		result, err = kc.callKanboardAPI(ctx, "getAllSwimlanes", nil)
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get swimlanes: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createSwimlaneHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id": project_id,
		"name":       name,
	}

	result, err := kc.callKanboardAPI(ctx, "addSwimlane", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create swimlane: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateSwimlaneHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	swimlane_id, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": swimlane_id}
	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	result, err := kc.callKanboardAPI(ctx, "updateSwimlane", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update swimlane: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) deleteSwimlaneHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	swimlane_id, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"swimlane_id": swimlane_id}
	result, err := kc.callKanboardAPI(ctx, "removeSwimlane", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete swimlane: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) assignTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]int{"task_id": task_id, "owner_id": user_id}
	result, err := kc.callKanboardAPI(ctx, "assignTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to assign task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) setTaskDueDateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	due_date, err := request.RequireString("due_date")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"task_id": task_id, "date_due": due_date}
	result, err := kc.callKanboardAPI(ctx, "updateTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set task due date: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) addTaskCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	comment, err := request.RequireString("comment")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"task_id": task_id,
		"user_id": user_id,
		"content": comment,
	}
	result, err := kc.callKanboardAPI(ctx, "addComment", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add task comment: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTaskCommentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id}
	result, err := kc.callKanboardAPI(ctx, "getAllComments", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get task comments: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) assignUserToProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id_str, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}

	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	role := request.GetString("role", "project-member") // Default to project-member

	params := map[string]interface{}{
		"project_id": project_id,
		"user_id":    user_id,
		"role":       role,
	}

	result, err := kc.callKanboardAPI(ctx, "addProjectUser", params) // Assuming addProjectUser API call
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to assign user to project: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}
