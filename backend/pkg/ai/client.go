package ai

import (
	"context"
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
	prompt := `Analyze the food in this image and return ONLY a JSON object with this exact structure (no markdown, no extra text):
{
  "description": "brief food description",
  "calories": 0,
  "protein": 0.0,
  "carbs": 0.0,
  "fat": 0.0
}
All numeric values must be estimates per portion shown. Calories as integer, macros as float with 1 decimal.`

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

// GenerateRecommendations asks GPT to generate personalized fitness recommendations.
func (cl *Client) GenerateRecommendations(ctx context.Context, profile UserProfile) ([]RecommendationItem, error) {
	weightProgress := ""
	if profile.TargetWeight > 0 && profile.Weight > 0 {
		delta := profile.Weight - profile.TargetWeight
		if delta > 0.5 {
			weightProgress = fmt.Sprintf(", needs to lose %.1fkg to reach target %.1fkg", delta, profile.TargetWeight)
		} else if delta < -0.5 {
			weightProgress = fmt.Sprintf(", needs to gain %.1fkg to reach target %.1fkg", -delta, profile.TargetWeight)
		} else {
			weightProgress = ", is at target weight"
		}
	}

	prompt := fmt.Sprintf(`You are a fitness coach. Generate 3-5 personalized recommendations for this user.
User profile: weight=%.1fkg, height=%.0fcm, age=%d, goal=%s, activity_level=%s%s.
Include at least one recommendation about their nutrition/workout plan relative to their weight progress toward target.
Return ONLY a JSON array (no markdown):
[{"type":"workout|nutrition|rest|general","description":"actionable advice","priority":1-3}]
Priority: 1=high, 2=medium, 3=low. Keep each description under 100 characters.`,
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
	prompt := fmt.Sprintf(`You are a nutritionist. A user has consumed today: %d kcal, %.1fg protein, %.1fg carbs, %.1fg fat.
Suggest 3-5 recipes for their next meal that complement what they've already eaten (balance the remaining macros).
Return ONLY a JSON array (no markdown, no extra text):
[{
  "name": "Recipe name",
  "description": "Brief taste/flavour description (1 sentence, max 80 chars)",
  "ingredients": [{"name": "Ingredient name", "grams": 150}],
  "macros": {"protein": 0.0, "fat": 0.0, "carbs": 0.0},
  "preparation_time": 10
}]
Rules: macros in grams (floats), preparation_time in minutes (integer), name max 50 chars, list 3-6 ingredients with realistic gram weights.`,
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
