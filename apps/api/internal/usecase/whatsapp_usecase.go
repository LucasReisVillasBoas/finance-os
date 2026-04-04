package usecase

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// BotState represents the state of a WhatsApp bot conversation.
type BotState string

const (
	StateIdle                 BotState = "idle"
	StateAwaitingConfirmation BotState = "awaiting_confirmation"

	sessionTimeoutMinutes = 15
)

// WhatsAppUseCase defines business logic for the WhatsApp bot.
type WhatsAppUseCase interface {
	HandleMessage(ctx context.Context, phone, message string) (string, error)
}

type whatsAppUseCase struct {
	whatsappRepo    domainrepo.WhatsAppRepository
	transactionRepo domainrepo.TransactionRepository
	accountRepo     domainrepo.AccountRepository
}

// NewWhatsAppUseCase creates a new WhatsAppUseCase implementation.
func NewWhatsAppUseCase(
	whatsappRepo domainrepo.WhatsAppRepository,
	transactionRepo domainrepo.TransactionRepository,
	accountRepo domainrepo.AccountRepository,
) WhatsAppUseCase {
	return &whatsAppUseCase{
		whatsappRepo:    whatsappRepo,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

func (uc *whatsAppUseCase) HandleMessage(ctx context.Context, phone, message string) (string, error) {
	message = strings.TrimSpace(message)

	// Find or create session
	session, err := uc.whatsappRepo.FindSessionByPhone(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("whatsAppUseCase.HandleMessage find session: %w", err)
	}

	if session == nil {
		// Try to find a user linked to this phone number
		user, err := uc.whatsappRepo.FindUserByPhone(ctx, phone)
		if err != nil {
			return "", fmt.Errorf("whatsAppUseCase.HandleMessage find user: %w", err)
		}

		userID := uuid.Nil
		if user != nil {
			userID = user.ID
		}

		session = &entity.WhatsAppSession{
			ID:           uuid.New(),
			UserID:       userID,
			PhoneNumber:  phone,
			State:        string(StateIdle),
			SessionData:  make(map[string]interface{}),
			LastActivity: time.Now(),
			IsActive:     true,
			CreatedAt:    time.Now(),
		}
		if err := uc.whatsappRepo.CreateSession(ctx, session); err != nil {
			return "", fmt.Errorf("whatsAppUseCase.HandleMessage create session: %w", err)
		}
	}

	// Check timeout — reset to idle if inactive > 15 minutes
	if time.Since(session.LastActivity) > sessionTimeoutMinutes*time.Minute {
		session.State = string(StateIdle)
		session.SessionData = make(map[string]interface{})
	}

	if session.UserID == uuid.Nil {
		session.LastActivity = time.Now()
		_ = uc.whatsappRepo.UpdateSession(ctx, session)
		return "Olá! Seu número ainda não está vinculado a uma conta FinanceOS. Acesse o app para configurar.", nil
	}

	var response string

	switch BotState(session.State) {
	case StateAwaitingConfirmation:
		response, err = uc.handleConfirmation(ctx, session, message)
	default:
		response, err = uc.handleIdleCommand(ctx, session, message)
	}

	if err != nil {
		return "", err
	}

	// Persist updated session
	session.LastActivity = time.Now()
	if updateErr := uc.whatsappRepo.UpdateSession(ctx, session); updateErr != nil {
		return "", fmt.Errorf("whatsAppUseCase.HandleMessage update session: %w", updateErr)
	}

	return response, nil
}

func (uc *whatsAppUseCase) handleIdleCommand(ctx context.Context, session *entity.WhatsAppSession, message string) (string, error) {
	lower := strings.ToLower(strings.TrimSpace(message))

	switch {
	case lower == "resumo":
		return uc.buildSummary(ctx, session.UserID)

	case strings.HasPrefix(lower, "gastei "):
		return uc.parseSpendCommand(session, message, "expense")

	case strings.HasPrefix(lower, "recebi "):
		return uc.parseSpendCommand(session, message, "income")

	case strings.HasPrefix(lower, "quanto gastei"):
		return uc.buildCategorySpend(ctx, session.UserID, message)

	case lower == "carteira":
		return uc.buildWalletSummary(ctx, session.UserID)

	default:
		return "Não entendi 🤔\n\nComandos disponíveis:\n• *resumo* — saldo e gastos do mês\n• *gastei X* ou *gastei X reais em Y* — registrar gasto\n• *recebi X* — registrar receita\n• *quanto gastei com [categoria]?* — consultar categoria\n• *carteira* — resumo de investimentos", nil
	}
}

func (uc *whatsAppUseCase) parseSpendCommand(session *entity.WhatsAppSession, message, txType string) (string, error) {
	lower := strings.ToLower(message)

	// Patterns: "gastei 50", "gastei 50 reais", "gastei 50 reais em supermercado", "gastei 50 em mercado"
	reAmount := regexp.MustCompile(`(\d+(?:[.,]\d+)?)`)
	reCategory := regexp.MustCompile(`(?i)em\s+(.+)$`)

	amounts := reAmount.FindStringSubmatch(lower)
	if len(amounts) < 2 {
		return "Não consegui identificar o valor. Tente: *gastei 50* ou *gastei 50 reais em mercado*", nil
	}

	amountStr := strings.ReplaceAll(amounts[1], ",", ".")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return "Valor inválido. Tente: *gastei 50.00*", nil
	}

	category := "Outros"
	if matches := reCategory.FindStringSubmatch(lower); len(matches) >= 2 {
		cat := strings.TrimSpace(matches[1])
		// Remove trailing "reais" or similar
		cat = regexp.MustCompile(`\s*(reais?|r\$)\s*`).ReplaceAllString(cat, "")
		if cat != "" {
			category = strings.Title(strings.TrimSpace(cat))
		}
	}

	action := "gasto"
	if txType == "income" {
		action = "receita"
	}

	// Store pending transaction in session
	session.SessionData["pending_amount"] = amount
	session.SessionData["pending_type"] = txType
	session.SessionData["pending_category"] = category
	session.SessionData["pending_date"] = time.Now().Format(time.RFC3339)
	session.State = string(StateAwaitingConfirmation)

	return fmt.Sprintf("Confirmar %s de *R$ %.2f* em *%s*?\n\nResponda *sim* para confirmar ou *não* para cancelar.", action, amount, category), nil
}

func (uc *whatsAppUseCase) handleConfirmation(ctx context.Context, session *entity.WhatsAppSession, message string) (string, error) {
	lower := strings.ToLower(strings.TrimSpace(message))

	confirm := lower == "sim" || lower == "confirmar" || lower == "s" || lower == "yes"
	cancel := lower == "não" || lower == "nao" || lower == "cancelar" || lower == "n" || lower == "no"

	if !confirm && !cancel {
		return "Responda *sim* para confirmar ou *não* para cancelar.", nil
	}

	// Always reset state after confirmation/cancel
	defer func() {
		session.State = string(StateIdle)
		session.SessionData = make(map[string]interface{})
	}()

	if cancel {
		return "Transação cancelada. ✅", nil
	}

	// Create the transaction
	amountRaw, ok := session.SessionData["pending_amount"]
	if !ok {
		return "Sessão expirada. Tente novamente.", nil
	}

	amount := toFloat64(amountRaw)
	txType, _ := session.SessionData["pending_type"].(string)
	category, _ := session.SessionData["pending_category"].(string)

	if amount <= 0 || txType == "" {
		return "Dados da transação inválidos. Tente novamente.", nil
	}

	// Get first active account for user
	accounts, err := uc.accountRepo.FindByUserID(ctx, session.UserID)
	if err != nil || len(accounts) == 0 {
		return "Não encontrei nenhuma conta ativa. Cadastre uma conta no app.", nil
	}

	now := time.Now()
	description := category
	tx := &entity.Transaction{
		ID:          uuid.New(),
		UserID:      session.UserID,
		AccountID:   accounts[0].ID,
		Type:        txType,
		Amount:      amount,
		Description: &description,
		Date:        now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.transactionRepo.Create(ctx, tx); err != nil {
		return "", fmt.Errorf("whatsAppUseCase.handleConfirmation create tx: %w", err)
	}

	action := "Gasto"
	if txType == "income" {
		action = "Receita"
	}

	return fmt.Sprintf("%s de *R$ %.2f* em *%s* registrado com sucesso! ✅", action, amount, category), nil
}

func (uc *whatsAppUseCase) buildSummary(ctx context.Context, userID uuid.UUID) (string, error) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	summary, err := uc.transactionRepo.GetSummary(ctx, userID, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("whatsAppUseCase.buildSummary: %w", err)
	}

	accounts, err := uc.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("whatsAppUseCase.buildSummary accounts: %w", err)
	}

	var totalBalance float64
	for _, acc := range accounts {
		totalBalance += acc.Balance
	}

	monthName := now.Format("January")
	msg := fmt.Sprintf("📊 *Resumo de %s*\n\n", monthName)
	msg += fmt.Sprintf("💰 Saldo total: *R$ %.2f*\n", totalBalance)
	msg += fmt.Sprintf("📈 Receitas: R$ %.2f\n", summary.TotalIncome)
	msg += fmt.Sprintf("📉 Gastos: R$ %.2f\n", summary.TotalExpense)
	msg += fmt.Sprintf("💵 Resultado: R$ %.2f\n", summary.Balance)

	return msg, nil
}

func (uc *whatsAppUseCase) buildCategorySpend(ctx context.Context, userID uuid.UUID, message string) (string, error) {
	re := regexp.MustCompile(`(?i)quanto gastei com\s+(.+?)[\?!.]*$`)
	matches := re.FindStringSubmatch(message)

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	summary, err := uc.transactionRepo.GetSummary(ctx, userID, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("whatsAppUseCase.buildCategorySpend: %w", err)
	}

	if len(matches) < 2 || summary == nil {
		return "Não encontrei dados de gastos por categoria este mês.", nil
	}

	categoryQuery := strings.ToLower(strings.TrimSpace(matches[1]))

	for _, cat := range summary.ByCategory {
		if strings.Contains(strings.ToLower(cat.CategoryName), categoryQuery) {
			return fmt.Sprintf("💳 Gastos com *%s* este mês: *R$ %.2f* (%d transações)", cat.CategoryName, cat.Total, cat.Count), nil
		}
	}

	return fmt.Sprintf("Não encontrei gastos com *%s* este mês.", categoryQuery), nil
}

func (uc *whatsAppUseCase) buildWalletSummary(ctx context.Context, userID uuid.UUID) (string, error) {
	accounts, err := uc.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("whatsAppUseCase.buildWalletSummary: %w", err)
	}

	var totalBalance float64
	for _, acc := range accounts {
		totalBalance += acc.Balance
	}

	msg := "💼 *Sua Carteira*\n\n"
	for _, acc := range accounts {
		msg += fmt.Sprintf("• %s: R$ %.2f\n", acc.Name, acc.Balance)
	}
	msg += fmt.Sprintf("\n💰 Total: *R$ %.2f*", totalBalance)

	return msg, nil
}

// toFloat64 safely converts an interface{} to float64.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return math.Abs(f)
	}
	return 0
}
