# CLI. Простой интерпретатор командой строки

Поддерживает команды:
- cat [FILE] - вывести на экран содержимое файла
- echo - вывести на экран свой аругмент (или аргументы)
- wc [FILE] - вывести количество строк, слов и байт в файле
- pwd - распечатать текущую директорию
- exit - выйти из интерпретатора

Дополнительно поддерживаются:
- Одинарыне и двойные кавычки (full и weak quoting)
- Окружение (команды вида "имя=значение), оператор $
- Вызов внешней программы через Process 
- Пайплайны (оператор "|")

### Как запустить

TODO


### Архитектура
Архитектура состоит из четырёх основных функциональных областей:
1. Контекст Сессии
    1. Shell: Главный цикл программы. Он отвечает за чтение пользовательского ввода и передачу его на исполнение
    2. Environment: Хранилище переменных окружения (`map[string]string`), доступное всем этапам обработки и исполнения
2. Анализ и Парсинг
    1. InputProcessor: Отвечает за всю работу с пользовательской строкой. Преобразует сырой ввод в структурированный список команд, готовых к запуску
3. Исполнение и Оркестрация
    1. PipelineRunner: управляет исполнением комманд (`[]CommandDescription`)
    * Создаёт и связывает пайпы между командами
    * Обрабатывает перенаправления в/из файлов.
    * Применяет переменные окружения перед запуском каждой команды
    2. CommandFactory: Фабрика возвращает конкретный объект, реализующий интерфейс `Command`
4. Команда (Интерфейс)
    1. Определяет единый контракт для всех команд (`Execute`)
    2. Включает реализации для команд `Cat`, `wc`, ... и `ExternalCommand` для запуска внешних исполняемых файлов

### Модель данных команды
```go
type CommandDescription struct {
    Name        string
    Arguments   []string
    FileOutPath *string // Используется для перенаправления >
    FileInPath  *string // Используется для перенаправления <
    IsPipedOut  bool    // Флаг: вывод направляется в pipe к следующей команде
    IsPipedIn   bool    // Флаг: ввод берётся из pipe от предыдущей команды
}
```

Для присваивание переменных окружений также используется структура `CommandDescription`, в которой `Name` равен '=' и массив `Arguments` из двух элементов: куда и что присваиваем.

## Диаграмма
```mermaid
classDiagram
    direction LR

    class Shell {
        +Run(): int
    }
    class Environment {
        Variables map<string, string>
        +Get(key): string
        +Set(key, value)
    }

    Shell --> Environment : uses

    class CommandDescription {
        <<data>>
        +Name: string
        +Arguments: []string
        +FileOutPath: *string
        +FileInPath: *string
        +IsPipedOut: bool
        +IsPipedIn: bool
    }

    class InputProcessor {
        +Process(input, env): []CommandDescription
    }

    Shell --> InputProcessor : uses
    InputProcessor "1" --> "*" CommandDescription : creates

    class PipelineRunner {
        +Execute(pipeline: []CommandDescription, env): int
        -createPipesAndRun(description, inputStream, outputStream)
    }
    
    class CommandFactory {
        +GetCommand(description: CommandDescription): Command
    }

    Shell --> PipelineRunner : uses
    PipelineRunner --> CommandFactory : resolves
    PipelineRunner ..> Command : invokes Execute()

    class Command {
        <<interface>>
        +Execute(in, out, env): int
    }

    class AssignEnvCommand
    class CatCommand
    class EchoCommand
    class WcCommand
    class PwdCommand
    class ExitCommand
    class ExternalCommand
    
    AssignEnvCommand --|> Command
    CatCommand --|> Command
    EchoCommand --|> Command
    WcCommand --|> Command
    PwdCommand --|> Command
    ExitCommand --|> Command
    ExternalCommand --|> Command

```

#### Особенности архитектуры
1. Если exit находится вначале или в середине пайплайна, то он игнорируется
   ```
    echo "1" | exit | echo "2"
   ```
2. Если exit находится в конце пайплайна, то процесс завершается

### Контакты
1. Алексей Лимонов (tg:@olexvp)
2. Эдуард Зарипов (tg:@edikgoose)
3. Артём Мурашко (tg:@art22m)