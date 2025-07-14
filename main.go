package main

import (
	"context"
	"fmt"
	"net/http"
	"bytes"
	"encoding/json"
	"os"

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

	// Add tool
	tool := mcp.NewTool("hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)
	s.AddTool(tool, helloHandler)

	// Initialize Kanboard API client
	apiEndpoint := os.Getenv("KANBOARD_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://your-kanboard-url/jsonrpc.php"
	}

	apiKey := os.Getenv("KANBOARD_API_KEY")
	if apiKey == "" {
		apiKey = "your-kanboard-api-key"
	}

	kbClient := newKanboardClient(apiEndpoint, apiKey)

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
		mcp.WithDescription("List all users"),
	)
	s.AddTool(tool, kbClient.getUsersHandler)

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

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}

type kanboardClient struct {
	apiEndpoint string
	apiKey      string
}

func newKanboardClient(apiEndpoint, apiKey string) *kanboardClient {
	return &kanboardClient{
		apiEndpoint: apiEndpoint,
		apiKey:      apiKey,
	}
}

func (kc *kanboardClient) callKanboardAPI(ctx context.Context, method string, params interface{}) (*mcp.CallToolResult, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      1,
		"params":  params,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal request: %v", err)), nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", kc.apiEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create request: %v", err)), nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Auth", kc.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to send request: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return mcp.NewToolResultError(fmt.Sprintf("API request failed with status code: %d", resp.StatusCode)), nil
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
		return mcp.NewToolResultError(fmt.Sprintf("Failed to decode API response: %v", err)), nil
	}

	if apiResponse.Error != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Kanboard API error: %s (Code: %d)", apiResponse.Error.Message, apiResponse.Error.Code)), nil
	}

	// Convert result to JSON string for display
	resultBytes, err := json.MarshalIndent(apiResponse.Result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return kc.callKanboardAPI(ctx, "getAllProjects", nil)
}

func (kc *kanboardClient) createProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"name": name}
	return kc.callKanboardAPI(ctx, "createProject", params)
}

func (kc *kanboardClient) getTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_name, err := request.RequireString("project_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// First, get the project ID from the project name
	projectParams := map[string]string{"name": project_name}
	projectResult, err := kc.callKanboardAPI(ctx, "getProjectByName", projectParams)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project ID: %v", err)), nil
	}

	// Assuming the result is a JSON string, parse it to extract the ID
	var projectInfo struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(projectResult.Text), &projectInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse project info: %v", err)), nil
	}

	if projectInfo.ID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found", project_name)), nil
	}

	params := map[string]interface{}{"project_id": projectInfo.ID}
	return kc.callKanboardAPI(ctx, "getAllTasks", params)
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
	projectParams := map[string]string{"name": project_name}
	projectResult, err := kc.callKanboardAPI(ctx, "getProjectByName", projectParams)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project ID: %v", err)), nil
	}

	var projectInfo struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(projectResult.Text), &projectInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse project info: %v", err)), nil
	}

	if projectInfo.ID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found", project_name)), nil
	}

	params := map[string]interface{}{
		"project_id": projectInfo.ID,
		"title":      title,
	}
	return kc.callKanboardAPI(ctx, "createTask", params)
}

func (kc *kanboardClient) updateTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": task_id}

	description, err := request.GetString("description")
	if err == nil && description != "" {
		params["description"] = description
	}

	title, err := request.GetString("title")
	if err == nil && title != "" {
		params["title"] = title
	}

	return kc.callKanboardAPI(ctx, "updateTask", params)
}

func (kc *kanboardClient) deleteTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id}
	return kc.callKanboardAPI(ctx, "removeTask", params)
}

func (kc *kanboardClient) getTaskDetailsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id}
	return kc.callKanboardAPI(ctx, "getTask", params)
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
	taskParams := map[string]int{"task_id": task_id}
	taskResult, err := kc.callKanboardAPI(ctx, "getTask", taskParams)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get task details: %v", err)), nil
	}

	var taskInfo struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal([]byte(taskResult.Text), &taskInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse task info: %v", err)), nil
	}

	if taskInfo.ProjectID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Could not determine project for task #%d", task_id)), nil
	}

	// Now, get the columns for that project to find the column_id by name.
	columnParams := map[string]string{"project_id": taskInfo.ProjectID}
	columnsResult, err := kc.callKanboardAPI(ctx, "getColumns", columnParams)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get columns for project: %v", err)), nil
	}

	var columns []struct {
		ID   string `json:"id"`
		Name string `json:"title"` // Kanboard API returns 'title' for column name
	}

	if err := json.Unmarshal([]byte(columnsResult.Text), &columns); err != nil {
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
	return kc.callKanboardAPI(ctx, "moveTaskPosition", moveParams)
}

func (kc *kanboardClient) getUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return kc.callKanboardAPI(ctx, "getAllUsers", nil)
}

func (kc *kanboardClient) getUserByNameHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, err := request.RequireString("username")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"username": username}
	return kc.callKanboardAPI(ctx, "getUserByName", params)
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

	name, err := request.GetString("name")
	if err == nil && name != "" {
		params["name"] = name
	}

	email, err := request.GetString("email")
	if err == nil && email != "" {
		params["email"] = email
	}

	return kc.callKanboardAPI(ctx, "createUser", params)
}

func (kc *kanboardClient) updateUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": user_id}

	username, err := request.GetString("username")
	if err == nil && username != "" {
		params["username"] = username
	}
	password, err := request.GetString("password")
	if err == nil && password != "" {
		params["password"] = password
	}
	name, err := request.GetString("name")
	if err == nil && name != "" {
		params["name"] = name
	}
	email, err := request.GetString("email")
	if err == nil && email != "" {
		params["email"] = email
	}

	return kc.callKanboardAPI(ctx, "updateUser", params)
}

func (kc *kanboardClient) removeUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": user_id}
	return kc.callKanboardAPI(ctx, "removeUser", params)
}
