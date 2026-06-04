# Simulador de Mercado em Go

Este projeto é uma simulação de um mercado financeiro distribuído desenvolvida em **Go**, com o objetivo de demonstrar conceitos de **concorrência**, **comunicação entre nós** e **benchmarking**.

O programa simula agentes que compram e vendem ativos financeiros em dois nós de mercado diferentes, que comunicam entre si através de TCP para sincronizar preços.

---

## Objetivo do Projeto

O objetivo principal é demonstrar, de forma prática, a utilização de:

- Goroutines
- Channels
- Mutexes
- Comunicação TCP
- Processamento concorrente de ordens
- Benchmark para análise de desempenho

---

## Lógica do Programa

O mercado trabalha com três ativos:

```txt
AAPL
GOOG
AMZN
```

Cada agente gera ordens aleatórias de compra ou venda. Essas ordens são enviadas para um nó do mercado através de um channel.

Cada nó valida a ordem recebida e decide se ela deve ser aceite ou rejeitada.

---

## Componentes Principais

### 1. Agentes

Os agentes são executados em goroutines.

Cada agente gera ordens com:

- ID do agente
- Tipo da ordem: compra ou venda
- Ativo
- Quantidade
- Preço

Exemplo de ordem:

```txt
{AgentID:0 Buy:true Asset:GOOG Quantity:3 Price:51}
```

Neste exemplo, o agente `0` tentou comprar `3` unidades de `GOOG` ao preço `51`.

---

### 2. Nós do Mercado

O programa utiliza dois nós principais:

```txt
Node 1
Node 2
```

Cada nó possui:

- Preço atual de cada ativo
- Histórico de transações válidas
- Processamento independente de ordens
- Servidor TCP para receber atualizações de preços

Cada nó funciona de forma independente, mas sincroniza os seus preços com o outro nó.

---

### 3. Validação de Ordens

Uma ordem só é processada se cumprir as regras do mercado.

#### Compra

Uma compra é válida se:

```txt
saldo do agente >= preço * quantidade
preço da ordem >= preço atual do ativo
```

#### Venda

Uma venda é válida se:

```txt
ativos do agente >= quantidade
preço da ordem <= preço atual do ativo
```

Se uma ordem não cumprir estas regras, é rejeitada.

As ordens rejeitadas não alteram:

- saldo do agente
- ativos do agente
- preço do ativo
- histórico de transações

---

## Sincronização entre Nós

Os nós comunicam entre si através de TCP.

Periodicamente, cada nó envia os seus preços ao outro nó.

Quando um nó recebe um preço externo, atualiza o seu preço usando a média:

```txt
preço atualizado = (preço local + preço recebido) / 2
```

Esta abordagem simula uma sincronização eventual entre nós distribuídos.

Os preços dos dois nós podem não ficar exatamente iguais, porque:

- cada nó processa ordens diferentes;
- a sincronização não é instantânea;
- o sistema usa média simples, não consenso perfeito.

---

## Concorrência

A concorrência é usada principalmente em três partes:

### Agentes

Cada agente corre numa goroutine e gera ordens em paralelo.

### Processamento de ordens

Cada nó processa ordens recebidas através de channels.

### Servidores TCP

Cada nó tem um servidor TCP ativo para receber atualizações de preços do outro nó.

Para evitar conflitos no acesso aos dados partilhados, o programa usa mutexes.

---

## Estrutura Principal do Código

### `agent()`

Gera ordens aleatórias de compra ou venda e envia essas ordens para um channel.

### `processOrders()`

Recebe ordens através do channel, valida cada ordem e atualiza o mercado se a ordem for válida.

### `syncPrice()`

Envia o preço atual de um ativo para outro nó através de TCP.

### `startServer()`

Inicia um servidor TCP que fica à espera de atualizações de preços vindas de outro nó.

### `benchmarkOrders()`

Executa um benchmark com 1000 ordens para medir o desempenho do sistema.

### `main()`

Inicializa:

- agentes;
- nós do mercado;
- preços iniciais;
- channels;
- goroutines;
- servidores TCP;
- benchmark final.

---

## Benchmark

No final da execução, o programa executa um benchmark com 1000 ordens.

Exemplo:

```txt
--- BENCHMARK ---
Tempo total: 3.0004ms
Ordens recebidas: 1000
Ordens válidas: 524
Ordens rejeitadas: 476
Ordens por segundo: 333288.89
```

As 1000 ordens do benchmark não são impressas individualmente para evitar um output demasiado grande e confuso.

Além disso, imprimir 1000 linhas no terminal iria afetar o tempo medido, porque escrever no terminal também consome tempo.

Por isso, o benchmark mostra apenas o resumo final:

- ordens recebidas;
- ordens válidas;
- ordens rejeitadas;
- tempo total;
- ordens por segundo.

---

## Exemplo de Output

```txt
Node 1 processou ordem válida: {AgentID:0 Buy:false Asset:AMZN Quantity:4 Price:50} | Novo preço AMZN: 50.00
Node 1 rejeitou ordem inválida: {AgentID:2 Buy:false Asset:AMZN Quantity:4 Price:53}
Node 2 rejeitou ordem inválida: {AgentID:1 Buy:false Asset:AMZN Quantity:2 Price:58}
Node 2 processou ordem válida: {AgentID:1 Buy:true Asset:GOOG Quantity:5 Price:53} | Novo preço GOOG: 53.00
Node 1 processou ordem válida: {AgentID:2 Buy:false Asset:GOOG Quantity:3 Price:41} | Novo preço GOOG: 41.00

[NODE 1] Preço sincronizado com localhost:8002
Node 2 recebeu atualização de preço: AAPL = 50.00
[NODE 2] Preço sincronizado com localhost:8001
Node 1 recebeu atualização de preço: AAPL = 50.00

--- Histórico Node 1 ---
{AgentID:0 Buy:false Asset:AMZN Quantity:4 Price:50}
{AgentID:0 Buy:true Asset:AMZN Quantity:2 Price:59}
{AgentID:2 Buy:false Asset:GOOG Quantity:3 Price:41}

--- Histórico Node 2 ---
{AgentID:1 Buy:true Asset:GOOG Quantity:5 Price:53}

--- Saldo e Ativos finais dos agentes ---
Agente 0 | Saldo: 479.00 | AAPL: 10.00 GOOG: 13.00 AMZN: 8.00
Agente 1 | Saldo: 285.00 | AAPL: 10.00 GOOG: 15.00 AMZN: 10.00
Agente 2 | Saldo: 713.00 | AAPL: 10.00 GOOG: 7.00 AMZN: 9.00

--- Preço final de cada ativo ---
AAPL | Node 1: 50.00 | Node 2: 50.00
GOOG | Node 1: 51.50 | Node 2: 52.00
AMZN | Node 1: 42.50 | Node 2: 45.00

--- BENCHMARK ---
Tempo total: 3.0004ms
Ordens recebidas: 1000
Ordens válidas: 524
Ordens rejeitadas: 476
Ordens por segundo: 333288.89
```

---

## Como Executar

No terminal, dentro da pasta do projeto:

```bash
go run mercado.go
```

---

## Observações Finais

Este projeto mostra uma simulação simples de um mercado distribuído.

A sincronização de preços entre nós é feita de forma simplificada, através de uma média entre o preço local e o preço recebido.

O objetivo não é criar um mercado financeiro real, mas sim demonstrar conceitos importantes de Go, como concorrência, channels, goroutines, mutexes, comunicação TCP e benchmark.
