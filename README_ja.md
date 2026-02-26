# memAI-go

脳型記憶システムのGoライブラリ。扁桃体（感情検出）・短期記憶（作業記憶）・長期記憶（ベクトル検索+感情プライミング）を提供する。

## インストール

```bash
go get github.com/ieee0824/memAI-go
```

## パッケージ構成

```
package memai

├── emotion.go   # 感情検出（扁桃体）
├── stm.go       # 短期記憶（作業記憶）
├── ltm.go       # 長期記憶（ベクトル検索）
├── feedback.go  # フィードバック検出
├── store.go     # ストレージインターフェース
└── types.go     # 共通型定義
```

## 使い方

### 感情検出

キーワードベースの日本語感情検出。LLM不要。

```go
es := memai.AnalyzeEmotion("嬉しい！ありがとう！")
// es.Primary   = EmotionJoy
// es.Intensity = 0.7  (0.0-1.0)
// es.Valence   = 0.56 (-1.0〜1.0)
```

6分類: 喜び・悲しみ・怒り・不安・驚き・中立

### 短期記憶 (STM)

活性化減衰モデルの作業記憶。感情的アイテムは減衰が遅い。

```go
stm := memai.NewSTM(memai.DefaultSTMConfig())

// アイテム追加
stm.Add(&memai.WorkingMemoryItem{
    Topic:      "会議の予定",
    Keywords:   []string{"会議", "予定"},
    Activation: 1.0,
    TurnCreated: 0,
    TurnAccessed: 0,
})

// ターン更新: 減衰 → 感情マーク → キーワードリフレッシュ → 除去
emotion := memai.AnalyzeEmotion(userMessage)
stm.Update(currentTurn, userMessage, emotion)

// フォーマット出力
fmt.Println(stm.Format())
// - [高] 会議の予定
```

設定:
- `MaxItems`: 作業記憶容量 (デフォルト: 7)
- `NormalDecayRate`: 通常の減衰率 (デフォルト: 0.15/ターン)
- `EmotionalDecayRate`: 感情的アイテムの減衰率 (デフォルト: 0.07/ターン)
- `RefreshBoost`: キーワード一致時のブースト (デフォルト: +0.3)
- `ActivationThreshold`: 除去閾値 (デフォルト: 0.1)

### 長期記憶 (LTM)

ベクトル検索 + 多要素スコアリング + 感情プライミング。

```go
ltm := memai.NewLTM(store, embeddingFn, memai.DefaultLTMConfig())

results, err := ltm.Search(ctx, memai.SearchQuery{
    Query:              "明日の会議",
    ThreadKey:          "thread-1",
    EmotionalIntensity: 0.7,
})

// フィードバック適用
ltm.ApplyFeedback(ctx, memoryIDs, memai.FeedbackBoostPositive)
```

スコアリング要素:
- **コサイン類似度**: embedding同士の類似度
- **感情ブースト**: 感情的な記憶を優先 (+0.12 × intensity)
- **スレッドブースト**: 同一スレッドの記憶を優先 (+0.1)
- **日付ブースト/ペナルティ**: 日付一致 (+0.15) / 不一致 (-0.2)
- **感情プライミング**: ユーザーが感情的なとき閾値を下げる (0.3 → 0.25)

### フィードバック検出

ユーザーの反応から記憶の正確性フィードバックを検出。

```go
delta := memai.DetectFeedback("ありがとう、そうそう！")
// delta = +0.05 (positive)

delta = memai.DetectFeedback("違うよ、それじゃない")
// delta = -0.05 (negative)
```

### ストレージインターフェース

`MemoryStore` を実装すれば任意のバックエンドを使える。

```go
type MemoryStore interface {
    GetMemories(ctx context.Context) ([]Memory, error)
    SaveMemory(ctx context.Context, mem *Memory) error
    DeleteMemory(ctx context.Context, id int64) error
    UpdateBoost(ctx context.Context, id int64, delta float64) error
}
```

`EmbeddingFunc` でembeddingプロバイダも差し替え可能。

```go
type EmbeddingFunc func(ctx context.Context, text string) ([]float64, error)
```

## ライセンス

MIT
