# Architecture

`ai-sensitive-files` の唯一の定義元は、
適用先リポジトリの `.ai-sensitive-files/sensitive-files.yaml` です。

このツール側のリポジトリではサンプルを `templates/` に置きます。
適用先リポジトリでは、実際に使うポリシーを
`.ai-sensitive-files/` に置きます。
`configs/` のようなアプリケーション側の設定ディレクトリと
混ざって見えないようにするためです。

CLI の責務は小さく分けています。

- `internal/domain`: ポリシー型と検証
- `internal/infra`: ポリシーファイルの読み込み
- `internal/usecase`: 生成ファイル作成、検査、一覧表示
- `internal/interface/cli`: コマンド解析と出力

任意のシークレットスキャン、鍵管理、
AI エージェント実行時の強制制御は扱いません。
