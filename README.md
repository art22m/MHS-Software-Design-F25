# CLI. Простой интерпретатор командой строки

Поддерживает команды:
- `cat FILE` - вывести на экран содержимое файла
- `echo [ARGS...]` - вывести на экран свой аргумент (или аргументы)
- `wc FILE` - вывести количество строк, слов и байт в файле
- `pwd` - распечатать текущую директорию
- `exit` - выйти из интерпретатора
- `VAR=value` - присвоить значение переменной окружения
- Внешние команды - запуск исполняемых файлов из системы

Дополнительно поддерживаются:
- Переменные окружения (синтаксис `$VAR` и `${VAR}`)
- Перенаправление ввода/вывода (`<` и `>`)
- Множественные команды через разделитель `;`
- Вызов внешних программ через `os/exec`

### Как запустить

```bash
go run cmd/main.go
```

Или собрать исполняемый файл:
```bash
go build -o cli cmd/main.go
./cli
```

Запуск тестов:
```bash
go test -v ./...
```


### Архитектура
Архитектура состоит из четырёх основных функциональных областей:

1. **Контекст Сессии**
    - **Shell**: Главный цикл программы (REPL). Отвечает за чтение пользовательского ввода и передачу его на исполнение
    - **Environment**: Хранилище переменных окружения (`map[string]string`), доступное всем этапам обработки и исполнения

2. **Анализ и Парсинг**
    - **InputProcessor**: Отвечает за всю работу с пользовательской строкой. Преобразует сырой ввод в структурированный список команд `[]CommandDescription`, готовых к запуску
    - Поддерживает разделение команд по `;`, присвоение переменных, перенаправления ввода/вывода

3. **Исполнение и Оркестрация**
    - **PipelineRunner**: управляет последовательным исполнением команд (`[]CommandDescription`)
      - Обрабатывает перенаправления в/из файлов (`<` и `>`)
      - Применяет подстановку переменных окружения (поддерживает `$VAR` и `${VAR}`)
      - Вызывает фабрику команд для получения конкретной реализации
    - **CommandFactory**: Фабрика возвращает конкретный объект, реализующий интерфейс `Command` на основе имени команды

4. **Команда (Интерфейс)**
    - Определяет единый контракт для всех команд: `Execute(in, out *os.File, env Env) (retCode int, exited bool)`
    - Включает реализации для команд `Cat`, `Echo`, `Wc`, `Pwd`, `Exit`, `EnvAssignment` и `ExternalCommand` для запуска внешних исполняемых файлов

### Модель данных команды
```go
type CommandDescription struct {
    name        CommandName  // Имя команды или тип операции
    arguments   []string     // Аргументы команды
    fileInPath  string       // Путь для перенаправления ввода (<)
    fileOutPath string       // Путь для перенаправления вывода (>)
    isPiped     bool         // Флаг: команда является частью pipeline (зарезервировано для будущей реализации)
}
```

Для присваивания переменных окружения используется структура `CommandDescription`, в которой `name` равен `EnvAssignmentCmd` (символ `$`) и массив `arguments` из двух элементов: имя переменной и значение.

## Диаграмма
```mermaid
classDiagram
    direction TB

    class Shell {
        -inputProcessor: InputProcessor
        -runner: PipelineRunner
        -env: Env
        +Run() int
    }

    class Env {
        <<interface>>
        +Get(key): (string, bool)
        +Set(key, value)
    }

    class Environment {
        -variables: map~string,string~
        +Get(key): (string, bool)
        +Set(key, value)
    }

    Environment ..|> Env

    Shell --> Env : uses
    Shell --> InputProcessor : uses
    Shell --> PipelineRunner : uses

    class CommandDescription {
        <<data>>
        -name: CommandName
        -arguments: []string
        -fileInPath: string
        -fileOutPath: string
        -isPiped: bool
    }

    class InputProcessor {
        <<interface>>
        +Parse(input): []CommandDescription
    }

    InputProcessor "1" --> "*" CommandDescription : creates

    class PipelineRunner {
        <<interface>>
        +Execute(pipeline, env): (int, bool)
    }

    class CommandFactory {
        <<interface>>
        +GetCommand(CommandDescription): (Command, error)
    }

    PipelineRunner --> CommandFactory : uses
    PipelineRunner ..> Command : invokes Execute()

    class Command {
        <<interface>>
        +Execute(in, out, env): (int, bool)
    }

    class envAssignmentCmd {
        -env: Env
        -key: string
        -value: string
    }
    class catCommand {
        -filePath: string
    }
    class echoCommand {
        -args: []string
    }
    class wcCommand {
        -filePath: string
    }
    class pwdCommand
    class exitCommand
    class externalCommand {
        -name: string
        -args: []string
    }

    envAssignmentCmd ..|> Command
    catCommand ..|> Command
    echoCommand ..|> Command
    wcCommand ..|> Command
    pwdCommand ..|> Command
    exitCommand ..|> Command
    externalCommand ..|> Command

    CommandFactory --> Command : creates

```

#### Особенности архитектуры
1. Команды выполняются последовательно, одна за другой
2. Команда `exit` завершает работу shell с кодом возврата последней выполненной команды
3. Переменные окружения видны всем командам в рамках одной сессии
4. Перенаправление ввода/вывода работает независимо для каждой команды
5. Неизвестные команды передаются в систему для запуска как внешние программы (через `os/exec`)

### Контакты
1. Алексей Лимонов (tg:@olexvp)
2. Эдуард Зарипов (tg:@edikgoose)
3. Артём Мурашко (tg:@art22m)