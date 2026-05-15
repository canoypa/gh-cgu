# gh-cgu Agent Instructions

git ユーザープロファイル（name/email）を短いキーで保存し、`gh cgu use <key>` でリポジトリ単位に切り替える gh CLI 拡張。プロファイルは Gist にバックアップされ、マルチマシン間で自動同期する。

**Guiding principle**: コードベースはシンプルに保つ。新しいコマンドは `cmd_<name>.go` に直接実装し、不要な抽象化・共有パッケージ・レイヤー追加は避ける。

## Build & Test

```sh
go build ./...
go test ./...
```

Requires Go 1.25+. Use `mise install` to set up the correct Go version.

## Design Philosophy

- **UX は gh CLI に倣う**: 成功時は `✓ ...` を stdout、エラーは `! ...` を stderr へ
- **シンプルを保つ**: コマンドは `cmd_*.go` に直接実装する。薄いラッパー関数で十分
- 致命的なエラーは `cobra.CheckErr(...)` で即終了。ユーザー向けエラーはメッセージ付き `fmt.Errorf` でラップする

## Critical Gotchas

**Viper はキーを小文字化する。** プロファイルキーの追加・削除・編集は `viper.AllSettings()` を経由し、`strings.ToLower(key)` でアクセスする。大文字小文字の混在でキーが見つからないバグが起きやすい。`cmd_remove.go` 参照。

**バックグラウンド同期は別プロセス。** `gh cgu add/edit/remove` 後に `_sync-gist` サブコマンドを独立プロセスとして spawn して Gist へ push する。Unix では `syscall.Flock` で排他制御。Windows ではバックグラウンド同期を無効化している。

## Testing Conventions

テストでは必ず `newTestViper(t)` を使う（`cmd_add_profile_test.go`）。メタファイルを直接操作するテストは `withTempMetaFile(t)` を使う（`meta_test.go`）。
