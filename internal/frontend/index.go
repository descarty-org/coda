package frontend

import (
	"coda/internal/llm"
	"coda/internal/llm/openai"
	"coda/internal/logger"
	"coda/internal/review"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Default values for code review parameters
const (
	defaultLanguage    = "python"
	defaultCode        = "print('Hello, World!')"
	defaultDetailLevel = "medium"
	defaultStrictness  = "medium"
)

// IndexHandler manages the index page and code review functionality.
// It handles rendering the main page, processing code review requests,
// and displaying results.
type IndexHandler struct {
	templates *TemplateManager
	completer llm.Completer
}

// newIndex creates a new IndexHandler with the given template manager and completer.
func newIndex(tpl *TemplateManager, completer llm.Completer) *IndexHandler {
	return &IndexHandler{
		templates: tpl,
		completer: completer,
	}
}

// RegisterRoutes registers the HTTP routes for the index handler.
func (h *IndexHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.getIndex)
	r.Get("/result", h.getResult)
	r.Post("/review", h.postReview)
}

// getIndex renders the index page.
func (h *IndexHandler) getIndex(w http.ResponseWriter, r *http.Request) {
	availableModels := h.completer.GetAvailableModels()
	var modelNames []string
	for _, model := range availableModels {
		modelNames = append(modelNames, model.DisplayName)
	}

	h.templates.Render(w, r, "index", struct {
		Models []string
	}{
		Models: modelNames,
	})
}

// getResult renders the result component with sample code and instructions.
func (h *IndexHandler) getResult(w http.ResponseWriter, r *http.Request) {
	const sampleInstructions = `# コードレビューAI

## 使い方
1. 左側のエディタにコードを入力してください
2. 言語を選択してください
3. "Review Code" ボタンをクリックしてください

AIがコードを解析して、以下の観点からレビューを行います:
- バグやエラーの可能性
- セキュリティ上の問題点
- パフォーマンスの最適化
- コーディング規約への準拠
- 可読性と保守性の向上

### サンプルコード
`

	const sampleCode = "```python\ndef calculate_sum(numbers):\n    total = 0\n    for num in numbers:\n        total += num\n    return total\n```"

	h.templates.RenderComponent(w, r, "components/results", struct {
		Result   string
		ReviewID string
	}{
		Result: sampleInstructions + sampleCode,
	})
}

// postReview handles the code review form submission.
func (h *IndexHandler) postReview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, r, http.StatusBadRequest, "フォームデータの解析に失敗しました。")
		return
	}

	// Extract form values with defaults
	code := getFormValueWithDefault(r, "code", "")
	language := getFormValueWithDefault(r, "language", defaultLanguage)
	detailLevel := getFormValueWithDefault(r, "detailLevel", defaultDetailLevel)
	strictness := getFormValueWithDefault(r, "strictness", defaultStrictness)
	modelName := getFormValueWithDefault(r, "model", "")

	if code == "" {
		h.handleError(w, r, http.StatusBadRequest, "コードが入力されていません。")
		return
	}

	if len(code) > 50_000 { // Limit input length to 50_000 characters for now
		h.handleError(w, r, http.StatusBadRequest, "入力が長すぎます。短縮して再試行してください。")
		return
	}

	// Get the selected model or use default
	var selectedModel llm.Model
	if modelName != "" {
		// Find the model by name
		availableModels := h.completer.GetAvailableModels()
		for _, model := range availableModels {
			if model.DisplayName == modelName {
				selectedModel = model
				break
			}
		}
	}

	// If no model was selected or found, use the default
	if selectedModel.Name == "" {
		selectedModel = openai.ModelGPT4o
		logger.Info(r.Context(), "using default model", "model", selectedModel.Name)
	}

	// Build the custom prompt for the AI
	customPrompt := buildCustomPrompt(language, detailLevel, strictness)

	// Call the AI service
	ret, err := h.completer.Complete(r.Context(), llm.CompleteParams{
		Messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: customPrompt,
			},
			{
				Role:    llm.RoleUser,
				Content: code,
			},
		},
	}, selectedModel)
	if err != nil {
		h.handleError(w, r, http.StatusInternalServerError, err)
		return
	}

	// Create a review object for persistence
	reviewObj := review.NewReview(code, language, detailLevel, strictness, ret.Messages[0].Content)

	// Render the results
	h.templates.RenderComponent(w, r, "components/results", struct {
		Result   string
		ReviewID string
	}{
		Result:   ret.Messages[0].Content,
		ReviewID: reviewObj.ID,
	})
}

// getFormValueWithDefault retrieves a form value or returns the default if empty.
func getFormValueWithDefault(r *http.Request, key, defaultValue string) string {
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// buildCustomPrompt constructs the AI prompt based on the review parameters.
func buildCustomPrompt(language, detailLevel, strictness string) string {
	// Base prompt with language specification
	customPrompt := systemPrompt + "\nprogramming language: " + language + "\n" +
		"Format your response in Markdown. Use headings, lists, code blocks, etc. to make your review clear and readable.\n\n"

	// Add detail level instructions
	customPrompt += getDetailLevelInstructions(detailLevel)

	// Add strictness instructions
	customPrompt += getStrictnessInstructions(strictness)

	return customPrompt
}

// getDetailLevelInstructions returns the prompt text for the specified detail level.
func getDetailLevelInstructions(detailLevel string) string {
	switch detailLevel {
	case "low":
		return "Detail level: Low - Provide a concise overview with only the most important points. Focus on major issues and skip minor details.\n"
	case "high":
		return "Detail level: High - Provide an in-depth analysis with detailed explanations and specific improvement suggestions for each issue found.\n"
	default: // medium
		return "Detail level: Medium - Provide a balanced review with reasonable detail on important issues.\n"
	}
}

// getStrictnessInstructions returns the prompt text for the specified strictness level.
func getStrictnessInstructions(strictness string) string {
	switch strictness {
	case "low":
		return "Strictness: Low - Focus only on critical issues like bugs, security problems, and major performance concerns. Ignore minor style issues.\n"
	case "high":
		return "Strictness: High - Apply strict best practices and standards. Point out all issues including minor style concerns, potential edge cases, and optimization opportunities.\n"
	default: // medium
		return "Strictness: Medium - Apply reasonable standards focusing on important issues while mentioning some style and optimization concerns.\n"
	}
}

// handleError renders an error message to the user.
func (h *IndexHandler) handleError(w http.ResponseWriter, r *http.Request, code int, err any) {
	message := determineErrorMessage(code, err)

	if code == http.StatusInternalServerError {
		// Log internal server errors
		logger.Error(r.Context(), "internal server error", "err", err)
	}

	h.templates.RenderComponent(w, r, "components/error", struct {
		Message string
	}{
		Message: message,
	})
}

// determineErrorMessage converts an error to an appropriate user-facing message.
func determineErrorMessage(code int, err any) string {
	// Handle string errors
	if errStr, ok := err.(string); ok {
		return errStr
	}

	// Handle Go errors with specific error types
	if goErr, ok := err.(error); ok {
		switch {
		case errors.Is(goErr, llm.ErrContextLengthExceeded):
			return "入力が長すぎます。短縮して再試行してください。"
		case errors.Is(goErr, llm.ErrServiceUnavailable):
			return "AIサービスが利用できません。しばらくしてから再試行してください。"
		case errors.Is(goErr, llm.ErrTooManyRequests):
			return "リクエストが多すぎます。しばらくしてから再試行してください。"
		}
	}

	// Default messages based on HTTP status code
	if code == http.StatusInternalServerError {
		return "エラーが発生しました。"
	}

	return "無効なリクエストです。"
}

// systemPrompt is the base prompt for the AI code review system.
var systemPrompt = `あなたはプログラミングとソフトウェア開発に特化したAIアシスタントです。

提供された入力に対して、以下のように応答してください：

1. コードが提示された場合は、以下の観点からコードレビューを行います：
   - バグやエラーの可能性
   - セキュリティ上の問題点
   - パフォーマンスの最適化
   - コーディング規約への準拠
   - 可読性と保守性の向上
   - ベストプラクティスの適用

2. プログラミングに関する質問の場合は、以下の点を意識して回答します：
   - 正確で最新の情報
   - わかりやすい説明と具体例
   - 複雑な概念の段階的な解説
   - 適切なコードサンプル（必要に応じて）

常に簡潔かつ具体的に回答し、余計な会話履歴を表示せず、常に新しい質問に対して直接的に応答してください。
回答はMarkdown形式で提供し、必要に応じてコードブロックやリストを使用して情報を整理してください。

[Important Note]
1. コードレビュー結果は必ず日本語で返してください
2. コードが提供されなかった場合は、コードレビューを行わずにエラーメッセージを返してください
3. コードレビュアーとしての役割に徹し、無関係な情報を提供しないようにしてください

[Settings]
`
