package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// MCPTool describes a tool advertised by an MCP server via tools/list.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// MCPContentItem represents a single content entry returned from tools/call.
type MCPContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int64      `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// MCPClient communicates with a remote or bridged Streamable HTTP MCP server.
type MCPClient struct {
	Endpoint   string
	Headers    map[string]string
	HTTPClient *http.Client
	sessionID  string
	reqCounter int64
}

// NewMCPClient constructs an MCP client for the given HTTP endpoint and auth headers.
func NewMCPClient(endpoint string, headers map[string]string) *MCPClient {
	hdrs := make(map[string]string, len(headers))
	for k, v := range headers {
		hdrs[k] = v
	}
	return &MCPClient{
		Endpoint: strings.TrimSpace(endpoint),
		Headers:  hdrs,
		HTTPClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *MCPClient) nextID() int64 {
	return atomic.AddInt64(&c.reqCounter, 1)
}

func (c *MCPClient) doRPC(ctx context.Context, method string, params interface{}, isNotification bool) (json.RawMessage, error) {
	if c.Endpoint == "" {
		return nil, fmt.Errorf("MCP endpoint URL is required")
	}

	reqPayload := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	var reqID int64
	if !isNotification {
		reqID = c.nextID()
		reqPayload.ID = &reqID
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("invalid MCP endpoint URL: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if c.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	for k, v := range c.Headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP request failed: %w", err)
	}
	defer resp.Body.Close()

	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.sessionID = sid
	}

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("MCP authentication failed (HTTP %d): check your API key / Bearer token", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(respBytes))
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		return nil, fmt.Errorf("MCP server returned HTTP %d: %s", resp.StatusCode, snippet)
	}

	if isNotification || len(bytes.TrimSpace(respBytes)) == 0 {
		return nil, nil
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	var rpcResp *jsonRPCResponse
	if strings.Contains(contentType, "text/event-stream") || bytes.Contains(respBytes, []byte("data:")) && !strings.HasPrefix(strings.TrimSpace(string(respBytes)), "{") {
		rpcResp, err = parseMCPSSE(respBytes, reqID)
	} else {
		var single jsonRPCResponse
		err = json.Unmarshal(respBytes, &single)
		if err == nil {
			rpcResp = &single
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode MCP JSON-RPC response: %w", err)
	}
	if rpcResp == nil {
		return nil, fmt.Errorf("empty MCP JSON-RPC response")
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("MCP error (%d): %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func parseMCPSSE(raw []byte, expectedID int64) (*jsonRPCResponse, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	var dataBuf strings.Builder
	var lastResp *jsonRPCResponse

	flushEvent := func() {
		payload := strings.TrimSpace(dataBuf.String())
		dataBuf.Reset()
		if payload == "" || payload == "[DONE]" {
			return
		}
		var candidate jsonRPCResponse
		if err := json.Unmarshal([]byte(payload), &candidate); err == nil {
			if candidate.Result != nil || candidate.Error != nil {
				lastResp = &candidate
			}
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			flushEvent()
			continue
		}
		if strings.HasPrefix(line, "data:") {
			chunk := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(chunk)
		}
	}
	flushEvent()

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if lastResp == nil {
		return nil, fmt.Errorf("no JSON-RPC payload found in SSE stream (id=%d)", expectedID)
	}
	return lastResp, nil
}

// Initialize performs the MCP initialize handshake and sends notifications/initialized.
func (c *MCPClient) Initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]string{
			"name":    "clarifai-note-taker",
			"version": "1.0.0",
		},
	}
	if _, err := c.doRPC(ctx, "initialize", params, false); err != nil {
		return err
	}
	// Best-effort initialized notification
	_, _ = c.doRPC(ctx, "notifications/initialized", map[string]interface{}{}, true)
	return nil
}

// ListTools lists all tools exposed by the MCP server.
func (c *MCPClient) ListTools(ctx context.Context) ([]MCPTool, error) {
	raw, err := c.doRPC(ctx, "tools/list", map[string]interface{}{}, false)
	if err != nil {
		return nil, err
	}
	var res struct {
		Tools []MCPTool `json:"tools"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("failed to parse tools/list response: %w", err)
	}
	return res.Tools, nil
}

// CallTool invokes a named MCP tool and returns its extracted text contents.
func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	if args == nil {
		args = map[string]interface{}{}
	}
	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}
	raw, err := c.doRPC(ctx, "tools/call", params, false)
	if err != nil {
		return "", err
	}

	var res struct {
		Content []MCPContentItem `json:"content"`
		IsError bool             `json:"isError"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return string(raw), nil
	}

	var parts []string
	for _, item := range res.Content {
		if strings.TrimSpace(item.Text) != "" {
			parts = append(parts, item.Text)
		}
	}
	combined := strings.Join(parts, "\n")
	if res.IsError {
		return "", fmt.Errorf("MCP tool %q returned error: %s", name, combined)
	}
	return combined, nil
}
