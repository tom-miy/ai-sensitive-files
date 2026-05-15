# AI Ignore Files

`.aiignore` は共通の意図表明です。
すべての AI エージェントに対する強制制御ではありません。

生成されるファイル:

- `.aiignore`
- `.cursorignore`
- `.copilotignore`
- `generated/claude-code-deny-read.json`
- `generated/ai-agent-guidance.md`

エージェントやエディタごとに対応状況は異なります。
ignore ファイルだけに依存せず、
コミット前チェック、エディタ設定、
ツール固有の deny list と併用してください。
生成物は手編集せず、ポリシーから再生成します。
