package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"ccdemo/src/pkg/logger"
	"ccdemo/src/pkg/storage"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// Server wraps the MCP server and provides hot search tools.
type Server struct {
	store  *storage.Storage
	server *server.MCPServer
}

// NewServer creates a new MCP server with registered tools.
func NewServer(store *storage.Storage) *Server {
	s := &Server{
		store: store,
	}

	s.server = server.NewMCPServer("hotsearch-bot", "1.0.0")

	s.server.AddTool(
		mcp.NewTool("get_hot_searches",
			mcp.WithDescription("Get today's aggregated hot search top 20"),
		),
		s.handleGetHotSearches,
	)

	s.server.AddTool(
		mcp.NewTool("get_hot_searches_by_platform",
			mcp.WithDescription("Get hot searches for a specific platform"),
			mcp.WithString("platform", mcp.Description("Platform name"), mcp.Required()),
		),
		s.handleGetHotSearchesByPlatform,
	)

	s.server.AddTool(
		mcp.NewTool("get_platforms",
			mcp.WithDescription("Get list of supported platforms"),
		),
		s.handleGetPlatforms,
	)

	return s
}

// Serve starts the MCP server using stdio transport.
func (s *Server) Serve() error {
	return server.ServeStdio(s.server)
}

func (s *Server) handleGetHotSearches(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	items, err := s.store.ListAll()
	if err != nil {
		logger.Error("list all hot searches failed", zapError(err))
		return nil, fmt.Errorf("list all hot searches: %w", err)
	}

	data, err := json.Marshal(items)
	if err != nil {
		logger.Error("marshal hot searches failed", zapError(err))
		return nil, fmt.Errorf("marshal hot searches: %w", err)
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleGetHotSearchesByPlatform(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	platform, err := request.RequireString("platform")
	if err != nil {
		return mcp.NewToolResultErrorf("missing required argument 'platform': %v", err), nil
	}

	items, err := s.store.ListByPlatform(platform)
	if err != nil {
		logger.Error("list hot searches by platform failed", zapError(err))
		return nil, fmt.Errorf("list hot searches by platform: %w", err)
	}

	data, err := json.Marshal(items)
	if err != nil {
		logger.Error("marshal hot searches failed", zapError(err))
		return nil, fmt.Errorf("marshal hot searches: %w", err)
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleGetPlatforms(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	platforms := []string{"weibo", "baidu", "zhihu", "douyin"}
	data, err := json.Marshal(platforms)
	if err != nil {
		logger.Error("marshal platforms failed", zapError(err))
		return nil, fmt.Errorf("marshal platforms: %w", err)
	}

	return mcp.NewToolResultText(string(data)), nil
}

// zapError creates a zap.Field from an error for structured logging.
func zapError(err error) zap.Field {
	return zap.String("error", err.Error())
}
