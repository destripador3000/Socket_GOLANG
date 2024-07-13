package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	PORT               = ":8080"
	HEADER             = 64
	FORMAT             = "utf-8"
	DISCONNECT_MESSAGE = "!DESCONECTAR"
)

type Server struct {
	clients  map[net.Conn]bool
	messages chan string
	mu       sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients:  make(map[net.Conn]bool),
		messages: make(chan string),
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	fmt.Printf("[NUEVA CONEXIÓN] %s conectado.\n", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	for {
		msgLength, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		msgLength = strings.TrimSpace(msgLength)
		msgLen := len(msgLength)
		if msgLen > HEADER {
			msgLen = HEADER
		}

		msg, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		msg = strings.TrimSpace(msg)
		if msg == DISCONNECT_MESSAGE {
			break
		}

		fmt.Printf("[MENSAJE] %s: %s\n", conn.RemoteAddr().String(), msg)
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
	fmt.Printf("[DESCONECTADO] %s desconectado.\n", conn.RemoteAddr().String())
}

func getServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error al obtener la IP del servidor:", err)
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "No se pudo obtener la IP del servidor"
}

func (s *Server) Start() {
	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}
	defer ln.Close()

	serverIP := getServerIP()
	fmt.Printf("[ESCUCHANDO] El servidor está escuchando en %s%s\n", serverIP, PORT)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error al aceptar la conexión:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func main() {
	server := NewServer()
	fmt.Println("[INICIANDO] El servidor está iniciando...")
	server.Start()
}
