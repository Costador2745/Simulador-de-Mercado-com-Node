package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Order struct {
	AgentID  int
	Buy      bool // true para compra, false para venda
	Quantity float64
	Price    float64
}

type Market struct {
	ID           int
	Price        float64
	Transactions []Order
	mu           sync.Mutex
}

var agentSaldo = map[int]float64{}
var agentAtivos = map[int]float64{}

func agent(id int, orders chan<- Order, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 5; i++ {
		order := Order{
			AgentID:  id,
			Buy:      rand.Intn(2) == 0,
			Quantity: float64(rand.Intn(5) + 1),
			Price:    float64(rand.Intn(50) + 1),
		}
		orders <- order
		time.Sleep(time.Millisecond * 500)
	}
}

func (m *Market) processOrders(orders <-chan Order, done chan bool) {
	for {
		select {
		case order := <-orders:
			m.mu.Lock()
			if order.Quantity <= 0 {
				continue // ignora ordens inválidas
			}
			valid := false
			if order.Buy {
				// Verifica saldo
				if agentSaldo[order.AgentID] >= order.Price*order.Quantity {
					agentSaldo[order.AgentID] -= order.Price * order.Quantity
					agentAtivos[order.AgentID] += order.Quantity
					valid = true
				}
			} else {
				// Verifica ativos
				if agentAtivos[order.AgentID] >= order.Quantity {
					agentAtivos[order.AgentID] -= order.Quantity
					agentSaldo[order.AgentID] += order.Price * order.Quantity
					valid = true
				}
			}
			if valid {
				m.Transactions = append(m.Transactions, order)
				// Ajusta preço baseado na ordem válida
				if order.Buy {
					m.Price = order.Price
				} else {
					m.Price = order.Price
				}
				fmt.Printf("Node %d processou ordem válida: %+v | Novo preço: %.2f\n", m.ID, order, m.Price)
			} else {
				fmt.Printf("Node %d rejeitou ordem inválida: %+v\n", m.ID, order)
			}
			m.mu.Unlock()
		case <-done:
			fmt.Printf("Node %d encerrando.\n", m.ID)
			return
		}
	}
}

func (m *Market) syncPrice(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()
	m.mu.Lock()
	fmt.Fprintf(conn, "%.2f\n", m.Price)
	m.mu.Unlock()
}

func startServer(node *Market, port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			priceStr, _ := bufio.NewReader(c).ReadString('\n')
			var receivedPrice float64
			fmt.Sscanf(priceStr, "%f", &receivedPrice)
			node.mu.Lock()
			node.Price = (node.Price + receivedPrice) / 2
			node.mu.Unlock()
			fmt.Printf("Node %d recebeu preço %.2f, novo preço: %.2f\n", node.ID, receivedPrice, node.Price)
		}(conn)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		agentSaldo[i] = 550.0 // cada agente começa com 550 de dinheiro
		agentAtivos[i] = 10.0 // cada agente começa com 10 ações
	}

	node1 := &Market{ID: 1, Price: 50.0}
	node2 := &Market{ID: 2, Price: 50.0}

	ordersnode1 := make(chan Order, 10)
	ordersnode2 := make(chan Order, 10)

	go startServer(node1, "8001")
	go startServer(node2, "8002")
	done1 := make(chan bool, 1)
	done2 := make(chan bool, 1)

	go node1.processOrders(ordersnode1, done1)
	go node2.processOrders(ordersnode2, done2)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		if i%2 == 0 {
			go agent(i, ordersnode1, &wg)
		} else {
			go agent(i, ordersnode2, &wg)
		}
	}

	go func() {
		for {
			time.Sleep(time.Second * 2)
			node1.syncPrice("localhost:8002")
			node2.syncPrice("localhost:8001")
		}
	}()

	wg.Wait()
	close(ordersnode1)
	close(ordersnode2)
	done1 <- true
	done2 <- true

	fmt.Println("\n=== Histórico Node 1 ===")
	for _, t := range node1.Transactions {
		fmt.Printf("%+v\n", t)
	}
	fmt.Println("\n=== Histórico Node 2 ===")
	for _, t := range node2.Transactions {
		fmt.Printf("%+v\n", t)
	}
	fmt.Printf("\nPreço final Node 1: %.2f | Node 2: %.2f\n", node1.Price, node2.Price)
}
