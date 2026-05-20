# Distributed Market Simulator in Go

Este projeto é uma simulação de mercado financeiro distribuído desenvolvida em Go, que demonstra conceitos de **concorrência** e **distribuição** de forma prática.

## Lógica do Código

O código simula um mercado com múltiplos ativos (`AAPL`, `GOOG`, `AMZN`) e múltiplos agentes que vão comprar e vender esses ativos ao mesmo tempo.

### Componentes principais

1. **Agentes (goroutines)**  
   - Cada agente envia ordens de compra ou venda para um nó(loja) do mercado.  
   - As ordens são geradas aleatoriamente, incluindo:  
     - Tipo (`Buy` ou `Sell`)  
     - Ativo (`AAPL`, `GOOG`, `AMZN`)  
     - Quantidade  
     - Preço por unidade  

2. **Nós do Mercado (MarketNode)**  
   - Cada nó mantém:
     - Preço atual de cada ativo (`map[string]float64`)
     - Histórico de ordens processadas
   - Processa as ordens recebidas via **canais** (`channels`) de forma concorrente, atualizando:
     - Saldo e ativos dos agentes
     - Preço do ativo
     - Histórico de transações

3. **Validação de ordens**  
   - Uma ordem é considerada válida se:
     - Compra: o agente tem saldo suficiente **e o preço da ordem ≥ preço atual do ativo**.
     - Venda: o agente tem ativos suficientes **e o preço da ordem ≤ preço atual do ativo**.
   - Ordens inválidas são rejeitadas e não afetam o histórico nem os preços.

4. **Sincronização entre nós**  
   - Cada nó envia periodicamente o preço de seus ativos para o outro nó via TCP.
   - O outro nó atualiza seu preço calculando a média:
     ```
     Preço atualizado = (Preço local + Preço recebido) / 2
     ```
   - Isso mantém os preços aproximados entre os nós e simula um mercado distribuído.

5. **Histórico e estado final**  
   - Cada nó mantém um histórico apenas das ordens válidas.  
   - No final da simulação, o código imprime:
     - Histórico de transações de cada nó  
     - Saldo e ativos finais de cada agente  
     - Preço final de cada ativo em cada nó

### Estrutura do Código

- `agent()` → goroutine que gera e envia ordens para um nó
- `processOrders()` → goroutine do nó que processa ordens recebidas
- `syncPrices()` → envia preços atuais para o outro nó
- `startServer()` → servidor TCP que recebe preços de outro nó
- `main()` → inicializa agentes, nós, canais e goroutines, e coordena a execução

### Concorrência e Distribuição

- Concorrência: cada agente é executado em **goroutine**, permitindo múltiplas ordens ao mesmo tempo.  
- Distribuição: dois nós independentes que sincronizam preços via TCP, mostrando comunicação entre processos diferentes.  

### Observações

- Pode ser expandido para incluir:
  - Agentes inteligentes que usam histórico para decidir ordens
  - Mais ativos ou nós
  - Benchmarking e análise de desempenho