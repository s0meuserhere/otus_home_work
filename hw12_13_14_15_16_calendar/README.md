#### Результатом выполнения следующих домашних заданий является сервис «Календарь»:
- [Домашнее задание №12 «Заготовка сервиса Календарь»](./docs/12_README.md)
- [Домашнее задание №13 «Внешние API от Календаря»](./docs/13_README.md)
- [Домашнее задание №14 «Кроликизация Календаря»](./docs/14_README.md)
- [Домашнее задание №15 «Докеризация и интеграционное тестирование Календаря»](./docs/15_README.md)

#### Ветки при выполнении
- `hw12_calendar` (от `master`) -> Merge Request в `master`
- `hw13_calendar` (от `hw12_calendar`) -> Merge Request в `hw12_calendar` (если уже вмержена, то в `master`)
- `hw14_calendar` (от `hw13_calendar`) -> Merge Request в `hw13_calendar` (если уже вмержена, то в `master`)
- `hw15_calendar` (от `hw14_calendar`) -> Merge Request в `hw14_calendar` (если уже вмержена, то в `master`)
- `hw16_calendar` (от `hw15_calendar`) -> Merge Request в `hw15_calendar` (если уже вмержена, то в `master`)


#### Кодогенерация API
Для `make generate` нужен компилятор `protoc` - `make install-proto-deps` ставит только плагины
`protoc-gen-go` и `protoc-gen-go-grpc`, сам `protoc` устанавливается отдельно:
- macOS: `brew install protobuf`
- Debian/Ubuntu: `apt install protobuf-compiler`
- другие ОС: https://protobuf.dev/installation/

Плагины для gRPC и HTTP (oapi-codegen): `make install-api-deps`, затем `make generate`.


#### Запуск в Docker
Нужен Docker Compose **2.24 или новее**: тестовый `deployments/docker-compose.test.yaml`
использует тег `!reset`, которого в старых версиях нет. Проверить версию: `docker compose version`.

- `make up` - поднять PostgreSQL, RabbitMQ, миграции и все сервисы. HTTP API: http://localhost:8888, gRPC: `localhost:9191`.
- `make down` - остановить окружение.
- `make integration-tests` - поднять отдельное окружение, прогнать интеграционные тесты и удалить его.


#### Запуск в Kubernetes
Нужны minikube, helm и kubectl. Проверить, что всё установлено: `make check-k8s-deps`.

Helm chart лежит в `deployments/calendar-chart`, настройки по умолчанию в `values.yaml`.

- `make k8s-up` - запустить minikube, собрать образы и установить chart. Повторный запуск подхватывает изменения в коде.
- `make k8s-down` - остановить кластер. Релиз и данные базы сохраняются.

HTTP API доступен через Ingress по хосту `calendar.local`, gRPC по хосту `grpc.calendar.local`:
```bash
kubectl -n ingress-nginx port-forward svc/ingress-nginx-controller 8081:80
curl -H 'Host: calendar.local' http://localhost:8081/hello
grpcurl -plaintext -authority grpc.calendar.local localhost:8081 list
```


**Домашнее задание не принимается, если не принято ДЗ, предшествующее ему.**
