package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
)

const (
	parseError     = -32700
	invalidRequest = -32600
	methodNotFound = -32601
	internalError  = -32603
)

type Server struct {
	reader      *messageReader
	writer      *messageWriter
	initialized bool
}

type rpcMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   json.RawMessage  `json:"error,omitempty"`
}

func NewServer(input io.Reader, output io.Writer) *Server {
	return &Server{
		reader: newMessageReader(input),
		writer: newMessageWriter(output),
	}
}

func (s *Server) Run() error {
	for {
		payload, err := s.reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			log.Printf("[mcp] transport error: %v", err)
			return err
		}

		var msg rpcMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("[mcp] failed to parse payload: %v", err)
			if err := s.respondWithError(nil, parseError, "Unable to parse JSON payload"); err != nil {
				return err
			}
			continue
		}

		if msg.JSONRPC != "2.0" {
			log.Printf("[mcp] dropping message with unsupported jsonrpc version: %q", msg.JSONRPC)
			continue
		}

		if msg.Method != "" {
			if msg.ID != nil {
				if err := s.handleRequest(&msg); err != nil {
					log.Printf("[mcp] request handling error: %v", err)
				}
			} else {
				s.handleNotification(&msg)
			}
			continue
		}

		if len(msg.Result) != 0 || len(msg.Error) != 0 {
			continue
		}

		log.Printf("[mcp] unrecognized message shape: %s", string(payload))
	}
}

func (s *Server) handleRequest(msg *rpcMessage) error {
	if msg.Method == "initialize" && !s.initialized {
		s.initialized = true
		capabilities := map[string]any{
			"capabilities": s.makeCapabilities(),
			"serverInfo": map[string]string{
				"name":    "go-mcp-server",
				"version": "0.1.0",
			},
		}
		return s.respond(msg.ID, capabilities)
	}
	if !s.initialized {
		return s.respondWithError(msg.ID, invalidRequest, "Server has not been initialized")
	}
	switch msg.Method {
		case "ping":
			return s.respond(msg.ID, map[string]string{"result": "pong"})
		case "tools/list":
			return s.respond(msg.ID, s.listTools())
		case "tools/call":
			if len(msg.Params) == 0 {
				return s.respondWithError(msg.ID, invalidRequest, "Missing params for tools/call")
			}
			result, err := s.callTool(msg.Params)
			if err != nil {
				return s.respondWithError(msg.ID, internalError, err.Error())
			}
			return s.respond(msg.ID, result)
		case "shutdown":
			return s.respond(msg.ID, map[string]any{})
		default:
			return s.respondWithError(msg.ID, methodNotFound, fmt.Sprintf("Method not implemented: %s", msg.Method))
	}
}

func (s *Server) handleNotification(msg *rpcMessage) {
	switch msg.Method {
	case "notifications/initialized":
		log.Printf("[mcp] client signalled that initialization is complete")
	default:
		log.Printf("[mcp] ignoring notification: %s", msg.Method)
	}
}

func (s *Server) respond(id *json.RawMessage, result any) error {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"result":  result,
	}
	if id == nil {
		payload["id"] = nil
	} else {
		payload["id"] = json.RawMessage(*id)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.writer.Write(data)
}

func (s *Server) respondWithError(id *json.RawMessage, code int, message string) error {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
	if id == nil {
		payload["id"] = nil
	} else {
		payload["id"] = json.RawMessage(*id)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.writer.Write(data)
}

func (s *Server) makeCapabilities() map[string]any {
	return map[string]any{
		"tools": map[string]bool{
			"list": true,
			"call": true,
		},
	}
}

func (s *Server) listTools() map[string]any {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        "echo",
				"description": "Return the same text that the caller provides.",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text": map[string]any{
							"type":        "string",
							"description": "Text to echo back to the caller.",
						},
					},
					"required": []string{"text"},
				},
			},
		},
	}
}

func (s *Server) callTool(params json.RawMessage) (map[string]any, error) {
	var request struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, err
	}
	if request.Name != "echo" {
		return nil, fmt.Errorf("Unknown tool: %s", request.Name)
	}
	text, _ := request.Arguments["text"].(string)
	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		},
	}, nil
}
