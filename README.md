# go-musthave-diploma-tpl

Шаблон репозитория для индивидуального дипломного проекта курса «Go-разработчик»

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.

# ToDo

Добавлена клиентская часть. Адрес: переменная окружения ОС RUN_ADDRESS или флаг -a;

### Found 56 TODO items in 4 files
#### gofermart
* cmd
* internal
* handlers
* api.go
* (23, 6) // 1 TODO: Регистрация пользователя
* (31, 9) // 1.1 TODO: Проверка на gzip
* (74, 9) // 1.2 TODO: Регистрация пользователя в БД.
* (75, 11) // 1.2.1 TODO: Ищем пользовотеля в таблице БД. Если находим, то не создаем. Пароль кэшируется
* (78, 9) // 1.3 TODO: Создание токена
* (85, 12) // 1.3.1 TODO: Если пользователь добавлен создаем токен
* (99, 9) // 1.4 TODO: Добавление токена в Header
* (106, 6) // 2 TODO: Аутентификации пользователя
* (115, 9) // 1.1 TODO: Проверка на gzip
* (150, 9) // 2.1 TODO: Аутентификации пользователя в БД
* (158, 9) // 2.2 TODO: Создание токена
* (165, 9) // 2.2 TODO: Добавление токена в Header
* (171, 6) // 3 TODO: Добавление нового ордера
* (209, 4) //TODO: Проверка на Луна
* (235, 9) // 3.1 TODO: Добавление нового ордера в БД.
* (236, 11) // 3.1.1 TODO: Ищем ордер по номеру. Если не находим, то создаем
* (252, 6) // 4 TODO: Списание баллов лояльности
* (297, 9) // 4.1 TODO: Списание баллов лояльности в БД
* (298, 11) // 4.1.1 TODO: Получаем баланс начисленных, списанных баллов
* (299, 11) // 4.1.2 TODO: Если начисленных баллов больше, чем списанных, то разрешаем спсание
* (300, 11) // 4.1.3 TODO: Добавляем запись с количеством списанных баллов
* (305, 6) // 5 TODO: Получение списка ордеров по токену
* (315, 9) // 5.1 TODO: Получение списка ордеров по токену в БД
* (316, 11) // 5.1.1 TODO: Из токена получаем имя пользователя
* (317, 11) // 5.1.2 TODO: По имени пользователя получаем ордера
* (337, 11) // 5.1.4 TODO: Выводим список ордеров
* (346, 6) // 6 TODO: Для тестирования сделал API для продвижения ордера на следующий (рандомный) этап (/api/user/orders-next-status)
* (354, 9) // 6.1 TODO: Двигаем ордер на следующий этап
* (355, 11) // 6.1.1 TODO: Получаем спсок ордеров по пользователю из токена
* (356, 11) // 6.1.2 TODO: Назначаем следующий этап. Если это статус PROCESSING, тогда выбираем рандомно INVALID или PROCESSED
* (357, 11) // 6.1.3 TODO: Устанавливаем статус и текущую дату соответствующей колонки
* (358, 11) // 6.1.4 TODO: Если это финальный этап (PROCESSED), рассчитываем баллы лояльности. Рандомное число между 100.10 и 501.98
* (359, 11) // 6.1.5 TODO: Добавляем баллы в ДБ
* (369, 6) // 7 TODO: получаем баланс пользователя
* (381, 9) // 7.1 TODO: Получаем баланс пользователя
* (382, 9) // 7.1 TODO: По токену получаем пользователя
* (383, 9) // 7.2 TODO: По пользовотелю получаем общий баланс начисленных и списанных баллов
* (400, 9) // 7.3 TODO: Выводим в формате JSON
* (408, 6) // 8 TODO: Получение информации о выводе средств
* (421, 9) // 8.1 TODO: Получение информации о выводе средств в разрезе ордера
* (432, 9) // 8.2 TODO: Упаковка овета в JSON
* (439, 9) // 8.2 TODO: Вывод овета в JSON
* (446, 6) // 9 TODO: Взаимодействие с системой расчёта начислений баллов лояльности
* (452, 9) // 9.1 TODO: Запускаем горутину с номером и каналом, где будет хранится ответ черного ящика
* (453, 11) // 9.1.1 TODO: Горутина запрашивает ответ от черного ящика.
* (454, 11) // 9.1.2 TODO: Если статус ответа не 429, то в канал пишется ответ горутина заканчивает свою работу
* (455, 11) // 9.1.3 TODO: Если статус ответа 429, то горутина засыпает на секунду и повторяет запрос к черному ящику
* (456, 13) // 9.1.3.1 TODO: так крутится пока не будет статус не 429
* (459, 9) // 9.2 TODO: Добавляет данные в БД. Вечный цикл с прослушиванием канала.
* (460, 11) // 9.2.1 TODO: Если в канале есть данные, то в БД добавляется запись начисления баллов ллояльности
* (461, 11) // 9.2.2 TODO: Если запись с начисление по ордеру есть в базе, то вторая запись не происходит
* init.go
* (12, 6) // 1 TODO: инициализация роутера и хендлеров
* (37, 6) // 2 TODO: инициализация базы данных
* (46, 6) // 3 TODO: инициализация конфигурации
* postgresql
* services.go
* (417, 4) //TODO: Добавляем спсанные баллы
* token
* token.go
* (52, 17) // IsAuthorized TODO: Проверка аутентификации


