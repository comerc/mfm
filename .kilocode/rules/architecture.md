# Архитектурные паттерны проекта

Проект следует Clean Architecture: Service (Use Cases) и Repo (Repositories).

## Service слой
- `type Service struct` с комментарием `// Service implements Service`
- Интерфейсы зависимостей определяются в Service пакете
- Конструктор: `func New(...) *Service`

## Repo слой
- `type Repo struct` с комментарием `// Repo - репозиторий для [описание]`
- Конструктор: `func New() *Repo`

## Принципы
- Dependency Inversion через интерфейсы
- Single Responsibility
- Testability
- Структуры: Service/Repo, конструкторы: New

## Go конвенции
- Meaningful имена, комментарии к публичным сущностям
- `go fmt`, тесты для новых функций
- Структура: `internal/service/` - бизнес-логика, `internal/repo/` - данные