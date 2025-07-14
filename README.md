# 🚀 Kanboard MCP Server

> **Model Context Protocol (MCP) Server for Kanboard Integration**

A powerful Go-based MCP server that enables seamless integration between AI assistants (like Claude Desktop, Cursor) and Kanboard project management system. Manage your Kanboard projects, tasks, users, and workflows directly through natural language commands.

![Go](https://img.shields.io/badge/Go-1.21+-blue?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![MCP](https://img.shields.io/badge/MCP-Protocol-orange?style=for-the-badge)

## 📋 Table of Contents

- [✨ Features](#-features)
- [🚀 Quick Start](#-quick-start)
- [⚙️ Configuration](#️-configuration)
- [🛠️ Available Tools](#️-available-tools)
- [📖 Usage Examples](#-usage-examples)
- [🔧 Development](#-development)
- [📄 License](#-license)

## ✨ Features

- 🔗 **Seamless Kanboard Integration** - Direct API communication with Kanboard
- 🤖 **Natural Language Processing** - Use plain English to manage your projects
- 📊 **Complete Project Management** - Handle projects, tasks, users, columns, and more
- 🔐 **Secure Authentication** - Support for both API key and username/password auth
- ⚡ **High Performance** - Built with Go for optimal performance
- 🎯 **MCP Standard** - Compatible with all MCP clients

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Kanboard instance with API access
- MCP-compatible client (Cursor, Claude Desktop, etc.)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/bivex/kanboard-mcp.git
   cd kanboard-mcp
   ```

2. **Build the executable:**

   **On Windows:**
   ```cmd
   build-release.bat
   ```

   **On Linux/macOS:**
   ```bash
   ./build-release.sh
   ```

   **Manual build:**
   ```bash
   go build -ldflags="-s -w" -o kanboard-mcp .
   ```

## ⚙️ Configuration

### 1. Environment Variables

Set up your Kanboard credentials using environment variables:

```bash
export KANBOARD_API_ENDPOINT="https://your-kanboard-url/jsonrpc.php"
export KANBOARD_API_KEY="your-kanboard-api-key"
export KANBOARD_USERNAME="your-kanboard-username"
export KANBOARD_PASSWORD="your-kanboard-password"
```

### 2. MCP Client Configuration

Create the MCP configuration file for your client:

**Location:**
- **Windows:** `C:\Users\YOUR_USERNAME\AppData\Roaming\Cursor\.cursor\mcp_config.json`
- **Linux/macOS:** `~/.cursor/mcp_config.json`

**Configuration:**
```json
{
  "mcpServers": {
    "kanboard-mcp-server": {
      "command": "/path/to/your/kanboard-mcp",
      "args": [],
      "env": {
        "KANBOARD_API_ENDPOINT": "https://your-kanboard-url/jsonrpc.php",
        "KANBOARD_API_KEY": "your-kanboard-api-key",
        "KANBOARD_USERNAME": "your-kanboard-username",
        "KANBOARD_PASSWORD": "your-kanboard-password"
      }
    }
  }
}
```

### 3. Restart Your Client

After saving the configuration, restart your MCP client (Cursor, Claude Desktop, etc.) for changes to take effect.

## 🛠️ Available Tools

### 📁 Project Management

| Tool | Description | Example |
|------|-------------|---------|
| `get_projects` | 📋 List all projects | "Show me all Kanboard projects" |
| `create_project` | ➕ Create new projects | "Create a project called 'Website Redesign'" |

### 📝 Task Management

| Tool | Description | Example |
|------|-------------|---------|
| `get_tasks` | 📋 Get project tasks | "Get tasks for 'Website Redesign' project" |
| `create_task` | ➕ Create new tasks | "Create task 'Design homepage' in 'Website Redesign'" |
| `update_task` | ✏️ Modify existing tasks | "Update task 123 with description 'New requirements'" |
| `delete_task` | 🗑️ Remove tasks | "Delete task with ID 456" |
| `get_task_details` | 🔍 Get detailed task info | "Get details for task 789" |
| `move_task` | ➡️ Move tasks between columns | "Move task 123 to 'Done' column" |
| `assign_task` | 👤 Assign tasks to users | "Assign the API task to John" |
| `set_task_due_date` | 📅 Set task deadlines | "Set due date for login task to 2024-01-15" |
| `add_task_comment` | 💬 Add task comments | "Add comment 'Testing completed' to task 123" |
| `get_task_comments` | 📝 Get task comments | "Show all comments on task 123" |

### 🏗️ Column Management

| Tool | Description | Example |
|------|-------------|---------|
| `get_columns` | 📋 List project columns | "Show me all columns in this project" |
| `create_column` | ➕ Add new columns | "Create a 'Testing' column with 5 task limit" |
| `update_column` | ✏️ Modify column settings | "Change the 'Review' column limit to 3 tasks" |
| `delete_column` | 🗑️ Remove columns | "Delete the unused 'Draft' column" |
| `reorder_columns` | 🔄 Change column positions | "Move 'Testing' column before 'Done'" |

### 🏷️ Category Management

| Tool | Description | Example |
|------|-------------|---------|
| `get_categories` | 📋 List project categories | "Show me all task categories" |
| `create_category` | ➕ Add task categories | "Create a 'Bug Fixes' category" |
| `update_category` | ✏️ Modify categories | "Rename 'Bug Fixes' to 'Critical Issues'" |
| `delete_category` | 🗑️ Remove categories | "Delete the unused 'Archive' category" |

### 🏊 Swimlane Management

| Tool | Description | Example |
|------|-------------|---------|
| `get_swimlanes` | 📋 List project swimlanes | "Show me all team swimlanes" |
| `create_swimlane` | ➕ Add team swimlanes | "Create a 'Frontend Team' swimlane" |
| `update_swimlane` | ✏️ Modify swimlanes | "Rename 'Mobile Team' to 'Cross-Platform Team'" |
| `delete_swimlane` | 🗑️ Remove swimlanes | "Delete the inactive 'Legacy Team' swimlane" |

### 👥 User Management

| Tool | Description | Example |
|------|-------------|---------|
| `get_users` | 📋 List all system users | "Show me all team members" |
| `get_user_by_name` | 🔍 Get user by name | "Find user 'john.doe'" |
| `create_user` | ➕ Create a new user | "Create user 'testuser' with password 'pass123'" |
| `update_user` | ✏️ Modify an existing user | "Update user 1 with email 'test@example.com'" |
| `remove_user` | 🗑️ Remove a user | "Remove user with ID 2" |
| `assign_user_to_project` | 🔗 Assign user to project | "Assign user 3 to project 10 as project-manager" |

## 📖 Usage Examples

### Project Workflow

```bash
# Create a new project
"Create a new project called 'Mobile App Development'"

# Add tasks to the project
"Create task 'Design UI mockups' in project 'Mobile App Development'"
"Create task 'Set up development environment' in project 'Mobile App Development'"

# Get all tasks
"Get tasks for 'Mobile App Development' project"

# Move tasks between columns
"Move task 1 to 'In Progress' column"
"Move task 2 to 'Done' column"
```

### Team Management

```bash
# Create a new team member
"Create user 'alice.smith' with password 'secure123' and email 'alice@company.com'"

# Assign user to project
"Assign user 'alice.smith' to project 'Mobile App Development' as project-member"

# Assign tasks to team members
"Assign task 1 to user 'alice.smith'"
```

### Task Organization

```bash
# Create categories for better organization
"Create category 'Critical Bugs'"
"Create category 'Feature Requests'"

# Add comments to tasks
"Add comment 'This needs urgent attention' to task 5"

# Set deadlines
"Set due date for task 3 to 2024-01-20"
```

## 🔧 Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/bivex/kanboard-mcp.git
cd kanboard-mcp

# Install dependencies
go mod download

# Build the application
go build -o kanboard-mcp .

# Run tests
go test ./...
```

### Project Structure

```
kanboard-mcp/
├── main.go              # Main application entry point
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── build-release.bat    # Windows build script
├── build-release.sh     # Unix build script
├── README.md            # This file
└── LICENSE.md           # License information
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

---

<div align="center">

**Made with ❤️ for the Kanboard community**

[![GitHub stars](https://img.shields.io/github/stars/bivex/kanboard-mcp?style=social)](https://github.com/bivex/kanboard-mcp/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/bivex/kanboard-mcp?style=social)](https://github.com/bivex/kanboard-mcp/network)
[![GitHub issues](https://img.shields.io/github/issues/bivex/kanboard-mcp)](https://github.com/bivex/kanboard-mcp/issues)

</div>
