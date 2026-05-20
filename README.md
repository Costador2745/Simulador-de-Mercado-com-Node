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

### Exemplo de output aleatorio

Node 2 rejeitou ordem inválida: {AgentID:1 Buy:false Asset:AAPL Quantity:5 Price:46}
Node 1 rejeitou ordem inválida: {AgentID:0 Buy:false Asset:AMZN Quantity:3 Price:6}
Node 1 rejeitou ordem inválida: {AgentID:2 Buy:false Asset:GOOG Quantity:3 Price:32}
Node 1 processou ordem válida: {AgentID:0 Buy:true Asset:GOOG Quantity:1 Price:9} | Novo preço GOOG: 9.00
Node 1 processou ordem válida: {AgentID:2 Buy:true Asset:AMZN Quantity:1 Price:13} | Novo preço AMZN: 13.00
Node 2 processou ordem válida: {AgentID:1 Buy:true Asset:AMZN Quantity:1 Price:20} | Novo preço AMZN: 20.00
Node 1 processou ordem válida: {AgentID:0 Buy:true Asset:AAPL Quantity:3 Price:9} | Novo preço AAPL: 9.00
Node 1 rejeitou ordem inválida: {AgentID:2 Buy:false Asset:AMZN Quantity:2 Price:23}
Node 2 processou ordem válida: {AgentID:1 Buy:true Asset:GOOG Quantity:1 Price:11} | Novo preço GOOG: 11.00
Node 1 rejeitou ordem inválida: {AgentID:0 Buy:false Asset:GOOG Quantity:4 Price:41}
Node 1 processou ordem válida: {AgentID:2 Buy:true Asset:AAPL Quantity:3 Price:11} | Novo preço AAPL: 11.00
Node 2 rejeitou ordem inválida: {AgentID:1 Buy:false Asset:AMZN Quantity:1 Price:5}
Node 1 rejeitou ordem inválida: {AgentID:2 Buy:false Asset:AMZN Quantity:1 Price:33}
Node 2 processou ordem válida: {AgentID:1 Buy:true Asset:AMZN Quantity:2 Price:30} | Novo preço AMZN: 30.00
Node 1 processou ordem válida: {AgentID:0 Buy:true Asset:AAPL Quantity:2 Price:47} | Novo preço AAPL: 47.00
Node 2 recebeu atualização de preço: AAPL = 48.50
Node 1 recebeu atualização de preço: AAPL = 47.75
Node 2 recebeu atualização de preço: GOOG = 10.00
Node 1 recebeu atualização de preço: GOOG = 9.50
Node 2 recebeu atualização de preço: AMZN = 21.50
Node 1 recebeu atualização de preço: AMZN = 17.25

--- Histórico Node 1 ---
Node 2 encerrando devido ao canal de ordens fechado.
Node 1 encerrando devido ao canal de ordens fechado.
{AgentID:0 Buy:true Asset:GOOG Quantity:1 Price:9}
{AgentID:2 Buy:true Asset:AMZN Quantity:1 Price:13}
{AgentID:0 Buy:true Asset:AAPL Quantity:3 Price:9}
{AgentID:2 Buy:true Asset:AAPL Quantity:3 Price:11}
{AgentID:0 Buy:true Asset:AAPL Quantity:2 Price:47}

--- Histórico Node 2 ---
{AgentID:1 Buy:true Asset:AMZN Quantity:1 Price:20}
{AgentID:1 Buy:true Asset:GOOG Quantity:1 Price:11}
{AgentID:1 Buy:true Asset:AMZN Quantity:2 Price:30}

--- Saldo e Ativos finais dos agentes ---
Agente 0 | Saldo: 492.00 | AAPL: 9.00 GOOG: 9.00 AMZN: 10.00 
Agente 1 | Saldo: 521.00 | AAPL: 10.00 GOOG: 9.00 AMZN: 11.00 
Agente 2 | Saldo: 530.00 | AAPL: 13.00 GOOG: 10.00 AMZN: 9.00 

--- Preço final de cada ativo ---
AAPL | Node 1: 47.75 | Node 2: 48.50
GOOG | Node 1: 9.50 | Node 2: 10.00
AMZN | Node 1: 17.25 | Node 2: 21.50