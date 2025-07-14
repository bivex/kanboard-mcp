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
            "KANBOARD_API_KEY": "your-kanboard-api-key"
          }
        }
      }
    }
    ```

3.  **Restart Cursor:**
    After saving the `mcp_config.json` file, restart your Cursor application for the changes to take effect.

Once configured, Cursor will be able to discover and interact with the `hello_world` tool exposed by this MCP server.
