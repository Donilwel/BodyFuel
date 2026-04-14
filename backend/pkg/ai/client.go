package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	c *openai.Client
}

func NewClient(apiKey string) *Client {
	return &Client{c: openai.NewClient(apiKey)}
}

// AnalyzeNutritionPhoto sends an image URL to GPT-4o Vision and returns parsed nutrition data.
func (cl *Client) AnalyzeNutritionPhoto(ctx context.Context, imageURL string) (*NutritionAnalysis, error) {
	prompt := `Проанализируй еду на изображении и верни ТОЛЬКО JSON-объект следующей структуры (без markdown, без лишнего текста):
{
  "description": "краткое описание блюда на русском",
  "calories": 0,
  "protein": 0.0,
  "carbs": 0.0,
  "fat": 0.0
}
Все числовые значения — оценки на порцию. Калории целым числом, макросы — float с 1 знаком. Описание писать на русском языке.`

	resp, err := cl.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{Type: openai.ChatMessagePartTypeText, Text: prompt},
					{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{URL: imageURL}},
				},
			},
		},
		MaxTokens: 200,
	})
	if err != nil {
		return nil, fmt.Errorf("openai vision: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai vision: empty response")
	}

	var result NutritionAnalysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("openai vision: parse response: %w", err)
	}

	return &result, nil
}

// AnalyzeNutritionPhotoData sends image bytes as base64 to GPT-4o Vision and returns parsed nutrition data.
// Use this instead of AnalyzeNutritionPhoto when the image is not publicly accessible (e.g. local MinIO).
func (cl *Client) AnalyzeNutritionPhotoData(ctx context.Context, data []byte, mediaType string) (*NutritionAnalysis, error) {
	prompt := `Проанализируй еду на изображении и верни ТОЛЬКО JSON-объект следующей структуры (без markdown, без лишнего текста):
{
  "description": "краткое описание блюда на русском",
  "calories": 0,
  "protein": 0.0,
  "carbs": 0.0,
  "fat": 0.0
}
Все числовые значения — оценки на порцию. Калории целым числом, макросы — float с 1 знаком. Описание писать на русском языке.`

	dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, base64.StdEncoding.EncodeToString(data))

	resp, err := cl.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{Type: openai.ChatMessagePartTypeText, Text: prompt},
					{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{URL: dataURL}},
				},
			},
		},
		MaxTokens: 200,
	})
	if err != nil {
		return nil, fmt.Errorf("openai vision: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai vision: empty response")
	}

	var result NutritionAnalysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("openai vision: parse response: %w", err)
	}

	return &result, nil
}

// GenerateRecommendations asks GPT to generate personalized fitness recommendations.
func (cl *Client) GenerateRecommendations(ctx context.Context, profile UserProfile) ([]RecommendationItem, error) {
	weightProgress := ""
	if profile.TargetWeight > 0 && profile.Weight > 0 {
		delta := profile.Weight - profile.TargetWeight
		if delta > 0.5 {
			weightProgress = fmt.Sprintf(", нужно сбросить %.1f кг до целевого веса %.1f кг", delta, profile.TargetWeight)
		} else if delta < -0.5 {
			weightProgress = fmt.Sprintf(", нужно набрать %.1f кг до целевого веса %.1f кг", -delta, profile.TargetWeight)
		} else {
			weightProgress = ", целевой вес достигнут"
		}
	}

	prompt := fmt.Sprintf(`Ты фитнес-тренер. Составь 3-5 персональных рекомендаций для пользователя.
Профиль: вес=%.1f кг, рост=%.0f см, возраст=%d лет, цель=%s, уровень активности=%s%s.
Включи хотя бы одну рекомендацию о питании/тренировках относительно прогресса к целевому весу.
Верни ТОЛЬКО JSON-массив (без markdown), все описания на русском языке:
[{"type":"workout|nutrition|rest|general","description":"конкретный совет","priority":1-3}]
Приоритет: 1=высокий, 2=средний, 3=низкий. Каждое описание не более 100 символов.`,
		profile.Weight, profile.Height, profile.Age, profile.Goal, profile.ActivityLevel, weightProgress)

	resp, err := cl.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		MaxTokens: 500,
	})
	if err != nil {
		return nil, fmt.Errorf("openai recommendations: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai recommendations: empty response")
	}

	var items []RecommendationItem
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &items); err != nil {
		return nil, fmt.Errorf("openai recommendations: parse response: %w", err)
	}

	return items, nil
}

// GenerateRecipes asks GPT to suggest recipes based on what the user has already eaten today.
func (cl *Client) GenerateRecipes(ctx context.Context, intake DailyIntake) ([]RecipeItem, error) {
	prompt := fmt.Sprintf(`Ты нутрициолог. Пользователь уже съел сегодня: %d ккал, %.1f г белка, %.1f г углеводов, %.1f г жиров.
Предложи 3-5 рецептов для следующего приёма пищи, которые дополнят уже съеденное (сбалансируй оставшиеся макросы).
Верни ТОЛЬКО JSON-массив (без markdown, без лишнего текста), все текстовые поля на русском языке:
[{
  "name": "Название рецепта",
  "description": "Краткое описание вкуса (1 предложение, не более 80 символов)",
  "ingredients": [{"name": "Название ингредиента", "grams": 150}],
  "macros": {"protein": 0.0, "fat": 0.0, "carbs": 0.0},
  "preparation_time": 10
}]
Правила: макросы в граммах (float), preparation_time в минутах (integer), название не более 50 символов, 3-6 ингредиентов с реальными граммовками.`,
		intake.ConsumedCalories, intake.ConsumedProtein, intake.ConsumedCarbs, intake.ConsumedFat)

	resp, err := cl.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		MaxTokens: 1400,
	})
	if err != nil {
		return nil, fmt.Errorf("openai recipes: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai recipes: empty response")
	}

	var items []RecipeItem
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &items); err != nil {
		return nil, fmt.Errorf("openai recipes: parse response: %w", err)
	}

	return items, nil
}

type NutritionAnalysis struct {
	Description string  `json:"description"`
	Calories    int     `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
}

type UserProfile struct {
	Weight        float64
	Height        float64
	Age           int
	Goal          string
	ActivityLevel string
	TargetWeight  float64
}

type RecommendationItem struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

type DailyIntake struct {
	ConsumedCalories int
	ConsumedProtein  float64
	ConsumedCarbs    float64
	ConsumedFat      float64
}

type Ingredient struct {
	Name  string  `json:"name"`
	Grams float64 `json:"grams"`
}

type RecipeItem struct {
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Ingredients     []Ingredient   `json:"ingredients"`
	Macros          MacroNutrients `json:"macros"`
	PreparationTime int            `json:"preparation_time"`
}

type MacroNutrients struct {
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
	Carbs   float64 `json:"carbs"`
}
