# Tech Lead Review — open-swag-go (master @ a389b91)

Tanggal: 2026-10-18. Build, `go vet`, dan `go test ./...` lulus di Go 1.24.12.

## Ringkasan

open-swag-go adalah library Go yang mengubah definisi `Endpoint` (struct literal, bukan komentar anotasi) menjadi
spec OpenAPI 3.1 lewat reflection, lalu menyajikannya dengan UI Scalar. Core-nya kecil dan mudah dibaca
(`openapi.go`, `pkg/schema`, `pkg/spec`, `handler.go`). Arah produknya jelas dan DX-nya bagus.

Masalah utamanya ada di tiga tempat: (1) beberapa bug correctness pada spec yang dihasilkan, (2) sekitar
40% kode di `pkg/` adalah scaffolding yang tidak dipakai oleh runtime mana pun, dan (3) README/CI/go.mod
menjanjikan hal yang tidak dilakukan kode. Belum ada yang menghalangi pemakaian internal, tapi ada beberapa
hal yang harus beres sebelum library ini layak dipromosikan sebagai v1 stabil.

## P0 — Bug yang menghasilkan spec salah (ditemukan dengan smoke test)

1. **Path `:id` tidak dinormalisasi ke `{id}`.** `extractPathParams` mengenali `:id`, tapi path yang ditulis
   ke `paths` tetap `/users/:id`. OpenAPI mewajibkan `{id}`; Scalar dan codegen (openapi-typescript, orval)
   akan memperlakukannya sebagai literal. File: `openapi.go` `addEndpointToSpec`.
2. **Embedded struct dengan tipe unexported hilang total.** `type base struct{ID string}` yang di-embed sebagai
   `base` gagal di `field.IsExported()` sebelum sampai ke cabang `field.Anonymous`, jadi `id` dan `created_at`
   tidak muncul. Ini pola yang sangat umum, dan berbeda dari perilaku `encoding/json` yang mem-promote field
   tersebut. File: `pkg/schema/converter.go` `fromStruct`.
3. **Duplikat method+path ditimpa diam-diam, method tak dikenal menghasilkan path item kosong.** Dua `Add`
   untuk `GET /users/{id}` menyisakan yang terakhir tanpa error. `HEAD`/`OPTIONS`/`TRACE` menghasilkan
   `"/health": {}`. `Add` perlu memvalidasi dan mengembalikan error, atau `BuildSpec` mengembalikan error.
4. **Pointer tidak menjadi nullable.** `*string` dipetakan ke `string` biasa. Untuk pitch utama README
   (generate TypeScript types), ini berarti `string | null` hilang di sisi klien. Catatan: spec menyatakan
   `openapi: "3.1.0"`, tapi struct schema memakai kata kunci `nullable` dari 3.0; di 3.1 yang benar adalah
   `type: ["string","null"]`. Pilih satu versi dan konsisten.
5. **Tidak ada `operationId`.** `Operation.OperationID` tidak pernah diisi. orval menolak spec tanpa
   `operationId`, padahal README merekomendasikannya. Default sederhana: `method + path` yang di-camelCase.

## P1 — Desain yang perlu diputuskan sebelum v1 stabil

6. **Tidak ada `$ref` / `components.schemas`.** Semua struct di-inline per endpoint. Akibatnya: spec besar,
   codegen tidak menghasilkan named type (`UserResponse` menjadi anonymous object), dan struct rekursif
   kehilangan schema-nya (jadi `{type: object, description: "Recursive reference to ..."}`). Ini satu-satunya
   cara yang benar untuk menangani rekursi, dan juga menyelesaikan masalah penamaan tipe di frontend.
7. **Satu `go.mod` untuk core + empat adapter framework.** Siapa pun yang `go get` core akan menarik gin, chi,
   echo, fiber, fasthttp, quic-go, dan sonic ke module graph-nya. SKILL.md bahkan mengklaim "core depends only
   on stdlib", yang tidak benar di level modul. Solusi standar: tiap adapter menjadi modul sendiri
   (`adapters/gin/go.mod`) atau pindahkan ke repo terpisah.
8. **Scalar dimuat dari CDN tanpa pin versi dan tanpa SRI** (`@scalar/api-reference` tanpa `@x.y.z`). Setiap
   rilis upstream langsung mengubah perilaku semua pengguna, dan ini supply-chain risk. Minimal pin versi;
   idealnya sediakan opsi self-host/embed.
9. **Inkonsistensi aturan tag antara body dan parameter.** `buildParamsFromStruct` di `openapi.go`
   hanya membaca `validate`/`binding` untuk required, jadi `swagger:"required"` pada query param diabaikan
   (terbukti di smoke test). Di `pkg/schema` ada `IsRequired` dan `ParseFieldTags` yang sudah benar.
   `ConvertExampleToType` juga diduplikasi di `pkg/schema/tags.go` dan `pkg/examples/generator.go`.
   Satu jalur parsing tag untuk semuanya.
10. **Validator tags tidak dipetakan.** `validate:"oneof=a b"` tidak menjadi `enum`, `min=18,max=99` tidak
    menjadi `minimum`/`maximum`. Field `Enum`, `Minimum`, `Pattern`, dll. sudah ada di struct `Schema` tapi
    tidak pernah diisi oleh converter. Ini quick win bernilai tinggi untuk pengguna Gin/validator.

## P2 — Kode mati dan klaim yang tidak sesuai

11. **Paket yang tidak terhubung ke runtime mana pun:** `pkg/auth` (CredentialStore, Playground),
    `pkg/tryit` (Console, Environment, History), `pkg/tryit/snippets`, `pkg/ui/themes.go`,
    `pkg/examples/templates.go` dan `faker.go`, `pkg/schema/validator.go`, `pkg/schema/types.go`.
    Tidak ada yang dipanggil dari `handler.go` atau template. Fitur "Auth playground", "Try-it console",
    "Code snippets" di README sebenarnya adalah fitur bawaan Scalar, bukan hasil dari paket ini. Field seperti
    `PersistConfig.Encrypt` dan `examples.Config.UseFaker` tidak berpengaruh apa pun. Rekomendasi: hapus, atau
    pindahkan ke branch `experimental` sampai benar-benar dipakai. Kode mati di public API adalah janji yang
    harus dijaga kompatibilitasnya.
12. **README tidak akurat di beberapa tempat.** Contoh co-located memakai `Handler: CreateUser`, tapi
    `Endpoint` tidak punya field `Handler` (compile error untuk pembaca). Import `pkg/snippets` dan
    `snippets.New()` tidak ada; yang ada `pkg/tryit/snippets` dan `NewManager()`, dan tidak ada generator PHP.
13. **`pkg/versioning` setengah jadi.** `breaking.go` mendefinisikan katalog rule (`BreakingChangeRule.Check`,
    `IsBreaking`) yang tidak pernah dipakai; klasifikasi di `getMigrationGuide` dan `createMigrationStep`
    bergantung pada prefix string `Description`. Contoh salah klasifikasi: "Parameter 'x' removed" cocok
    ke cabang `"removed"` dulu di `migration.go`, sehingga berjudul "Handle removed endpoint". Diff hanya
    membandingkan properti top-level request body; perubahan response schema (menghapus field yang dipakai
    klien) tidak terdeteksi. Urutan `Changes` nondeterministik karena iterasi map, menyulitkan diff di CI.
    Beri `Change` field bertipe `BreakingChangeType`, lalu switch di situ.

## P3 — Keamanan dan operasional

14. **DocsAuth via `?key=`** bocor ke access log, browser history, dan header Referer, lalu template
    meneruskan `window.location.search` ke URL spec. Juga: `DocsAuth.Enabled=true` tanpa `APIKey` dan tanpa
    username/password mengunci semua request dengan 401 tanpa peringatan. Pertimbangkan cookie setelah
    satu kali validasi, atau setidaknya dokumentasikan risikonya.
15. **Template HTML tidak di-escape.** `{{TITLE}}`, `{{CUSTOM_CSS}}`, `{{SPEC_URL}}` disisipkan via
    `strings.ReplaceAll`. Sumbernya config developer, jadi risikonya rendah, tapi gunakan `html/template`
    agar tidak jadi masalah saat `Title` suatu hari datang dari konfigurasi eksternal.
16. **`Access-Control-Allow-Origin: *`** pada spec handler tidak bisa dimatikan. Jadikan opsional.

## P4 — CI dan hygiene

17. **CI tidak pernah jalan.** Workflow trigger pada branch `main`, sedangkan repo memakai `master`.
    Go di CI adalah 1.22 sementara `go.mod` meminta 1.24.0. golangci-lint action v4 dengan `version: latest`
    tanpa file konfigurasi akan rapuh.
18. **Coverage test tipis.** Hanya 4 file test, semuanya regression untuk bug O1–O6. Tidak ada test untuk
    `Mount`/`Handler`/`DocsAuth`, adapter, `buildParamsFromStruct`, `extractPathParams`, embedded struct,
    map, pointer, `time.Time`, `omitempty`. Satu golden-file test yang membandingkan output `SpecJSON()`
    dari `examples/full-featured` akan menangkap sebagian besar regresi di atas sekaligus.
19. `gofmt -l` menandai `examples/with-auth/main.go` dan `pkg/spec/server.go`. `.gitignore` berisi baris
    nyasar `dan`. `w.Write` di handler mengabaikan error.

## Yang sudah baik

- Cycle guard dengan `visiting` map, penanganan `omitempty` vs required, `time.Time`/`Duration`/`[]byte`/UUID.
- `AuthConfig.Schemes` sebagai sumber kebenaran dan penolakan fallback diam-diam ke bearer (O3) adalah
  keputusan yang benar dan dikomentari dengan baik.
- `subtle.ConstantTimeCompare` untuk DocsAuth.
- Semua adapter konsisten melakukan redirect `/docs` ke `/docs/` agar relative URL `./openapi.json` benar.
- SKILL.md untuk integrasi AI adalah ide bagus dan ditulis rapi.

## Urutan yang saya sarankan

1. Perbaiki P0 (1–5): kecil, terisolasi, bisa satu PR dengan golden test.
2. Putuskan P1 nomor 6 dan 7 sebelum tag berikutnya karena keduanya mengubah bentuk output dan struktur modul.
3. Bersihkan P2 (hapus kode mati, perbaiki README) agar permukaan API yang dijaga mengecil.
4. Betulkan CI (P4 nomor 17) supaya langkah 1–3 terjaga.

## Status (update 2026-10-18, dikerjakan oleh dua agent Opus, direview Fable)

| Item | Status | Catatan |
|------|--------|---------|
| P0 1 path `:id` | Selesai | `normalizePath`, juga `*name` dan regex chi `{id:[0-9]+}` |
| P0 2 embedded unexported | Selesai | Mengikuti aturan promosi `encoding/json`, termasuk shadowing |
| P0 3 duplikat / method asing | Selesai | `Docs.Validate() error`; HEAD/OPTIONS/TRACE didukung |
| P0 4 nullable 3.1 | Selesai | `spec.Schema.MarshalJSON` menulis `type: [T,"null"]`; differ versioning disesuaikan |
| P0 5 operationId | Selesai | Default `getUsersById`, override via `Endpoint.OperationID` |
| P1 6 `$ref`/components | Terbuka | Butuh keputusan: mengubah bentuk output |
| P1 7 go.mod adapter terpisah | Terbuka | Butuh keputusan: mengubah struktur modul; SKILL.md sudah jujur soal ini |
| P1 8 pin Scalar CDN | Selesai | `@scalar/api-reference@1.75.0`, override via `UIConfig.ScalarVersion` |
| P1 9 unifikasi tag param | Selesai | `swagger:"required"` kini berlaku di query param |
| P1 10 validator tags | Selesai | `oneof`, `min/max/len/gte/lte/gt/lt`, format, `dive`, tag `enum` |
| P2 11 hapus kode mati | Terbuka | Keputusan API publik |
| P2 12 README | Selesai | `Handler` dihapus, snippets dibetulkan, fitur Scalar dijelaskan, API baru didokumentasikan |
| P2 13 refactor versioning | Terbuka | Hanya `schemaTypeSignature` yang disesuaikan untuk type array 3.1 |
| P3 14 DocsAuth `?key=` | Terbuka | Misconfig tanpa kredensial kini 500, bukan 401 |
| P3 15 escape template | Selesai | `html/template`; CustomCSS tetap trusted |
| P3 16 CORS opsional | Selesai | `UIConfig.SpecCORSOrigin` |
| P4 17 CI | Selesai | Trigger `master`, Go dari go.mod, gofmt + vet + golangci-lint v1.64 |
| P4 18 coverage | Sebagian | 25 test baru (handler, validate, path, nullable, embedded, validator tags) |
| P4 19 hygiene | Selesai | gofmt bersih, `.gitignore` dibetulkan |
