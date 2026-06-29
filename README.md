# ☠️ Pirates Gold

**EN** | [RU](#-pirates-gold-ru)

> *The ocean of addresses is vast. The treasure is real. The odds are yours to defy.*

Every Bitcoin wallet is unlocked by a sequence of 12 random words — a seed phrase. Lose the words, lose the coins forever. There are millions of such forgotten and abandoned wallets on the blockchain, their funds locked away with no one to claim them.

**Pirates Gold** generates random 12-word combinations, turns each one into a Bitcoin address, and checks its balance against the public blockchain. Get lucky — and the program finds a wallet whose key just fits. Think of it as a digital treasure hunt: the odds are near zero, but the treasure is real.

A lightweight, portable Bitcoin wallet scanner. Generates random [BIP39](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki) seed phrases, derives P2PKH addresses via the BIP44 path `m/44'/0'/0'/0/0`, and checks their balance against public blockchain APIs — with no blockchain download required.

---

## Philosophy

There are an estimated **3.7 million Bitcoin permanently lost** — forgotten passwords, destroyed hard drives, wallets of the deceased whose keys were never passed on. According to [Chainalysis](https://www.chainalysis.com), the leading blockchain analytics firm, this figure represents roughly **17–20% of all Bitcoin ever mined**, worth hundreds of billions of dollars at any recent market price. These coins sit in addresses that are mathematically valid, publicly visible on the blockchain, yet functionally unreachable.

**Pirates Gold** does not pretend to be a practical tool for finding wealth. It is, at its core, a meditation on probability and fate — a digital lottery ticket written in Go. The seed space of BIP39 is 2¹²⁸ ≈ 3.4 × 10³⁸ combinations. The chance of a match is, by any honest reckoning, essentially zero.

And yet.

Pirates did not sail because treasure was guaranteed. They sailed because it existed.

This program runs for the same reason. Not as a tool for theft — the addresses it might ever find are those whose owners are long gone, keys destroyed, coins stranded forever in the ledger. It runs as a monument to fortune, to the mathematics of impossibility, to the ghost ships of the blockchain drifting with unclaimed cargo.

**The program is named for what it seeks: not certainty, but the chance of it.**

---

## How It Works

```
Random entropy (128 bit)
    │
    ▼
BIP39 mnemonic (12 words from 2048-word standard wordlist)
    │
    ▼
BIP39 seed derivation (PBKDF2-HMAC-SHA512, 2048 iterations)
    │
    ▼
BIP32 master key → BIP44 path m/44'/0'/0'/0/0
    │
    ▼
secp256k1 → compressed public key → SHA256 → RIPEMD160 → Base58Check
    │
    ▼
Bitcoin address (starts with "1")
    │
    ▼
Balance check via two independent public APIs
    │
    ├─ balance > 0      → found.txt
    ├─ tx count > 0     → used.txt  (spent, drained)
    └─ empty (never used) → discarded
```

**API providers** (no registration, no API keys):
- [blockstream.info](https://blockstream.info) — Esplora format, fixed 0.15 req/s (respects their 700 req/hour IP limit)
- [blockchain.info](https://blockchain.info) — configurable via `-rate` flag

Each provider has dedicated workers. Blockstream runs at a hardcoded safe rate. Blockchain.info rate is fully configurable.

---

## Features

- **Zero runtime dependencies** — single static binary, no install required
- **Dual-provider API** — blockstream.info + blockchain.info, each with dedicated workers
- **Session statistics** — per-minute logging to `stats.txt`, survives restarts
- **Systemd integration** — install/remove as a system service with one flag
- **Fully configurable** — workers, rate limit, output directory
- **Cross-platform build** — static Linux x86-64 binary via `make build`

---

## Build

Requires Go 1.21+.

```bash
git clone https://github.com/Ground_Zerro/pirates-gold
cd pirates-gold
make build
```

Produces a fully static, stripped binary (~6.5 MB):

```bash
./pirates-gold -h
```

---

## Usage

```
pirates-gold [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-workers` | `2` | Workers for blockchain.info (blockstream.info always uses 1) |
| `-rate` | `1.0` | Max req/s for blockchain.info |
| `-out` | `.` | Output directory for result files |
| `-count` | `0` | Stop after N checks (0 = infinite) |
| `-service` | — | Install or remove systemd service |
| `-stats` | — | Show total statistics across all sessions |
| `-version` | — | Show version and exit |
| `-h` | — | Show help |

**Examples:**

```bash
# Run with defaults (~1.15 req/s total: 1.0 blockchain.info + 0.15 blockstream.info)
./pirates-gold

# Higher throughput via blockchain.info
./pirates-gold -workers 4 -rate 2.0 -out /data/results

# Install as systemd service with autostart
./pirates-gold -service

# View cumulative statistics
./pirates-gold -stats

# Quick test run
./pirates-gold -count 20
```

---

## Output Files

| File | Contents |
|------|----------|
| `found.txt` | Wallets with **positive balance** — seed phrase, address, balance, tx count |
| `used.txt` | Wallets with **transaction history** but zero balance |
| `stats.txt` | Per-session statistics, updated every minute |

Empty wallets (zero balance, zero transactions) are discarded — they carry no information and would fill disk space indefinitely.

---

## Systemd Service

```bash
# Install (autostart enabled)
./pirates-gold -service

# Then control with:
systemctl start pirates-gold
systemctl stop pirates-gold
systemctl status pirates-gold
journalctl -u pirates-gold -f

# Remove
./pirates-gold -service
```

---

## Statistics Format (`stats.txt`)

Each run appends a new session block. Existing data is never overwritten.

```
=== Session 29.06.2026 11:52 ===
[11:53] Time: 0h 01m 00s | Checked: 900  | Used: 0 | Found: 0
[11:54] Time: 0h 02m 00s | Checked: 1800 | Used: 1 | Found: 0
=== End of session 29.06.2026 11:54 ===

=== Session 30.06.2026 09:00 ===
...
```

---

## Technical Stack

| Component | Implementation |
|-----------|---------------|
| Seed generation | BIP39, `crypto/rand` |
| Key derivation | BIP32/BIP44, HMAC-SHA512 |
| Elliptic curve | secp256k1 via `btcsuite/btcd` |
| Address encoding | SHA256 + RIPEMD160 + Base58Check |
| Seed derivation | PBKDF2-HMAC-SHA512 (`golang.org/x/crypto`) |
| Rate limiting | Token bucket (`golang.org/x/time/rate`) |

---

## Author

**Ground_Zerro** · Pirates Gold v.1.1

---
---

# ☠️ Pirates Gold <sup>RU</sup>

> *Океан адресов бесконечен. Сокровище существует. Вероятность — на вашей стороне.*

Каждый биткоин-кошелёк открывается набором из 12 случайных слов — seed-фразой. Потерял слова — потерял деньги навсегда. Таких забытых и брошенных кошельков на блокчейне миллионы, и монеты на них недоступны никому.

**Pirates Gold** перебирает случайные комбинации из 12 слов, превращает каждую в биткоин-адрес и проверяет его баланс через публичный блокчейн. Если повезёт — программа найдёт кошелёк, к которому подошёл ключ. Это цифровой аналог поиска пиратского клада: вероятность ничтожна, но сокровище существует.

Лёгкий портативный сканер Bitcoin-кошельков. Генерирует случайные [BIP39](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki) seed-фразы, восстанавливает P2PKH адреса по пути деривации `m/44'/0'/0'/0/0` и проверяет их баланс через публичные API блокчейна — без загрузки блокчейна.

---

## Философия

По оценкам компании [Chainalysis](https://www.chainalysis.com) — ведущей аналитической фирмы в области блокчейна — около **3,7 миллиона Bitcoin утеряны навсегда**: забытые пароли, уничтоженные носители, кошельки умерших владельцев, ключи от которых не были переданы. Это составляет примерно **17–20% от всех когда-либо добытых Bitcoin** — сотни миллиардов долларов, которые видны в блокчейне, но недостижимы.

**Pirates Gold** не притворяется практическим инструментом для поиска богатства. По своей сути это медитация на вероятность и судьбу — цифровой лотерейный билет, написанный на Go. Пространство seed-фраз BIP39 составляет 2¹²⁸ ≈ 3,4 × 10³⁸ комбинаций. Шанс совпадения, по любой честной оценке, стремится к нулю.

И всё же.

Пираты выходили в море не потому, что сокровище было гарантировано. Они выходили, потому что оно существовало.

Эта программа работает по той же причине. Не как инструмент кражи — адреса, которые она теоретически может найти, принадлежат тем, чьи ключи давно уничтожены, а монеты навечно застыли в реестре. Она работает как памятник удаче, математике невозможного, кораблям-призракам блокчейна с невостребованным грузом.

**Программа названа в честь того, что ищет: не уверенности, а её шанса.**

---

## Как это работает

```
Случайная энтропия (128 бит)
    │
    ▼
BIP39 мнемоника (12 слов из стандартного списка 2048 слов)
    │
    ▼
Деривация seed (PBKDF2-HMAC-SHA512, 2048 итерации)
    │
    ▼
BIP32 мастер-ключ → путь BIP44 m/44'/0'/0'/0/0
    │
    ▼
secp256k1 → сжатый публичный ключ → SHA256 → RIPEMD160 → Base58Check
    │
    ▼
Bitcoin-адрес (начинается с "1")
    │
    ▼
Проверка баланса через два независимых публичных API
    │
    ├─ баланс > 0          → found.txt
    ├─ количество транзакций > 0 → used.txt  (использован, опустошён)
    └─ пустой (никогда не использован) → отбрасывается
```

**API-провайдеры** (без регистрации, без ключей):
- [blockstream.info](https://blockstream.info) — формат Esplora, фиксированные 0.15 req/s (уважает лимит 700 req/hour на IP)
- [blockchain.info](https://blockchain.info) — настраивается флагом `-rate`

Каждый провайдер имеет выделенные воркеры. Blockstream работает на жёстко заданной безопасной скорости. Скорость blockchain.info полностью настраивается.

---

## Возможности

- **Ноль зависимостей в рантайме** — один статичный бинарь, без установки
- **Два API-провайдера** — blockstream.info + blockchain.info, каждый с выделенными воркерами
- **Статистика сессий** — запись каждую минуту в `stats.txt`, не сбрасывается при перезапуске
- **Интеграция с systemd** — установка/удаление сервиса одним флагом
- **Полная настраиваемость** — воркеры, rate limit, директория вывода
- **Кросс-платформенная сборка** — статичный Linux x86-64 бинарь через `make build`

---

## Сборка

Требуется Go 1.21+.

```bash
git clone https://github.com/Ground_Zerro/pirates-gold
cd pirates-gold
make build
```

Создаёт полностью статичный, stripped бинарь (~6.5 МБ):

```bash
./pirates-gold -h
```

---

## Использование

```
pirates-gold [флаги]
```

| Флаг | По умолчанию | Описание |
|------|-------------|----------|
| `-workers` | `2` | Воркеры для blockchain.info (blockstream.info всегда использует 1) |
| `-rate` | `1.0` | Лимит запросов/сек для blockchain.info |
| `-out` | `.` | Директория для выходных файлов |
| `-count` | `0` | Остановиться после N проверок (0 = бесконечно) |
| `-service` | — | Установить или удалить systemd-сервис |
| `-stats` | — | Показать общую статистику за все сессии |
| `-version` | — | Показать версию и выйти |
| `-h` | — | Показать справку |

**Примеры:**

```bash
# Запуск с настройками по умолчанию (~1,15 req/s: blockchain.info + blockstream.info)
./pirates-gold

# Повышенная производительность через blockchain.info
./pirates-gold -workers 4 -rate 2.0 -out /data/results

# Установка как systemd-сервис с автозапуском
./pirates-gold -service

# Просмотр накопленной статистики
./pirates-gold -stats

# Быстрый тестовый запуск
./pirates-gold -count 20
```

---

## Выходные файлы

| Файл | Содержимое |
|------|-----------|
| `found.txt` | Кошельки с **положительным балансом** — фраза, адрес, баланс, транзакции |
| `used.txt` | Кошельки с **историей транзакций**, но нулевым балансом |
| `stats.txt` | Статистика по сессиям, обновляется каждую минуту |

Пустые кошельки (нулевой баланс, ноль транзакций) отбрасываются — они не несут информации и заполнили бы диск бесконечным балластом.

---

## Формат статистики (`stats.txt`)

Каждый запуск добавляет новый блок сессии. Существующие данные не затираются.

```
=== Session 29.06.2026 11:52 ===
[11:53] Time: 0h 01m 00s | Checked: 900  | Used: 0 | Found: 0
[11:54] Time: 0h 02m 00s | Checked: 1800 | Used: 1 | Found: 0
=== End of session 29.06.2026 11:54 ===

=== Session 30.06.2026 09:00 ===
...
```

---

## Технический стек

| Компонент | Реализация |
|-----------|-----------|
| Генерация seed | BIP39, `crypto/rand` |
| Деривация ключей | BIP32/BIP44, HMAC-SHA512 |
| Эллиптическая кривая | secp256k1 через `btcsuite/btcd` |
| Кодирование адреса | SHA256 + RIPEMD160 + Base58Check |
| Деривация seed | PBKDF2-HMAC-SHA512 (`golang.org/x/crypto`) |
| Rate limiting | Token bucket (`golang.org/x/time/rate`) |

---

## Автор

**Ground_Zerro** · Pirates Gold v.1.2
