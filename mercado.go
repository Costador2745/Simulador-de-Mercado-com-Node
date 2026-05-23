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
	Asset    string
	Quantity float64
	Price    float64
}

type Market struct {
	ID           int
	Price        map[string]float64
	Transactions []Order
	mu           sync.Mutex
}

var agentSaldo = map[int]float64{}
var agentAtivos = map[int]map[string]float64{}
var assets = []string{"AAPL", "GOOG", "AMZN"}
var accountMutex sync.Mutex

func agent(id int, orders chan<- Order, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 5; i++ {
		order := Order{
			AgentID:  id,
			Buy:      rand.Intn(2) == 0,
			Asset:    assets[rand.Intn(len(assets))],
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
		case order, ok := <-orders:
			if !ok {
				fmt.Printf("Node %d encerrando devido ao canal de ordens fechado.\n", m.ID)
				return
			}
			m.mu.Lock()
			if order.Quantity <= 0 {
				m.mu.Unlock()
				continue // ignora ordens inválidas
			}
			valid := false
			accountMutex.Lock()
			if order.Buy {
				// Verifica saldo e se o preço é >= preço atual do ativo
				if agentSaldo[order.AgentID] >= order.Price*order.Quantity && order.Price >= m.Price[order.Asset] {
					agentSaldo[order.AgentID] -= order.Price * order.Quantity
					agentAtivos[order.AgentID][order.Asset] += order.Quantity
					valid = true
				}
			} else {
				// Verifica ativos e se o preço é <= preço atual do ativo
				if agentAtivos[order.AgentID][order.Asset] >= order.Quantity && order.Price <= m.Price[order.Asset] {
					agentAtivos[order.AgentID][order.Asset] -= order.Quantity
					agentSaldo[order.AgentID] += order.Price * order.Quantity
					valid = true
				}
			}
			accountMutex.Unlock()
			if valid {
				m.Transactions = append(m.Transactions, order)
				m.Price[order.Asset] = order.Price
				fmt.Printf("Node %d processou ordem válida: %+v | Novo preço %s: %.2f\n", m.ID, order, order.Asset, m.Price[order.Asset])
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

func (m *Market) syncPrice(address string, asset string) {
	for retries := 0; retries < 3; retries++ { // ADICIONADO: retry automático

		conn, err := net.DialTimeout("tcp", address, 2*time.Second) // ADICIONADO
		if err != nil {

			fmt.Printf("[NODE %d] Erro ao conectar (%d/3): %v\n",
				m.ID,
				retries+1,
				err,
			)

			time.Sleep(time.Millisecond * 500)
			continue
		}

		defer conn.Close()

		conn.SetDeadline(time.Now().Add(2 * time.Second))

		m.mu.Lock()
		fmt.Fprintf(conn, "%s %.2f\n", asset, m.Price[asset])
		m.mu.Unlock()

		fmt.Printf("[NODE %d] Preço sincronizado com %s\n", m.ID, address)

		return
	}
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
			scanner := bufio.NewScanner(c)
			for scanner.Scan() {
				var asset string
				var price float64
				fmt.Sscanf(scanner.Text(), "%s %f", &asset, &price)
				node.mu.Lock()
				node.Price[asset] = (node.Price[asset] + price) / 2
				node.mu.Unlock()
				fmt.Printf("Node %d recebeu atualização de preço: %s = %.2f\n", node.ID, asset, node.Price[asset])
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("Node %d erro no scanner: %v\n", node.ID, err)
			}
		}(conn)
	}
}

func benchmarkOrders() {

	start := time.Now()

	var wg sync.WaitGroup

	orders := make(chan Order, 100)

	node := &Market{
		ID: 99,
		Price: map[string]float64{
			"AAPL": 50,
			"GOOG": 50,
			"AMZN": 50,
		},
	}

	done := make(chan bool, 1)

	go node.processOrders(orders, done)

	// simular carga grande
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			orders <- Order{
				AgentID:  id,
				Buy:      rand.Intn(2) == 0,
				Asset:    assets[rand.Intn(len(assets))],
				Quantity: float64(rand.Intn(5) + 1),
				Price:    float64(rand.Intn(50) + 1),
			}
		}(i)
	}

	wg.Wait()

	close(orders)

	done <- true

	elapsed := time.Since(start)

	fmt.Println("\n--- BENCHMARK ---")
	fmt.Printf("Tempo total: %s\n", elapsed)
	fmt.Printf("Ordens processadas: %d\n", 1000)
	fmt.Printf("Ordens por segundo: %.2f\n", float64(1000)/elapsed.Seconds())
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		agentSaldo[i] = 550.0 // cada agente começa com 550 de dinheiro
		agentAtivos[i] = make(map[string]float64)
		for _, asset := range assets {
			agentAtivos[i][asset] = 10.0
		}
	}

	node1 := &Market{ID: 1, Price: map[string]float64{"AAPL": 50, "GOOG": 50, "AMZN": 50}}
	node2 := &Market{ID: 2, Price: map[string]float64{"AAPL": 50, "GOOG": 50, "AMZN": 50}}

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
			for _, asset := range assets {
				node1.syncPrice("localhost:8002", asset)
				node2.syncPrice("localhost:8001", asset)
			}
		}
	}()

	wg.Wait()
	close(ordersnode1)
	close(ordersnode2)
	done1 <- true
	done2 <- true

	fmt.Println("\n--- Histórico Node 1 ---")
	for _, t := range node1.Transactions {
		fmt.Printf("%+v\n", t)
	}
	fmt.Println("\n--- Histórico Node 2 ---")
	for _, t := range node2.Transactions {
		fmt.Printf("%+v\n", t)
	}
	fmt.Println("\n--- Saldo e Ativos finais dos agentes ---")
	for i := 0; i < 3; i++ {
		fmt.Printf("Agente %d | Saldo: %.2f | ", i, agentSaldo[i])
		for _, asset := range assets {
			fmt.Printf("%s: %.2f ", asset, agentAtivos[i][asset])
		}
		fmt.Println()
	}

	fmt.Println("\n--- Preço final de cada ativo ---")
	for _, asset := range assets {
		fmt.Printf("%s | Node 1: %.2f | Node 2: %.2f\n", asset, node1.Price[asset], node2.Price[asset])
	}
	benchmarkOrders()
}
