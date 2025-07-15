package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"KanboardMCP",
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
		mcp.WithString("description",
			mcp.Description("Description of the project (optional)"),
		),
		mcp.WithNumber("owner_id",
			mcp.Description("ID of the project owner (optional)"),
		),
		mcp.WithString("identifier",
			mcp.Description("Alphanumeric project identifier (optional)"),
		),
		mcp.WithString("start_date",
			mcp.Description("Project start date in ISO8601 format (optional)"),
		),
		mcp.WithString("end_date",
			mcp.Description("Project end date in ISO8601 format (optional)"),
		),
		mcp.WithNumber("priority_default",
			mcp.Description("Default task priority (optional)"),
		),
		mcp.WithNumber("priority_start",
			mcp.Description("Start priority (optional)"),
		),
		mcp.WithNumber("priority_end",
			mcp.Description("End priority (optional)"),
		),
		mcp.WithString("email",
			mcp.Description("Project email address (optional)"),
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
		mcp.WithString("color_id",
			mcp.Description("Color ID for the task (optional)"),
		),
		mcp.WithNumber("column_id",
			mcp.Description("ID of the column to add the task to (optional)"),
		),
		mcp.WithNumber("owner_id",
			mcp.Description("ID of the task owner (optional)"),
		),
		mcp.WithNumber("creator_id",
			mcp.Description("ID of the task creator (optional)"),
		),
		mcp.WithString("date_due",
			mcp.Description("Due date in YYYY-MM-DD HH:MM format (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("Markdown content for the task description (optional)"),
		),
		mcp.WithNumber("category_id",
			mcp.Description("ID of the task category (optional)"),
		),
		mcp.WithNumber("score",
			mcp.Description("Complexity score of the task (optional)"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Description("ID of the swimlane to add the task to (optional)"),
		),
		mcp.WithNumber("priority",
			mcp.Description("Priority of the task (optional)"),
		),
		mcp.WithNumber("recurrence_status",
			mcp.Description("Recurrence status of the task (optional)"),
		),
		mcp.WithNumber("recurrence_trigger",
			mcp.Description("Recurrence trigger of the task (optional)"),
		),
		mcp.WithNumber("recurrence_factor",
			mcp.Description("Recurrence factor of the task (optional)"),
		),
		mcp.WithNumber("recurrence_timeframe",
			mcp.Description("Recurrence timeframe of the task (optional)"),
		),
		mcp.WithNumber("recurrence_basedate",
			mcp.Description("Recurrence base date of the task (optional)"),
		),
		mcp.WithString("reference",
			mcp.Description("External reference for the task (optional)"),
		),
		mcp.WithArray("tags",
			mcp.WithStringItems(),
			mcp.Description("List of tags (array of strings) (optional)"),
		),
		mcp.WithString("date_started",
			mcp.Description("Start date in YYYY-MM-DD HH:MM format (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createTaskHandler)

	tool = mcp.NewTool("update_task",
		mcp.WithDescription("Update a task"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("ID of the task to update"),
		),
		mcp.WithString("title",
			mcp.Description("New title for the task (optional)"),
		),
		mcp.WithString("color_id",
			mcp.Description("New color ID for the task (optional)"),
		),
		mcp.WithNumber("owner_id",
			mcp.Description("New owner ID for the task (optional)"),
		),
		mcp.WithString("date_due",
			mcp.Description("New due date in YYYY-MM-DD HH:MM format (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New Markdown content for the task description (optional)"),
		),
		mcp.WithNumber("category_id",
			mcp.Description("New ID of the task category (optional)"),
		),
		mcp.WithNumber("score",
			mcp.Description("New complexity score of the task (optional)"),
		),
		mcp.WithNumber("priority",
			mcp.Description("New priority of the task (optional)"),
		),
		mcp.WithNumber("recurrence_status",
			mcp.Description("New recurrence status of the task (optional)"),
		),
		mcp.WithNumber("recurrence_trigger",
			mcp.Description("New recurrence trigger of the task (optional)"),
		),
		mcp.WithNumber("recurrence_factor",
			mcp.Description("New recurrence factor of the task (optional)"),
		),
		mcp.WithNumber("recurrence_timeframe",
			mcp.Description("New recurrence timeframe of the task (optional)"),
		),
		mcp.WithNumber("recurrence_basedate",
			mcp.Description("New recurrence base date of the task (optional)"),
		),
		mcp.WithString("reference",
			mcp.Description("New external reference for the task (optional)"),
		),
		mcp.WithArray("tags",
			mcp.WithStringItems(),
			mcp.Description("New list of tags (array of strings) (optional)"),
		),
		mcp.WithString("date_started",
			mcp.Description("New start date in YYYY-MM-DD HH:MM format (optional)"),
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

	tool = mcp.NewTool("get_task",
		mcp.WithDescription("Get task by the unique id"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get details for"),
		),
	)
	s.AddTool(tool, kbClient.getTaskHandler)

	tool = mcp.NewTool("move_task_position",
		mcp.WithDescription("Move a task to another column, position or swimlane inside the same board"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project containing the task"),
		),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to move"),
		),
		mcp.WithNumber("column_id",
			mcp.Required(),
			mcp.Description("ID of the column to move the task to"),
		),
		mcp.WithNumber("position",
			mcp.Required(),
			mcp.Description("New position for the task (must be >= 1)"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to move the task to"),
		),
	)
	s.AddTool(tool, kbClient.moveTaskPositionHandler)

	tool = mcp.NewTool("get_users",
		mcp.WithDescription("List all system users"),
	)
	s.AddTool(tool, kbClient.getUsersHandler)

	tool = mcp.NewTool("create_user",
		mcp.WithDescription("Create a new user"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("Username for the new user (must be unique)"),
		),
		mcp.WithString("password",
			mcp.Required(),
			mcp.Description("Password for the new user (must have at least 6 characters)"),
		),
		mcp.WithString("name",
			mcp.Description("Full name of the new user (optional)"),
		),
		mcp.WithString("email",
			mcp.Description("Email address of the new user (optional)"),
		),
		mcp.WithString("role",
			mcp.Description("Role for the user (app-admin, app-manager, app-user) (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createUserHandler)

	tool = mcp.NewTool("create_ldap_user",
		mcp.WithDescription("Create a new user authenticated by LDAP"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("Username for the LDAP user"),
		),
	)
	s.AddTool(tool, kbClient.createLdapUserHandler)

	tool = mcp.NewTool("get_user",
		mcp.WithDescription("Get user information by ID"),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getUserHandler)

	tool = mcp.NewTool("get_user_by_name",
		mcp.WithDescription("Get user information by username"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("Username of the user to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getUserByNameHandler)

	tool = mcp.NewTool("update_user",
		mcp.WithDescription("Update a user"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("ID of the user to update"),
		),
		mcp.WithString("username",
			mcp.Description("New username (optional)"),
		),
		mcp.WithString("name",
			mcp.Description("New full name (optional)"),
		),
		mcp.WithString("email",
			mcp.Description("New email address (optional)"),
		),
		mcp.WithString("role",
			mcp.Description("New role (app-admin, app-manager, app-user) (optional)"),
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

	tool = mcp.NewTool("disable_user",
		mcp.WithDescription("Disable a user"),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to disable"),
		),
	)
	s.AddTool(tool, kbClient.disableUserHandler)

	tool = mcp.NewTool("enable_user",
		mcp.WithDescription("Enable a user"),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to enable"),
		),
	)
	s.AddTool(tool, kbClient.enableUserHandler)

	tool = mcp.NewTool("is_active_user",
		mcp.WithDescription("Check if a user is active"),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to check"),
		),
	)
	s.AddTool(tool, kbClient.isActiveUserHandler)

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

	tool = mcp.NewTool("create_comment",
		mcp.WithDescription("Create a new comment"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to add a comment to"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user adding the comment"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Markdown content for the comment"),
		),
		mcp.WithString("reference",
			mcp.Description("External reference for the comment"),
		),
		mcp.WithString("visibility",
			mcp.Description("Visibility of the comment (app-user, app-manager, app-admin)"),
		),
	)
	s.AddTool(tool, kbClient.createCommentHandler)

	tool = mcp.NewTool("get_task_comments",
		mcp.WithDescription("Get task comments"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get comments for"),
		),
	)
	s.AddTool(tool, kbClient.getTaskCommentsHandler)

	tool = mcp.NewTool("get_comment",
		mcp.WithDescription("Get comment information"),
		mcp.WithNumber("comment_id",
			mcp.Required(),
			mcp.Description("ID of the comment to get details for"),
		),
	)
	s.AddTool(tool, kbClient.getCommentHandler)

	tool = mcp.NewTool("update_comment",
		mcp.WithDescription("Update a comment"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("ID of the comment to update"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("New Markdown content for the comment"),
		),
	)
	s.AddTool(tool, kbClient.updateCommentHandler)

	tool = mcp.NewTool("remove_comment",
		mcp.WithDescription("Remove a comment"),
		mcp.WithNumber("comment_id",
			mcp.Required(),
			mcp.Description("ID of the comment to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeCommentHandler)

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

	tool = mcp.NewTool("get_me",
		mcp.WithDescription("Get logged user session"),
	)
	s.AddTool(tool, kbClient.getMeHandler)

	tool = mcp.NewTool("get_my_dashboard",
		mcp.WithDescription("Get the dashboard of the logged user without pagination"),
	)
	s.AddTool(tool, kbClient.getMyDashboardHandler)

	tool = mcp.NewTool("get_my_activity_stream",
		mcp.WithDescription("Get the last 100 events for the logged user"),
	)
	s.AddTool(tool, kbClient.getMyActivityStreamHandler)

	tool = mcp.NewTool("create_my_private_project",
		mcp.WithDescription("Create a private project for the logged user"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the private project to create"),
		),
		mcp.WithString("description",
			mcp.Description("Description of the private project (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createMyPrivateProjectHandler)

	tool = mcp.NewTool("get_my_projects_list",
		mcp.WithDescription("Get projects of the connected user"),
	)
	s.AddTool(tool, kbClient.getMyProjectsListHandler)

	tool = mcp.NewTool("get_my_overdue_tasks",
		mcp.WithDescription("Get my overdue tasks"),
	)
	s.AddTool(tool, kbClient.getMyOverdueTasksHandler)

	tool = mcp.NewTool("get_my_projects",
		mcp.WithDescription("Get projects of connected user with full details"),
	)
	s.AddTool(tool, kbClient.getMyProjectsHandler)

	tool = mcp.NewTool("get_external_task_link_types",
		mcp.WithDescription("Get all registered external link providers"),
	)
	s.AddTool(tool, kbClient.getExternalTaskLinkTypesHandler)

	tool = mcp.NewTool("get_ext_link_provider_deps",
		mcp.WithDescription("Get available dependencies for a given provider"),
		mcp.WithString("provider",
			mcp.Required(),
			mcp.Description("Provider name"),
		),
	)
	s.AddTool(tool, kbClient.getExternalTaskLinkProviderDependenciesHandler)

	tool = mcp.NewTool("create_external_task_link",
		mcp.WithDescription("Create a new external link"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task"),
		),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("URL of the external link"),
		),
		mcp.WithString("dependency",
			mcp.Required(),
			mcp.Description("Dependency of the external link"),
		),
		mcp.WithString("type",
			mcp.Description("Type of the external link (optional)"),
		),
		mcp.WithString("title",
			mcp.Description("Title of the external link (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createExternalTaskLinkHandler)

	tool = mcp.NewTool("update_external_task_link",
		mcp.WithDescription("Update external task link"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task"),
		),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the external link to update"),
		),
		mcp.WithString("title",
			mcp.Description("New title for the external link"),
		),
		mcp.WithString("url",
			mcp.Description("New URL for the external link")),
		mcp.WithString("dependency",
			mcp.Description("New dependency for the external link"),
		),
	)
	s.AddTool(tool, kbClient.updateExternalTaskLinkHandler)

	tool = mcp.NewTool("get_external_task_link_by_id",
		mcp.WithDescription("Get an external task link by ID"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task"),
		),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the external link to retrieve")),
	)
	s.AddTool(tool, kbClient.getExternalTaskLinkByIdHandler)

	tool = mcp.NewTool("get_all_external_task_links",
		mcp.WithDescription("Get all external links attached to a task"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get external links for"),
		),
	)
	s.AddTool(tool, kbClient.getAllExternalTaskLinksHandler)

	tool = mcp.NewTool("remove_external_task_link",
		mcp.WithDescription("Remove an external link"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task"),
		),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the external link to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeExternalTaskLinkHandler)

	tool = mcp.NewTool("get_columns",
		mcp.WithDescription("List project columns"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get columns from"),
		),
	)
	s.AddTool(tool, kbClient.getColumnsHandler)

	tool = mcp.NewTool("get_column",
		mcp.WithDescription("Get a single column"),
		mcp.WithNumber("column_id",
			mcp.Required(),
			mcp.Description("ID of the column to get details for"),
		),
	)
	s.AddTool(tool, kbClient.getColumnHandler)

	tool = mcp.NewTool("create_column",
		mcp.WithDescription("Add new columns"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to add the column to"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Title of the column to create"),
		),
		mcp.WithNumber("task_limit",
			mcp.Description("Task limit for the new column"),
		),
		mcp.WithString("description",
			mcp.Description("Description for the new column"),
		),
	)
	s.AddTool(tool, kbClient.createColumnHandler)

	tool = mcp.NewTool("update_column",
		mcp.WithDescription("Modify column settings"),
		mcp.WithNumber("column_id",
			mcp.Required(),
			mcp.Description("ID of the column to update"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("New title for the column"),
		),
		mcp.WithNumber("task_limit",
			mcp.Description("New task limit for the column"),
		),
		mcp.WithString("description",
			mcp.Description("New description for the column"),
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
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project containing the columns"),
		),
		mcp.WithNumber("column_id",
			mcp.Required(),
			mcp.Description("ID of the column to reorder"),
		),
		mcp.WithNumber("position",
			mcp.Required(),
			mcp.Description("New position for the column (must be >= 1)"),
		),
	)
	s.AddTool(tool, kbClient.reorderColumnsHandler)

	tool = mcp.NewTool("get_categories",
		mcp.WithDescription("List project categories"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get categories from"),
		),
	)
	s.AddTool(tool, kbClient.getCategoriesHandler)

	tool = mcp.NewTool("create_category",
		mcp.WithDescription("Add task categories"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the category to create"),
		),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to add the category to"),
		),
		mcp.WithString("color_id",
			mcp.Description("Color ID for the category (e.g., 'blue', 'green')"),
		),
	)
	s.AddTool(tool, kbClient.createCategoryHandler)

	tool = mcp.NewTool("get_category",
		mcp.WithDescription("Get category information"),
		mcp.WithNumber("category_id",
			mcp.Required(),
			mcp.Description("ID of the category to get details for"),
		),
	)
	s.AddTool(tool, kbClient.getCategoryHandler)

	tool = mcp.NewTool("update_category",
		mcp.WithDescription("Modify categories"),
		mcp.WithNumber("category_id",
			mcp.Required(),
			mcp.Description("ID of the category to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the category"),
		),
		mcp.WithString("color_id",
			mcp.Description("Color ID for the category (e.g., 'blue', 'green')"),
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
		mcp.WithDescription("List all swimlanes of a project (enabled or disabled) and sorted by position"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get swimlanes from"),
		),
	)
	s.AddTool(tool, kbClient.getAllSwimlanesHandler)

	tool = mcp.NewTool("get_active_swimlanes",
		mcp.WithDescription("Get the list of enabled swimlanes of a project (include default swimlane if enabled)"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get active swimlanes from"),
		),
	)
	s.AddTool(tool, kbClient.getActiveSwimlanesHandler)

	tool = mcp.NewTool("get_swimlane",
		mcp.WithDescription("Get a swimlane by ID"),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getSwimlaneHandler)

	tool = mcp.NewTool("get_swimlane_by_id",
		mcp.WithDescription("Get a swimlane by ID"),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getSwimlaneByIdHandler)

	tool = mcp.NewTool("get_swimlane_by_name",
		mcp.WithDescription("Get a swimlane by name"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project the swimlane belongs to"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the swimlane to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getSwimlaneByNameHandler)

	tool = mcp.NewTool("change_swimlane_position",
		mcp.WithDescription("Move a swimlane's position"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project containing the swimlane"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to reorder"),
		),
		mcp.WithNumber("position",
			mcp.Required(),
			mcp.Description("New position for the swimlane (must be >= 1)"),
		),
	)
	s.AddTool(tool, kbClient.changeSwimlanePositionHandler)

	tool = mcp.NewTool("create_swimlane",
		mcp.WithDescription("Add a new swimlane"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to add the swimlane to"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the swimlane to create"),
		),
		mcp.WithString("description",
			mcp.Description("Description of the swimlane (optional)"),
		),
	)
	s.AddTool(tool, kbClient.addSwimlaneHandler)

	tool = mcp.NewTool("update_swimlane",
		mcp.WithDescription("Update swimlane properties"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project the swimlane belongs to"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the swimlane (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New description for the swimlane (optional)"),
		),
	)
	s.AddTool(tool, kbClient.updateSwimlaneHandler)

	tool = mcp.NewTool("remove_swimlane",
		mcp.WithDescription("Remove a swimlane"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project the swimlane belongs to"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeSwimlaneHandler)

	tool = mcp.NewTool("disable_swimlane",
		mcp.WithDescription("Disable a swimlane"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project the swimlane belongs to"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to disable"),
		),
	)
	s.AddTool(tool, kbClient.disableSwimlaneHandler)

	tool = mcp.NewTool("enable_swimlane",
		mcp.WithDescription("Enable a swimlane"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project the swimlane belongs to"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Required(),
			mcp.Description("ID of the swimlane to enable"),
		),
	)
	s.AddTool(tool, kbClient.enableSwimlaneHandler)

	tool = mcp.NewTool("get_board",
		mcp.WithDescription("Get all necessary information to display a board"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get board details for"),
		),
	)
	s.AddTool(tool, kbClient.getBoardHandler)

	// Task Metadata Management
	tool = mcp.NewTool("get_task_metadata",
		mcp.WithDescription("Get all metadata related to a task by task unique id"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get metadata from"),
		),
	)
	s.AddTool(tool, kbClient.getTaskMetadataHandler)

	tool = mcp.NewTool("get_task_metadata_by_name",
		mcp.WithDescription("Get metadata related to a task by task unique id and metakey (name)"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get metadata from"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the metadata key"),
		),
	)
	s.AddTool(tool, kbClient.getTaskMetadataByNameHandler)

	tool = mcp.NewTool("save_task_metadata",
		mcp.WithDescription("Save/update task metadata"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to save/update metadata for"),
		),
		mcp.WithObject("values",
			mcp.Required(),
			mcp.Description("Dictionary of metadata values (key-value pairs)"),
		),
	)
	s.AddTool(tool, kbClient.saveTaskMetadataHandler)

	tool = mcp.NewTool("remove_task_metadata",
		mcp.WithDescription("Remove task metadata by name"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to remove metadata from"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the metadata key to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeTaskMetadataHandler)

	tool = mcp.NewTool("create_group",
		mcp.WithDescription("Create a new group"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the group to create"),
		),
		mcp.WithString("external_id",
			mcp.Description("External ID for the group (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createGroupHandler)

	tool = mcp.NewTool("update_group",
		mcp.WithDescription("Update a group"),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the group (optional)"),
		),
		mcp.WithString("external_id",
			mcp.Description("New external ID for the group (optional)"),
		),
	)
	s.AddTool(tool, kbClient.updateGroupHandler)

	tool = mcp.NewTool("remove_group",
		mcp.WithDescription("Remove a group"),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeGroupHandler)

	tool = mcp.NewTool("get_group",
		mcp.WithDescription("Get one group"),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getGroupHandler)

	tool = mcp.NewTool("get_all_groups",
		mcp.WithDescription("Get all groups"),
	)
	s.AddTool(tool, kbClient.getAllGroupsHandler)

	tool = mcp.NewTool("get_member_groups",
		mcp.WithDescription("Get all groups for a given user"),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user"),
		),
	)
	s.AddTool(tool, kbClient.getMemberGroupsHandler)

	tool = mcp.NewTool("get_group_members",
		mcp.WithDescription("Get all members of a group"),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group"),
		),
	)
	s.AddTool(tool, kbClient.getGroupMembersHandler)

	tool = mcp.NewTool("add_group_member",
		mcp.WithDescription("Add a user to a group"),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to add"),
		),
	)
	s.AddTool(tool, kbClient.addGroupMemberHandler)

	tool = mcp.NewTool("remove_group_member",
		mcp.WithDescription("Remove a user from a group"),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group")),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeGroupMemberHandler)

	tool = mcp.NewTool("is_group_member",
		mcp.WithDescription("Check if a user is member of a group"),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user"),
		),
	)
	s.AddTool(tool, kbClient.isGroupMemberHandler)

	tool = mcp.NewTool("create_task_link",
		mcp.WithDescription("Create a link between two tasks"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the first task"),
		),
		mcp.WithNumber("opposite_task_id",
			mcp.Required(),
			mcp.Description("ID of the opposite task"),
		),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the link type"),
		),
	)
	s.AddTool(tool, kbClient.createTaskLinkHandler)

	tool = mcp.NewTool("update_task_link",
		mcp.WithDescription("Update task link"),
		mcp.WithNumber("task_link_id",
			mcp.Required(),
			mcp.Description("ID of the task link to update"),
		),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the first task"),
		),
		mcp.WithNumber("opposite_task_id",
			mcp.Required(),
			mcp.Description("ID of the opposite task"),
		),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the link type"),
		),
	)
	s.AddTool(tool, kbClient.updateTaskLinkHandler)

	tool = mcp.NewTool("get_task_link_by_id",
		mcp.WithDescription("Get a task link by ID"),
		mcp.WithNumber("task_link_id",
			mcp.Required(),
			mcp.Description("ID of the task link to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getTaskLinkByIdHandler)

	tool = mcp.NewTool("get_all_task_links",
		mcp.WithDescription("Get all links related to a task"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get links for"),
		),
	)
	s.AddTool(tool, kbClient.getAllTaskLinksHandler)

	tool = mcp.NewTool("remove_task_link",
		mcp.WithDescription("Remove a link between two tasks"),
		mcp.WithNumber("task_link_id",
			mcp.Required(),
			mcp.Description("ID of the task link to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeTaskLinkHandler)

	// Link Management
	tool = mcp.NewTool("get_all_links",
		mcp.WithDescription("Get the list of possible relations between tasks"),
	)
	s.AddTool(tool, kbClient.getAllLinksHandler)

	tool = mcp.NewTool("get_opposite_link_id",
		mcp.WithDescription("Get the opposite link id of a task link"),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the link to get the opposite ID for"),
		),
	)
	s.AddTool(tool, kbClient.getOppositeLinkIdHandler)

	tool = mcp.NewTool("get_link_by_label",
		mcp.WithDescription("Get a link by label"),
		mcp.WithString("label",
			mcp.Required(),
			mcp.Description("Label of the link to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getLinkByLabelHandler)

	tool = mcp.NewTool("get_link_by_id",
		mcp.WithDescription("Get a link by id"),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the link to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getLinkByIdHandler)

	tool = mcp.NewTool("create_link",
		mcp.WithDescription("Create a new task relation"),
		mcp.WithString("label",
			mcp.Required(),
			mcp.Description("Label of the new link"),
		),
		mcp.WithString("opposite_label",
			mcp.Description("Label of the opposite link (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createLinkHandler)

	tool = mcp.NewTool("update_link",
		mcp.WithDescription("Update a link"),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the link to update"),
		),
		mcp.WithNumber("opposite_link_id",
			mcp.Required(),
			mcp.Description("ID of the opposite link"),
		),
		mcp.WithString("label",
			mcp.Required(),
			mcp.Description("New label for the link"),
		),
	)
	s.AddTool(tool, kbClient.updateLinkHandler)

	tool = mcp.NewTool("remove_link",
		mcp.WithDescription("Remove a link"),
		mcp.WithNumber("link_id",
			mcp.Required(),
			mcp.Description("ID of the link to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeLinkHandler)

	// Project Management

	tool = mcp.NewTool("get_project_by_id",
		mcp.WithDescription("Get project information by ID"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getProjectByIdHandler)

	tool = mcp.NewTool("get_project_by_name",
		mcp.WithDescription("Get project information by name"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the project to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getProjectByNameHandler)

	tool = mcp.NewTool("get_project_by_identifier",
		mcp.WithDescription("Get project information by identifier"),
		mcp.WithString("identifier",
			mcp.Required(),
			mcp.Description("Identifier of the project to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getProjectByIdentifierHandler)

	tool = mcp.NewTool("get_project_by_email",
		mcp.WithDescription("Get project information by email"),
		mcp.WithString("email",
			mcp.Required(),
			mcp.Description("Email of the project to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getProjectByEmailHandler)

	tool = mcp.NewTool("get_all_projects",
		mcp.WithDescription("Get all available projects"),
	)
	s.AddTool(tool, kbClient.getAllProjectsHandler)

	tool = mcp.NewTool("update_project",
		mcp.WithDescription("Update a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to update"),
		),
		mcp.WithString("name",
			mcp.Description("New name for the project (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New description for the project (optional)"),
		),
		mcp.WithNumber("owner_id",
			mcp.Description("New owner ID for the project (optional)"),
		),
		mcp.WithString("identifier",
			mcp.Description("New alphanumeric identifier for the project (optional)"),
		),
		mcp.WithString("start_date",
			mcp.Description("New start date in ISO8601 format (optional)"),
		),
		mcp.WithString("end_date",
			mcp.Description("New end date in ISO8601 format (optional)"),
		),
		mcp.WithNumber("priority_default",
			mcp.Description("New default task priority (optional)"),
		),
		mcp.WithNumber("priority_start",
			mcp.Description("New start priority (optional)"),
		),
		mcp.WithNumber("priority_end",
			mcp.Description("New end priority (optional)"),
		),
		mcp.WithString("email",
			mcp.Description("New project email address (optional)"),
		),
	)
	s.AddTool(tool, kbClient.updateProjectHandler)

	tool = mcp.NewTool("remove_project",
		mcp.WithDescription("Remove a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeProjectHandler)

	tool = mcp.NewTool("enable_project",
		mcp.WithDescription("Enable a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to enable"),
		),
	)
	s.AddTool(tool, kbClient.enableProjectHandler)

	tool = mcp.NewTool("disable_project",
		mcp.WithDescription("Disable a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to disable"),
		),
	)
	s.AddTool(tool, kbClient.disableProjectHandler)

	tool = mcp.NewTool("enable_project_public_access",
		mcp.WithDescription("Enable public access for a given project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to enable public access for"),
		),
	)
	s.AddTool(tool, kbClient.enableProjectPublicAccessHandler)

	tool = mcp.NewTool("disable_project_public_access",
		mcp.WithDescription("Disable public access for a given project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to disable public access for"),
		),
	)
	s.AddTool(tool, kbClient.disableProjectPublicAccessHandler)

	tool = mcp.NewTool("get_project_activity",
		mcp.WithDescription("Get activity stream for a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get activity for"),
		),
	)
	s.AddTool(tool, kbClient.getProjectActivityHandler)

	tool = mcp.NewTool("get_project_activities",
		mcp.WithDescription("Get Activityfeed for Project(s)"),
		mcp.WithArray("project_ids",
			mcp.Required(),
			mcp.WithNumberItems(),
			mcp.Description("Array of project IDs to get activities for"),
		),
	)
	s.AddTool(tool, kbClient.getProjectActivitiesHandler)

	// Project File Management
	tool = mcp.NewTool("create_project_file",
		mcp.WithDescription("Create and upload a new project attachment"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to attach the file to"),
		),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Name of the file"),
		),
		mcp.WithString("blob",
			mcp.Required(),
			mcp.Description("File content encoded in base64"),
		),
	)
	s.AddTool(tool, kbClient.createProjectFileHandler)

	tool = mcp.NewTool("get_all_project_files",
		mcp.WithDescription("Get all files attached to a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get files from"),
		),
	)
	s.AddTool(tool, kbClient.getAllProjectFilesHandler)

	tool = mcp.NewTool("get_project_file",
		mcp.WithDescription("Get file information"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("file_id",
			mcp.Required(),
			mcp.Description("ID of the file to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getProjectFileHandler)

	tool = mcp.NewTool("download_project_file",
		mcp.WithDescription("Download project file contents (encoded in base64)"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("file_id",
			mcp.Required(),
			mcp.Description("ID of the file to download"),
		),
	)
	s.AddTool(tool, kbClient.downloadProjectFileHandler)

	tool = mcp.NewTool("remove_project_file",
		mcp.WithDescription("Remove a file associated to a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("file_id",
			mcp.Required(),
			mcp.Description("ID of the file to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeProjectFileHandler)

	tool = mcp.NewTool("remove_all_project_files",
		mcp.WithDescription("Remove all files associated to a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to remove all files from"),
		),
	)
	s.AddTool(tool, kbClient.removeAllProjectFilesHandler)

	// Project Metadata Management
	tool = mcp.NewTool("get_project_metadata",
		mcp.WithDescription("Get Project metadata"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get metadata from"),
		),
	)
	s.AddTool(tool, kbClient.getProjectMetadataHandler)

	tool = mcp.NewTool("get_project_metadata_by_name",
		mcp.WithDescription("Fetch single metadata value"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the metadata key"),
		),
	)
	s.AddTool(tool, kbClient.getProjectMetadataByNameHandler)

	tool = mcp.NewTool("save_project_metadata",
		mcp.WithDescription("Add or update metadata"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithObject("values",
			mcp.Required(),
			mcp.Description("Dictionary of metadata values (key-value pairs)"),
		),
	)
	s.AddTool(tool, kbClient.saveProjectMetadataHandler)

	tool = mcp.NewTool("remove_project_metadata",
		mcp.WithDescription("Remove a project metadata"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the metadata key to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeProjectMetadataHandler)

	// Project Permission Management
	tool = mcp.NewTool("get_project_users",
		mcp.WithDescription("Get all members of a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get users from"),
		),
	)
	s.AddTool(tool, kbClient.getProjectUsersHandler)

	tool = mcp.NewTool("get_assignable_users",
		mcp.WithDescription("Get users that can be assigned to a task for a project (all members except viewers)"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithBoolean("prepend_unassigned",
			mcp.Description("Prepend the 'Unassigned' option (optional, default is false)"),
		),
	)
	s.AddTool(tool, kbClient.getAssignableUsersHandler)

	tool = mcp.NewTool("add_project_user",
		mcp.WithDescription("Grant access to a project for a user"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user"),
		),
		mcp.WithString("role",
			mcp.Description("Role to assign (optional)"),
		),
	)
	s.AddTool(tool, kbClient.addProjectUserHandler)

	tool = mcp.NewTool("add_project_group",
		mcp.WithDescription("Grant access to a project for a group"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group"),
		),
		mcp.WithString("role",
			mcp.Description("Role to assign (optional)"),
		),
	)
	s.AddTool(tool, kbClient.addProjectGroupHandler)

	tool = mcp.NewTool("remove_project_user",
		mcp.WithDescription("Revoke user access to a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user"),
		),
	)
	s.AddTool(tool, kbClient.removeProjectUserHandler)

	tool = mcp.NewTool("remove_project_group",
		mcp.WithDescription("Revoke group access to a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group"),
		),
	)
	s.AddTool(tool, kbClient.removeProjectGroupHandler)

	tool = mcp.NewTool("change_project_user_role",
		mcp.WithDescription("Change role of a user for a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user"),
		),
		mcp.WithString("role",
			mcp.Required(),
			mcp.Description("New role to assign"),
		),
	)
	s.AddTool(tool, kbClient.changeProjectUserRoleHandler)

	tool = mcp.NewTool("change_project_group_role",
		mcp.WithDescription("Change role of a group for a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("group_id",
			mcp.Required(),
			mcp.Description("ID of the group"),
		),
		mcp.WithString("role",
			mcp.Required(),
			mcp.Description("New role to assign"),
		),
	)
	s.AddTool(tool, kbClient.changeProjectGroupRoleHandler)

	tool = mcp.NewTool("get_project_user_role",
		mcp.WithDescription("Get the role of a user for a given project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("user_id",
			mcp.Required(),
			mcp.Description("ID of the user"),
		),
	)
	s.AddTool(tool, kbClient.getProjectUserRoleHandler)

	// Subtask Management
	tool = mcp.NewTool("create_subtask",
		mcp.WithDescription("Create a new subtask"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to associate the subtask with"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Title of the subtask"),
		),
		mcp.WithNumber("user_id",
			mcp.Description("ID of the user assigned to the subtask (optional)"),
		),
		mcp.WithNumber("time_estimated",
			mcp.Description("Estimated time for the subtask in hours (optional)"),
		),
		mcp.WithNumber("time_spent",
			mcp.Description("Time spent on the subtask in hours (optional)"),
		),
		mcp.WithNumber("status",
			mcp.Description("Status of the subtask (0: Todo, 1: In Progress, 2: Done) (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createSubtaskHandler)

	tool = mcp.NewTool("get_subtask",
		mcp.WithDescription("Get subtask information"),
		mcp.WithNumber("subtask_id",
			mcp.Required(),
			mcp.Description("ID of the subtask to retrieve"),
		),
	)
	s.AddTool(tool, kbClient.getSubtaskHandler)

	tool = mcp.NewTool("get_all_subtasks",
		mcp.WithDescription("Get all available subtasks for a task"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to get subtasks for"),
		),
	)
	s.AddTool(tool, kbClient.getAllSubtasksHandler)

	tool = mcp.NewTool("update_subtask",
		mcp.WithDescription("Update a subtask"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("ID of the subtask to update"),
		),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the associated task"),
		),
		mcp.WithString("title",
			mcp.Description("New title for the subtask (optional)"),
		),
		mcp.WithNumber("user_id",
			mcp.Description("New user ID assigned to the subtask (optional)"),
		),
		mcp.WithNumber("time_estimated",
			mcp.Description("New estimated time for the subtask in hours (optional)"),
		),
		mcp.WithNumber("time_spent",
			mcp.Description("New time spent on the subtask in hours (optional)"),
		),
		mcp.WithNumber("status",
			mcp.Description("New status of the subtask (0: Todo, 1: In Progress, 2: Done) (optional)"),
		),
	)
	s.AddTool(tool, kbClient.updateSubtaskHandler)

	tool = mcp.NewTool("remove_subtask",
		mcp.WithDescription("Remove a subtask"),
		mcp.WithNumber("subtask_id",
			mcp.Required(),
			mcp.Description("ID of the subtask to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeSubtaskHandler)

	// Subtask Time Tracking
	tool = mcp.NewTool("has_subtask_timer",
		mcp.WithDescription("Check if a timer is started for the given subtask and user"),
		mcp.WithNumber("subtask_id",
			mcp.Required(),
			mcp.Description("ID of the subtask"),
		),
		mcp.WithNumber("user_id",
			mcp.Description("ID of the user (optional)"),
		),
	)
	s.AddTool(tool, kbClient.hasSubtaskTimerHandler)

	tool = mcp.NewTool("set_subtask_start_time",
		mcp.WithDescription("Start subtask timer for a user"),
		mcp.WithNumber("subtask_id",
			mcp.Required(),
			mcp.Description("ID of the subtask"),
		),
		mcp.WithNumber("user_id",
			mcp.Description("ID of the user (optional)"),
		),
	)
	s.AddTool(tool, kbClient.setSubtaskStartTimeHandler)

	tool = mcp.NewTool("set_subtask_end_time",
		mcp.WithDescription("Stop subtask timer for a user"),
		mcp.WithNumber("subtask_id",
			mcp.Required(),
			mcp.Description("ID of the subtask"),
		),
		mcp.WithNumber("user_id",
			mcp.Description("ID of the user (optional)"),
		),
	)
	s.AddTool(tool, kbClient.setSubtaskEndTimeHandler)

	tool = mcp.NewTool("get_subtask_time_spent",
		mcp.WithDescription("Get time spent on a subtask for a user"),
		mcp.WithNumber("subtask_id",
			mcp.Required(),
			mcp.Description("ID of the subtask"),
		),
		mcp.WithNumber("user_id",
			mcp.Description("ID of the user (optional)"),
		),
	)
	s.AddTool(tool, kbClient.getSubtaskTimeSpentHandler)

	// Tag Management
	tool = mcp.NewTool("get_all_tags",
		mcp.WithDescription("Get all tags"),
	)
	s.AddTool(tool, kbClient.getAllTagsHandler)

	tool = mcp.NewTool("get_tags_by_project",
		mcp.WithDescription("Get all tags for a given project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get tags for"),
		),
	)
	s.AddTool(tool, kbClient.getTagsByProjectHandler)

	tool = mcp.NewTool("create_tag",
		mcp.WithDescription("Create a new tag"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to associate the tag with"),
		),
		mcp.WithString("tag",
			mcp.Required(),
			mcp.Description("Name of the tag"),
		),
		mcp.WithNumber("color_id",
			mcp.Description("ID of the color for the tag (optional)"),
		),
	)
	s.AddTool(tool, kbClient.createTagHandler)

	tool = mcp.NewTool("update_tag",
		mcp.WithDescription("Rename a tag"),
		mcp.WithNumber("tag_id",
			mcp.Required(),
			mcp.Description("ID of the tag to update"),
		),
		mcp.WithString("tag",
			mcp.Required(),
			mcp.Description("New name for the tag"),
		),
		mcp.WithNumber("color_id",
			mcp.Description("New color ID for the tag (optional)"),
		),
	)
	s.AddTool(tool, kbClient.updateTagHandler)

	tool = mcp.NewTool("remove_tag",
		mcp.WithDescription("Remove a tag"),
		mcp.WithNumber("tag_id",
			mcp.Required(),
			mcp.Description("ID of the tag to remove"),
		),
	)
	s.AddTool(tool, kbClient.removeTagHandler)

	tool = mcp.NewTool("set_task_tags",
		mcp.WithDescription("Assign/Create/Update tags for a task"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task"),
		),
		mcp.WithArray("tags",
			mcp.Required(),
			mcp.WithStringItems(),
			mcp.Description("List of tags (array of strings)"),
		),
	)
	s.AddTool(tool, kbClient.setTaskTagsHandler)

	tool = mcp.NewTool("get_task_tags",
		mcp.WithDescription("Get assigned tags to a task"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task"),
		),
	)
	s.AddTool(tool, kbClient.getTaskTagsHandler)

	tool = mcp.NewTool("create_task_file",
		mcp.WithDescription("Create and upload a new task attachment"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("The project ID"),
		),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("The task ID"),
		),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("The filename"),
		),
		mcp.WithString("blob",
			mcp.Required(),
			mcp.Description("File content encoded in base64"),
		),
	)
	s.AddTool(tool, kbClient.createTaskFileHandler)

	tool = mcp.NewTool("get_all_task_files",
		mcp.WithDescription("Get all files attached to task"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("The task ID"),
		),
	)
	s.AddTool(tool, kbClient.getAllTaskFilesHandler)

	tool = mcp.NewTool("get_task_file",
		mcp.WithDescription("Get file information"),
		mcp.WithNumber("file_id",
			mcp.Required(),
			mcp.Description("The file ID"),
		),
	)
	s.AddTool(tool, kbClient.getTaskFileHandler)

	tool = mcp.NewTool("download_task_file",
		mcp.WithDescription("Download file contents (encoded in base64)"),
		mcp.WithNumber("file_id",
			mcp.Required(),
			mcp.Description("The file ID"),
		),
	)
	s.AddTool(tool, kbClient.downloadTaskFileHandler)

	tool = mcp.NewTool("remove_task_file",
		mcp.WithDescription("Remove file"),
		mcp.WithNumber("file_id",
			mcp.Required(),
			mcp.Description("The file ID"),
		),
	)
	s.AddTool(tool, kbClient.removeTaskFileHandler)

	tool = mcp.NewTool("remove_all_task_files",
		mcp.WithDescription("Remove all files associated to a task"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("The task ID"),
		),
	)
	s.AddTool(tool, kbClient.removeAllTaskFilesHandler)

	// Application API Procedures
	tool = mcp.NewTool("get_version",
		mcp.WithDescription("Get the application version"),
	)
	s.AddTool(tool, kbClient.getVersionHandler)

	tool = mcp.NewTool("get_timezone",
		mcp.WithDescription("Get the timezone of the connected user"),
	)
	s.AddTool(tool, kbClient.getTimezoneHandler)

	tool = mcp.NewTool("get_default_task_colors",
		mcp.WithDescription("Get all default task colors"),
	)
	s.AddTool(tool, kbClient.getDefaultTaskColorsHandler)

	tool = mcp.NewTool("get_default_task_color",
		mcp.WithDescription("Get default task color"),
	)
	s.AddTool(tool, kbClient.getDefaultTaskColorHandler)

	tool = mcp.NewTool("get_color_list",
		mcp.WithDescription("Get the list of task colors"),
	)
	s.AddTool(tool, kbClient.getColorListHandler)

	tool = mcp.NewTool("get_application_roles",
		mcp.WithDescription("Get the application roles"),
	)
	s.AddTool(tool, kbClient.getApplicationRolesHandler)

	tool = mcp.NewTool("get_project_roles",
		mcp.WithDescription("Get the project roles"),
	)
	s.AddTool(tool, kbClient.getProjectRolesHandler)

	// Automatic Actions API Procedures
	tool = mcp.NewTool("get_available_actions",
		mcp.WithDescription("Get list of available automatic actions"),
	)
	s.AddTool(tool, kbClient.getAvailableActionsHandler)

	tool = mcp.NewTool("get_available_action_events",
		mcp.WithDescription("Get list of available events for actions"),
	)
	s.AddTool(tool, kbClient.getAvailableActionEventsHandler)

	tool = mcp.NewTool("get_compatible_action_events",
		mcp.WithDescription("Get list of events compatible with an action"),
		mcp.WithString("action_name",
			mcp.Required(),
			mcp.Description("Action name"),
		),
	)
	s.AddTool(tool, kbClient.getCompatibleActionEventsHandler)

	tool = mcp.NewTool("get_actions",
		mcp.WithDescription("Get list of actions for a project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("Project ID"),
		),
	)
	s.AddTool(tool, kbClient.getActionsHandler)

	tool = mcp.NewTool("create_action",
		mcp.WithDescription("Create an action"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("Project ID"),
		),
		mcp.WithString("event_name",
			mcp.Required(),
			mcp.Description("Event name"),
		),
		mcp.WithString("action_name",
			mcp.Required(),
			mcp.Description("Action name"),
		),
		mcp.WithObject("params",
			mcp.Required(),
			mcp.Description("Key/value parameters"),
		),
	)
	s.AddTool(tool, kbClient.createActionHandler)

	tool = mcp.NewTool("remove_action",
		mcp.WithDescription("Remove an action"),
		mcp.WithNumber("action_id",
			mcp.Required(),
			mcp.Description("Action ID"),
		),
	)
	s.AddTool(tool, kbClient.removeActionHandler)

	tool = mcp.NewTool("get_task_by_reference",
		mcp.WithDescription("Get task by the external reference"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
		mcp.WithString("reference",
			mcp.Required(),
			mcp.Description("External reference for the task"),
		),
	)
	s.AddTool(tool, kbClient.getTaskByReferenceHandler)

	tool = mcp.NewTool("get_all_tasks",
		mcp.WithDescription("Get all available tasks"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get tasks from"),
		),
		mcp.WithNumber("status_id",
			mcp.Required(),
			mcp.Description("The value 1 for active tasks and 0 for inactive"),
		),
	)
	s.AddTool(tool, kbClient.getAllTasksHandler)

	tool = mcp.NewTool("get_overdue_tasks",
		mcp.WithDescription("Get all overdue tasks"),
	)
	s.AddTool(tool, kbClient.getOverdueTasksHandler)

	tool = mcp.NewTool("get_overdue_tasks_by_project",
		mcp.WithDescription("Get all overdue tasks for a special project"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project"),
		),
	)
	s.AddTool(tool, kbClient.getOverdueTasksByProjectHandler)

	tool = mcp.NewTool("open_task",
		mcp.WithDescription("Set a task to the status open"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to open"),
		),
	)
	s.AddTool(tool, kbClient.openTaskHandler)

	tool = mcp.NewTool("close_task",
		mcp.WithDescription("Set a task to the status close"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to close"),
		),
	)
	s.AddTool(tool, kbClient.closeTaskHandler)

	tool = mcp.NewTool("move_task_to_project",
		mcp.WithDescription("Move a task to another project"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to move"),
		),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to move the task to"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Description("ID of the swimlane (optional)"),
		),
		mcp.WithNumber("column_id",
			mcp.Description("ID of the column (optional)"),
		),
		mcp.WithNumber("category_id",
			mcp.Description("ID of the category (optional)"),
		),
		mcp.WithNumber("owner_id",
			mcp.Description("ID of the owner (optional)"),
		),
	)
	s.AddTool(tool, kbClient.moveTaskToProjectHandler)

	tool = mcp.NewTool("duplicate_task_to_project",
		mcp.WithDescription("Duplicate a task to another project"),
		mcp.WithNumber("task_id",
			mcp.Required(),
			mcp.Description("ID of the task to duplicate"),
		),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to duplicate the task to"),
		),
		mcp.WithNumber("swimlane_id",
			mcp.Description("ID of the swimlane (optional)"),
		),
		mcp.WithNumber("column_id",
			mcp.Description("ID of the column (optional)"),
		),
		mcp.WithNumber("category_id",
			mcp.Description("ID of the category (optional)"),
		),
		mcp.WithNumber("owner_id",
			mcp.Description("ID of the owner (optional)"),
		),
	)
	s.AddTool(tool, kbClient.duplicateTaskToProjectHandler)

	tool = mcp.NewTool("search_tasks",
		mcp.WithDescription("Find tasks by using the search engine"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to search tasks in"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query string"),
		),
	)
	s.AddTool(tool, kbClient.searchTasksHandler)

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
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", kc.apiEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
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
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d, Response: %s", resp.StatusCode, resp.Status)
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
		return nil, fmt.Errorf("failed to decode API response: %v", err)
	}

	if apiResponse.Error != nil {
		return nil, fmt.Errorf("kanboard API error: %s (Code: %d)", apiResponse.Error.Message, apiResponse.Error.Code)
	}

	return apiResponse.Result, nil
}

func (kc *kanboardClient) getProjectsHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	params := map[string]interface{}{"name": name}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}

	ownerId := request.GetInt("owner_id", 0)
	if ownerId != 0 {
		params["owner_id"] = ownerId
	}

	identifier := request.GetString("identifier", "")
	if identifier != "" {
		params["identifier"] = identifier
	}

	startDate := request.GetString("start_date", "")
	if startDate != "" {
		params["start_date"] = startDate
	}

	endDate := request.GetString("end_date", "")
	if endDate != "" {
		params["end_date"] = endDate
	}

	priorityDefault := request.GetInt("priority_default", 0)
	if priorityDefault != 0 {
		params["priority_default"] = priorityDefault
	}

	priorityStart := request.GetInt("priority_start", 0)
	if priorityStart != 0 {
		params["priority_start"] = priorityStart
	}

	priorityEnd := request.GetInt("priority_end", 0)
	if priorityEnd != 0 {
		params["priority_end"] = priorityEnd
	}

	email := request.GetString("email", "")
	if email != "" {
		params["email"] = email
	}

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
	projectName, err := request.RequireString("project_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// First, get the project ID from the project name
	result, err := kc.callKanboardAPI(ctx, "getProjectByName", map[string]string{"name": projectName})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project ID: %v", err)), nil
	}

	// Kanboard API sometimes returns a boolean false or true, or an empty array []
	// instead of a project object when the project is not found.
	// We need to handle these cases before attempting to unmarshal into a struct.
	projectMap, ok := result.(map[string]interface{})
	if !ok {
		// If it's not a map, check if it's a boolean (false or true).
		// Kanboard API might return `false` for not found, or `true` if it's a "success" response without data.
		if b, isBool := result.(bool); isBool {
			// If boolean false, it's definitively not found.
			// If boolean true, and it's not a map, it's also not a valid project object.
			return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found or API returned an unexpected boolean value: %v", projectName, b)), nil
		}

		// If it's not a map and not a boolean, check if it's nil or an empty array.
		if result == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned nil", projectName)), nil
		}
		if arr, isArray := result.([]interface{}); isArray && len(arr) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned empty array", projectName)), nil
		}

		// For any other unexpected type
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned unexpected type %T", projectName, result)), nil
	}

	// If projectMap is empty, treat as not found. (e.g., Kanboard returns {} for not found)
	if len(projectMap) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned empty object", projectName)), nil
	}

	var projectInfo struct {
		ID int `json:"id"` // Change type to int
	}
	// Marshal and unmarshal from the confirmed map to ensure correct type conversion
	tempBytes, err := json.Marshal(projectMap) // Marshal the map, not the raw interface{} result
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal project info for parsing: %v", err)), nil
	}
	if err := json.Unmarshal(tempBytes, &projectInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse project info: %v", err)), nil
	}

	if projectInfo.ID == 0 { // Check for 0 instead of empty string
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found or ID is zero", projectName)), nil
	}

	params := map[string]interface{}{"project_id": strconv.Itoa(projectInfo.ID)} // Convert int to string here
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
	projectName, err := request.RequireString("project_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// First, get the project ID from the project name
	result, err := kc.callKanboardAPI(ctx, "getProjectByName", map[string]string{"name": projectName})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project ID: %v", err)), nil
	}

	// Kanboard API sometimes returns a boolean false or true, or an empty array []
	// instead of a project object when the project is not found.
	// We need to handle these cases before attempting to unmarshal into a struct.
	projectMap, ok := result.(map[string]interface{})
	if !ok {
		// If it's not a map, check if it's a boolean (false or true).
		// Kanboard API might return `false` for not found, or `true` if it's a "success" response without data.
		if b, isBool := result.(bool); isBool {
			// If boolean false, it's definitively not found.
			// If boolean true, and it's not a map, it's also not a valid project object.
			return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found or API returned an unexpected boolean value: %v", projectName, b)), nil
		}

		// If it's not a map and not a boolean, check if it's nil or an empty array.
		if result == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned nil", projectName)), nil
		}
		if arr, isArray := result.([]interface{}); isArray && len(arr) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned empty array", projectName)), nil
		}

		// For any other unexpected type
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned unexpected type %T", projectName, result)), nil
	}

	// If projectMap is empty, treat as not found. (e.g., Kanboard returns {} for not found)
	if len(projectMap) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found: API returned empty object", projectName)), nil
	}

	var projectInfo struct {
		ID int `json:"id"` // Change type to int
	}
	tempBytes, err := json.Marshal(projectMap)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal project info for parsing: %v", err)), nil
	}
	if err := json.Unmarshal(tempBytes, &projectInfo); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse project info: %v", err)), nil
	}

	if projectInfo.ID == 0 { // Check for 0 instead of empty string
		return mcp.NewToolResultError(fmt.Sprintf("Project '%s' not found or ID is zero", projectName)), nil
	}

	params := map[string]interface{}{
		"project_id": strconv.Itoa(projectInfo.ID), // Convert int to string here
		"title":      title,
	}

	colorId := request.GetString("color_id", "")
	if colorId != "" {
		params["color_id"] = colorId
	}

	columnId := request.GetInt("column_id", 0)
	if columnId != 0 {
		params["column_id"] = columnId
	}

	ownerId := request.GetInt("owner_id", 0)
	if ownerId != 0 {
		params["owner_id"] = ownerId
	}

	creatorId := request.GetInt("creator_id", 0)
	if creatorId != 0 {
		params["creator_id"] = creatorId
	}

	dateDue := request.GetString("date_due", "")
	if dateDue != "" {
		params["date_due"] = dateDue
	}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}

	categoryId := request.GetInt("category_id", 0)
	if categoryId != 0 {
		params["category_id"] = categoryId
	}

	score := request.GetInt("score", 0)
	if score != 0 {
		params["score"] = score
	}

	swimlaneId := request.GetInt("swimlane_id", 0)
	if swimlaneId != 0 {
		params["swimlane_id"] = swimlaneId
	}

	priority := request.GetInt("priority", 0)
	if priority != 0 {
		params["priority"] = priority
	}

	recurrenceStatus := request.GetInt("recurrence_status", 0)
	if recurrenceStatus != 0 {
		params["recurrence_status"] = recurrenceStatus
	}

	recurrenceTrigger := request.GetInt("recurrence_trigger", 0)
	if recurrenceTrigger != 0 {
		params["recurrence_trigger"] = recurrenceTrigger
	}

	recurrenceFactor := request.GetInt("recurrence_factor", 0)
	if recurrenceFactor != 0 {
		params["recurrence_factor"] = recurrenceFactor
	}

	recurrenceTimeframe := request.GetInt("recurrence_timeframe", 0)
	if recurrenceTimeframe != 0 {
		params["recurrence_timeframe"] = recurrenceTimeframe
	}

	recurrenceBasedate := request.GetInt("recurrence_basedate", 0)
	if recurrenceBasedate != 0 {
		params["recurrence_basedate"] = recurrenceBasedate
	}

	reference := request.GetString("reference", "")
	if reference != "" {
		params["reference"] = reference
	}

	tags := request.GetStringSlice("tags", []string{})
	if len(tags) > 0 {
		params["tags"] = tags
	}

	dateStarted := request.GetString("date_started", "")
	if dateStarted != "" {
		params["date_started"] = dateStarted
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
	id, err := request.RequireInt("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": id}

	title := request.GetString("title", "")
	if title != "" {
		params["title"] = title
	}

	colorId := request.GetString("color_id", "")
	if colorId != "" {
		params["color_id"] = colorId
	}

	ownerId := request.GetInt("owner_id", 0)
	if ownerId != 0 {
		params["owner_id"] = ownerId
	}

	dateDue := request.GetString("date_due", "")
	if dateDue != "" {
		params["date_due"] = dateDue
	}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}

	categoryId := request.GetInt("category_id", 0)
	if categoryId != 0 {
		params["category_id"] = categoryId
	}

	score := request.GetInt("score", 0)
	if score != 0 {
		params["score"] = score
	}

	priority := request.GetInt("priority", 0)
	if priority != 0 {
		params["priority"] = priority
	}

	recurrenceStatus := request.GetInt("recurrence_status", 0)
	if recurrenceStatus != 0 {
		params["recurrence_status"] = recurrenceStatus
	}

	recurrenceTrigger := request.GetInt("recurrence_trigger", 0)
	if recurrenceTrigger != 0 {
		params["recurrence_trigger"] = recurrenceTrigger
	}

	recurrenceFactor := request.GetInt("recurrence_factor", 0)
	if recurrenceFactor != 0 {
		params["recurrence_factor"] = recurrenceFactor
	}

	recurrenceTimeframe := request.GetInt("recurrence_timeframe", 0)
	if recurrenceTimeframe != 0 {
		params["recurrence_timeframe"] = recurrenceTimeframe
	}

	recurrenceBasedate := request.GetInt("recurrence_basedate", 0)
	if recurrenceBasedate != 0 {
		params["recurrence_basedate"] = recurrenceBasedate
	}

	reference := request.GetString("reference", "")
	if reference != "" {
		params["reference"] = reference
	}

	tags := request.GetStringSlice("tags", []string{})
	if len(tags) > 0 {
		params["tags"] = tags
	}

	dateStarted := request.GetString("date_started", "")
	if dateStarted != "" {
		params["date_started"] = dateStarted
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
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": taskId}
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

func (kc *kanboardClient) getTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": taskId}
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

func (kc *kanboardClient) moveTaskPositionHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	columnId, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	position, err := request.RequireInt("position")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id":  projectId,
		"task_id":     taskId,
		"column_id":   columnId,
		"position":    position,
		"swimlane_id": swimlaneId,
	}

	result, err := kc.callKanboardAPI(ctx, "moveTaskPosition", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getUsersHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	role := request.GetString("role", "")
	if role != "" {
		params["role"] = role
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
	id, err := request.RequireInt("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": id}

	username := request.GetString("username", "")
	if username != "" {
		params["username"] = username
	}

	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	email := request.GetString("email", "")
	if email != "" {
		params["email"] = email
	}

	role := request.GetString("role", "")
	if role != "" {
		params["role"] = role
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
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": userId}
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

func (kc *kanboardClient) getMeHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getMe", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getMyDashboardHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getMyDashboard", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getMyActivityStreamHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getMyActivityStream", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createMyPrivateProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"name": name}
	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}
	result, err := kc.callKanboardAPI(ctx, "createMyPrivateProject", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getMyProjectsListHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getMyProjectsList", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getMyOverdueTasksHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getMyOverdueTasks", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getMyProjectsHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getMyProjects", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getExternalTaskLinkTypesHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getExternalTaskLinkTypes", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getExternalTaskLinkProviderDependenciesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	providerName, err := request.RequireString("provider_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"providerName": providerName}
	result, err := kc.callKanboardAPI(ctx, "getExternalTaskLinkProviderDependencies", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createExternalTaskLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	url, err := request.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dependency, err := request.RequireString("dependency")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"task_id":    taskId,
		"url":        url,
		"dependency": dependency,
	}

	typeName := request.GetString("type", "")
	if typeName != "" {
		params["type"] = typeName
	}

	title := request.GetString("title", "")
	if title != "" {
		params["title"] = title
	}

	result, err := kc.callKanboardAPI(ctx, "createExternalTaskLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateExternalTaskLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"task_id": taskId, "link_id": linkId}

	title := request.GetString("title", "")
	if title != "" {
		params["title"] = title
	}
	url := request.GetString("url", "")
	if url != "" {
		params["url"] = url
	}
	dependency := request.GetString("dependency", "")
	if dependency != "" {
		params["dependency"] = dependency
	}

	result, err := kc.callKanboardAPI(ctx, "updateExternalTaskLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getExternalTaskLinkByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": taskId, "link_id": linkId}
	result, err := kc.callKanboardAPI(ctx, "getExternalTaskLinkById", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllExternalTaskLinksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": taskId}
	result, err := kc.callKanboardAPI(ctx, "getAllExternalTaskLinks", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeExternalTaskLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": taskId, "link_id": linkId}
	result, err := kc.callKanboardAPI(ctx, "removeExternalTaskLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getColumnsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIdStr, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}

	params := map[string]int{"project_id": projectId}
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

func (kc *kanboardClient) getColumnHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	columnId, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"column_id": columnId}
	result, err := kc.callKanboardAPI(ctx, "getColumn", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get column details: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createColumnHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIdStr, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}
	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id": projectId,
		"title":      title,
	}

	taskLimit := request.GetInt("task_limit", 0)
	if taskLimit != 0 {
		params["task_limit"] = taskLimit
	}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
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
	columnId, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": columnId, "title": title}

	taskLimit := request.GetInt("task_limit", 0)
	if taskLimit != 0 {
		params["task_limit"] = taskLimit
	}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
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
	columnId, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"column_id": columnId}
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
	projectIdStr, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}
	columnId, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	position, err := request.RequireInt("position")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id": projectId,
		"column_id":  columnId,
		"position":   position,
	}

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
	projectIdStr, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}

	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getAllCategories", params)
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
	projectIdStr, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"name": name, "project_id": projectId}

	colorId := request.GetString("color_id", "")
	if colorId != "" {
		params["color_id"] = colorId
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

func (kc *kanboardClient) getCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	categoryId, err := request.RequireInt("category_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"category_id": categoryId}
	result, err := kc.callKanboardAPI(ctx, "getCategory", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get category details: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	categoryId, err := request.RequireInt("category_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": categoryId}
	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	colorId := request.GetString("color_id", "")
	if colorId != "" {
		params["color_id"] = colorId
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
	categoryId, err := request.RequireInt("category_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"category_id": categoryId}
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

func (kc *kanboardClient) getOldSwimlanesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId := request.GetString("project_id", "") // Swimlanes can be global or project-specific

	var result interface{}
	var err error
	if projectId != "" {
		params := map[string]string{"project_id": projectId}
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

func (kc *kanboardClient) getBoardHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := []int{projectId}
	result, err := kc.callKanboardAPI(ctx, "getBoard", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get board details: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) assignTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]int{"task_id": taskId, "owner_id": userId}
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
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dueDate, err := request.RequireString("due_date")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"task_id": taskId, "date_due": dueDate}
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

func (kc *kanboardClient) createCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"task_id": taskId,
		"user_id": userId,
		"content": content,
	}

	reference := request.GetString("reference", "")
	if reference != "" {
		params["reference"] = reference
	}

	visibility := request.GetString("visibility", "")
	if visibility != "" {
		// Validate visibility
		validVisibilities := []string{"app-user", "app-manager", "app-admin"}
		visibilityValid := false
		for _, validVisibility := range validVisibilities {
			if visibility == validVisibility {
				visibilityValid = true
				break
			}
		}
		if !visibilityValid {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid visibility '%s'. Valid visibilities are: %v", visibility, validVisibilities)), nil
		}
		params["visibility"] = visibility
	}

	result, err := kc.callKanboardAPI(ctx, "createComment", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create comment: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTaskCommentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": taskId}
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

func (kc *kanboardClient) getCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	commentId, err := request.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"comment_id": commentId}
	result, err := kc.callKanboardAPI(ctx, "getComment", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get comment details: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireInt("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": id, "content": content}
	result, err := kc.callKanboardAPI(ctx, "updateComment", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update comment: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	commentId, err := request.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"comment_id": commentId}
	result, err := kc.callKanboardAPI(ctx, "removeComment", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove comment: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) assignUserToProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIdStr, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}

	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	role := request.GetString("role", "project-member") // Default to project-member

	// Validate role
	validRoles := []string{"project-member", "project-manager", "project-viewer"}
	roleValid := false
	for _, validRole := range validRoles {
		if role == validRole {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid role '%s'. Valid roles are: %v", role, validRoles)), nil
	}

	params := map[string]interface{}{
		"project_id": projectId,
		"user_id":    userId,
		"role":       role,
	}

	result, err := kc.callKanboardAPI(ctx, "addProjectUser", params)
	if err != nil {
		// Provide more helpful error message for 403 errors
		if err.Error() == "API request failed with status code: 403, Response: 403 Forbidden" {
			return mcp.NewToolResultError(fmt.Sprintf("Permission denied (403 Forbidden). The API user does not have sufficient privileges to assign users to projects. Please ensure the API user has 'app-admin' role in Kanboard. Project ID: %d, User ID: %d, Role: %s", projectId, userId, role)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to assign user to project: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"name": name}
	externalId := request.GetString("external_id", "")
	if externalId != "" {
		params["external_id"] = externalId
	}

	result, err := kc.callKanboardAPI(ctx, "createGroup", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": groupId}
	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}
	externalId := request.GetString("external_id", "")
	if externalId != "" {
		params["external_id"] = externalId
	}

	result, err := kc.callKanboardAPI(ctx, "updateGroup", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": groupId}
	result, err := kc.callKanboardAPI(ctx, "removeGroup", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": groupId}
	result, err := kc.callKanboardAPI(ctx, "getGroup", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllGroupsHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getAllGroups", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getMemberGroupsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "getMemberGroups", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getGroupMembersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": groupId}
	result, err := kc.callKanboardAPI(ctx, "getGroupMembers", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) addGroupMemberHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": groupId, "user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "addGroupMember", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeGroupMemberHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": groupId, "user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "removeGroupMember", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) isGroupMemberHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": groupId, "user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "isGroupMember", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createTaskLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	oppositeTaskId, err := request.RequireInt("opposite_task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{
		"task_id":          taskId,
		"opposite_task_id": oppositeTaskId,
		"link_id":          linkId,
	}
	result, err := kc.callKanboardAPI(ctx, "createTaskLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateTaskLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskLinkId, err := request.RequireInt("task_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	oppositeTaskId, err := request.RequireInt("opposite_task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{
		"id":               taskLinkId,
		"task_id":          taskId,
		"opposite_task_id": oppositeTaskId,
		"link_id":          linkId,
	}
	result, err := kc.callKanboardAPI(ctx, "updateTaskLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTaskLinkByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskLinkId, err := request.RequireInt("task_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_link_id": taskLinkId}
	result, err := kc.callKanboardAPI(ctx, "getTaskLinkById", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllTaskLinksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": taskId}
	result, err := kc.callKanboardAPI(ctx, "getAllTaskLinks", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeTaskLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskLinkId, err := request.RequireInt("task_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_link_id": taskLinkId}
	result, err := kc.callKanboardAPI(ctx, "removeTaskLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllLinksHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getAllLinks", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getOppositeLinkIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"link_id": linkId}
	result, err := kc.callKanboardAPI(ctx, "getOppositeLinkId", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getLinkByLabelHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	label, err := request.RequireString("label")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"label": label}
	result, err := kc.callKanboardAPI(ctx, "getLinkByLabel", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getLinkByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"link_id": linkId}
	result, err := kc.callKanboardAPI(ctx, "getLinkById", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	label, err := request.RequireString("label")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	oppositeLabel := request.GetString("opposite_label", "")

	params := map[string]interface{}{
		"label":          label,
		"opposite_label": oppositeLabel,
	}

	result, err := kc.callKanboardAPI(ctx, "createLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	oppositeLinkId, err := request.RequireInt("opposite_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	label, err := request.RequireString("label")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"id":               linkId,
		"opposite_link_id": oppositeLinkId,
		"label":            label,
	}

	result, err := kc.callKanboardAPI(ctx, "updateLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	linkId, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"link_id": linkId}
	result, err := kc.callKanboardAPI(ctx, "removeLink", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getProjectById", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectByNameHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"name": name}
	result, err := kc.callKanboardAPI(ctx, "getProjectByName", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project by name: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectByIdentifierHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	identifier, err := request.RequireString("identifier")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"identifier": identifier}
	result, err := kc.callKanboardAPI(ctx, "getProjectByIdentifier", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project by identifier: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectByEmailHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	email, err := request.RequireString("email")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"email": email}
	result, err := kc.callKanboardAPI(ctx, "getProjectByEmail", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project by email: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllProjectsHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (kc *kanboardClient) updateProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": projectId}

	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}

	ownerId := request.GetInt("owner_id", 0)
	if ownerId != 0 {
		params["owner_id"] = ownerId
	}

	identifier := request.GetString("identifier", "")
	if identifier != "" {
		params["identifier"] = identifier
	}

	startDate := request.GetString("start_date", "")
	if startDate != "" {
		params["start_date"] = startDate
	}

	endDate := request.GetString("end_date", "")
	if endDate != "" {
		params["end_date"] = endDate
	}

	priorityDefault := request.GetInt("priority_default", 0)
	if priorityDefault != 0 {
		params["priority_default"] = priorityDefault
	}

	priorityStart := request.GetInt("priority_start", 0)
	if priorityStart != 0 {
		params["priority_start"] = priorityStart
	}

	priorityEnd := request.GetInt("priority_end", 0)
	if priorityEnd != 0 {
		params["priority_end"] = priorityEnd
	}

	email := request.GetString("email", "")
	if email != "" {
		params["email"] = email
	}

	result, err := kc.callKanboardAPI(ctx, "updateProject", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update project: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "removeProject", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove project: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) enableProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "enableProject", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to enable project: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) disableProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "disableProject", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to disable project: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) enableProjectPublicAccessHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "enableProjectPublicAccess", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to enable project public access: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) disableProjectPublicAccessHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "disableProjectPublicAccess", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to disable project public access: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectActivityHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getProjectActivity", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project activity: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectActivitiesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIds, err := request.RequireIntSlice("project_ids")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_ids": projectIds}
	result, err := kc.callKanboardAPI(ctx, "getProjectActivities", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get project activities: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createProjectFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	filename, err := request.RequireString("filename")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	blob, err := request.RequireString("blob")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id": projectId,
		"filename":   filename,
		"blob":       blob,
	}

	result, err := kc.callKanboardAPI(ctx, "createProjectFile", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllProjectFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getAllProjectFiles", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fileId, err := request.RequireInt("file_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId, "file_id": fileId}
	result, err := kc.callKanboardAPI(ctx, "getProjectFile", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) downloadProjectFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fileId, err := request.RequireInt("file_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId, "file_id": fileId}
	result, err := kc.callKanboardAPI(ctx, "downloadProjectFile", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeProjectFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fileId, err := request.RequireInt("file_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId, "file_id": fileId}
	result, err := kc.callKanboardAPI(ctx, "removeProjectFile", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeAllProjectFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "removeAllProjectFiles", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectMetadataHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getProjectMetadata", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectMetadataByNameHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "name": name}
	result, err := kc.callKanboardAPI(ctx, "getProjectMetadataByName", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) saveProjectMetadataHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	args := request.GetArguments()
	values, ok := args["values"].(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Missing or invalid 'values' parameter"), nil
	}

	params := map[string]interface{}{"project_id": projectId, "values": values}
	result, err := kc.callKanboardAPI(ctx, "saveProjectMetadata", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeProjectMetadataHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "name": name}
	result, err := kc.callKanboardAPI(ctx, "removeProjectMetadata", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getProjectUsers", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAssignableUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	prependUnassigned := request.GetBool("prepend_unassigned", false)

	params := map[string]interface{}{"project_id": projectId, "prepend_unassigned": prependUnassigned}
	result, err := kc.callKanboardAPI(ctx, "getAssignableUsers", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) addProjectUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	role := request.GetString("role", "")
	params := map[string]interface{}{"project_id": projectId, "user_id": userId}
	if role != "" {
		params["role"] = role
	}
	result, err := kc.callKanboardAPI(ctx, "addProjectUser", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) addProjectGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	role := request.GetString("role", "")
	params := map[string]interface{}{"project_id": projectId, "group_id": groupId}
	if role != "" {
		params["role"] = role
	}
	result, err := kc.callKanboardAPI(ctx, "addProjectGroup", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeProjectUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "removeProjectUser", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeProjectGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "group_id": groupId}
	result, err := kc.callKanboardAPI(ctx, "removeProjectGroup", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) changeProjectUserRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	role, err := request.RequireString("role")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "user_id": userId, "role": role}
	result, err := kc.callKanboardAPI(ctx, "changeProjectUserRole", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) changeProjectGroupRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	groupId, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	role, err := request.RequireString("role")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "group_id": groupId, "role": role}
	result, err := kc.callKanboardAPI(ctx, "changeProjectGroupRole", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectUserRoleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "getProjectUserRole", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

// Subtask Management
func (kc *kanboardClient) createSubtaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"task_id": taskId, "title": title}

	userId := request.GetInt("user_id", 0)
	if userId != 0 {
		params["user_id"] = userId
	}
	timeEstimated := request.GetInt("time_estimated", 0)
	if timeEstimated != 0 {
		params["time_estimated"] = timeEstimated
	}
	timeSpent := request.GetInt("time_spent", 0)
	if timeSpent != 0 {
		params["time_spent"] = timeSpent
	}
	status := request.GetInt("status", 0)
	if status != 0 {
		params["status"] = status
	}

	result, err := kc.callKanboardAPI(ctx, "createSubtask", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getSubtaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subtaskId, err := request.RequireInt("subtask_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"subtask_id": subtaskId}
	result, err := kc.callKanboardAPI(ctx, "getSubtask", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllSubtasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"task_id": taskId}
	result, err := kc.callKanboardAPI(ctx, "getAllSubtasks", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateSubtaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireInt("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"id": id, "task_id": taskId}

	title := request.GetString("title", "")
	if title != "" {
		params["title"] = title
	}
	userId := request.GetInt("user_id", 0)
	if userId != 0 {
		params["user_id"] = userId
	}
	timeEstimated := request.GetInt("time_estimated", 0)
	if timeEstimated != 0 {
		params["time_estimated"] = timeEstimated
	}
	timeSpent := request.GetInt("time_spent", 0)
	if timeSpent != 0 {
		params["time_spent"] = timeSpent
	}
	status := request.GetInt("status", 0)
	if status != 0 {
		params["status"] = status
	}

	result, err := kc.callKanboardAPI(ctx, "updateSubtask", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeSubtaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subtaskId, err := request.RequireInt("subtask_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"subtask_id": subtaskId}
	result, err := kc.callKanboardAPI(ctx, "removeSubtask", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) hasSubtaskTimerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subtaskId, err := request.RequireInt("subtask_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"subtask_id": subtaskId}
	userId := request.GetInt("user_id", 0)
	if userId != 0 {
		params["user_id"] = userId
	}
	result, err := kc.callKanboardAPI(ctx, "hasSubtaskTimer", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) setSubtaskStartTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subtaskId, err := request.RequireInt("subtask_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"subtask_id": subtaskId}
	userId := request.GetInt("user_id", 0)
	if userId != 0 {
		params["user_id"] = userId
	}
	result, err := kc.callKanboardAPI(ctx, "setSubtaskStartTime", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) setSubtaskEndTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subtaskId, err := request.RequireInt("subtask_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"subtask_id": subtaskId}
	userId := request.GetInt("user_id", 0)
	if userId != 0 {
		params["user_id"] = userId
	}
	result, err := kc.callKanboardAPI(ctx, "setSubtaskEndTime", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getSubtaskTimeSpentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subtaskId, err := request.RequireInt("subtask_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"subtask_id": subtaskId}
	userId := request.GetInt("user_id", 0)
	if userId != 0 {
		params["user_id"] = userId
	}
	result, err := kc.callKanboardAPI(ctx, "getSubtaskTimeSpent", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllTagsHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getAllTags", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTagsByProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getTagsByProject", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createTagHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	tag, err := request.RequireString("tag")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "tag": tag}
	colorId := request.GetInt("color_id", 0)
	if colorId != 0 {
		params["color_id"] = colorId
	}
	result, err := kc.callKanboardAPI(ctx, "createTag", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) updateTagHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tagId, err := request.RequireInt("tag_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	tag, err := request.RequireString("tag")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"id": tagId, "name": tag}
	colorId := request.GetInt("color_id", 0)
	if colorId != 0 {
		params["color_id"] = colorId
	}
	result, err := kc.callKanboardAPI(ctx, "updateTag", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeTagHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tagId, err := request.RequireInt("tag_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"id": tagId}
	result, err := kc.callKanboardAPI(ctx, "removeTag", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) setTaskTagsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	tags, err := request.RequireStringSlice("tags")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "task_id": taskId, "tags": tags}
	result, err := kc.callKanboardAPI(ctx, "setTaskTags", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTaskTagsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"task_id": taskId}
	result, err := kc.callKanboardAPI(ctx, "getTaskTags", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createTaskFileHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetInt("project_id", 0)
	if projectID == 0 {
		return mcp.NewToolResultError("project_id is required"), nil
	}
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return mcp.NewToolResultError("task_id is required"), nil
	}
	filename := request.GetString("filename", "")
	if filename == "" {
		return mcp.NewToolResultError("filename is required"), nil
	}
	blob := request.GetString("blob", "")
	if blob == "" {
		return mcp.NewToolResultError("blob is required"), nil
	}

	result, err := kc.CreateTaskFile(
		projectID,
		taskID,
		filename,
		blob,
	)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllTaskFilesHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return mcp.NewToolResultError("task_id is required"), nil
	}

	result, err := kc.GetAllTaskFiles(taskID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTaskFileHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fileID := request.GetInt("file_id", 0)
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required"), nil
	}

	result, err := kc.GetTaskFile(fileID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) downloadTaskFileHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fileID := request.GetInt("file_id", 0)
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required"), nil
	}

	result, err := kc.DownloadTaskFile(fileID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}

func (kc *kanboardClient) removeTaskFileHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fileID := request.GetInt("file_id", 0)
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required"), nil
	}

	result, err := kc.RemoveTaskFile(fileID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) removeAllTaskFilesHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID := request.GetInt("task_id", 0)
	if taskID == 0 {
		return mcp.NewToolResultError("task_id is required"), nil
	}

	result, err := kc.RemoveAllTaskFiles(taskID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) CreateTaskFile(projectID, taskID int, filename, blob string) (int, error) {
	params := []interface{}{projectID, taskID, filename, blob}
	result, err := kc.callKanboardAPI(context.Background(), "createTaskFile", params)
	if err != nil {
		return 0, err
	}

	// Kanboard API returns 1 on success for some operations like creating a file
	if fileID, ok := result.(float64); ok {
		return int(fileID), nil
	}
	return 0, fmt.Errorf("unexpected result type for CreateTaskFile: %T", result)
}

func (kc *kanboardClient) GetAllTaskFiles(taskID int) ([]interface{}, error) {
	params := map[string]interface{}{"task_id": taskID}
	result, err := kc.callKanboardAPI(context.Background(), "getAllTaskFiles", params)
	if err != nil {
		return nil, err
	}

	if files, ok := result.([]interface{}); ok {
		return files, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetAllTaskFiles: %T", result)
}

func (kc *kanboardClient) GetTaskFile(fileID int) (interface{}, error) {
	params := []interface{}{fileID}
	result, err := kc.callKanboardAPI(context.Background(), "getTaskFile", params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (kc *kanboardClient) DownloadTaskFile(fileID int) (string, error) {
	params := []interface{}{fileID}
	result, err := kc.callKanboardAPI(context.Background(), "downloadTaskFile", params)
	if err != nil {
		return "", err
	}

	if content, ok := result.(string); ok {
		return content, nil
	}
	return "", fmt.Errorf("unexpected result type for DownloadTaskFile: %T", result)
}

func (kc *kanboardClient) RemoveTaskFile(fileID int) (bool, error) {
	params := []interface{}{fileID}
	result, err := kc.callKanboardAPI(context.Background(), "removeTaskFile", params)
	if err != nil {
		return false, err
	}

	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for RemoveTaskFile: %T", result)
}

func (kc *kanboardClient) RemoveAllTaskFiles(taskID int) (bool, error) {
	params := map[string]interface{}{"task_id": taskID}
	result, err := kc.callKanboardAPI(context.Background(), "removeAllTaskFiles", params)
	if err != nil {
		return false, err
	}

	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for RemoveAllTaskFiles: %T", result)
}

func (kc *kanboardClient) GetVersion() (string, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getVersion", nil)
	if err != nil {
		return "", err
	}

	if version, ok := result.(string); ok {
		return version, nil
	}
	return "", fmt.Errorf("unexpected result type for GetVersion: %T", result)
}

func (kc *kanboardClient) GetTimezone() (string, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getTimezone", nil)
	if err != nil {
		return "", err
	}

	if timezone, ok := result.(string); ok {
		return timezone, nil
	}
	return "", fmt.Errorf("unexpected result type for GetTimezone: %T", result)
}

func (kc *kanboardClient) GetDefaultTaskColors() (map[string]interface{}, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getDefaultTaskColors", nil)
	if err != nil {
		return nil, err
	}

	if colors, ok := result.(map[string]interface{}); ok {
		return colors, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetDefaultTaskColors: %T", result)
}

func (kc *kanboardClient) GetDefaultTaskColor() (string, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getDefaultTaskColor", nil)
	if err != nil {
		return "", err
	}

	if colorID, ok := result.(string); ok {
		return colorID, nil
	}
	return "", fmt.Errorf("unexpected result type for GetDefaultTaskColor: %T", result)
}

func (kc *kanboardClient) GetColorList() (map[string]interface{}, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getColorList", nil)
	if err != nil {
		return nil, err
	}

	if colorList, ok := result.(map[string]interface{}); ok {
		return colorList, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetColorList: %T", result)
}

func (kc *kanboardClient) GetApplicationRoles() (map[string]interface{}, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getApplicationRoles", nil)
	if err != nil {
		return nil, err
	}

	if roles, ok := result.(map[string]interface{}); ok {
		return roles, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetApplicationRoles: %T", result)
}

func (kc *kanboardClient) GetProjectRoles() (map[string]interface{}, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getProjectRoles", nil)
	if err != nil {
		return nil, err
	}

	if roles, ok := result.(map[string]interface{}); ok {
		return roles, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetProjectRoles: %T", result)
}

func (kc *kanboardClient) getVersionHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetVersion()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}

func (kc *kanboardClient) getTimezoneHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetTimezone()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}

func (kc *kanboardClient) getDefaultTaskColorsHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetDefaultTaskColors()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getDefaultTaskColorHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetDefaultTaskColor()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}

func (kc *kanboardClient) getColorListHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetColorList()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getApplicationRolesHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetApplicationRoles()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getProjectRolesHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetProjectRoles()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) GetAvailableActions() (map[string]interface{}, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getAvailableActions", nil)
	if err != nil {
		return nil, err
	}

	if actions, ok := result.(map[string]interface{}); ok {
		return actions, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetAvailableActions: %T", result)
}

func (kc *kanboardClient) GetAvailableActionEvents() (map[string]interface{}, error) {
	result, err := kc.callKanboardAPI(context.Background(), "getAvailableActionEvents", nil)
	if err != nil {
		return nil, err
	}

	if events, ok := result.(map[string]interface{}); ok {
		return events, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetAvailableActionEvents: %T", result)
}

func (kc *kanboardClient) GetCompatibleActionEvents(actionName string) (map[string]interface{}, error) {
	params := []interface{}{actionName}
	result, err := kc.callKanboardAPI(context.Background(), "getCompatibleActionEvents", params)
	if err != nil {
		return nil, err
	}

	if events, ok := result.(map[string]interface{}); ok {
		return events, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetCompatibleActionEvents: %T", result)
}

func (kc *kanboardClient) GetActions(projectID int) ([]interface{}, error) {
	params := []interface{}{projectID}
	result, err := kc.callKanboardAPI(context.Background(), "getActions", params)
	if err != nil {
		return nil, err
	}

	if actions, ok := result.([]interface{}); ok {
		return actions, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetActions: %T", result)
}

func (kc *kanboardClient) CreateAction(projectID int, eventName, actionName string, params map[string]interface{}) (int, error) {
	realParams := map[string]interface{}{
		"project_id":  projectID,
		"event_name":  eventName,
		"action_name": actionName,
		"params":      params,
	}
	result, err := kc.callKanboardAPI(context.Background(), "createAction", realParams)
	if err != nil {
		return 0, err
	}

	if actionID, ok := result.(float64); ok {
		return int(actionID), nil
	}
	return 0, fmt.Errorf("unexpected result type for CreateAction: %T", result)
}

func (kc *kanboardClient) RemoveAction(actionID int) (bool, error) {
	params := []interface{}{actionID}
	result, err := kc.callKanboardAPI(context.Background(), "removeAction", params)
	if err != nil {
		return false, err
	}

	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for RemoveAction: %T", result)
}

func (kc *kanboardClient) getAvailableActionsHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetAvailableActions()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAvailableActionEventsHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.GetAvailableActionEvents()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getCompatibleActionEventsHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	actionName := request.GetString("action_name", "")
	if actionName == "" {
		return mcp.NewToolResultError("action_name is required"), nil
	}
	result, err := kc.GetCompatibleActionEvents(actionName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getActionsHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetInt("project_id", 0)
	if projectID == 0 {
		return mcp.NewToolResultError("project_id is required"), nil
	}
	result, err := kc.GetActions(projectID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createActionHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetInt("project_id", 0)
	if projectID == 0 {
		return mcp.NewToolResultError("project_id is required"), nil
	}
	eventName := request.GetString("event_name", "")
	if eventName == "" {
		return mcp.NewToolResultError("event_name is required"), nil
	}
	actionName := request.GetString("action_name", "")
	if actionName == "" {
		return mcp.NewToolResultError("action_name is required"), nil
	}

	// Retrieve all arguments and then extract 'params'
	allArgs := request.GetArguments()
	params, ok := allArgs["params"].(map[string]interface{})
	if !ok {
		// If 'params' is not provided or not a map, return an error or an empty map
		return mcp.NewToolResultError("params must be a map or omitted"), nil
	}

	actionID, err := kc.CreateAction(projectID, eventName, actionName, params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.Itoa(actionID)), nil
}

func (kc *kanboardClient) removeActionHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	actionID := request.GetInt("action_id", 0)
	if actionID == 0 {
		return mcp.NewToolResultError("action_id is required"), nil
	}
	result, err := kc.RemoveAction(actionID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

func (kc *kanboardClient) GetActiveSwimlanes(projectID int) ([]interface{}, error) {
	params := []interface{}{projectID}
	result, err := kc.callKanboardAPI(context.Background(), "getActiveSwimlanes", params)
	if err != nil {
		return nil, err
	}

	if swimlanes, ok := result.([]interface{}); ok {
		return swimlanes, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetActiveSwimlanes: %T", result)
}

func (kc *kanboardClient) GetAllSwimlanes(projectID int) ([]interface{}, error) {
	params := []interface{}{projectID}
	result, err := kc.callKanboardAPI(context.Background(), "getAllSwimlanes", params)
	if err != nil {
		return nil, err
	}

	if swimlanes, ok := result.([]interface{}); ok {
		return swimlanes, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetAllSwimlanes: %T", result)
}

func (kc *kanboardClient) GetSwimlaneById(swimlaneID int) (interface{}, error) {
	params := []interface{}{swimlaneID}
	result, err := kc.callKanboardAPI(context.Background(), "getSwimlaneById", params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (kc *kanboardClient) GetSwimlaneByName(projectID int, name string) (interface{}, error) {
	params := []interface{}{projectID, name}
	result, err := kc.callKanboardAPI(context.Background(), "getSwimlaneByName", params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (kc *kanboardClient) ChangeSwimlanePosition(projectID, swimlaneID, position int) (bool, error) {
	params := []interface{}{projectID, swimlaneID, position}
	result, err := kc.callKanboardAPI(context.Background(), "changeSwimlanePosition", params)
	if err != nil {
		return false, err
	}
	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for ChangeSwimlanePosition: %T", result)
}

func (kc *kanboardClient) UpdateSwimlane(projectID, swimlaneID int, name, description string) (bool, error) {
	params := map[string]interface{}{
		"project_id": projectID,
		"id":         swimlaneID,
		"name":       name,
	}
	if description != "" {
		params["description"] = description
	}
	result, err := kc.callKanboardAPI(context.Background(), "updateSwimlane", params)
	if err != nil {
		return false, err
	}
	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for UpdateSwimlane: %T", result)
}

func (kc *kanboardClient) AddSwimlane(projectID int, name, description string) (int, error) {
	params := map[string]interface{}{
		"project_id": projectID,
		"name":       name,
	}
	if description != "" {
		params["description"] = description
	}
	result, err := kc.callKanboardAPI(context.Background(), "addSwimlane", params)
	if err != nil {
		return 0, err
	}
	if swimlaneID, ok := result.(float64); ok {
		return int(swimlaneID), nil
	}
	return 0, fmt.Errorf("unexpected result type for AddSwimlane: %T", result)
}

func (kc *kanboardClient) RemoveSwimlane(projectID, swimlaneID int) (bool, error) {
	params := []interface{}{projectID, swimlaneID}
	result, err := kc.callKanboardAPI(context.Background(), "removeSwimlane", params)
	if err != nil {
		return false, err
	}
	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for RemoveSwimlane: %T", result)
}

func (kc *kanboardClient) DisableSwimlane(projectID, swimlaneID int) (bool, error) {
	params := []interface{}{projectID, swimlaneID}
	result, err := kc.callKanboardAPI(context.Background(), "disableSwimlane", params)
	if err != nil {
		return false, err
	}
	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for DisableSwimlane: %T", result)
}

func (kc *kanboardClient) EnableSwimlane(projectID, swimlaneID int) (bool, error) {
	params := []interface{}{projectID, swimlaneID}
	result, err := kc.callKanboardAPI(context.Background(), "enableSwimlane", params)
	if err != nil {
		return false, err
	}
	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for EnableSwimlane: %T", result)
}

func (kc *kanboardClient) GetSwimlane(swimlaneID int) (interface{}, error) {
	params := []interface{}{swimlaneID}
	result, err := kc.callKanboardAPI(context.Background(), "getSwimlane", params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (kc *kanboardClient) getSwimlaneHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.GetSwimlane(swimlaneId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getSwimlaneByIdHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.GetSwimlaneById(swimlaneId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getActiveSwimlanesHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.GetActiveSwimlanes(projectId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllSwimlanesHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.GetAllSwimlanes(projectId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getSwimlaneByNameHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.GetSwimlaneByName(projectId, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) changeSwimlanePositionHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	position, err := request.RequireInt("position")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.ChangeSwimlanePosition(projectId, swimlaneId, position)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

func (kc *kanboardClient) addSwimlaneHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	description := request.GetString("description", "")
	result, err := kc.AddSwimlane(projectId, name, description)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.Itoa(result)), nil
}

func (kc *kanboardClient) updateSwimlaneHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name := request.GetString("name", "")
	description := request.GetString("description", "")
	result, err := kc.UpdateSwimlane(projectId, swimlaneId, name, description)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

func (kc *kanboardClient) removeSwimlaneHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.RemoveSwimlane(projectId, swimlaneId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

func (kc *kanboardClient) disableSwimlaneHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.DisableSwimlane(projectId, swimlaneId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

func (kc *kanboardClient) enableSwimlaneHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId, err := request.RequireInt("swimlane_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.EnableSwimlane(projectId, swimlaneId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

// GetTaskMetadata Task Metadata API Procedures
func (kc *kanboardClient) GetTaskMetadata(taskID int) (map[string]interface{}, error) {
	params := []interface{}{taskID}
	result, err := kc.callKanboardAPI(context.Background(), "getTaskMetadata", params)
	if err != nil {
		return nil, err
	}

	if metadata, ok := result.(map[string]interface{}); ok {
		return metadata, nil
	}

	if result == nil {
		return map[string]interface{}{}, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetTaskMetadata: %T", result)
}

func (kc *kanboardClient) GetTaskMetadataByName(taskID int, name string) (string, error) {
	params := []interface{}{taskID, name}
	result, err := kc.callKanboardAPI(context.Background(), "getTaskMetadataByName", params)
	if err != nil {
		return "", err
	}

	if value, ok := result.(string); ok {
		return value, nil
	}

	if result == nil {
		return "", nil // Kanboard returns null for empty string
	}
	return "", fmt.Errorf("unexpected result type for GetTaskMetadataByName: %T", result)
}

func (kc *kanboardClient) SaveTaskMetadata(taskID int, values map[string]string) (bool, error) {
	params := map[string]interface{}{
		"task_id": taskID,
		"values":  values,
	}
	result, err := kc.callKanboardAPI(context.Background(), "saveTaskMetadata", params)
	if err != nil {
		return false, err
	}

	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for SaveTaskMetadata: %T", result)
}

func (kc *kanboardClient) RemoveTaskMetadata(taskID int, name string) (bool, error) {
	params := []interface{}{taskID, name}
	result, err := kc.callKanboardAPI(context.Background(), "removeTaskMetadata", params)
	if err != nil {
		return false, err
	}

	if success, ok := result.(bool); ok {
		return success, nil
	}
	return false, fmt.Errorf("unexpected result type for RemoveTaskMetadata: %T", result)
}

// Task Metadata Handlers
func (kc *kanboardClient) getTaskMetadataHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.GetTaskMetadata(taskId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getTaskMetadataByNameHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.GetTaskMetadataByName(taskId, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}

func (kc *kanboardClient) saveTaskMetadataHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Extract 'values' as a map[string]string
	args := request.GetArguments()
	valuesMap, ok := args["values"].(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Missing or invalid 'values' parameter. Expected a map."), nil
	}

	stringValues := make(map[string]string)
	for key, val := range valuesMap {
		strVal, isString := val.(string)
		if !isString {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid value for metadata key '%s'. Expected string.", key)), nil
		}
		stringValues[key] = strVal
	}

	result, err := kc.SaveTaskMetadata(taskId, stringValues)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

func (kc *kanboardClient) removeTaskMetadataHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, err := kc.RemoveTaskMetadata(taskId, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(strconv.FormatBool(result)), nil
}

func (kc *kanboardClient) getTaskByReferenceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	reference, err := request.RequireString("reference")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "reference": reference}
	result, err := kc.callKanboardAPI(ctx, "getTaskByReference", params)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getAllTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	statusId, err := request.RequireInt("status_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId, "status_id": statusId}
	result, err := kc.callKanboardAPI(ctx, "getAllTasks", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get tasks: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getOverdueTasksHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := kc.callKanboardAPI(ctx, "getOverdueTasks", nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get overdue tasks: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getOverdueTasksByProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": projectId}
	result, err := kc.callKanboardAPI(ctx, "getOverdueTasksByProject", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get overdue tasks: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) openTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"task_id": taskId, "status": "open"}
	result, err := kc.callKanboardAPI(ctx, "updateTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open task: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) closeTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"task_id": taskId, "status": "closed"}
	result, err := kc.callKanboardAPI(ctx, "updateTask", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to close task: %v", err)), nil
	}
	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) moveTaskToProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId := request.GetInt("swimlane_id", 0)
	columnId := request.GetInt("column_id", 0)
	categoryId := request.GetInt("category_id", 0)
	ownerId := request.GetInt("owner_id", 0)

	params := map[string]interface{}{
		"task_id":    taskId,
		"project_id": projectId,
	}
	if swimlaneId != 0 {
		params["swimlane_id"] = swimlaneId
	}
	if columnId != 0 {
		params["column_id"] = columnId
	}
	if categoryId != 0 {
		params["category_id"] = categoryId
	}
	if ownerId != 0 {
		params["owner_id"] = ownerId
	}

	result, err := kc.callKanboardAPI(ctx, "moveTaskToProject", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) duplicateTaskToProjectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskId, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	swimlaneId := request.GetInt("swimlane_id", 0)
	columnId := request.GetInt("column_id", 0)
	categoryId := request.GetInt("category_id", 0)
	ownerId := request.GetInt("owner_id", 0)

	params := map[string]interface{}{
		"task_id":    taskId,
		"project_id": projectId,
	}
	if swimlaneId != 0 {
		params["swimlane_id"] = swimlaneId
	}
	if columnId != 0 {
		params["column_id"] = columnId
	}
	if categoryId != 0 {
		params["category_id"] = categoryId
	}
	if ownerId != 0 {
		params["owner_id"] = ownerId
	}

	result, err := kc.callKanboardAPI(ctx, "duplicateTaskToProject", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to duplicate task: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) searchTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectId, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id": projectId,
		"query":      query,
	}

	result, err := kc.callKanboardAPI(ctx, "searchTasks", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search tasks: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) createLdapUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, err := request.RequireString("username")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]string{"username": username}
	result, err := kc.callKanboardAPI(ctx, "createLdapUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create LDAP user: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) getUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "getUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get user: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) disableUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "disableUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to disable user: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) enableUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "enableUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to enable user: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}

func (kc *kanboardClient) isActiveUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userId, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": userId}
	result, err := kc.callKanboardAPI(ctx, "isActiveUser", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to check if user is active: %v", err)), nil
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal API result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultBytes)), nil
}
