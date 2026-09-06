# go-mcr-srvc

Goでマイクロサービスの構成要素を検証するためのサンプルバックエンドです。ユーザー検索と名前のバリデーションルール管理を題材に、Clean Architecture、DDDの一部、HTTP/OpenAPI、gRPC、PostgreSQL、Redis、外部サービス連携を実装しています。

本番サービスとして完成したアプリケーションではなく、層間の依存関係、ドメインロジックの配置、リクエストコンテキストのサービス間伝搬を確認するためのリファレンス実装です。

コントリビューション方法は[`CONTRIBUTING.md`](CONTRIBUTING.md)、脆弱性の報告方法は[`SECURITY.md`](SECURITY.md)を参照してください。このプロジェクトは[MIT License](LICENSE)で公開されています。

## 主な機能

- EchoとOpenAPIによるHTTP API
- ユーザーの名前またはメールアドレスによる検索
- 名前のバリデーションルールの登録・更新・削除
- PostgreSQLを永続化先、Redisをcache-aside方式のキャッシュとして使用
- gRPC server/client、Auth0 client、Kafka publisher/consumerの実装例
- HTTP、gRPC metadata、Pub/Sub header間のリクエストコンテキスト変換
- domain、usecase、adapter、framework各層のユニットテスト

## 使用技術

- Go 1.27 / Echo / OpenAPI / oapi-codegen
- gRPC / Protocol Buffers
- PostgreSQL / GORM / Redis
- Kafka（Confluent Kafka Go）/ Auth0
- Docker Compose

## セットアップ

### 必要なもの

- Docker 20.10.5以降
- Docker Compose v2
- Go 1.27（Docker外でテストやコード生成を実行する場合）
- `make`

### 環境変数

サンプルをコピーしてローカル用の設定ファイルを作成します。

```bash
cp backend/internal/env/local.env.example backend/internal/env/local.env
```

| 変数 | 用途 | サンプル既定値 |
| --- | --- | --- |
| `GO_ECHO_PORT` | HTTP serverの公開ポート | `53234` |
| `GRPC_PORT` | gRPC serverの公開ポート | `43456` |
| `BACKEND_HOST` | backendのサービス名 | `backend_app` |
| `POSTGRES_HOST` | PostgreSQLのサービス名 | `postgres_db` |
| `POSTGRES_FRONT_PORT` | ホスト側のPostgreSQLポート | `54932` |
| `POSTGRES_BACK_PORT` | コンテナ側のPostgreSQLポート | `5432` |
| `POSTGRES_DB` | データベース名 | `app` |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` | PostgreSQL認証情報 | 無効な開発用サンプル |
| `REDIS_ADDR` | Redis接続先 | `redis:6379` |
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka broker接続先 | `kafka:9092` |
| `AUTH0_*` | Auth0接続・認証情報 | 無効なサンプル |
| `AWS_*` | LocalStack / Secrets Manager設定 | 無効なサンプル |

すべての項目は[`local.env.example`](backend/internal/env/local.env.example)を参照してください。実際の認証情報はコミットしないでください。

### 起動

```bash
make install-tools
make build
make up
```

`make up`はforegroundで起動します。停止は別のterminalから実行します。

```bash
make stop
```

既定ではHTTP APIが`http://localhost:53234`、gRPC serverが`localhost:43456`で待ち受けます。

## API

OpenAPI定義は[`openapi.yaml`](backend/internal/1_framework/in/go-echo/openapi/openapi.yaml)にあります。

| Method | Path | 内容 | 状態 |
| --- | --- | --- | --- |
| `GET` | `/v1/health` | ヘルスチェック | 実装済み |
| `GET` | `/v1/users` | 名前またはメールアドレスでユーザーを検索 | 実装済み |
| `POST` | `/v1/users` | ユーザーを作成 | 実装済み |
| `GET` | `/v1/validation-word-rules` | バリデーションルール一覧 | 実装済み |
| `POST` | `/v1/validation-word-rules` | バリデーションルール追加 | 実装済み |
| `PUT` | `/v1/validation-word-rules` | バリデーションルール更新 | 実装済み |
| `DELETE` | `/v1/validation-word-rules` | バリデーションルール削除 | 実装済み |
| `GET` | `/v1/to-pubsub` | Kafkaへのテストメッセージ送信 | 現在無効 |

### HTTP APIの実行例

```bash
curl http://localhost:53234/v1/health

curl "http://localhost:53234/v1/users?name=a"

curl "http://localhost:53234/v1/validation-word-rules?targetType=name&isBlacklist=true"

curl -X POST http://localhost:53234/v1/validation-word-rules \
  -H 'Content-Type: application/json' \
  -d '{"targetType":"name","isBlacklist":true,"word":"禁止語"}'
```

より多くの例は各handler配下の`rest.http`ファイルにあります。

### gRPCの実行例

```bash
grpcurl -plaintext \
  -import-path backend/internal/1_framework/parameter/grpc \
  -proto user.proto \
  -d '{"v1UserParameter":{"name":"a"}}' \
  localhost:43456 grpc_parameter.v1.UserService/GetUserListByCondition
```

定義は[`user.proto`](backend/internal/1_framework/parameter/grpc/user.proto)にあります。

## テストと静的解析

```bash
# ユニットテスト
make gotest

# race detector
cd backend && GOTOOLCHAIN=go1.27.0 go test -race ./...

# lint
make lint

# 既知の脆弱性を検査
make govulncheck
```

GitHub Actionsでも、`main`へのpushと`main`を対象とするPull Requestに対して、上記の検査を実行します。

## アーキテクチャ

`backend/internal`以下を、外側から内側へ依存する4つの層に分割しています。

```text
1_framework → 2_adapter → 3_usecase → 4_domain
```

![Clean Architecture](doc/ca.drawio.png)

### HTTPからDBまで

```text
HTTP request
  → Echo middleware（RequestContextを生成）
  → OpenAPI handler
  → controller
  → usecase（処理順序・transaction境界・domain rule適用）
  → DB gateway
  ├→ Redis client（cache lookup / invalidation）
  └→ PostgreSQL client（永続化）
  → HTTP response DTO
```

HTTP requestの`context.Context`はcontroller、usecase、gatewayを経由し、GORMの`WithContext`やRedis commandまで伝搬します。

### gRPC連携

```text
HTTP / gRPC request
  → controller → usecase → external gateway → gRPC client
  → gRPC metadataへRequestContextを設定
  → downstream gRPC server
```

### Pub/Sub連携

```text
HTTP / application request
  → controller → usecase → external gateway → publisher
  → Kafka headerへRequestContextを設定
  → consumer → headerからRequestContextを復元
```

Kafka接続は現在無効化されているため、このフローは実装例として配置されています。

## DDDの取り入れ方

DDDの考え方を取り入れていますが、package名はプロジェクト内で扱いやすいようにカスタマイズしています。

| このプロジェクトの用語 | DDDで近い概念 | 役割 |
| --- | --- | --- |
| `primitive_object` | Primitive / base type wrapper | Goの基本型に対する共通ルール |
| `type_object` | Value Object | ドメイン固有の値と値単体のルール |
| `group_object` | Entity / Aggregate | 識別子、ライフサイクル、状態変更 |
| `service_object` | Domain Service | 複数オブジェクト間の業務ルール |

ドメインロジックは[`backend/internal/4_domain`](backend/internal/4_domain)に凝集し、DB、HTTP、gRPC、context、loggerなどの外部都合を持ち込みません。

- [`primitive_object`](backend/internal/4_domain/primitive_object/_README.md)：基本データ型の共通ルール
- [`type_object`](backend/internal/4_domain/type_object/_README.md)：`Email`、`UserID`などの値オブジェクト
- [`group_object`](backend/internal/4_domain/group_object/_README.md)：`User`、`UserEmployment`などのEntity / Aggregate相当
- [`service_object`](backend/internal/4_domain/service_object/_README.md)：複数オブジェクト間の業務ルール

### usecaseとdomainの境界

usecaseはtransaction境界、gatewayの呼び出し順序、domain ruleの適用など、外部I/Oを含む処理のorchestrationを担当します。domainは値の妥当性、Entityの更新条件、Aggregateの不変条件など、外部I/Oに依存しない業務判断を担当します。

### あえて実装していないDDDの設計

| 設計 | このプロジェクトでの扱い |
| --- | --- |
| DDD標準用語をそのまま使うpackage分割 | 独自の4分類として配置 |
| Repository / Unit of Workをdomain層に配置 | gatewayとtransaction境界をusecase側で制御 |
| Domain Event / Event Sourcing | 必要になった段階で追加 |
| Bounded Contextごとのdomain分割 | 業務領域やサービス境界が増えた段階で検討 |
| Aggregate Rootの共通基底型 | 具体型と小さなinterfaceを使用 |
| 状態を持つDomain Service | `service_object`はstatelessにする |

## リクエストコンテキスト

HTTP middlewareで`RequestStartTime`、`TraceID`、`ClientIP`、`UserAgent`、`UserID`、`AccessToken`、`TenantID`、`Locale`、`TimeZone`、`PermissionList`を生成・取得し、gRPC metadataまたはPub/Sub headerへ変換します。

`TimeOutMillSecond`は`RequestStartTime`から内部計算するtimeout判定用の値で、現在はサービス間に伝搬しません。詳細は[`request_context/model.go`](backend/internal/1_framework/middleware/request_context/model.go)を参照してください。

## プロジェクト固有のコーディング規約

- 変数名は長さではなく、スコープ内で実態が読み取れることを優先する
- index変数は`i`ではなく`index`とする
- 関数とメソッドの戻り値は原則として名前付き戻り値にする
- 引数と戻り値は1項目ずつ改行する
- 層をまたぐinterfaceは依存先を表す`ToXXXX`形式とする
- type / method名にはDBやtransportではなく業務語彙を反映する

## 現時点の制約・未実装

- KafkaとLocalStackのDocker Composeサービスはコメントアウトされている
- Pub/Sub producerの初期化とconsumerの起動は無効化されており、`GET /v1/to-pubsub`は利用できない
- Auth0用handlerは実装例があるが、現在のHTTP serverにはroute登録されていない
- OpenAPI定義と生成コードの更新は手動で行う
- PostgreSQLとRedisのテストはmock中心で、実サービスを使うE2E testはない
- production向けのデプロイ設定、監視、メトリクス、分散トレーシングは未実装
- 単一サービス内のサンプルであり、実際に複数サービスを展開する構成は含まれない
