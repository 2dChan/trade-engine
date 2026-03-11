# Trade Engine - Техническая спецификация

## 1. Обзор

Trade Engine — CLI-приложение для автоматической торговли через российских брокеров (БКС, Т-Инвестиции и др.). Пользователь описывает ботов в YAML-конфиге, загружает стратегии на любом языке (Go/Rust, скомпилированные в Wasm, или Python/JS/любой через subprocess), а оркестратор управляет их жизненным циклом, безопасностью и исполнением ордеров.

### Ключевые цели

- **Мультиброкерность** через единый интерфейс-адаптер `Broker` (БКС REST+WS, Т-Инвестиции gRPC+WS и др.)
- **Мультиязычные стратегии** через гибридный runtime: Wasm (in-process) для Go/Rust/C и subprocess (stdin/stdout JSON-lines) для Python/JS/любого языка
- **Событийная архитектура** — стратегии подписываются на рыночные события и реагируют в реальном времени, не только по cron-расписанию
- **Безопасность by design** — стратегии никогда не имеют прямого доступа к учётным данным брокера или интерфейсу Broker
- **Масштабируемость до 1000+ ботов** с общими потоками рыночных данных, пулингом процессов и эффективным управлением ресурсами
- **Бэктестинг** — тот же код стратегии запускается на исторических данных через mock-адаптер брокера
- **Hot-reload** конфигурации и стратегий без перезапуска приложения

---

## 2. Архитектура

### 2.1 Высокоуровневая диаграмма

```
+------------------------------------------------------------------+
|                      CLI (cmd/trade-engine)                      |
|                 YAML Config + fsnotify hot-reload                |
+-------------------------------+----------------------------------+
                                |
                                v
+------------------------------------------------------------------+
|                          Orchestrator                            |
|                                                                  |
|  +------------+  +------------+  +------------+                  |
|  |   Bot #1   |  |   Bot #2   |  |   Bot #N   |                  |
|  | (goroutine)|  | (goroutine)|  | (goroutine)|                  |
|  |            |  |            |  |            |                  |
|  | +--------+ |  | +--------+ |  | +--------+ |                  |
|  | |Strategy| |  | |Strategy| |  | |Strategy| |                  |
|  | | (Wasm) | |  | |(subproc)||  | | (Wasm) | |                  |
|  | +---+----+ |  | +---+----+ |  | +---+----+ |                  |
|  |     |      |  |     |      |  |     |      |                  |
|  |     v      |  |     v      |  |     v      |                  |
|  | +--------+ |  | +--------+ |  | +--------+ |                  |
|  | | Broker | |  | | Broker | |  | | Broker | |                  |
|  | | Proxy  | |  | | Proxy  | |  | | Proxy  | |                  |
|  | +---+----+ |  | +---+----+ |  | +---+----+ |                  |
|  +-----+------+  +-----+------+  +-----+------+                  |
|        |               |               |                         |
|        +-------+-------+-------+-------+                         |
|                |                                                 |
|                v                                                 |
|  +---------------------------+                                   |
|  |   Shared MarketData Hub   |                                   |
|  |                           |                                   |
|  |  BCS WS ----+             |                                   |
|  |  TInvest ---+-> TickRouter ---> fan-out подписанным ботам     |
|  +---------------------------+                                   |
|                                                                  |
|  +-------------------------------+                               |
|  |     Broker Registry           |                               |
|  |   (имя брокера -> реализация) |                               |
|  +-------------+-----------------+                               |
+----------------|-------------------------------------------------+
                 |
     +-----------+-----------+
     v           v           v
+---------+ +---------+ +---------+
|   BCS   | | TInvest | |Backtest |
| Adapter | | Adapter | | Adapter |
+---------+ +---------+ +---------+
```

### 2.2 Основные принципы

1. **Гексагональная архитектура (Ports & Adapters)** — доменные типы (`trade`) не имеют зависимостей; `broker.Broker` — порт; `bcs.Adapter`, `tinvest.Adapter` — адаптеры.
2. **Песочница стратегий** — стратегии общаются только через `State` (read-only вход) и `Intent` (выход). Они никогда не обращаются к `Broker` напрямую.
3. **Общие ресурсы** — одно WebSocket-соединение на тикер разделяется между всеми ботами через MarketData Hub.
4. **Graceful Lifecycle** — каждый компонент поддерживает отмену через `context.Context` и корректное завершение.

---

## 3. Компоненты

### 3.1 Конфигурация (`internal/config/`)

#### Структура YAML

```yaml
brokers:
  bcs:
    type: bcs
    credentials:
      refresh_token: "${BCS_TOKEN}" # интерполяция переменных окружения

  tinkoff:
    type: tinvest
    credentials:
      api_token: "${TINVEST_TOKEN}"

bots:
  - name: rebalance-conservative
    broker: bcs
    account_id: "ABC123"
    strategy:
      runtime: subprocess # subprocess | wasm
      command: "python3" # только для subprocess
      path: "./strategies/rebalance.py"
      params:
        allocations:
          SBER: 33
          TATNP: 33
          GOLD: 34
    schedule:
      type: cron # active | cron | event
      expression: "0 0 1 * *" # 1-е число каждого месяца

  - name: momentum-bot
    broker: tinkoff
    account_id: "XYZ789"
    strategy:
      runtime: wasm
      path: "./strategies/momentum.wasm"
      params:
        lookback_days: 30
    schedule:
      type: active # постоянно слушает рынок

  - name: signal-trader
    broker: bcs
    account_id: "ABC123"
    strategy:
      runtime: subprocess
      command: "node"
      path: "./strategies/signal.js"
      params:
        rsi_threshold: 30
    schedule:
      type: event # реагирует на рыночные события
      triggers: # декларативный пре-фильтр (опционально)
        - type: price_cross
          ticker: SBER
          threshold: 300
          direction: above

backtest:
  data_dir: "./data/historical"
```

#### Go-типы

```go
type Config struct {
    Brokers  map[string]BrokerConfig `yaml:"brokers"`
    Bots     []BotConfig             `yaml:"bots"`
    Backtest BacktestConfig          `yaml:"backtest"`
}

type BrokerConfig struct {
    Type        string            `yaml:"type"`        // "bcs" | "tinvest"
    Credentials map[string]string `yaml:"credentials"` // поддержка ${ENV_VAR}
}

type BotConfig struct {
    Name      string         `yaml:"name"`
    Broker    string         `yaml:"broker"`       // ссылка на brokers[key]
    AccountID string         `yaml:"account_id"`
    Strategy  StrategyConfig `yaml:"strategy"`
    Schedule  ScheduleConfig `yaml:"schedule"`
}

type StrategyConfig struct {
    Runtime string         `yaml:"runtime"`  // "wasm" | "subprocess"
    Command string         `yaml:"command"`  // для subprocess: "python3", "node"
    Path    string         `yaml:"path"`     // путь к .wasm или .py/.js
    Params  map[string]any `yaml:"params"`   // передаётся стратегии как JSON
}

type ScheduleConfig struct {
    Type       string          `yaml:"type"`       // "active" | "cron" | "event"
    Expression string          `yaml:"expression"` // cron-выражение (если type=cron)
    Triggers   []TriggerConfig `yaml:"triggers"`   // декларативные фильтры (если type=event)
}

type TriggerConfig struct {
    Type      string `yaml:"type"`      // "price_cross" | "drawdown" | "volume_spike"
    Ticker    string `yaml:"ticker"`
    Threshold string `yaml:"threshold"` // парсится в зависимости от type
    Direction string `yaml:"direction"` // "above" | "below"
}

type BacktestConfig struct {
    DataDir string `yaml:"data_dir"`
}
```

#### Интерполяция переменных окружения

Значения конфига, содержащие `${VAR_NAME}`, подставляются из переменных окружения при загрузке. Это позволяет хранить секреты вне конфиг-файла.

#### Hot-Reload

`fsnotify` отслеживает `config.yaml` и файлы стратегий. При изменении:

1. Парсинг и валидация нового конфига.
2. Diff с текущим конфигом.
3. Graceful-остановка удалённых/изменённых ботов.
4. Запуск новых/изменённых ботов с обновлённым конфигом.
5. Неизменённые боты продолжают работать без прерывания.

---

### 3.2 Интерфейс брокера (`internal/broker/`)

#### Порт: `Broker`

```go
type Broker interface {
    Name() string
    Accounts(ctx context.Context) ([]trade.Account, error)
    Portfolio(ctx context.Context, accountID string) (trade.Portfolio, error)
    Orders(ctx context.Context, accountID string) ([]trade.OrderState, error)
    OrderState(ctx context.Context, accountID string, orderID string) (trade.OrderState, error)
    PlaceOrder(ctx context.Context, accountID string, order trade.Order) (string, error)
    CancelOrder(ctx context.Context, accountID string, orderID string) error
    InstrumentByTicker(ctx context.Context, ticker string) (trade.Instrument, error)
    InstrumentsByTickers(ctx context.Context, tickers []string) ([]trade.Instrument, error)
}
```

#### Порт: `MarketData`

```go
type MarketData interface {
    Subscribe(ctx context.Context, tickers []string) (<-chan Tick, error)
    Candles(ctx context.Context, ticker string, from, to time.Time,
        interval CandleInterval) ([]Candle, error)
    Close() error
}

type Tick struct {
    Ticker    string
    Price     decimal.Decimal
    Volume    decimal.Decimal
    Timestamp time.Time
}

type Candle struct {
    Open, High, Low, Close decimal.Decimal
    Volume                 decimal.Decimal
    Timestamp              time.Time
}

type CandleInterval int

const (
    CandleInterval1m CandleInterval = iota
    CandleInterval5m
    CandleInterval15m
    CandleInterval1h
    CandleInterval1d
)
```

#### Sentinel-ошибки

```go
var (
    // Доступ
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")

    // Логика
    ErrInvalidRequest   = errors.New("invalid request")
    ErrInvalidAccountID = fmt.Errorf("%w: invalid account ID", ErrInvalidRequest)
    ErrNotFound         = errors.New("not found")
    ErrConflict         = errors.New("conflict")

    // Инфраструктура
    ErrTimeout     = errors.New("timeout")
    ErrRateLimited = errors.New("rate limited")
    ErrUnavailable = errors.New("unavailable")

    // Реализация
    ErrNotSupported       = errors.New("not supported")
    ErrUnexpectedResponse = errors.New("unexpected response")
)
```

#### Реестр брокеров

```go
type Registry struct {
    brokers map[string]Broker
}

func (r *Registry) Register(name string, b Broker)
func (r *Registry) Get(name string) (Broker, error)
```

Сопоставляет имена брокеров из конфига с конкретными реализациями. Заполняется при запуске на основе `config.Brokers`.

---

### 3.3 Broker Proxy (`internal/broker/proxy.go`)

Медиатор безопасности между стратегией и реальным брокером. Стратегии **никогда** не получают интерфейс `Broker` напрямую.

```go
type BrokerProxy struct {
    broker    Broker
    accountID string
    botID     string           // для аудит-логирования
    limits    ProxyLimits
    logger    *slog.Logger
}

type ProxyLimits struct {
    MaxOrdersPerMinute int
    MaxOrderValue      decimal.Decimal
    AllowedTickers     []string          // пустой = все разрешены
}

func (p *BrokerProxy) ExecuteIntents(
    ctx context.Context,
    intents []strategy.Intent,
) ([]IntentResult, error)
```

Поток выполнения `ExecuteIntents`:

1. **Валидация** каждого intent'а (тикер существует, количество > 0, список разрешённых тикеров).
2. **Rate-limit** проверка (скользящее окно, счётчик на бота).
3. **Конвертация** `Intent` в `trade.Order`.
4. **Исполнение** через `p.broker.PlaceOrder()`.
5. **Логирование** каждого действия с ID бота для аудита.
6. **Возврат** результатов (ID ордера или ошибка по каждому intent'у).

---

### 3.4 Runtime стратегий (`internal/strategy/`)

#### Основные типы

```go
// Runtime — единый интерфейс для всех бэкендов стратегий.
type Runtime interface {
    Start(ctx context.Context) error
    OnTick(ctx context.Context, state *State) (*Response, error)
    OnSchedule(ctx context.Context, state *State) (*Response, error)
    Stop(ctx context.Context) error
}

// State — read-only снимок, передаваемый стратегии.
type State struct {
    Portfolio trade.Portfolio            `json:"portfolio"`
    Prices    map[string]decimal.Decimal `json:"prices"`
    Params    map[string]any             `json:"params"`
    Timestamp time.Time                  `json:"timestamp"`
}

// Response — то, что стратегия возвращает после обработки события.
type Response struct {
    Intents   []Intent `json:"intents"`
    Subscribe []string `json:"subscribe,omitempty"` // возвращается только при init
    Log       string   `json:"log,omitempty"`
}

// Intent — торговое намерение (не ордер).
type Intent struct {
    Action    IntentAction    `json:"action"`
    Ticker    string          `json:"ticker"`
    Quantity  decimal.Decimal `json:"quantity"`
    OrderType trade.OrderType `json:"order_type"`
    Price     decimal.Decimal `json:"price,omitempty"` // для лимитных ордеров
}

type IntentAction string

const (
    IntentBuy    IntentAction = "buy"
    IntentSell   IntentAction = "sell"
    IntentCancel IntentAction = "cancel"
)
```

#### Wasm Runtime (`internal/strategy/wasm/`)

```go
type WasmRuntime struct {
    engine   wazero.Runtime
    compiled wazero.CompiledModule
    instance api.Module
    params   map[string]any
}
```

- Загружает `.wasm`-файл через `wazero` (pure Go, zero CGO).
- Экспортирует host-функции в wasm-модуль: `get_portfolio`, `get_prices`, `log_message`.
- Вызывает экспортированные guest-функции: `on_init`, `on_tick`, `on_schedule`.
- Обмен данными через shared linear memory в формате JSON.
- **Изоляция:** wasm-модуль не имеет доступа к файловой системе, сети или переменным окружения.
- **Детерминизм:** одинаковые входные данные дают одинаковый результат — критично для бэктестинга.

Поддерживаемые исходные языки: Go (через TinyGo), Rust, C/C++, AssemblyScript, Zig.

#### Subprocess Runtime (`internal/strategy/subprocess/`)

```go
type SubprocessRuntime struct {
    cmd     *exec.Cmd
    stdin   io.WriteCloser
    stdout  *bufio.Scanner
    stderr  io.ReadCloser
    params  map[string]any
}
```

**Протокол: JSON-lines через stdin/stdout**

Оркестратор -> стратегия (stdin):

```jsonl
{"type": "init", "params": {"allocations": {"SBER": 33, "TATNP": 33, "GOLD": 34}}}
{"type": "tick", "state": {"portfolio": {...}, "prices": {"SBER": "295.50"}, "timestamp": "..."}}
{"type": "schedule", "state": {"portfolio": {...}, "prices": {...}}}
{"type": "stop"}
```

Стратегия -> оркестратор (stdout):

```jsonl
{"intents": [], "subscribe": ["tick:SBER", "tick:GAZP", "candle:SBER:1h"]}
{"intents": [{"action": "buy", "ticker": "SBER", "quantity": 10, "order_type": "market"}]}
{"intents": []}
{"log": "Ребалансировка завершена, изменений не требуется"}
```

**Пример стратегии на Python:**

```python
import json
import sys

def on_event(event):
    if event["type"] == "init":
        return {
            "intents": [],
            "subscribe": ["tick:SBER", "tick:GAZP"]
        }

    state = event["state"]
    portfolio = state["portfolio"]
    prices = state["prices"]
    params = event.get("params", {})
    allocations = params.get("allocations", {})

    # ... логика ребалансировки ...

    return {
        "intents": [
            {"action": "buy", "ticker": "SBER", "quantity": 5, "order_type": "market"}
        ]
    }

for line in sys.stdin:
    event = json.loads(line.strip())
    if event["type"] == "stop":
        break
    result = on_event(event)
    print(json.dumps(result), flush=True)
```

**Тестируемость:** стратегии можно тестировать автономно через pipe:

```bash
echo '{"type":"init","params":{}}' | python3 strategy.py
```

---

### 3.5 Система событий и планирование (`internal/bot/`)

#### Типы расписания

| Тип      | Поведение                                                | Применение                                                |
| -------- | -------------------------------------------------------- | --------------------------------------------------------- |
| `cron`   | Запуск по cron-выражению                                 | Ежемесячная ребалансировка, ежедневные отчёты             |
| `active` | Получает каждый рыночный тик по подписанным инструментам | Активный трейдинг, скальпинг, моментум                    |
| `event`  | Получает тики, но фильтрует через декларативные триггеры | Покупка при пересечении цены, ребалансировка при просадке |

#### Модель подписки на события

При `init` стратегия возвращает список `subscribe`, указывая, какие события ей нужны:

```json
{
  "subscribe": ["tick:SBER", "tick:GAZP", "candle:SBER:1h", "portfolio_change"]
}
```

Поддерживаемые типы событий:

| Событие            | Формат                       | Описание                              |
| ------------------ | ---------------------------- | ------------------------------------- |
| Тик цены           | `tick:<TICKER>`              | Обновление цены в реальном времени    |
| Закрытие свечи     | `candle:<TICKER>:<INTERVAL>` | Свеча закрылась (1m, 5m, 15m, 1h, 1d) |
| Изменение портфеля | `portfolio_change`           | Изменилась позиция или баланс         |
| Исполнение ордера  | `order_filled`               | Ордер полностью исполнен              |

Для типа расписания `event` **декларативные триггеры** в конфиге действуют как пре-фильтр. Оркестратор пересылает события стратегии только при выполнении условия триггера. Это позволяет не будить процесс стратегии для нерелевантных тиков.

#### Общий MarketData Hub

```
MarketData Hub
+------------------------------------------------------+
|                                                       |
|  BCS WebSocket ----+                                  |
|  TInvest Stream ---+--> Tick Normalizer --> TickRouter |
|                                              |        |
|         +----------+----------+---------+    |        |
|         v          v          v         v    |        |
|      Bot#1      Bot#2      Bot#47    Bot#N   |        |
|   (tick:SBER) (tick:SBER) (tick:GAZP)        |        |
|                                               |        |
|  - Одно WS-соединение на тикер, не на бота   |        |
|  - Fan-out через Go-каналы                   |        |
|  - Backpressure: медленные боты пропускают   |        |
|    тики                                       |        |
+------------------------------------------------------+
```

Одна WebSocket-подписка на тикер разделяется между всеми ботами. TickRouter хранит маппинг `тикер -> []chan Tick` и раздаёт входящие тики всем подписанным каналам ботов. Это критически важно для поддержки 1000+ ботов без исчерпания лимитов WebSocket-соединений брокера.

---

### 3.6 Бот (`internal/bot/`)

```go
type Bot struct {
    ID       string
    Config   config.BotConfig
    Runtime  strategy.Runtime
    Proxy    *broker.BrokerProxy
    Logger   *slog.Logger
    cancel   context.CancelFunc
}

func (b *Bot) Run(ctx context.Context) error {
    if err := b.Runtime.Start(ctx); err != nil {
        return fmt.Errorf("strategy start: %w", err)
    }
    defer b.Runtime.Stop(ctx)

    // Init: получение подписок от стратегии
    initState := b.buildState(ctx)
    resp, err := b.Runtime.OnTick(ctx, initState)
    // ... сохранение resp.Subscribe для MarketData Hub ...

    switch b.Config.Schedule.Type {
    case "active":
        return b.runActive(ctx)
    case "cron":
        return b.runCron(ctx)
    case "event":
        return b.runEvent(ctx)
    }
    return nil
}
```

**runActive:** подписывается на каналы MarketData Hub, вызывает `Runtime.OnTick` на каждый тик, передаёт intent'ы в `BrokerProxy.ExecuteIntents`.

**runCron:** спит между cron-тригерами (через `robfig/cron` или `time.Timer`), вызывает `Runtime.OnSchedule` при каждом срабатывании.

**runEvent:** подписывается на каналы MarketData Hub, применяет декларативные фильтры триггеров, вызывает `Runtime.OnTick` только при срабатывании триггера.

---

### 3.7 Оркестратор (`internal/orchestrator/`)

```go
type Orchestrator struct {
    bots        map[string]*bot.Bot
    brokerReg   *broker.Registry
    marketHub   *market.Hub
    config      *config.Config
    processPool *subprocess.Pool   // общий пул процессов для cron-ботов
    mu          sync.RWMutex
    logger      *slog.Logger
    wg          sync.WaitGroup
}
```

#### Жизненный цикл

```go
func (o *Orchestrator) Start(ctx context.Context) error {
    // 1. Инициализация адаптеров брокеров из конфига
    // 2. Запуск MarketData Hub
    // 3. Запуск всех ботов
    for _, botCfg := range o.config.Bots {
        o.startBot(ctx, botCfg)
    }
    // 4. Запуск наблюдателя конфига для hot-reload
    // 5. Блокировка до отмены ctx
    <-ctx.Done()
    return o.gracefulShutdown()
}

func (o *Orchestrator) startBot(ctx context.Context, cfg config.BotConfig) error {
    brk, _ := o.brokerReg.Get(cfg.Broker)
    proxy := broker.NewBrokerProxy(brk, cfg.AccountID, cfg.Name)

    var rt strategy.Runtime
    switch cfg.Strategy.Runtime {
    case "wasm":
        rt = wasm.New(cfg.Strategy.Path, cfg.Strategy.Params)
    case "subprocess":
        rt = subprocess.New(cfg.Strategy.Command, cfg.Strategy.Path, cfg.Strategy.Params)
    }

    b := bot.New(cfg, rt, proxy, o.logger)

    botCtx, cancel := context.WithCancel(ctx)
    b.SetCancel(cancel)

    o.mu.Lock()
    o.bots[cfg.Name] = b
    o.mu.Unlock()

    o.wg.Add(1)
    go func() {
        defer o.wg.Done()
        if err := b.Run(botCtx); err != nil {
            o.logger.Error("bot stopped with error", "bot", cfg.Name, "error", err)
        }
    }()
    return nil
}

func (o *Orchestrator) Reload(newCfg *config.Config) error {
    // 1. Diff старого конфига с новым
    // 2. Остановка удалённых ботов (отмена контекста)
    // 3. Остановка изменённых ботов, перезапуск с новым конфигом
    // 4. Запуск новых ботов
    // 5. Обновление реестра брокеров при изменении конфигов брокеров
}

func (o *Orchestrator) gracefulShutdown() error {
    o.mu.RLock()
    for _, b := range o.bots {
        b.Cancel()
    }
    o.mu.RUnlock()

    // Ожидание завершения всех ботов с таймаутом
    done := make(chan struct{})
    go func() { o.wg.Wait(); close(done) }()

    select {
    case <-done:
        return nil
    case <-time.After(30 * time.Second):
        return errors.New("graceful shutdown timed out")
    }
}
```

---

### 3.8 Движок бэктестинга (`internal/backtest/`)

```go
type Engine struct {
    dataDir string
    logger  *slog.Logger
}

type Result struct {
    TotalReturn decimal.Decimal   // проценты
    MaxDrawdown decimal.Decimal   // проценты
    SharpeRatio decimal.Decimal
    TotalTrades int
    WinRate     decimal.Decimal   // проценты
    TradeLog    []TradeRecord
    EquityCurve []EquityPoint
}

type TradeRecord struct {
    Timestamp time.Time
    Ticker    string
    Direction trade.OrderDirection
    Quantity  decimal.Decimal
    Price     decimal.Decimal
    Value     decimal.Decimal
}

type EquityPoint struct {
    Timestamp time.Time
    Value     decimal.Decimal
}
```

#### Как это работает

Движок бэктестинга переиспользует **тот же самый** runtime стратегии и broker proxy, что и live-торговля. Единственное отличие — адаптер брокера:

```
Strategy Runtime (идентичный) --> BrokerProxy (идентичный) --> MockBroker (вместо BCS/TInvest)
```

```go
func (e *Engine) Run(ctx context.Context, cfg config.BotConfig,
    from, to time.Time) (*Result, error) {

    // 1. Загрузка исторических данных для нужных тикеров
    data := e.loadHistoricalData(cfg, from, to)

    // 2. Создание MockBroker на основе исторических данных
    mock := NewMockBroker(data)

    // 3. Создание BrokerProxy -> MockBroker
    proxy := broker.NewBrokerProxy(mock, "backtest-account", cfg.Name)

    // 4. Создание runtime стратегии (тот же, что и для live)
    var rt strategy.Runtime
    switch cfg.Strategy.Runtime {
    case "wasm":
        rt = wasm.New(cfg.Strategy.Path, cfg.Strategy.Params)
    case "subprocess":
        rt = subprocess.New(cfg.Strategy.Command, cfg.Strategy.Path, cfg.Strategy.Params)
    }
    rt.Start(ctx)
    defer rt.Stop(ctx)

    // 5. Воспроизведение исторических тиков через стратегию
    for _, tick := range data.Ticks() {
        mock.AdvanceTo(tick.Timestamp)
        state := mock.CurrentState()

        resp, _ := rt.OnTick(ctx, state)
        if len(resp.Intents) > 0 {
            proxy.ExecuteIntents(ctx, resp.Intents)
        }
    }

    // 6. Вычисление и возврат метрик
    return mock.ComputeResults(), nil
}
```

**MockBroker** реализует `broker.Broker`:

- `Portfolio()` возвращает состояние виртуального портфеля.
- `PlaceOrder()` исполняет по исторической цене с настраиваемой моделью проскальзывания.
- `InstrumentByTicker()` возвращает метаданные инструментов из файлов исторических данных.

**Детерминизм:** Wasm-стратегии дают идентичные выходные данные для идентичных входных, что делает бэктесты полностью воспроизводимыми. Subprocess-стратегии детерминистичны, если сам код стратегии детерминистичен (без random, без сетевых вызовов).

---

## 4. Масштабируемость (1000+ ботов)

### 4.1 Анализ ресурсов

| Тип бота                  | Количество (пример) | Ресурс на бота                                   | Итого         |
| ------------------------- | ------------------- | ------------------------------------------------ | ------------- |
| Active (wasm)             | 50                  | 1 горутина + ~10-50 MB wasm-памяти               | ~2.5 GB       |
| Active (subprocess)       | 20                  | 1 горутина + 1 процесс (~30 MB)                  | ~600 MB       |
| Event-driven (wasm)       | 100                 | 1 горутина + wasm-инстанс (по требованию)        | ~500 MB       |
| Event-driven (subprocess) | 80                  | 1 горутина + worker из пула процессов            | Общий пул     |
| Cron (wasm)               | 400                 | 1 горутина (спит) + wasm-инстанс (по требованию) | ~4 MB горутин |
| Cron (subprocess)         | 350                 | 1 горутина (спит) + worker из пула процессов     | ~4 MB горутин |

### 4.2 Стратегии оптимизации

#### Пул subprocess-процессов

Cron и event-driven subprocess-боты **не** держат выделенный процесс постоянно. Вместо этого общий пул процессов управляет worker'ами:

```go
type Pool struct {
    workers  chan *Worker
    maxSize  int
}

type Worker struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout *bufio.Scanner
    lang   string // "python3", "node" и т.д.
}
```

- Cron-бот запрашивает worker из пула перед выполнением.
- После выполнения worker возвращается в пул.
- Active-боты получают выделенный долгоживущий процесс.
- Размер пула настраивается; избыточные запросы встают в очередь.

#### Пулинг Wasm-инстансов

```go
type InstancePool struct {
    compiled wazero.CompiledModule  // компилируется один раз, разделяется
    pool     chan api.Module         // пре-инстанцированные инстансы
}
```

- `.wasm`-модуль компилируется один раз и кэшируется.
- Для cron-ботов инстансы создаются по требованию из скомпилированного модуля и уничтожаются после использования.
- Для active-ботов инстансы долгоживущие.

#### Общий MarketData Hub

- Одно WebSocket-соединение на тикер, независимо от количества подписанных ботов.
- Fan-out через буферизированные Go-каналы.
- Медленные потребители (боты, не успевающие обрабатывать) пропускают тики с предупреждением в логе, не блокируя fan-out.

#### Тюнинг GC

- `sync.Pool` для переиспользования объектов `State` и `Intent` для снижения аллокаций.
- Тюнинг `GOGC` и `GOMEMLIMIT` для контроля частоты GC под нагрузкой.
- При 500 объектов State/сек паузы GC в Go остаются менее 1мс.

---

## 5. Модель безопасности

### 5.1 Уровни изоляции стратегий

| Уровень                     | Механизм                                                              | Что предотвращает                               |
| --------------------------- | --------------------------------------------------------------------- | ----------------------------------------------- |
| **Граница API**             | Стратегия получает `State` (read-only), возвращает `Intent`           | Нет прямого доступа к брокеру                   |
| **Broker Proxy**            | Валидирует intent'ы, rate-limit, проверяет список разрешённых тикеров | Несанкционированные ордера, чрезмерная торговля |
| **Изоляция процессов**      | Subprocess-стратегии работают в отдельных процессах ОС                | Нет доступа к памяти/env оркестратора           |
| **Изоляция памяти**         | Wasm-стратегии имеют изолированную linear memory (wazero)             | Нет доступа к памяти хоста                      |
| **Изоляция учётных данных** | Токены брокеров живут только в процессе оркестратора                  | Стратегии не могут украсть токены               |
| **Изоляция между ботами**   | Каждый бот имеет свой BrokerProxy, привязанный к одному аккаунту      | Бот A не может получить данные бота B           |

### 5.2 Будущее: мультипользовательность

Когда Web UI будет поддерживать несколько пользователей:

- Каждый пользователь владеет набором ботов; RBAC контролирует доступ.
- Учётные данные брокеров шифруются at-rest (AES-256-GCM) с ключами per-user.
- Subprocess-стратегии запускаются под ограниченным unix-пользователем (без сети, ограниченная FS).
- Аудит-лог всех операций с брокером с атрибуцией пользователя/бота.

---

## 6. Доменные типы (`internal/trade/`)

Эти типы — ядро доменной модели. Они не имеют **никаких внешних зависимостей** и используются во всех компонентах.

```go
// Account
type Account struct {
    Name string
    ID   string
}

// Instrument
type Instrument struct {
    Name      string
    Ticker    string
    ClassCode string
    Type      InstrumentType
    Currency  CurrencyCode
    Lot       decimal.Decimal
}

type InstrumentType int
const (
    InstrumentUnspecified InstrumentType = iota
    InstrumentBond
    InstrumentShare
    InstrumentCurrency
    InstrumentEtf
    InstrumentFutures
    InstrumentSp
    InstrumentOption
    InstrumentClearingCertificate
    InstrumentIndex
    InstrumentCommodity
)

// Position
type Position struct {
    Name         string
    Ticker       string
    Type         InstrumentType
    Currency     CurrencyCode
    AveragePrice decimal.Decimal
    CurrentPrice decimal.Decimal
    Quantity     decimal.Decimal
}

// Portfolio
type Portfolio struct {
    Name      string
    Currency  CurrencyCode
    Positions []Position
}

// Order
type OrderDirection int
const ( Sell OrderDirection = iota; Buy )

type OrderType int
const ( Limit OrderType = iota; Market )

type OrderStatus int
const ( New OrderStatus = iota; Fill; PartiallyFill; Cancelled; Rejected )

type Order struct {
    Ticker    string
    Type      OrderType
    Direction OrderDirection
    Quantity  decimal.Decimal
    Price     decimal.Decimal
}

type OrderState struct {
    ID        string
    Ticker    string
    Status    OrderStatus
    Type      OrderType
    Direction OrderDirection
    Price     decimal.Decimal
    Quantity  decimal.Decimal
}

// Currency
type CurrencyCode string
const ( RUB CurrencyCode = "RUB"; USD CurrencyCode = "USD" )
```

---

## 7. Реализации адаптеров

### 7.1 БКС Trade-API (`internal/bcs/`)

**Статус:** Реализован (REST API). WebSocket для рыночных данных запланирован.

- **Авторизация:** OAuth2 refresh token flow через БКС Keycloak.
- **Транспорт:** HTTP/JSON к `https://be.broker.ru/`.
- **Эндпоинты:** Портфель, Ордера (поиск/размещение/отмена/состояние), Инструменты по тикерам.
- **Особенности:** Позиции фильтруются по терму T0; инструменты фильтруются по бирже MOEX; автоматический поиск classCode при размещении ордера.

### 7.2 Т-Инвестиции (`internal/tinvest/`)

**Статус:** Запланирован.

- **Авторизация:** API-токен (bearer).
- **Транспорт:** gRPC для торговых операций; WebSocket (или gRPC streaming) для рыночных данных.
- **Примечание:** Богатый API с рыночными данными, фундаменталом и стримингом свечей.

### 7.3 Backtest Mock (`internal/backtest/`)

**Статус:** Запланирован.

- **Авторизация:** Нет.
- **Транспорт:** In-memory.
- **Поведение:** Воспроизводит исторические данные; исполняет ордера по историческим ценам с настраиваемым проскальзыванием.

---

## 8. Структура проекта (целевая)

```
trade-engine/
|-- cmd/
|   `-- trade-engine/
|       `-- main.go                     # Точка входа CLI
|-- internal/
|   |-- broker/
|   |   |-- broker.go                   # Интерфейс Broker (порт)
|   |   |-- market.go                   # Интерфейс MarketData (порт)
|   |   |-- errors.go                   # Sentinel-ошибки
|   |   |-- proxy.go                    # BrokerProxy (медиатор безопасности)
|   |   `-- registry.go                 # BrokerRegistry (имя -> реализация)
|   |-- trade/
|   |   |-- types.go                    # Account, Instrument, Position, Portfolio
|   |   |-- order.go                    # Order, OrderState, перечисления
|   |   `-- currency.go                 # CurrencyCode
|   |-- bcs/
|   |   |-- adapter.go                  # Реализация Broker для БКС
|   |   |-- market.go                   # MarketData для БКС (WebSocket)
|   |   |-- models.go                   # JSON-модели БКС
|   |   |-- constants.go                # Константы-перечисления БКС
|   |   |-- utils.go                    # Конвертеры BCS <-> trade
|   |   `-- errors.go                   # Обработка ошибок БКС
|   |-- tinvest/                        # Адаптер Т-Инвестиций (будущее)
|   |   |-- adapter.go
|   |   |-- market.go
|   |   `-- ...
|   |-- strategy/
|   |   |-- strategy.go                 # Интерфейс Runtime, State, Intent, Response
|   |   |-- wasm/
|   |   |   `-- runtime.go              # Wasm runtime (wazero)
|   |   `-- subprocess/
|   |       |-- runtime.go              # Subprocess runtime (JSON-lines)
|   |       `-- pool.go                 # Пул процессов для cron-ботов
|   |-- bot/
|   |   |-- bot.go                      # Структура Bot + цикл Run
|   |   `-- scheduler.go                # Планирование: active, cron, event
|   |-- orchestrator/
|   |   |-- orchestrator.go             # Управление жизненным циклом
|   |   `-- reload.go                   # Diff конфига + hot-reload
|   |-- backtest/
|   |   |-- engine.go                   # Движок бэктестинга
|   |   |-- mock_broker.go              # Mock Broker (исторические данные)
|   |   `-- result.go                   # Метрики: доходность, просадка, Шарп
|   |-- market/
|   |   |-- hub.go                      # Общий MarketData Hub + TickRouter
|   |   `-- feed.go                     # Агрегация потоков MarketData
|   |-- config/
|   |   |-- config.go                   # Парсинг YAML + интерполяция env
|   |   `-- watcher.go                  # fsnotify hot-reload
|   `-- logging/
|       `-- logger.go                   # Настройка slog
|-- strategies/
|   `-- examples/
|       |-- rebalance.py                # Пример: стратегия ребалансировки на Python
|       `-- momentum/
|           |-- main.go                 # Пример: Go моментум (TinyGo -> wasm)
|           `-- Makefile
|-- configs/
|   `-- config.example.yaml
|-- data/
|   `-- historical/                     # Исторические данные для бэктестинга
`-- docs/
    |-- tech-spec.md                    # Этот документ
    `-- roadmap.md                      # Дорожная карта разработки
```

---

## 9. Технологический стек

| Компонент           | Технология             | Обоснование                                                      |
| ------------------- | ---------------------- | ---------------------------------------------------------------- |
| Язык                | Go 1.24+               | Горутины для конкурентности, низкие паузы GC, быстрая компиляция |
| Wasm runtime        | wazero                 | Pure Go, zero CGO, production-ready песочница                    |
| YAML конфиг         | gopkg.in/yaml.v3       | Де-факто стандарт для Go YAML                                    |
| Cron                | robfig/cron/v3         | Зрелая библиотека, поддержка секундной точности                  |
| Отслеживание файлов | fsnotify/fsnotify      | Стандарт для hot-reload в Go                                     |
| Логирование         | log/slog (stdlib)      | Структурированное логирование, встроено в Go 1.24+               |
| Точная арифметика   | govalues/decimal       | Уже используется, точная финансовая арифметика                   |
| OAuth2              | golang.org/x/oauth2    | Уже используется для БКС                                         |
| UUID                | google/uuid            | Уже используется для clientOrderId                               |
| gRPC                | google.golang.org/grpc | Для адаптера Т-Инвестиций                                        |
| CLI (будущее)       | spf13/cobra            | Если потребуются подкоманды                                      |
| Web UI (будущее)    | Svelte + SvelteKit     | Админ-панель для оркестратора                                    |

---

## 10. Почему Go достаточен (Rust не нужен)

Система **ограничена I/O**, а не CPU. Ключевые задержки:

| Операция                          | Задержка                 |
| --------------------------------- | ------------------------ |
| HTTP-запрос к API брокера         | 50-500 мс                |
| WebSocket-тик брокера             | ~1-10 мс                 |
| Пауза GC в Go (1.24+)             | < 1 мс (обычно ~100 мкс) |
| Вызов Wasm-функции (wazero)       | ~1-5 мкс                 |
| Обмен JSON-lines через subprocess | ~50-200 мкс              |
| Переключение контекста горутины   | ~200 нс                  |

Узкое место — всегда сетевой вызов к брокеру. Паузы GC в Go на 3-5 порядков меньше задержки API брокера. При 1000+ ботах горутины (4 KB стека каждая) значительно эффективнее потоков ОС.

**Когда Rust имел бы значение:**

- Co-located HFT-серверы с требованиями к задержке менее 10 мкс (другой продукт).
- Вычислительно тяжёлая логика стратегий (ML-инференс) — но это решается компиляцией стратегии в Wasm, который может быть написан на Rust.

**Вывод:** Go обрабатывает оркестратор, адаптеры брокеров, управление конфигурацией и жизненный цикл ботов. Rust могут использовать авторы стратегий, компилирующие в Wasm для максимальной производительности — но сам движок остаётся на Go.
