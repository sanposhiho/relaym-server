# Relaym Server
RelaymのサーバーサイドAPIを管理するリポジトリです。

![test_and_lint](https://github.com/camphor-/relaym-server/workflows/test_and_lint/badge.svg)
[![codecov](https://codecov.io/gh/camphor-/relaym-server/branch/master/graph/badge.svg)](https://codecov.io/gh/camphor-/relaym-server)

API仕様は[こちら](docs/api.md)

## 開発に参加するには？

1. [CAMPHOR- Code of Conduct](https://github.com/camphor-/code-of-conduct)に同意してください。
1. [CAMPHOR- Lab](https://lab.camph.net/)に参加しましょう!


## 開発を始める前に

### STEP1 サービス内容を理解する

詳しくはこちらのリンクを参照してください。
[Relaym PRD (プロダクト要求仕様書)](docs/prd.md)

単語の定義も[docs/definition.md](docs/definition.md)からご覧ください。

### STEP2 アーキテクチャを理解する

APIリクエストはHTTPで受け付けています。曲の操作に[Spotify Web API](https://developer.spotify.com/documentation/web-api/)を使用しています。

詳しくは [docs/architecture.md](docs/architecture.md)をご覧ください。

### STEP3 データベースの設計を理解する

RelaymではMySQLをデータベースとして採用しています。

詳しくは [docs/database.md](docs/database.md)をご覧ください。


### STEP4 アプリケーションアーキテクチャを理解する

DDDやClean Architectureライクなアーキテクチャを採用しています。

詳しくは [docs/application_architecture.md](docs/application_architecture.md)をご覧ください。

## 開発

### ローカル開発環境のセットアップ

詳しくは [docs/development.md](docs/development.md)をご覧ください。

### 一般的な開発ルール

- GitHub Flowを用いたブランチ管理を行います。
- 設計はGitHub Issueを使って行います。
    - 新規機能の場合は必ずIssueを立てます。
- **PR, Issue, Commit, Commentは全て日本語を可とします。**

### レビュー

PRのマージにはレビュワーのApproveを必要とします。

レビューで確認すべき一般的な事項は[Google エンジニアリング・プラクティス ドキュメント](http://shuuji3.xyz/eng-practices/)に従います。

Goに関する作法は以下のドキュメントに従います。

- [Go Codereview Comments](https://knsh14.github.io/translations/go-codereview-comments/)
- [Effective Go](https://golang.org/doc/effective_go.html)
