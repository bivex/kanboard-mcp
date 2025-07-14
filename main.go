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

	tool = mcp.NewTool("get_external_task_link_provider_dependencies",
		mcp.WithDescription("Get available dependencies for a given provider"),
		mcp.WithString("provider_name",
			mcp.Required(),
			mcp.Description("Name of the provider"),
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
			mcp.Description("New URL for the external link"),		),
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
			mcp.Description("ID of the external link to retrieve"),		),
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
		mcp.WithDescription("List project swimlanes"),
		mcp.WithString("project_id",
			mcp.Description("ID of the project to get swimlanes from (optional)"),
		),
	)
	s.AddTool(tool, kbClient.getSwimlanesHandler)

	tool = mcp.NewTool("get_board",
		mcp.WithDescription("Get all necessary information to display a board"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("ID of the project to get board details for"),
		),
	)
	s.AddTool(tool, kbClient.getBoardHandler)

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
			mcp.Description("ID of the group"),		),
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

	params := map[string]interface{}{"name": name}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}

	owner_id := request.GetInt("owner_id", 0)
	if owner_id != 0 {
		params["owner_id"] = owner_id
	}

	identifier := request.GetString("identifier", "")
	if identifier != "" {
		params["identifier"] = identifier
	}

	start_date := request.GetString("start_date", "")
	if start_date != "" {
		params["start_date"] = start_date
	}

	end_date := request.GetString("end_date", "")
	if end_date != "" {
		params["end_date"] = end_date
	}

	priority_default := request.GetInt("priority_default", 0)
	if priority_default != 0 {
		params["priority_default"] = priority_default
	}

	priority_start := request.GetInt("priority_start", 0)
	if priority_start != 0 {
		params["priority_start"] = priority_start
	}

	priority_end := request.GetInt("priority_end", 0)
	if priority_end != 0 {
		params["priority_end"] = priority_end
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

func (kc *kanboardClient) getMeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (kc *kanboardClient) getMyDashboardHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (kc *kanboardClient) getMyActivityStreamHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (kc *kanboardClient) getMyProjectsListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (kc *kanboardClient) getMyOverdueTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (kc *kanboardClient) getMyProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (kc *kanboardClient) getExternalTaskLinkTypesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	task_id, err := request.RequireInt("task_id")
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
		"task_id":    task_id,
		"url":        url,
		"dependency": dependency,
	}

	type_name := request.GetString("type", "")
	if type_name != "" {
		params["type"] = type_name
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
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"task_id": task_id, "link_id": link_id}

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
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id, "link_id": link_id}
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
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id}
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
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id, "link_id": link_id}
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

func (kc *kanboardClient) getColumnHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	column_id, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"column_id": column_id}
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
	project_id_str, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}
	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id": project_id,
		"title":      title,
	}

	task_limit := request.GetInt("task_limit", 0)
	if task_limit != 0 {
		params["task_limit"] = task_limit
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
	column_id, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": column_id, "title": title}

	task_limit := request.GetInt("task_limit", 0)
	if task_limit != 0 {
		params["task_limit"] = task_limit
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
	column_id, err := request.RequireInt("column_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	position, err := request.RequireInt("position")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"project_id":  project_id,
		"column_id":   column_id,
		"position":    position,
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
	project_id_str, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	project_id, err := strconv.Atoi(project_id_str)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid project_id: %v", err)), nil
	}

	params := map[string]int{"project_id": project_id}
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

	params := map[string]interface{}{"name": name, "project_id": project_id}

	color_id := request.GetString("color_id", "")
	if color_id != "" {
		params["color_id"] = color_id
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
	category_id, err := request.RequireInt("category_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"category_id": category_id}
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
	category_id, err := request.RequireInt("category_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": category_id}
	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	color_id := request.GetString("color_id", "")
	if color_id != "" {
		params["color_id"] = color_id
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

func (kc *kanboardClient) getBoardHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := []int{project_id}
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

func (kc *kanboardClient) createCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"task_id": task_id,
		"user_id": user_id,
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

func (kc *kanboardClient) getCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	comment_id, err := request.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"comment_id": comment_id}
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
	comment_id, err := request.RequireInt("comment_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"comment_id": comment_id}
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
		"project_id": project_id,
		"user_id":    user_id,
		"role":       role,
	}

	result, err := kc.callKanboardAPI(ctx, "addProjectUser", params)
	if err != nil {
		// Provide more helpful error message for 403 errors
		if err.Error() == "API request failed with status code: 403, Response: 403 Forbidden" {
			return mcp.NewToolResultError(fmt.Sprintf("Permission denied (403 Forbidden). The API user does not have sufficient privileges to assign users to projects. Please ensure the API user has 'app-admin' role in Kanboard. Project ID: %d, User ID: %d, Role: %s", project_id, user_id, role)), nil
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
	external_id := request.GetString("external_id", "")
	if external_id != "" {
		params["external_id"] = external_id
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
	group_id, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": group_id}
	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}
	external_id := request.GetString("external_id", "")
	if external_id != "" {
		params["external_id"] = external_id
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
	group_id, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": group_id}
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
	group_id, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": group_id}
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

func (kc *kanboardClient) getAllGroupsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"user_id": user_id}
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
	group_id, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": group_id}
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
	group_id, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": group_id, "user_id": user_id}
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
	group_id, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": group_id, "user_id": user_id}
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
	group_id, err := request.RequireInt("group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	user_id, err := request.RequireInt("user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"group_id": group_id, "user_id": user_id}
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
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opposite_task_id, err := request.RequireInt("opposite_task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{
		"task_id":        task_id,
		"opposite_task_id": opposite_task_id,
		"link_id":        link_id,
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
	task_link_id, err := request.RequireInt("task_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opposite_task_id, err := request.RequireInt("opposite_task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{
		"id":             task_link_id,
		"task_id":        task_id,
		"opposite_task_id": opposite_task_id,
		"link_id":        link_id,
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
	task_link_id, err := request.RequireInt("task_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_link_id": task_link_id}
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
	task_id, err := request.RequireInt("task_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_id": task_id}
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
	task_link_id, err := request.RequireInt("task_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"task_link_id": task_link_id}
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

func (kc *kanboardClient) getAllLinksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"link_id": link_id}
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
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"link_id": link_id}
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
	opposite_label := request.GetString("opposite_label", "")

	params := map[string]interface{}{
		"label":         label,
		"opposite_label": opposite_label,
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
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opposite_link_id, err := request.RequireInt("opposite_link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	label, err := request.RequireString("label")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{
		"id":             link_id,
		"opposite_link_id": opposite_link_id,
		"label":          label,
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
	link_id, err := request.RequireInt("link_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"link_id": link_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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

func (kc *kanboardClient) getAllProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := map[string]interface{}{"id": project_id}

	name := request.GetString("name", "")
	if name != "" {
		params["name"] = name
	}

	description := request.GetString("description", "")
	if description != "" {
		params["description"] = description
	}

	owner_id := request.GetInt("owner_id", 0)
	if owner_id != 0 {
		params["owner_id"] = owner_id
	}

	identifier := request.GetString("identifier", "")
	if identifier != "" {
		params["identifier"] = identifier
	}

	start_date := request.GetString("start_date", "")
	if start_date != "" {
		params["start_date"] = start_date
	}

	end_date := request.GetString("end_date", "")
	if end_date != "" {
		params["end_date"] = end_date
	}

	priority_default := request.GetInt("priority_default", 0)
	if priority_default != 0 {
		params["priority_default"] = priority_default
	}

	priority_start := request.GetInt("priority_start", 0)
	if priority_start != 0 {
		params["priority_start"] = priority_start
	}

	priority_end := request.GetInt("priority_end", 0)
	if priority_end != 0 {
		params["priority_end"] = priority_end
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_ids, err := request.RequireIntSlice("project_ids")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_ids": project_ids}
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
	project_id, err := request.RequireInt("project_id")
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
		"project_id": project_id,
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	file_id, err := request.RequireInt("file_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id, "file_id": file_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	file_id, err := request.RequireInt("file_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id, "file_id": file_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	file_id, err := request.RequireInt("file_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id, "file_id": file_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]int{"project_id": project_id}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": project_id, "name": name}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	args := request.GetArguments()
	values, ok := args["values"].(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Missing or invalid 'values' parameter"), nil
	}

	params := map[string]interface{}{"project_id": project_id, "values": values}
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
	project_id, err := request.RequireInt("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	params := map[string]interface{}{"project_id": project_id, "name": name}
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

