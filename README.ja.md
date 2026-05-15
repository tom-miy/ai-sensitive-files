# ai-sensitive-files

`ai-sensitive-files` は、`.env` や `.aws/credentials` などの
機密ファイルの扱いを 1 つの YAML にまとめるツールです。

うれしい点は、AI に見せない場所、Git に載せない場所、
暗号化する場所、コミット時に止める場所を
別々に管理しなくてよくなることです。

たとえば `.env` を追加したときに、
`.aiignore`、`.gitignore`、SOPS/age の対象、
Lefthook のチェックをそれぞれ手で直す必要がありません。
`.ai-sensitive-files/sensitive-files.yaml` を直して生成し直せば、
同じ判断基準から必要なファイルを作れます。

その結果、次のような状態を減らせます。

- AI ignore には入っているが `.gitignore` にはない
- 暗号化対象なのに平文ファイルだけ更新されている
- パスワードマネージャーから出した平文ファイルが残っている
- 生成済みファイルがポリシーより古い

このリポジトリは、その小さな運用問題を扱う
ポートフォリオ用のセキュリティ向け開発ツールです。

このプロジェクトで示していること:

- ignore ファイルを手書きでばらばらに保守せず、
  ポリシーから生成する設計
- 利用者が管理する設定を勝手に編集せず、
  次に実行するコマンドを表示する保守的な導入手順
- SOPS/age と 1Password/Bitwarden への接続点。
  ただし鍵管理は持たない
- 平文ファイルが Git 管理されていないかを検査する流れ
- 生成物が古くないかをコミット前に検査する流れ
- プロンプトの匿名化、危険コマンドのブロック、
  secret scanner とは責務を分ける設計

English: [README.md](README.md)

## 生成されるもの

`.ai-sensitive-files/sensitive-files.yaml` から以下を生成します。

- [`.aiignore`](.aiignore), [`.cursorignore`](.cursorignore),
  [`.copilotignore`](.copilotignore):
  AI/editor 向け ignore の意図表明
- [`.gitignore.ai-sensitive-files`](.gitignore.ai-sensitive-files):
  `.gitignore` に追記する候補
- [`generated/claude-code-deny-read.json`](generated/claude-code-deny-read.json):
  Claude Code の `denyRead` スニペット
- [`generated/ai-agent-guidance.md`](generated/ai-agent-guidance.md):
  Codex / Cursor / Copilot 向けの案内スニペット
- [`generated/ai-sensitive-files.summary.md`](generated/ai-sensitive-files.summary.md):
  ポリシーの要約
- [`generated/encryption-targets.txt`](generated/encryption-targets.txt):
  SOPS/age の暗号化対象
- [`generated/decryption-targets.txt`](generated/decryption-targets.txt):
  暗号化済みの場所と復号後の場所の対応表
- [`generated/secret-sources.txt`](generated/secret-sources.txt):
  1Password / Bitwarden から平文ファイルを作るための参照
- [`generated/crypto-plan.md`](generated/crypto-plan.md):
  暗号化/復号コマンドと手動編集ポリシー

既存の `.gitignore` は直接変更しません。
生成内容を確認してから、手動で取り込む運用にしています。
`check` は `.gitignore.ai-sensitive-files` の項目が、
実際の `.gitignore` に反映されているかも確認します。

## コマンド

開発・連携用ツールは mise で揃えます。

```bash
mise trust .
mise install
mise run install-cli
```

`mise trust .` で、このリポジトリのツール定義を信頼します。
`mise install` で Go ツールチェーンを入れます。
SOPS、age、Lefthook もここで入ります。
用途は検証とサンプル実行です。
`mise run install-cli` はこのリポジトリの CLI を
`.bin/ai-sensitive-files` としてビルドします。
mise はこのリポジトリ内で `.bin` を PATH に追加します。
これらのコマンドは暗号鍵の作成や、
適用先フック設定の信頼/有効化は行いません。

ポリシーの構文と必須項目を確認します。

```bash
ai-sensitive-files validate --config .ai-sensitive-files/sensitive-files.yaml
```

ポリシーから ignore/config ファイルを生成します。

```bash
ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out .
```

`commit_block: true` の対象が
Git 管理されていないかを確認します。
`encrypt: true` の `decrypted_path` が
ポリシーに従っているかも確認します。
`.gitignore` に生成項目が反映されているか、
生成物が古くないかも検査します。

```bash
ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml
```

レビューしやすい形式でポリシーを表示します。

```bash
ai-sensitive-files list --config .ai-sensitive-files/sensitive-files.yaml
```

各コマンドは `--json` に対応しています。

## 導入

### 1. 適用先リポジトリへサンプルポリシーを入れる

```bash
bash install.sh --target /path/to/app
```

このコマンドは
`/path/to/app/.ai-sensitive-files/` を作ります。
`templates/sensitive-files.example.yaml` を
`/path/to/app/.ai-sensitive-files/sensitive-files.yaml`
としてコピーします。
`configs/` は作りません。

### 2. 適用先リポジトリでポリシーを確認する

```bash
cd /path/to/app
ai-sensitive-files validate --config .ai-sensitive-files/sensitive-files.yaml
```

### 3. ポリシーから ignore ファイルと生成物を作る

```bash
ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out .
```

### 4. 必要な追記を手動で行う

`install.sh` は既存ポリシーを上書きしません。
`.gitignore` の自動追記、SOPS/age 初期化、
Lefthook 設定の自動編集もしません。
生成後に表示される案内を見て、
`.gitignore` や `lefthook.yml` へ、
必要な内容だけを追加してください。

## Policy

この例では、1 つの機密情報を
2 つの場所で管理します。

- `.env`:
  アプリと開発者がローカルで使う平文ファイル。
  AI には見せず、コミットも止める
- `.env.sops.yaml`:
  同じ内容を暗号化したファイル。
  Git に載せる対象

`action` ブロックは、
この場所をどこに使うかを指定します。
対象は AI ignore 出力、gitignore 出力、
暗号化/復号チェック、コミットブロックです。

```yaml
sensitive_files:
  - path: ".env"                      # AI に見せずコミットも止めたい平文ファイルの場所
    encrypted_path: ".env.sops.yaml"  # SOPS/age で管理する暗号化済みファイル
    decrypted_path: ".env"            # 復号後にローカルへ出る平文ファイル
    reason: "local environment secrets"
    tags: ["env", "secret"]
    crypto:
      method: "sops-age"
      recipients: ["age1exampleteampublickey...", "age1examplecipublickey..."]
      encrypt_command: "sops --encrypt --output {encrypted_path} {decrypted_path}"
      decrypt_command: "sops --decrypt --output {decrypted_path} {encrypted_path}"
      manual_edit: "decrypted"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

YAML を唯一の定義元にすることで、
AI ignore、gitignore 候補、暗号化対象、
復号後の平文ファイル、暗号化/復号コマンド、
手動編集ルールも同じ判断基準を参照できます。
コミット前チェックも同じ定義を使います。

Git に暗号化済みファイルを置かない場合は、
`encrypted_path` ではなく `crypto.secret_ref` を使います。
これは 1Password / Bitwarden を
定義元にする場合の形です。

```yaml
sensitive_files:
  - path: ".env.ci"
    decrypted_path: ".env.ci"
    reason: "CI-only environment secrets fetched from 1Password"
    crypto:
      method: "1password"
      secret_ref: "op://Engineering/App CI/.env"
      decrypt_command: "op read {secret_ref} > {decrypted_path}"
      manual_edit: "none"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

この形では、定義元は
リポジトリ内の暗号化済みファイルではありません。
パスワードマネージャーの項目です。
生成される計画には、
ローカルの平文ファイルを作るコマンドが残ります。
`check` は、その平文ファイルが
コミットされないように止めます。

## Lefthook

既存の `lefthook.yml` には、
必要に応じて以下を手動で追記します。

```yaml
pre-commit:
  commands:
    ai-sensitive-files:
      run: ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml
```

このコマンドは、日常的に手で毎回実行するためのものでは
ありません。
コミット前のゲートとして使うものです。

## Secret Storage

`encrypt: true` は、
その項目を暗号化の対象にする指定です。
外部シークレット管理の対象にする場合にも使います。

`crypto.method` には、使う方式を書きます。
例: `sops-age`, `1password`, `bitwarden`

`crypto.encrypt_command` と `crypto.decrypt_command` には、
暗号化と復号の実行方法を書きます。
これにより、ローカルの平文ファイルを
どう作るかをレビューできます。
コミット前にどう暗号化し直すかも確認できます。

1Password / Bitwarden のように、
リポジトリ内へ暗号化済みファイルを置かない場合は
`crypto.secret_ref` を使います。
`secret_ref` は、保護された値が
外部サービスのどこにあるかを示す参照です。

想定するポリシーの形:

- SOPS/age:
  `encrypted_path` がリポジトリ内の暗号化済みファイル、
  `decrypted_path` がローカルの平文出力
- 1Password / Bitwarden:
  `secret_ref` が外部シークレットの参照、
  `decrypt_command` が `decrypted_path` をローカルに作るコマンド

より強く分離したい場合は devcontainer を使います。
復号後の env ファイルはワークスペース外に置きます。

```yaml
sensitive_files:
  - path: ".env"                               # リポジトリ内に出てはいけない場所
    encrypted_path: ".env.sops.yaml"           # リポジトリ内に置く暗号化済みファイル
    decrypted_path: "/workspaces/.runtime-secrets/app.env" # devcontainer が読む env ファイル
    crypto:
      method: "sops-age"
      decrypt_command: "umask 077; mkdir -p /workspaces/.runtime-secrets; sops --decrypt {encrypted_path} > {decrypted_path}"
      encrypt_command: "umask 077; sops --encrypt --output {encrypted_path} {decrypted_path}"
      manual_edit: "decrypted"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

この構成では、devcontainer が env ファイルを読みます。
指定は `runArgs: ["--env-file", "/workspaces/.runtime-secrets/app.env"]` です。

プロジェクトのワークスペースには、
暗号化済みファイルと生成されたポリシー関連ファイルだけを
置きます。
`ai-sensitive-files` はワークスペース外の平文ファイルも検査します。
ただし、その外部パスを `.gitignore` や
AI ignore ファイルには混ぜません。

`crypto.manual_edit` は、
ローカルで何を手編集してよいかを表します。

- `decrypted`:
  平文ファイルのローカル編集を許可する。
  平文ファイルが暗号化済みファイルより新しい場合、
  `check` は失敗する。
  コミット前に暗号化し直す必要があることを示す
- `encrypted`:
  暗号化済みファイルを編集対象にする。
  平文ファイルが残っている場合、`check` は失敗する
- `none`:
  どちらも手動編集しない。
  平文ファイルが残っている、または更新されている場合、
  `check` は失敗する

暗号化済みファイルが復号後ファイルより新しい場合、
`check` は復号し直すよう警告します。
これは pull 後の更新に気づくための警告です。

パスワードマネージャーを使う場合、
生成計画には `op` や `bw` のコマンドを残します。
このツールはログイン、取得、認証情報の保存を行いません。

SOPS/age を使う場合は `mise install` で SOPS と age を導入します。
1Password / Bitwarden を使う場合は、
`op` や `bw` の導入と認証を別途行ってください。

このリポジトリは鍵管理を実装しません。
age の公開鍵、つまり recipient は公開情報なので
Git 管理してかまいません。
一方で、age identity やその他の秘密鍵は Git 管理しないでください。

Lefthook では、秘密鍵ファイルのコミットを
止めることはできます。
ただし、秘密鍵を作らないこと、配布しないこと、
権限を正しく保つことは SOPS/age や運用側の責務です。
`.sops.yaml` の例では、
チーム用と CI 用のダミー recipient を使っています。

## Demo

```bash
bash scripts/demo.sh
```

mise 経由でも実行できます。

```bash
mise run demo
```

デモはサンプルポリシーの検証、生成、
`.agent-privacy-guard/entities.local.yaml` の平文検知、
Claude Code `denyRead` スニペット表示までを一度に確認します。

## 責務分離

| Repository | Responsibility |
|---|---|
| agent-privacy-guard | 外部 LLM / MCP に送る prompt の検査、匿名化、policy enforcement |
| secure-dev-hooks | AI agent のローカル操作、危険コマンド、ファイルアクセス制御 |
| ai-sensitive-files | AI に見せないファイル、暗号化対象、commit block 対象を YAML で一元管理 |

## Project Boundary

このプロジェクトは、すでに機密だと分かっている場所を
リポジトリ内で管理するためのものです。

答えるのは、たとえば次のような実務上の問いです。

- `.env` を AI ツールから隠すべきか
- `.env` を git ignore すべきか
- 暗号化済みファイルはどこか
- パスワードマネージャー側の定義元はどこか
- ローカル開発時に平文ファイルはどこへ出るか
- その平文ファイルが Git 管理されたら止めるべきか
- 暗号化済みファイルと同期ずれしたら止めるべきか

未知の漏えい認証情報を見つける仕事は別です。
その用途には gitleaks、trufflehog、
GitHub secret scanning などを使います。
SOPS、age、1Password、Bitwarden、
リポジトリ権限の制御も、
それぞれのシステムが担当します。
`ai-sensitive-files` はリポジトリ側のファイルルールを明示し、
検査できる形にします。
