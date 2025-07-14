# kanboard-mcp

This project implements a basic Model Context Protocol (MCP) server in Go.

## Connecting to Cursor

To connect this MCP server to Cursor (or any compatible MCP client like Claude Desktop), you need to compile the Go application and then configure your client to run the executable.

1.  **Build the Go executable:**
    Open your terminal in the project root directory and run:
    ```bash
    go build -o kanboard-mcp.exe .
    ```
    (Note: On Linux/macOS, the executable will be `kanboard-mcp` without the `.exe` extension.)

2.  **Configure Cursor:**
    Create a file named `mcp_config.json` inside a `.cursor` directory in your user's configuration directory (e.g., `~/.cursor/mcp_config.json` on Linux/macOS, or `C:\Users\YOUR_USERNAME\AppData\Roaming\Cursor\.cursor\mcp_config.json` on Windows).

    Add the following content to the `mcp_config.json` file, replacing `/path/to/your/kanboard-mcp.exe` with the actual absolute path to the executable you built in the previous step:

    ```json
    {
      "mcpServers": {
        "kanboard-mcp-server": {
          "command": "/path/to/your/kanboard-mcp.exe",
          "args": [],
          "env": {
            "KANBOARD_API_ENDPOINT": "https://t.b-b.top/jsonrpc.php",
            "KANBOARD_API_KEY": "your-kanboard-api-key",
            "KANBOARD_USERNAME": "your-kanboard-username",
            "KANBOARD_PASSWORD": "your-kanboard-password"
          }
        }
      }
    }
    ```

3.  **Restart Cursor:**
    After saving the `mcp_config.json` file, restart your Cursor application for the changes to take effect.

Once configured, Cursor will be able to discover and interact with the following tools exposed by this MCP server:

| Tool Name          | Description               | NLP Call Example                                   |
|--------------------|---------------------------|----------------------------------------------------|
| `get_projects`     | List all projects         | `"List all Kanboard projects for me"`              |
| `create_project`   | Create new projects       | `"Create a new project called 'My New Project'"`   |
| `get_tasks`        | Get project tasks         | `"Get tasks for the project 'My Project'"`         |
| `create_task`      | Create new tasks          | `"Create a task 'Finish report' in project 'My Project'"` |
| `update_task`      | Modify existing tasks     | `"Update task 123 with description 'New description'"` |
| `delete_task`      | Remove tasks              | `"Delete task with ID 456"`                        |
| `get_task_details` | Get detailed task info    | `"Get details for task 789"`                       |
| `move_task`        | Move tasks between columns| `"Move task 123 to column 'Done'"`                 |
| `get_columns`      | List project columns      | `"Show me all columns in this project"`            |
| `create_column`    | Add new columns           | `"Create a Testing column with 5 task limit"`      |
| `update_column`    | Modify column settings    | `"Change the Review column limit to 3 tasks"`      |
| `delete_column`    | Remove columns            | `"Delete the unused Draft column"`                 |
| `reorder_columns`  | Change column positions   | `"Move Testing column before Done"`                |
| `get_categories`   | List project categories   | `"Show me all task categories"`                    |
| `create_category`  | Add task categories       | `"Create a Bug Fixes category"`                    |
| `update_category`  | Modify categories         | `"Rename Bug Fixes to Critical Issues"`            |
| `delete_category`  | Remove categories         | `"Delete the unused category"`                     |
| `get_swimlanes`    | List project swimlanes    | `"Show me all team swimlanes"`                     |
| `create_swimlane`  | Add team swimlanes        | `"Create a Frontend Team swimlane"`                |
| `update_swimlane`  | Modify swimlanes          | `"Rename Mobile Team to Cross-Platform Team"`      |
| `delete_swimlane`  | Remove swimlanes          | `"Delete the inactive team swimlane"`              |
| `get_users`        | List all system users     | `"Show me all team members"`                       |
| `get_user_by_name` | Get user by name          | `"Find user 'john.doe'"`                           |
| `create_user`      | Create a new user         | `"Create a new user 'testuser' with password 'pass123'"` |
| `update_user`      | Modify an existing user   | `"Update user 1 with new email 'test@example.com'"` |
| `remove_user`      | Remove a user             | `"Remove user with ID 2"`                          |
| `assign_user_to_project` | Assign a user to a project with a specific role | `"Assign user 3 to project 10 as project-manager"` |
| `assign_task`      | Assign tasks to users     | `"Assign the API task to John"`                    |
| `set_task_due_date`| Set task deadlines        | `"Set due date for login task to next Friday"`     |
| `add_task_comment` | Add task comments         | `"Add comment about testing requirements"`         |
| `get_task_comments`| Get task comments         | `"Show all comments on this task"`                 |
