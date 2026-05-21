# AI 向け除外ファイル

このページは、AI ツールに読ませたくないパスを
どの生成物に出すのかを判断するためのものです。

`.aiignore` は共通の意図表明です。
すべての AI エージェントに対する強制制御ではありません。
そのため、ここで生成されるファイルだけで安全になるとは
考えないでください。

生成されるファイル:

- [`.aiignore`](../.aiignore)
- [`.cursorignore`](../.cursorignore)
- [`.copilotignore`](../.copilotignore)
- [`generated/claude-code-deny-read.json`](../generated/claude-code-deny-read.json)
- [`generated/ai-agent-guidance.md`](../generated/ai-agent-guidance.md)

エージェントやエディタごとに対応状況は異なります。
除外ファイルだけに依存せず、
コミット前チェック、エディタ設定、
ツール固有の読み取り拒否リストと併用してください。
生成物は手編集せず、ポリシーから再生成します。

判断の目安:

- AI に読ませたくないパスには `action.aiignore: true` を付ける
- 実際にコミットも止めたいパスには `commit_block: true` も付ける
- ツール固有の強制設定が必要な場合は、生成物を設定例として使う
