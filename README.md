# Redux & Gunpack Fix — Rockstar GTA 5

<p align="center">
  <img src="assets/logo.png" alt="Logo" width="200"/>
</p>

<p align="center">
  Утилита для игроков <strong>MAJESTIC RP</strong> / <strong>GTA 5 RP</strong> на Rockstar-версии игры.<br/>
  Автоматически подставляет файлы <strong>Redux</strong> и <strong>Gunpack</strong> в папку игры пока GTA5 не запущена.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Windows-10%2F11-0078D4?logo=windows" alt="Windows"/>
  <img src="https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go" alt="Go"/>
  <img src="https://img.shields.io/badge/version-2.0.0-brightgreen" alt="Version"/>
</p>

---

> **Только для Rockstar Launcher.** На Steam и Epic Games версиях это не требуется.

---

## Зачем это нужно

На **Rockstar-версии** GTA 5 файлы игры перед запуском проверяются и восстанавливаются из папки backup. Из-за этого моды (Redux, Gunpack) слетают при каждом запуске игры.

Эта программа следит за GTA5.exe и как только игра закрывается — автоматически возвращает нужные файлы на место, чтобы к следующему запуску всё было готово.

---

## Быстрый старт

1. Скачайте `auto-redux-gunpack.exe` со страницы [Releases](../../releases)
2. Запустите **от имени администратора**
3. Введите пути к папкам (программа подскажет)
4. Готово — программа уходит в фон и работает сама

---

## Первый запуск — мастер настройки

При первом запуске программа задаст несколько вопросов:

```
  +------------------------------------------+
  |  Redux & Gunpack Fix  --  Первая настройка  |
  +------------------------------------------+

  gunpack-new (папка с новыми файлами gunpack): C:\mods\gunpack
  gunpack-old (куда копировать gunpack):        C:\Games\GTA5\gunpack
  redux-new   (папка с новыми файлами redux):   C:\mods\redux
  redux-old   (куда копировать redux):          C:\Games\GTA5\redux
  Путь к GTA5.exe [C:\Users\...\backup\GTA5.exe]: (Enter — оставить по умолчанию)

  Добавить в автозапуск Windows? (y/n) [n]: y
```

После этого консоль закрывается и программа работает тихо в фоне.

### Изменить настройки

```bat
auto-redux-gunpack.exe --config
```

---

## Как работает

Rockstar Launcher перед запуском игры восстанавливает оригинальные файлы из папки backup — это перезаписывает Redux и Gunpack. Программа перехватывает этот момент.

**Схема работы:**

```
[Rockstar открыт, GTA не запущена]
         опрос каждые 300мс
              ↓
         GTA5.exe появилась?  ──НЕТ──► продолжаем ждать
              │ ДА
              ▼
     немедленно копируем файлы  (все три операции параллельно)
              │
     gunpack-new  ──►  gunpack-old
     redux-new    ──►  redux-old
     redux-new    ──►  altv-majestic\backup
              │
    уведомление в трее
              ↓
     ждём пока GTA работает  (опрос каждые 2 сек)
              ↓
     GTA закрылась → снова ждём запуска
```

Копирование срабатывает **ровно один раз** при каждом запуске GTA5 — пока игра ещё грузится. Файлы, которые не изменились, пропускаются. Определение процесса через Win32 API — без запуска внешних команд, мгновенно.

---

## config.json

Настройки хранятся в файле `config.json` рядом с `.exe`:

```json
{
  "gunpack_new":  "C:\\mods\\gunpack",
  "gunpack_old":  "C:\\Games\\GTA5\\gunpack",
  "redux_new":    "C:\\mods\\redux",
  "redux_old":    "C:\\Games\\GTA5\\redux",
  "gta_exe_path": "C:\\Users\\ИМЯ\\AppData\\Local\\altv-majestic\\backup\\GTA5.exe",
  "auto_run":     true
}
```

| Поле | Описание |
|---|---|
| `gunpack_new` | Папка с актуальными файлами gunpack |
| `gunpack_old` | Куда копировать gunpack (папка игры) |
| `redux_new` | Папка с актуальными файлами redux |
| `redux_old` | Куда копировать redux (папка игры) |
| `gta_exe_path` | Полный путь к GTA5.exe для точного определения запуска |
| `auto_run` | `true` — запускать вместе с Windows |

---

## Автозапуск

При включении автозапуска программа прописывается в реестр:

```
HKCU\Software\Microsoft\Windows\CurrentVersion\Run
  auto-redux-gunpack = "C:\...\auto-redux-gunpack.exe -autostart"
```

### Как отключить автозапуск

**Через перенастройку** (рекомендуется):
```bat
auto-redux-gunpack.exe --config
```
На вопрос об автозапуске ответить `n`.

**Через командную строку:**
```bat
reg delete HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v auto-redux-gunpack /f
```

**Через диспетчер задач:** вкладка «Автозагрузка» → отключить `auto-redux-gunpack`.

---

## Флаги

| Флаг | Описание |
|---|---|
| `--config` | Сбросить настройки и запустить мастер заново |
| `--autostart` | Запуск в фоне без консоли (используется реестром) |
| `--version` | Показать версию |

---

## Логи

Все события записываются в `app.log` рядом с `.exe`. Если что-то не копируется — смотрите туда.

---

## Сборка из исходников

Требуется Go 1.21+:

```bat
go build -o auto-redux-gunpack.exe .
```

Для пересоздания иконки / манифеста:
```bat
go install github.com/akavel/rsrc@latest
rsrc -manifest assets/manifest.xml -ico assets/app.ico -o app.syso
```

---

## Структура проекта

```
rockstar-redux-gunpack-fix/
├── internal/
│   ├── config.go       # Конфиг: загрузка, сохранение, ввод
│   ├── fileops.go      # Параллельное копирование, пропуск неизменённых файлов
│   ├── process.go      # Определение запуска GTA5.exe
│   ├── windows.go      # Реестр, скрытие консоли, уведомления
│   └── logger.go       # Логирование в файл + консоль
├── assets/
│   ├── app.ico
│   ├── logo.png
│   └── manifest.xml    # Исходник для app.syso (запрос прав администратора)
├── scripts/
│   └── remove_autorun.bat
├── main.go
├── app.syso            # Скомпилированные иконка + манифест
├── go.mod
└── .gitignore
```

---

## Требования

- Windows 10 / 11
- Права администратора
- GTA 5 Rockstar Launcher версия
