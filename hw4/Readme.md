# Домашнее задание №4: Поиск данных по XML и тестовое покрытие

## Цель задания

Задание направлено на изучение: 
- Отправки HTTP-запросов и обработки ответов в Go. 
- Работы с GET-параметрами и заголовками (headers). 
- Чтения и парсинга XML-файлов. 
- Реализации полноценного HTTP-сервера для поиска данных. 
- Написания тестов с табличным подходом для покрытия логики клиента на 100%. 
- Генерации отчёта о покрытии кода.

<details>
<summary><strong>Оригинальная формулировка задания</strong></summury>

> Это комбинированное задание по тому, как отправлять запросы, получать ответы, работать с параметрами, хедерами, а так же писать тесты.
> 
> Задание не сложное, основной объёма работы - написание разных условий и тестов, чтобы эти условия удовлетворить.
> 
> У нас есть какой-то поисковый сервис:
> * SearchClient - структура с методом FindUsers, который отправляет запрос во внешнюю систему и возвращает результат, немного преобразуя его. Находится в файле client.go, править нельзя.
> * SearchServer - своего рода внешняя система. Непосредственно занимается поиском данных в файле `dataset.xml`. В продакшене бы запускалась в виде отдельного веб-сервиса, но в вашем колде запустится как отдельный хендлер.
> 
> Требуется:
> * Написать функцию SearchServer в файле `client_test.go`, который вы будете запускать в тесте через тестовый сервер (`httptest.NewServer`, пример использования в `4/http/server_test.go`)
> * Покрыть тестами метод FindUsers, чтобы покрытие файла `client.go` было максимально возможным, а именно - 100%. Тесты писать в `client_test.go`. Но когда вы будете запускать тесты с флагом покрытия - там будет писаться общий процент, какой процент в `client.go` - смотрите в отчете.
> * Так же требуется сгенерировать html-отчет с покрытием. См. пример построения тестового покрытия и отчета в `3/testing/coverage_test.go`.
> * Тесты надо писать полноценные, те не чтобы получить покрытие, а которые реально тестируют ваш код, проверяют возвращаемый результат, граничные случаи и тд. Они должны показывать что SearchServer работает правильно.
> * Из предыдущего пункта вытекает что SearchServer тоже надо писать полноценный
> 
> SearchServer принимает GET-параметры:
> * `query` - что искать. Ищем по полям записи `Name` и `About` просто подстроку, без регулярок. `Name` - это first_name + last_name из xml (вам надо руками пройтись в цикле по записям и сделать такой, автоматом нельзя). Если поле пустое - то возвращаем все записи (поиск пустой подстроки всегда возвращает true), т.е. делаем только логику сортировки
> * `order_field` - по какому полю сортировать. Работает по полям `Id`, `Age`, `Name`, если пустой - то сортируем по `Name`, если что-то другое - SearchServer ругается ошибкой. 
> * `order_by` - направление сортировки (как есть, по убыванию, по возрастанию), в client.go есть соответствующие константы
> * `limit` - сколько записей вернуть
> * `offset` - начиня с какой записи вернуть (сколько пропустить с начала) - нужно для огранизации постраничной навигации
> 
> Дополнительно:
> * Данные для работы лежаит в файле `dataset.xml`
> * Как работать с XML - почти так же как с JSON, смотрите доку https://golang.org/pkg/encoding/xml/ и пример в боте
> * Запускать как `go test -cover`
> * Можно начать с того что вы просто напишите сервер в `main.go` который реализует логику, а потом уже унесите это в `client_test.go`
> * Построение покрытия: `go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html`
> * Документация https://golang.org/pkg/net/http/ может помочь
> * Пользуйтесь табличным тестированием - это когда у вас есть слайс тест-кейсок, которые отличаются параметрами.
> * Вы можете не ограничиваться функцией SearchServer при тестировании, если вам надо проверить какой-то совсем отдельный хитрый кейс, вроде ошибки. Но таких случаев будет немного. В основном всё будет в SearchServer
> * Для покрытия тестом одной из ошибок придётся залезть в исходники функции, которая возвращает эту ошибку, и посмотреть при каких условиях работы или входных данных это происходит. Это клиентская ошибка, т.е. запрос в этом случае в сервер уходить не будет.
> * Не пытайтесь реализовать таймаут подключением к неизвестному IPшнику
> * Блок c NextPage на строке 121 в client.go используется для создания постраничной навигации - я заглядываю в сервер на +1 запись - если она есть - я могу показать следующую страницу
> 
> Объем кода:
> * SearchServer со всеми структурами и всем-всем будет 170-200 строк
> * Тесты 200-300-400 строк, в зависимости от формы - основное там будет список тест-кейсов
> 
> Рекомендуемый план работы:
> 1. Напишите в функции main код который просто по фиксированным параметрам реализует логику SearchServer и выводит в консоль, без http
> 2. Теперь оформите ваш код в http-хендлер, параметры уже не хардкодом, а берите из запроса
> 3. Проверьте запросами из браузера что код отрабатывает
> 4. Теперь начинайте писать тесты в client_test.go
> 5. Реализуйте сначала один тест, который просто делает запрос через SearchClient-а в ваш хттп-хендлер, запущенный через тестовый сервер 
> 6. Теперь постройте отчет и смотрите какой код у вас был вызван, а какой нет
> 7. Начинайте дописывать тест кейсы
> 8. Для ошибок реализуйте отдельный хендлер или хендлеры

</details>

## Компоненты

### 1 client.go (не изменяется) 
SearchClient: структура с методом FindUsers, отправляющим GET-запрос на внешний сервер и преобразующим результат. 
SearchRequest: параметры поиска (Limit, Offset, Query, OrderField, OrderBy).
SearchResponse: список пользователей и флаг NextPage. 

Логика FindUsers включает: 
- Валидацию Limit и Offset. 
- Обработка лимита до 25. 
- Таймауты и сетевые ошибки. 
- Разбор JSON-ответов и ошибок сервера (400, 401, 500). 
- Поддержку NextPage для постраничной навигации. 

### client_test.go 
Реализация SearchServer: 
- Чтение данных из dataset.xml. 
- Объединение first_name + last_name в поле Name.
- Фильтрация по query (подстрока в Name или About).
- Сортировка по Id, Age или Name.
- Поддержка параметров: 
    - limit сколько записей вернуть. 
    - offset с какой записи начинать. 
    - order_field и order_by. 
- Эмуляция ошибок:
    - 401 Unauthorized (нет токена). 
    - 500 InternalServerError (файл не найден или ошибка чтения). 
    - 400 BadRequest (ErrorBadOrderField).

## Результаты

Все тесты проходят успешно:
```
=== RUN   TestSearchClient_AllCases
=== RUN   TestSearchClient_AllCases/Limit_negative
=== RUN   TestSearchClient_AllCases/Offset_negative
=== RUN   TestSearchClient_AllCases/Bad_order_field
=== RUN   TestSearchClient_AllCases/Valid_request_small_limit
=== RUN   TestSearchClient_AllCases/Valid_request_with_query
--- PASS: TestSearchClient_AllCases (0.02s)
    --- PASS: TestSearchClient_AllCases/Limit_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Offset_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Bad_order_field (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_small_limit (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
=== RUN   TestSearchClient_AllCases
=== RUN   TestSearchClient_AllCases/Limit_negative
=== RUN   TestSearchClient_AllCases/Offset_negative
=== RUN   TestSearchClient_AllCases/Bad_order_field
=== RUN   TestSearchClient_AllCases/Valid_request_small_limit
=== RUN   TestSearchClient_AllCases/Valid_request_with_query
--- PASS: TestSearchClient_AllCases (0.02s)
    --- PASS: TestSearchClient_AllCases/Limit_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Offset_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Bad_order_field (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_small_limit (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
=== RUN   TestSearchClient_Timeout
--- PASS: TestSearchClient_Timeout (1.00s)
=== RUN   TestSearchClient_Unauthorized
--- PASS: TestSearchClient_Unauthorized (0.00s)
=== RUN   TestSearchClient_InternalServerError
=== RUN   TestSearchClient_AllCases
=== RUN   TestSearchClient_AllCases/Limit_negative
=== RUN   TestSearchClient_AllCases/Offset_negative
=== RUN   TestSearchClient_AllCases/Bad_order_field
=== RUN   TestSearchClient_AllCases/Valid_request_small_limit
=== RUN   TestSearchClient_AllCases/Valid_request_with_query
--- PASS: TestSearchClient_AllCases (0.02s)
    --- PASS: TestSearchClient_AllCases/Limit_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Offset_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Bad_order_field (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_small_limit (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
=== RUN   TestSearchClient_Timeout
--- PASS: TestSearchClient_Timeout (1.00s)
=== RUN   TestSearchClient_Unauthorized
=== RUN   TestSearchClient_AllCases/Offset_negative
=== RUN   TestSearchClient_AllCases/Bad_order_field
=== RUN   TestSearchClient_AllCases/Valid_request_small_limit
=== RUN   TestSearchClient_AllCases/Valid_request_with_query
--- PASS: TestSearchClient_AllCases (0.02s)
    --- PASS: TestSearchClient_AllCases/Limit_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Offset_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Bad_order_field (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_small_limit (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
--- PASS: TestSearchClient_AllCases (0.02s)
    --- PASS: TestSearchClient_AllCases/Limit_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Offset_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Bad_order_field (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_small_limit (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
    --- PASS: TestSearchClient_AllCases/Offset_negative (0.00s)
    --- PASS: TestSearchClient_AllCases/Bad_order_field (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_small_limit (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
    --- PASS: TestSearchClient_AllCases/Valid_request_small_limit (0.01s)
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
=== RUN   TestSearchClient_Timeout
    --- PASS: TestSearchClient_AllCases/Valid_request_with_query (0.00s)
=== RUN   TestSearchClient_Timeout
=== RUN   TestSearchClient_Timeout
--- PASS: TestSearchClient_Timeout (1.00s)
--- PASS: TestSearchClient_Timeout (1.00s)
=== RUN   TestSearchClient_Unauthorized
--- PASS: TestSearchClient_Unauthorized (0.00s)
=== RUN   TestSearchClient_InternalServerError
--- PASS: TestSearchClient_InternalServerError (0.00s)
PASS
ok      hw4     1.609s
```

- Проверка покрытия (go test -cover) показывает высокий процент покрытия client.go и тесты затрагивают все ключевые ветки. 
- Таймаут, ошибки 401 и 500, а также граничные значения проверены. 
- Реализован поиск по имени и описанию, корректная сортировка и фильтрация по параметрам.