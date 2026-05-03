package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dangrier/ocm-client/ocm"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start an MCP server exposing OCM tools",
	Long: `Starts a Model Context Protocol (MCP) server over stdio, exposing
QPS Online Crime Map data as tools for use with Claude and other MCP clients.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := server.NewMCPServer(
			"ocm",
			"1.0.0",
			server.WithToolCapabilities(true),
		)

		srv.AddTool(toolGetLocations(), handleGetLocations)
		srv.AddTool(toolGetLocation(), handleGetLocation)
		srv.AddTool(toolGetLocationByName(), handleGetLocationByName)
		srv.AddTool(toolGetOffences(), handleGetOffences)

		return server.ServeStdio(srv)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

// --- tool definitions ---

func toolGetLocations() mcp.Tool {
	return mcp.NewTool("get_locations",
		mcp.WithDescription("List Queensland Police Service crime map locations. Returns all known locations, optionally filtered by type. Use this to find location codes for use with get_offences."),
		mcp.WithString("type",
			mcp.Description("Filter by location type. One of: suburb, postcode, lga, nhw, region, district, patrol, division"),
			mcp.Enum("suburb", "postcode", "lga", "nhw", "region", "district", "patrol", "division"),
		),
	)
}

func toolGetLocation() mcp.Tool {
	return mcp.NewTool("get_location",
		mcp.WithDescription("Get a single QPS crime map location by its numeric code."),
		mcp.WithInteger("code",
			mcp.Required(),
			mcp.Description("Numeric location code"),
		),
	)
}

func toolGetLocationByName() mcp.Tool {
	return mcp.NewTool("get_location_by_name",
		mcp.WithDescription("Look up a QPS crime map location by type and name (case-insensitive). Returns the location code and details."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Location type. One of: suburb, postcode, lga, nhw, region, district, patrol, division"),
			mcp.Enum("suburb", "postcode", "lga", "nhw", "region", "district", "patrol", "division"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Location name, e.g. \"Brisbane\", \"4000\", \"Gold Coast\""),
		),
	)
}

func toolGetOffences() mcp.Tool {
	return mcp.NewTool("get_offences",
		mcp.WithDescription("Retrieve crime offences from the QPS Online Crime Map for specified locations and date range. Returns a list of offences with category, time, and coordinates."),
		mcp.WithArray("location_codes",
			mcp.Required(),
			mcp.Description("List of numeric location codes to query. Use get_locations or get_location_by_name to find codes."),
			mcp.WithStringItems(),
		),
		mcp.WithInteger("days",
			mcp.Description("Number of days to look back from today. Defaults to 90. Ignored if date_from and date_to are provided."),
			mcp.DefaultNumber(90),
		),
		mcp.WithString("date_from",
			mcp.Description("Start date in YYYY-MM-DD format. If provided, date_to is also required."),
		),
		mcp.WithString("date_to",
			mcp.Description("End date in YYYY-MM-DD format. If provided, date_from is also required."),
		),
	)
}

// --- handlers ---

func handleGetLocations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	typeStr := req.GetString("type", "")

	locs, err := s.client.GetLocations(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	list := make([]*ocm.Location, 0, len(locs))
	for _, l := range locs {
		if typeStr != "" {
			lt, err := ocm.ParseLocationType(typeStr)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if l.Type != lt {
				continue
			}
		}
		list = append(list, l)
	}

	return jsonResult(list)
}

func handleGetLocation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := req.RequireInt("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	l, err := s.client.GetLocation(ctx, code)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return jsonResult(l)
}

func handleGetLocationByName(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	typeStr, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lt, err := ocm.ParseLocationType(typeStr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	l, err := s.client.GetLocationByName(ctx, lt, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return jsonResult(l)
}

func handleGetOffences(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codeStrs := req.GetStringSlice("location_codes", nil)
	if len(codeStrs) == 0 {
		return mcp.NewToolResultError("location_codes is required and must not be empty"), nil
	}

	// Convert string codes to Location objects via client lookup.
	locations := make([]*ocm.Location, 0, len(codeStrs))
	for _, cs := range codeStrs {
		var code int
		if _, err := fmt.Sscan(cs, &code); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid location code %q: %v", cs, err)), nil
		}
		l, err := s.client.GetLocation(ctx, code)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		locations = append(locations, l)
	}

	dateFrom, dateTo, err := resolveDateRange(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	offences, err := s.client.GetOffences(ctx, dateFrom, dateTo, locations)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return jsonResult(offences)
}

// resolveDateRange parses date_from/date_to or falls back to the days parameter.
func resolveDateRange(req mcp.CallToolRequest) (time.Time, time.Time, error) {
	fromStr := req.GetString("date_from", "")
	toStr := req.GetString("date_to", "")

	if fromStr != "" || toStr != "" {
		if fromStr == "" || toStr == "" {
			return time.Time{}, time.Time{}, fmt.Errorf("both date_from and date_to must be provided together")
		}
		from, err := time.ParseInLocation("2006-01-02", fromStr, time.UTC)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid date_from %q: %v", fromStr, err)
		}
		to, err := time.ParseInLocation("2006-01-02", toStr, time.UTC)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid date_to %q: %v", toStr, err)
		}
		return from, to, nil
	}

	days := req.GetInt("days", 90)
	now := time.Now().UTC()
	return now.AddDate(0, 0, -days), now, nil
}

// jsonResult marshals v as JSON and returns a text tool result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
