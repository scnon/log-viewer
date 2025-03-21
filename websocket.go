package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client 表示一个WebSocket客户端连接
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// WebSocketServer WebSocket服务器结构
type WebSocketServer struct {
	upgrader  websocket.Upgrader
	clients   map[*Client]bool
	broadcast chan []byte
	mutex     sync.Mutex
	handler   func([]byte) ([]byte, error)
}

// NewWebSocketServer 创建新的WebSocket服务器
func NewWebSocketServer(handler func([]byte) ([]byte, error)) *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:   make(map[*Client]bool),
		broadcast: make(chan []byte),
		handler:   handler,
	}
}

// HandleWebSocket 处理WebSocket连接请求
func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级连接失败: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.mutex.Lock()
	s.clients[client] = true
	s.mutex.Unlock()

	log.Printf("新客户端连接: %s", conn.RemoteAddr().String())

	// 启动读取和写入的 goroutines
	go s.handleClientRead(client)
	go s.handleClientWrite(client)
}

// handleClientRead 处理客户端的读取消息
func (s *WebSocketServer) handleClientRead(client *Client) {
	defer func() {
		s.mutex.Lock()
		delete(s.clients, client)
		s.mutex.Unlock()
		close(client.send) // 关闭发送通道
		client.conn.Close()
		log.Printf("客户端断开连接: %s", client.conn.RemoteAddr().String())
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("读取消息错误: %v", err)
			}
			break
		}

		// 处理消息
		response, err := s.handler(message)
		if err != nil {
			log.Printf("处理消息错误: %v", err)
			continue
		}

		// 发送响应到客户端的发送通道
		client.send <- response
	}
}

// handleClientWrite 处理客户端的写入消息
func (s *WebSocketServer) handleClientWrite(client *Client) {
	defer func() {
		client.conn.Close()
	}()

	for message := range client.send {
		err := client.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("发送消息失败: %v", err)
			return
		}
	}

	// 通道关闭时发送关闭消息
	client.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// BroadcastMessages 广播消息给所有连接的客户端
func (s *WebSocketServer) BroadcastMessages() {
	for {
		message := <-s.broadcast
		s.mutex.Lock()
		for client := range s.clients {
			select {
			case client.send <- message:
			default:
				// 如果客户端的发送缓冲区已满，关闭连接
				close(client.send)
				delete(s.clients, client)
				client.conn.Close()
			}
		}
		s.mutex.Unlock()
	}
}

// Start 启动WebSocket服务器
func (s *WebSocketServer) Start(addr string) error {
	go s.BroadcastMessages()
	http.HandleFunc("/ws", s.HandleWebSocket)
	log.Printf("WebSocket服务器启动在 %s", addr)
	return http.ListenAndServe(addr, nil)
}
