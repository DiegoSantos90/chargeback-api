# 🧪 Guia de Teste - Chargeback API

Este guia fornece instruções passo a passo para testar a API Chargeback localmente usando DynamoDB Local e ferramentas como Insomnia ou Postman.

## 📋 Pré-requisitos

- Docker instalado
- Go 1.21+ instalado
- AWS CLI instalado
- Insomnia ou Postman instalado

## 🚀 Passo a Passo

### Passo 1: Preparar o Ambiente

#### 1.1 Clone o projeto (se ainda não fez)
```bash
git clone https://github.com/DiegoSantos90/chargeback-api.git
cd chargeback-api
```

#### 1.2 Instalar dependências
```bash
make deps
```

### Passo 2: Configurar DynamoDB Local

#### 2.1 Iniciar DynamoDB Local via Docker
```bash
make setup-local-db
```

Ou manualmente:
```bash
docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
```

#### 2.2 Criar a tabela de chargebacks
```bash
make create-table
```

#### 2.3 Verificar se a tabela foi criada
```bash
aws dynamodb list-tables --endpoint-url http://localhost:8000
```

**Saída esperada:**
```json
{
    "TableNames": [
        "chargebacks"
    ]
}
```

### Passo 3: Configurar Variáveis de Ambiente

#### 3.1 Exportar variáveis necessárias
```bash
export PORT=8080
export AWS_REGION=us-east-1
export DYNAMODB_ENDPOINT=http://localhost:8000
export DYNAMODB_TABLE=chargebacks
```

#### 3.2 Verificar configuração
```bash
echo "Port: $PORT"
echo "Region: $AWS_REGION" 
echo "DynamoDB Endpoint: $DYNAMODB_ENDPOINT"
echo "Table: $DYNAMODB_TABLE"
```

### Passo 4: Iniciar a API

#### 4.1 Executar em modo desenvolvimento
```bash
make dev
```

Ou diretamente:
```bash
go run cmd/api/main.go
```

#### 4.2 Verificar se a API está funcionando
```bash
curl http://localhost:8080/health
```

**Saída esperada:**
```json
{
  "service":"chargeback-api",  
  "status": "ok",
  "timestamp": "2024-10-15T12:00:00Z"
}
```

### Passo 5: Testar com Insomnia/Postman

#### 5.1 Configuração Base

**Base URL:** `http://localhost:8080`

**Headers necessários:**
- `Content-Type: application/json`

#### 5.2 Endpoint: Health Check

**Método:** `GET`
**URL:** `http://localhost:8080/health`

**Resposta esperada:**
```json
{
  "status": "healthy",
  "timestamp": "2024-10-15T12:00:00Z"
}
```

#### 5.3 Endpoint: Criar Chargeback

**Método:** `POST`
**URL:** `http://localhost:8080/chargebacks`
**Headers:** `Content-Type: application/json`

## 📦 Massa de Teste

### Caso 1: Chargeback por Fraude (Sucesso)

```json
{
  "transaction_id": "txn_001_2024_fraud",
  "merchant_id": "merchant_amazon_br",
  "amount": 299.99,
  "currency": "BRL",
  "card_number": "4111111111111111",
  "reason": "fraud",
  "description": "Transação não autorizada identificada pelo cliente",
  "transaction_date": "2024-10-14T15:30:00Z"
}
```

**Resposta esperada (201 Created):**
```json
{
  "id": "cb_1729001234567890123",
  "transaction_id": "txn_001_2024_fraud",
  "merchant_id": "merchant_amazon_br",
  "amount": 299.99,
  "currency": "BRL",
  "card_number": "****-****-****-1111",
  "reason": "fraud",
  "status": "pending",
  "description": "Transação não autorizada identificada pelo cliente",
  "transaction_date": "2024-10-14T15:30:00Z",
  "chargeback_date": "2024-10-15T12:00:00Z",
  "created_at": "2024-10-15T12:00:00Z",
  "updated_at": "2024-10-15T12:00:00Z"
}
```

### Caso 2: Chargeback por Cobrança Duplicada

```json
{
  "transaction_id": "txn_002_2024_duplicate",
  "merchant_id": "merchant_shopify_store",
  "amount": 89.90,
  "currency": "USD",
  "card_number": "5555555555554444",
  "reason": "duplicate",
  "description": "Cliente foi cobrado duas vezes pelo mesmo produto",
  "transaction_date": "2024-10-13T09:15:00Z"
}
```

### Caso 3: Chargeback por Produto Não Recebido

```json
{
  "transaction_id": "txn_003_2024_not_received",
  "merchant_id": "merchant_ecommerce_xyz",
  "amount": 1299.00,
  "currency": "BRL",
  "card_number": "4000000000000002",
  "reason": "product_not_received",
  "description": "Produto não foi entregue após 30 dias da compra",
  "transaction_date": "2024-09-15T14:22:00Z"
}
```

### Caso 4: Chargeback por Assinatura Indevida

```json
{
  "transaction_id": "txn_004_2024_subscription",
  "merchant_id": "merchant_streaming_service",
  "amount": 29.90,
  "currency": "BRL",
  "card_number": "4242424242424242",
  "reason": "subscription",
  "description": "Cobrança de assinatura após cancelamento",
  "transaction_date": "2024-10-01T00:05:00Z"
}
```

### Caso 5: Chargeback por Crédito Não Processado

```json
{
  "transaction_id": "txn_005_2024_credit",
  "merchant_id": "merchant_travel_agency",
  "amount": 2500.00,
  "currency": "USD",
  "card_number": "4000000000000069",
  "reason": "credit_not_processed",
  "description": "Reembolso prometido não foi processado",
  "transaction_date": "2024-09-20T11:45:00Z"
}
```

## 🔧 Casos de Teste de Erro

### Erro 1: Transação Duplicada

Tente criar um chargeback com o mesmo `transaction_id` de um caso anterior:

```json
{
  "transaction_id": "txn_001_2024_fraud",
  "merchant_id": "merchant_another",
  "amount": 100.00,
  "currency": "USD",
  "card_number": "4111111111111111",
  "reason": "fraud",
  "transaction_date": "2024-10-15T10:00:00Z"
}
```

**Resposta esperada (400 Bad Request):**
```json
{
  "error": "chargeback already exists for transaction txn_001_2024_fraud"
}
```

### Erro 2: Dados Inválidos - Transaction ID Vazio

```json
{
  "transaction_id": "",
  "merchant_id": "merchant_test",
  "amount": 100.00,
  "currency": "USD",
  "card_number": "4111111111111111",
  "reason": "fraud",
  "transaction_date": "2024-10-15T10:00:00Z"
}
```

**Resposta esperada (400 Bad Request):**
```json
{
  "error": "validation errors: transaction ID is required"
}
```

### Erro 3: Reason Inválido

```json
{
  "transaction_id": "txn_invalid_reason",
  "merchant_id": "merchant_test",
  "amount": 100.00,
  "currency": "USD",
  "card_number": "4111111111111111",
  "reason": "invalid_reason",
  "transaction_date": "2024-10-15T10:00:00Z"
}
```

### Erro 4: Método HTTP Inválido

**Método:** `GET`
**URL:** `http://localhost:8080/chargebacks`

**Resposta esperada (405 Method Not Allowed):**
```json
{
  "error": "Method not allowed"
}
```

### Erro 5: Content-Type Inválido

**Método:** `POST`
**URL:** `http://localhost:8080/chargebacks`
**Headers:** `Content-Type: text/plain`

**Resposta esperada (415 Unsupported Media Type):**
```json
{
  "error": "Content-Type must be application/json"
}
```

## 📊 Validação dos Dados

### Campos Obrigatórios
- `transaction_id`: String não vazia
- `merchant_id`: String não vazia
- `amount`: Número maior que 0
- `currency`: String não vazia
- `card_number`: String não vazia
- `reason`: Um dos valores válidos (fraud, duplicate, subscription, product_not_received, credit_not_processed)
- `transaction_date`: Data válida no formato ISO 8601

### Valores de Reason Válidos
- `fraud`: Transação fraudulenta
- `duplicate`: Cobrança duplicada
- `subscription`: Disputa relacionada à assinatura
- `product_not_received`: Produto não recebido
- `credit_not_processed`: Crédito não processado

### Status Automático
- Todos os chargebacks são criados com status `pending`

## 🏃‍♂️ Scripts de Automação

### Script para Teste Rápido (Bash)

Crie um arquivo `test_api.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "🔍 Testando Health Check..."
curl -s "$BASE_URL/health" | jq

echo -e "\n📝 Criando Chargeback de Teste..."
curl -s -X POST "$BASE_URL/chargebacks" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "txn_script_test_'$(date +%s)'",
    "merchant_id": "merchant_test_script",
    "amount": 99.99,
    "currency": "USD",
    "card_number": "4111111111111111",
    "reason": "fraud",
    "description": "Teste automatizado via script",
    "transaction_date": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
  }' | jq

echo -e "\n✅ Teste concluído!"
```

Execute:
```bash
chmod +x test_api.sh
./test_api.sh
```

## 🔍 Verificação no DynamoDB

### Listar todos os chargebacks criados

```bash
aws dynamodb scan \
  --table-name chargebacks \
  --endpoint-url http://localhost:8000
```

### Buscar chargeback específico por ID

```bash
aws dynamodb get-item \
  --table-name chargebacks \
  --key '{"id":{"S":"SEU_CHARGEBACK_ID_AQUI"}}' \
  --endpoint-url http://localhost:8000
```

## 🛑 Limpeza do Ambiente

### Parar a API
```bash
# Ctrl+C no terminal onde a API está rodando
```

### Parar e remover DynamoDB Local
```bash
make stop-local-db
```

Ou manualmente:
```bash
docker stop dynamodb-local
docker rm dynamodb-local
```

## 📝 Dicas Importantes

1. **Porta ocupada**: Se a porta 8080 estiver ocupada, altere a variável `PORT`
2. **DynamoDB Local**: Dados são perdidos quando o container é removido
3. **Transaction ID único**: Cada chargeback deve ter um `transaction_id` único
4. **Formato de data**: Use sempre formato ISO 8601 (`YYYY-MM-DDTHH:mm:ssZ`)
5. **Card number masking**: O número do cartão é automaticamente mascarado na resposta

## 🚨 Troubleshooting

### Problema: "Connection refused" na API
**Solução:** Verificar se a API está rodando na porta correta

### Problema: "UnknownOperationException" no DynamoDB
**Solução:** Verificar se o DynamoDB Local está rodando

### Problema: "ResourceNotFoundException" 
**Solução:** Verificar se a tabela foi criada corretamente

### Problema: Chargeback duplicado
**Solução:** Usar um `transaction_id` diferente

Este guia deve cobrir todos os cenários de teste da sua API Chargeback! 🎉