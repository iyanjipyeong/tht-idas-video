# IDAS Video Subscription API

Golang REST API untuk manajemen akses video berdasarkan tier subscription dengan PostgreSQL dan Clean Architecture.

## Requirements

- Go sesuai versi di `go.mod`
- PostgreSQL

## Environment

Salin `.env.example` menjadi `.env`, lalu sesuaikan nilainya:

```bash
cp .env.example .env
```

Isi utama yang perlu diperhatikan:

```env
APP_ADDR=:8080
DATABASE_DRIVER=pgx
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=idas_video
DATABASE_USERNAME=postgres
DATABASE_PASSWORD=postgres
DATABASE_SSLMODE=disable
DATABASE_SCHEMA=public
DATABASE_TIMEZONE=Asia/Jakarta
MIGRATION_DIR=migrations
```

Seeder menyediakan akun demo untuk login:

```json
{
  "email": "demo@example.com",
  "password": "password"
}
```


## Command Reference

| Command | Purpose | Notes |
| --- | --- | --- |
| `make serve` | Menjalankan HTTP API server. | Membutuhkan database PostgreSQL yang bisa diakses dari `.env`. |
| `make build` | Build binary production-ready dengan metadata build. | Output default: `bin/idas-video-api`. |
| `make utest` | Menjalankan seluruh unit test dengan output verbose. | Menggunakan `go test -short -v ./...`. |
| `make mocking` | Generate ulang mock repository context. | Output tunggal: `internal/usecase/outbound/mock_i_repository_context.go`. |
| `make migrate-up` | Menjalankan semua migration `.up.sql`. | Membutuhkan binary `migrate` dan koneksi PostgreSQL. |

## Migration

```bash
make migrate-up
```

Atau langsung:

```bash
migrate -source "file://migrations" -database "postgres://postgres:postgres@localhost:5432/idas_video?sslmode=disable&search_path=public" up
```

## Run

```bash
make serve
```

Server default berjalan di `:8080`. Ubah alamat dengan environment variable `APP_ADDR`.

Command `make serve` menyuntikkan `buildId` dan `buildTime` ke health response melalui `-ldflags`, sehingga metadata build bisa berasal dari pipeline build/CI.



Untuk build binary production-ready:

```bash
make build
```

Output default binary ada di `bin/idas-video-api` dan memakai `-ldflags` yang sama seperti `make serve`.

## Docs

- Database schema: `migrations/*.up.sql` dan `migrations/*.down.sql`
- OpenAPI draft: `docs/openapi.yaml`
- Swagger UI: `http://localhost:8080/docs`
- Raw OpenAPI: `http://localhost:8080/openapi.yaml`

## Project Structure

```text
cmd/api                              Application entrypoint
internal/app                         Composition root, dependency wiring, and HTTP server bootstrap
internal/entity                      Enterprise entities, enums, constants, errors, and business rules
internal/usecase                     Application business logic and use case orchestration
internal/usecase/inbound             Input ports and request/result models per use case
internal/usecase/outbound            Output ports for repositories, services, and cross-boundary logging
internal/adapter/inbound/http        HTTP router, handlers, middleware, and transport response mapping
internal/adapter/inbound/http/observability    HTTP adapter logging helper without infrastructure dependency
internal/adapter/outbound/postgres   GORM-backed PostgreSQL outbound adapter and persistence models
internal/infrastructure/buildinfo    Build metadata used by health response
internal/infrastructure/config       Environment configuration loader and validation
internal/infrastructure/logger       Concrete application logger and usecase logger adapter
migrations                           SQL schema and seed data
docs                                 API documentation
```

## Test

```bash
make utest
```

| Package | Test Name | Purpose |
| --- | --- | --- |
| `internal` | `TestCleanArchitectureImportBoundaries` | Mengunci boundary import agar struktur clean architecture tetap konsisten. |
| `internal/adapter/inbound/http/handler` | `TestHealthHandlerReturnsSuccessEnvelope` | Memastikan health handler selalu mengembalikan success envelope yang konsisten. |
| `internal/adapter/inbound/http/handler` | `TestAuthHandlerLoginReturnsSuccessEnvelope` | Memastikan endpoint login mengembalikan success envelope saat autentikasi berhasil. |
| `internal/adapter/inbound/http/handler` | `TestPaymentHandlerCallbackReturnsSuccessEnvelope` | Memastikan endpoint callback payment mengembalikan success envelope saat payload valid diproses. |
| `internal/adapter/inbound/http/handler` | `TestSubscriptionHandlerGetActiveReturnsSuccessEnvelope` | Memastikan endpoint active subscription mengembalikan success envelope untuk subscription aktif user. |
| `internal/adapter/inbound/http/handler` | `TestSubscriptionHandlerSubscribeReturnsPendingTransactionEnvelope` | Memastikan endpoint subscribe mengembalikan envelope transaksi pending saat request valid. |
| `internal/adapter/inbound/http/handler` | `TestTransactionHandlerListReturnsListEnvelope` | Memastikan endpoint daftar transaksi mengembalikan list envelope yang konsisten. |
| `internal/adapter/inbound/http/handler` | `TestTransactionHandlerGetDetailReturnsSuccessEnvelope` | Memastikan endpoint detail transaksi mengembalikan success envelope untuk transaksi milik user. |
| `internal/adapter/inbound/http/handler` | `TestTransactionHandlerGetTransactionRejectsOtherUserAccess` | Memastikan detail transaksi menolak akses ke transaksi milik user lain. |
| `internal/adapter/inbound/http/handler` | `TestVideoHandlerListReturnsListEnvelope` | Memastikan endpoint daftar video mengembalikan list envelope yang konsisten. |
| `internal/adapter/inbound/http/handler` | `TestVideoHandlerGetVideoReturnsDetailEnvelope` | Memastikan endpoint detail video mengembalikan detail envelope untuk video yang bisa diakses. |
| `internal/adapter/inbound/http/handler` | `TestVideoHandlerGetVideoReturnsUnauthorizedEnvelope` | Memastikan detail video mengembalikan unauthorized envelope saat user belum terautentikasi. |
| `internal/adapter/inbound/http/handler` | `TestVideoHandlerGetVideoReturnsForbiddenEnvelope` | Memastikan detail video mengembalikan forbidden envelope saat tier user tidak cukup. |
| `internal/adapter/inbound/http/handler` | `TestVideoHandlerGetVideoReturnsNotFoundEnvelope` | Memastikan detail video mengembalikan not found envelope saat video tidak ditemukan. |
| `internal/adapter/inbound/http/handler` | `TestVideoHandlerGetVideoRejectsInvalidUUID` | Memastikan `GET /videos/{id}` menolak UUID path yang invalid dengan status `400`. |
| `internal/adapter/outbound/postgres` | `TestMapConstraintErrorMapsTransactionUserForeignKey` | Memastikan foreign key user pada transaksi dipetakan ke domain error yang benar. |
| `internal/adapter/outbound/postgres` | `TestMapConstraintErrorMapsTransactionTierForeignKey` | Memastikan foreign key tier pada transaksi dipetakan ke domain error yang benar. |
| `internal/entity` | `TestCanAccessVideo` | Memastikan aturan akses tier video sesuai business rule Bronze/Silver/Gold. |
| `internal/entity` | `TestIsUUID` | Memastikan validator UUID hanya menerima format UUID yang valid. |
| `internal/usecase` | `TestAuthUsecaseLoginSuccess` | Memastikan login sukses menghasilkan token saat email/password valid. |
| `internal/usecase` | `TestAuthUsecaseAuthenticateAccessTokenSuccess` | Memastikan access token valid bisa di-autentikasi kembali ke `user_id`. |
| `internal/usecase` | `TestPaymentCallbackUsecaseProcessPaymentCallbackPaidSuccess` | Memastikan callback payment `paid` membuat log callback dan mengaktifkan subscription. |
| `internal/usecase` | `TestPaymentCallbackUsecaseRejectsDuplicateTransaction` | Memastikan transaksi callback duplikat ditolak. |
| `internal/usecase` | `TestPaymentCallbackUsecaseProcessesReservedTransaction` | Memastikan callback payment memproses transaksi reserved yang cocok tanpa membuat transaksi baru. |
| `internal/usecase` | `TestPaymentCallbackUsecaseRejectsReservedTransactionMismatch` | Memastikan callback payment ditolak saat payload tidak cocok dengan reserved transaction. |
| `internal/usecase` | `TestPaymentCallbackUsecaseRejectsReservedTransactionWithoutMatchingOrderID` | Memastikan callback payment ditolak saat `order_id` tidak cocok dengan transaksi reserved. |
| `internal/usecase` | `TestPaymentCallbackUsecaseRejectsReservedTransactionWithDifferentGrossAmount` | Memastikan callback payment ditolak saat `gross_amount` berbeda dari transaksi reserved. |
| `internal/usecase` | `TestSubscriptionUsecaseActivateSubscriptionSuccess` | Memastikan aktivasi subscription valid membuat subscription aktif. |
| `internal/usecase` | `TestSubscriptionUsecaseActivateSubscriptionRejectsInvalidInput` | Memastikan input subscription invalid ditolak. |
| `internal/usecase` | `TestSubscriptionUsecaseCreateSubscriptionTransactionSuccess` | Memastikan pembuatan transaksi subscription berhasil dan membentuk snapshot transaksi yang benar. |
| `internal/usecase` | `TestSubscriptionUsecaseCreateSubscriptionTransactionIsIdempotentForSamePendingTier` | Memastikan request transaksi subscription yang sama mengembalikan transaksi pending yang sudah ada. |
| `internal/usecase` | `TestSubscriptionUsecaseRejectsRenewWithoutActiveSubscription` | Memastikan renew subscription ditolak jika user belum punya subscription aktif. |
| `internal/usecase` | `TestSubscriptionUsecaseRejectsUpgradeWhenTierNotHigher` | Memastikan upgrade ditolak jika tier target tidak lebih tinggi dari tier aktif. |
| `internal/usecase` | `TestTierUsecaseListTiers` | Memastikan usecase tier mengembalikan daftar tier sesuai repository. |
| `internal/usecase` | `TestTransactionUsecaseListTransactionsByUserID` | Memastikan usecase transaksi mengembalikan daftar transaksi milik user. |
| `internal/usecase` | `TestTransactionUsecaseGetTransactionByID` | Memastikan usecase transaksi mengembalikan detail transaksi berdasarkan `transaction_id`. |
| `internal/usecase` | `TestEveryUsecaseFileHasMatchingUnitTest` | Guardrail agar setiap file `*_usecase.go` wajib punya `*_usecase_test.go`. |
| `internal/usecase` | `TestVideoUsecaseListAccessibleVideos` | Memastikan usecase video hanya mengembalikan video yang accessible untuk tier user. |

