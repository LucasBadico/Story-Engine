# LLM Executor + Parallel Phases Plan

## Objetivo
Centralizar chamadas de LLM em um único "entrypoint" para controlar:
- rate limit e concorrência por provider
- fila e backpressure
- uso por múltiplos módulos (workers, HTTP handlers, pipelines)

Além disso, paralelizar fases (Phase2/Phase3) sem quebrar limites de provider.

## Componentes
- LLMExecutor (novo pacote)
  - API: `Submit(ctx, Request) (Response, error)`
  - Interno: `scheduler`, `provider pools`, `worker loops`
  - Métricas: fila, tempo médio, falhas por provider

- ProviderAdapter (já existe: gemini/openai/ollama)
  - Precisará expor:
    - `ProviderName()`
    - `Capabilities()` (max output tokens, supports json mode, etc)

## Contratos
### Request
```
type LLMRequest struct {
  Provider string // "gemini", "openai", "ollama" ou "auto"
  Model string
  Prompt string
  MaxOutputTokens int
  Temperature float64
  JSONSchema string // opcional (para validação)
  RequestID string
  ResponseCh chan LLMResponse
  ErrorCh chan error
}
```

### Response
```
type LLMResponse struct {
  Text string
  Provider string
  Model string
  FinishReason string
  UsageTokens int
}
```

## Fluxo do Executor
1. `Submit` recebe `LLMRequest` e decide provider:
   - Provider explícito ou `auto` (estratégia futura).
2. Enfileira em `inCh` (buffer configurável).
3. Scheduler direciona para o pool do provider.
4. Worker do provider executa chamada e publica em `ResponseCh`.
5. Quem chamou lê **uma resposta** e fecha o channel.

## Concorrência / Limites
Config por provider:
```
LLM_EXECUTOR_PROVIDERS=gemini,openai,ollama
LLM_EXECUTOR_GEMINI_MAX_PARALLEL=4
LLM_EXECUTOR_GEMINI_QPS=2
LLM_EXECUTOR_OPENAI_MAX_PARALLEL=2
LLM_EXECUTOR_OLLAMA_MAX_PARALLEL=1
LLM_EXECUTOR_QUEUE_SIZE=100
```

## Paralelização das Fases
### Phase2 (extractors)
- Paralelizar por chunk **e** por tipo.
- Cada task chama `LLMExecutor.Submit`.
- Executor aplica limites do provider.

### Phase3 (matcher)
- Paralelizar por entidade.
- Cada match chama `LLMExecutor.Submit`.

### Deduplicação
- Mantém o merge final **após** completar todas as tasks.

## Backpressure
- Fila com buffer para não bloquear pequenas bursts.
- Se fila cheia:
  - modo default: erro rápido (HTTP 429 ou erro no worker).
  - modo alternativo: bloquear até vaga (config).

## Observabilidade
Log básico:
- request_id, provider, model, queue_len, latency_ms, finish_reason

## Integração
1. Criar `internal/platform/llm/executor` (ou `internal/ports/llm/executor`).
2. `gemini/openai/ollama` continuam como adapters.
3. Pipelines (extract/ingest) passam a depender **somente** do Executor.

## Impacto
- Reduz duplicação de rate limit/queue.
- Facilita adicionar novo provider (só registrar no Executor).
- Controla custo/latência e evita overload.

## Próximos passos
- Implementar Executor + provider registry.
- Substituir chamadas diretas nas fases (Phase1/2/3).
- Adicionar testes com providers mockados.
