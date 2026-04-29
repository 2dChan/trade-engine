# Для разработчиков

Спасибо, что хотите внести вклад в этот проект!

## Оглавление

1. [Структура проекта](#структура-проекта)
2. [Добавление нового модуля](#добавление-нового-модуля)
3. [Политика управления историей](#политика-управления-историей)
4. [Кодстайл](#кодстайл)
5. [Лицензия](#лицензия)

## Структура проекта

**core**

- Language: Go
- Shared core libraries

**adapters**

- Language: Go
- Broker adapters implementing `core/broker` interfaces
- Current modules: `adapters/bcs`, `adapters/tinvest`

**botkit**

- Language: Go
- Public runtime/SDK for building bots outside this monorepo

**bots**

- Language: Go
- First-party trading bots maintained in this monorepo

**dashboard/backend**

- Language: Go

**dashboard/frontend**

- Language: TypeScript
- Package Manager: pnpm
- Frameworks: Svelte, SvelteKit

## Добавление нового модуля

- Добавьте ваш модуль в раздел "Структура проекта" файла CONTRIBUTING.md с указанием используемого
  стека.
- Если модуль на Go — создайте `go.mod` и добавьте путь к модулю в `go.work`.
- Настройте интеграцию с CI: добавьте модуль в существующий workflow либо создайте новый workflow с
  этапами проверки линта, кодстайла, сборки и тестирования. Добавьте модуль в Dependabot для
  автоматического отслеживания и обновления зависимостей через pull request.

## Политика управления историей

- **Запрещены merge-коммиты** — используйте только rebase- или squash-merge при объединении
  изменений.
- **Запрещён прямой push в ветку main** — все изменения должны проходить через Pull Request.
- **Запрещено принимать в ветку main файлы без лицензионного заголовка** —
  [шаблон заголовка](#шаблон-лицензионного-заголовка).
- Для сообщений коммитов используйте стандарт
  [Conventional Commits](https://www.conventionalcommits.org/ru/v1.0.0/).
- Для названий Pull Request используйте стандарт
  [Conventional Commits](https://www.conventionalcommits.org/ru/v1.0.0/).
- Для именования веток используйте шаблон `type/<название>-<номер_issues>`
  - `type` — категория изменений: feature, fix, chore, ci, docs, question
  - `<название>` — название issue или краткое описание задачи (используйте "-" вместо пробелов)
  - `<номер_issues>` — номер соответствующего issue (при наличии)

## Кодстайл

### Общие

- Все комментарии в коде должны быть написаны только на английском языке.
- Каждый исходный файл должен начинаться с лицензионного заголовка (см. [шаблон](#шаблон-лицензионного-заголовка)).

### Go

Принят стандартный стиль [Google](https://google.github.io/styleguide/go/).

Дополнительные ресурсы:

- [Effective Go](https://go.dev/doc/effective_go)
- [Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Test Comments](https://go.dev/wiki/TestComments)

### TypeScript

[Eslint](https://eslint.org) обнаружит большинство проблем со стилем, которые могут быть в вашем
коде. Вы можете проверить состояние стиля кода, просто запустив команду `pnpm lint`.

## Лицензия

Проект распространяется по лицензии **GNU Affero General Public License v3.0** (см. файл `LICENSE`).

### Шаблон лицензионного заголовка

Правила:

- Каждый исходный файл должен начинаться с лицензионного заголовка.
- Нужно указывать имя, фамилию или псевдоним.
- Если заголовок уже существует, вы можете добавить себя в Copyright (C) через запятую.
- Указывается год первого появления файла (или диапазон, если файл поддерживается долго).

```text
Copyright (C) 2026 <name or nickname>
Licensed under the GNU Affero General Public License v3.0 or later.
See the LICENSE file in the project root for the full license text.
```
