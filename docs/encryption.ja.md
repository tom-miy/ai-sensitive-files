# Encryption

`encrypt: true` は、その項目を暗号化と
平文チェックの対象にするという意味です。

場所を表す項目は以下を使います。

- `encrypted_path`: 暗号化済みファイルの保存先
- `decrypted_path`: 復号後に平文ファイルが書き出される先
- `crypto.secret_ref`:
  外部シークレット管理サービスにある保護された値の参照

`crypto` ブロックで暗号化方式とコマンドを管理します。

```yaml
crypto:
  method: "sops-age"
  recipients: ["age1exampleteampublickey...", "age1examplecipublickey..."]
  encrypt_command: "sops --encrypt --output {encrypted_path} {decrypted_path}"
  decrypt_command: "sops --decrypt --output {decrypted_path} {encrypted_path}"
  manual_edit: "decrypted"
```

`recipients` には、そのファイルを復号できる
age 公開鍵をすべて書きます。
たとえばチーム用の公開鍵と CI 用の公開鍵を並べます。
これらは秘密鍵ではなく公開鍵なので、Git 管理してかまいません。
実際の recipient は SOPS の `.sops.yaml` にも保存されます。
ポリシー側にも同じ一覧を持たせることで、
レビューと生成される暗号化計画が分かりやすくなります。

`manual_edit` は同期検査の挙動を決めます。
`decrypted` はローカルの平文編集を許可します。
平文ファイルが暗号化済みファイルより新しい場合、
`check` は失敗します。
コミット前に暗号化し直す必要があることを示します。
`encrypted` と `none` では、平文ファイルが残っている場合に
`check` が失敗します。
暗号化済みファイルが平文ファイルより新しければ、
復号し直すよう `check` が警告を出します。

1Password / Bitwarden では `encrypted_path` を省略し、
`secret_ref` を使います。

```yaml
crypto:
  method: "1password"
  secret_ref: "op://Engineering/App CI/.env"
  decrypt_command: "op read {secret_ref} > {decrypted_path}"
  manual_edit: "none"
```

この方式では、パスワードマネージャーの項目が定義元です。
`ai-sensitive-files` はコマンドと
ローカル平文ファイルの扱いを記録・検査します。
パスワードマネージャーへのログインや
シークレット取得は行いません。

このリポジトリでは SOPS と age を mise で入れる想定です。

```bash
mise trust .
mise install
```

1Password / Bitwarden を使う場合は、
`op` や `bw` の導入と認証を別途行ってください。

このツールは SOPS 初期化、鍵作成、鍵ローテーション、
秘密 identity 管理を行いません。
`templates/sops/.sops.yaml.example` の recipient はダミーです。

Lefthook では、秘密鍵ファイルがコミットされることは止められます。
ただし、秘密鍵を安全に作ること、配布しないこと、
ローカル権限を保つことは別の運用で管理します。

運用方針:

- 秘密の age identity は Git 管理しない
- age identity や秘密鍵に一致するパスは
  `commit_block: true` にする
- `.sops.yaml` やポリシーレビューに必要な
  age 公開鍵 recipient は Git 管理してよい
- チームメンバーや CI が復号する必要がある場合は、
  共有前に必要な age 公開鍵をすべて recipient に追加する
- 暗号化済みファイルをコミットする場合は、
  復旧とローテーションの手順をチームで決める
- 例外がない限り `decrypted_path` の平文ファイルは Git 管理しない
- 組織の鍵ローテーションは別途定義する
