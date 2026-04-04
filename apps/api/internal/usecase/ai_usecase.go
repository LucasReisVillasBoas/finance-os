package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/pkg/claude"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// AIUseCase defines AI-powered financial features.
type AIUseCase interface {
	Chat(ctx context.Context, userID uuid.UUID, message string) (string, error)
	GetSpendingForecast(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetPortfolioAnalysis(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
}

type aiUseCase struct {
	claudeClient    *claude.Client
	transactionRepo domainrepo.TransactionRepository
	rdb             *redis.Client
}

// NewAIUseCase creates a new AIUseCase.
func NewAIUseCase(claudeClient *claude.Client, transactionRepo domainrepo.TransactionRepository, rdb *redis.Client) AIUseCase {
	return &aiUseCase{
		claudeClient:    claudeClient,
		transactionRepo: transactionRepo,
		rdb:             rdb,
	}
}

func (uc *aiUseCase) Chat(ctx context.Context, userID uuid.UUID, message string) (string, error) {
	system := "Você é um assistente financeiro pessoal do FinanceOS, um aplicativo de controle financeiro. " +
		"Responda em português brasileiro de forma concisa e útil. " +
		"Ajude o usuário com questões sobre finanças pessoais, orçamento, investimentos e economias."

	response, err := uc.claudeClient.Complete(ctx, system, message)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Chat: %w", err)
	}
	return response, nil
}

func (uc *aiUseCase) GetSpendingForecast(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("ai:forecast:%s", userID)

	// Try cache first
	cached, err := uc.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var result map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(cached), &result); jsonErr == nil {
			return result, nil
		}
	}

	// Fetch last 6 months of transactions
	now := time.Now()
	startDate := now.AddDate(0, -6, 0)
	summary, err := uc.transactionRepo.GetSummary(ctx, userID, startDate, now)
	if err != nil {
		return nil, fmt.Errorf("aiUseCase.GetSpendingForecast get summary: %w", err)
	}

	// Build prompt
	var catLines []string
	for _, cat := range summary.ByCategory {
		catLines = append(catLines, fmt.Sprintf("- %s: R$ %.2f", cat.CategoryName, cat.Total))
	}

	userMsg := fmt.Sprintf(
		"Analise os gastos dos últimos 6 meses:\n"+
			"Total de receitas: R$ %.2f\n"+
			"Total de despesas: R$ %.2f\n"+
			"Por categoria:\n%s\n\n"+
			"Com base nesses dados, faça uma previsão de gastos para o próximo mês e dê 3 sugestões práticas de economia. "+
			"Retorne um JSON com campos: 'previsao_proximos_30_dias' (número), 'sugestoes' (lista de strings), 'analise' (string).",
		summary.TotalIncome,
		summary.TotalExpense,
		strings.Join(catLines, "\n"),
	)

	system := "Você é um analista financeiro especializado. Responda SEMPRE com JSON válido, sem markdown."
	response, err := uc.claudeClient.Complete(ctx, system, userMsg)
	if err != nil {
		return nil, fmt.Errorf("aiUseCase.GetSpendingForecast claude: %w", err)
	}

	// Parse JSON response
	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(response), &result); jsonErr != nil {
		// If not parseable, wrap in a map
		result = map[string]interface{}{
			"analise":                    response,
			"previsao_proximos_30_dias":  summary.TotalExpense / 6,
			"sugestoes":                  []string{},
		}
	}

	// Cache for 24h
	if cacheData, jsonErr := json.Marshal(result); jsonErr == nil {
		uc.rdb.Set(ctx, cacheKey, cacheData, 24*time.Hour) //nolint:errcheck
	}

	return result, nil
}

func (uc *aiUseCase) GetPortfolioAnalysis(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("ai:portfolio:%s", userID)

	// Try cache first
	cached, err := uc.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var result map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(cached), &result); jsonErr == nil {
			return result, nil
		}
	}

	// Fetch recent income as proxy for investment capacity
	now := time.Now()
	startDate := now.AddDate(0, -3, 0)
	summary, err := uc.transactionRepo.GetSummary(ctx, userID, startDate, now)
	if err != nil {
		return nil, fmt.Errorf("aiUseCase.GetPortfolioAnalysis: %w", err)
	}

	monthlyBalance := (summary.TotalIncome - summary.TotalExpense) / 3
	userMsg := fmt.Sprintf(
		"Saldo médio mensal disponível para investimentos: R$ %.2f\n"+
			"Receita média mensal: R$ %.2f\n"+
			"Gasto médio mensal: R$ %.2f\n\n"+
			"Faça uma análise de portfólio com sugestões de alocação de ativos adequada para um investidor brasileiro. "+
			"Retorne um JSON com: 'perfil_sugerido' (string), 'alocacao_sugerida' (objeto com percentuais), 'recomendacoes' (lista de strings), 'analise' (string).",
		monthlyBalance,
		summary.TotalIncome/3,
		summary.TotalExpense/3,
	)

	system := "Você é um assessor de investimentos. Responda SEMPRE com JSON válido, sem markdown."
	response, err := uc.claudeClient.Complete(ctx, system, userMsg)
	if err != nil {
		return nil, fmt.Errorf("aiUseCase.GetPortfolioAnalysis claude: %w", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(response), &result); jsonErr != nil {
		result = map[string]interface{}{
			"analise":         response,
			"recomendacoes":   []string{},
		}
	}

	if cacheData, jsonErr := json.Marshal(result); jsonErr == nil {
		uc.rdb.Set(ctx, cacheKey, cacheData, 24*time.Hour) //nolint:errcheck
	}

	return result, nil
}

// CategorizationResult holds the result of AI categorization.
type CategorizationResult struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	CategoryName  string    `json:"category_name"`
	Confidence    float64   `json:"confidence"`
}

// CategorizeTransaction uses Claude to suggest a category for a transaction description.
func CategorizeTransaction(ctx context.Context, client *claude.Client, description string) (string, error) {
	system := "Você é um categorizador financeiro. Retorne APENAS o nome da categoria em português, sem explicações."
	userMsg := fmt.Sprintf(
		"Categorize esta transação: '%s'. "+
			"Categorias disponíveis: Alimentação, Transporte, Moradia, Saúde, Educação, Lazer, Roupas, Tecnologia, Assinaturas, Outros, Salário, Freelance, Investimentos",
		description,
	)
	result, err := client.Complete(ctx, system, userMsg)
	if err != nil {
		return "", fmt.Errorf("CategorizeTransaction: %w", err)
	}
	return strings.TrimSpace(result), nil
}

// AISuggestion holds an AI-generated financial suggestion.
type AISuggestion struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// BuildCategoryMapping returns a case-insensitive map from category name to entity.
func BuildCategoryMapping(categories []*entity.Category) map[string]*entity.Category {
	m := make(map[string]*entity.Category, len(categories))
	for _, c := range categories {
		m[strings.ToLower(c.Name)] = c
	}
	return m
}
