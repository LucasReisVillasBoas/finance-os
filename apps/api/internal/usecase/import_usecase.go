package usecase

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/google/uuid"
)

// OFXTransaction represents a parsed transaction from an OFX/QFX file.
type OFXTransaction struct {
	Type   string    // DEBIT or CREDIT
	Date   time.Time
	Amount float64
	FitID  string // unique ID for deduplication
	Name   string
	Memo   string
}

// CSVMapping defines the column mapping for CSV import.
type CSVMapping struct {
	DateCol        int
	AmountCol      int
	DescriptionCol int
	TypeCol        int
	DateFormat     string
}

// CSVPreviewRow represents a parsed preview row from CSV.
type CSVPreviewRow struct {
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
}

// ImportUseCase defines business logic for importing transactions.
type ImportUseCase interface {
	ImportOFX(ctx context.Context, userID, accountID uuid.UUID, data []byte) (*entity.ImportResult, error)
	ImportCSV(ctx context.Context, userID, accountID uuid.UUID, data []byte, mapping CSVMapping) (*entity.ImportResult, error)
	PreviewCSV(ctx context.Context, data []byte) ([]CSVPreviewRow, error)
}

type importUseCase struct {
	transactionRepo domainrepo.TransactionRepository
}

// NewImportUseCase creates a new ImportUseCase implementation.
func NewImportUseCase(transactionRepo domainrepo.TransactionRepository) ImportUseCase {
	return &importUseCase{transactionRepo: transactionRepo}
}

// ParseOFX parses OFX/QFX format data and returns a list of transactions.
// The format is SGML-like; we parse line by line looking for STMTTRN blocks.
func ParseOFX(data []byte) ([]OFXTransaction, error) {
	var transactions []OFXTransaction

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var current *OFXTransaction

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		upperLine := strings.ToUpper(line)

		switch {
		case upperLine == "<STMTTRN>" || strings.HasPrefix(upperLine, "<STMTTRN>"):
			current = &OFXTransaction{}

		case upperLine == "</STMTTRN>" || strings.HasPrefix(upperLine, "</STMTTRN>"):
			if current != nil {
				transactions = append(transactions, *current)
				current = nil
			}

		case current != nil:
			tag, value := parseOFXTag(line)
			switch strings.ToUpper(tag) {
			case "TRNTYPE":
				current.Type = strings.ToUpper(value)
			case "DTPOSTED":
				parsed, err := parseOFXDate(value)
				if err == nil {
					current.Date = parsed
				}
			case "TRNAMT":
				clean := strings.ReplaceAll(value, ",", ".")
				amount, err := strconv.ParseFloat(clean, 64)
				if err == nil {
					current.Amount = math.Abs(amount)
					// Preserve sign info via Type if not already set
					if amount < 0 && current.Type == "" {
						current.Type = "DEBIT"
					} else if amount > 0 && current.Type == "" {
						current.Type = "CREDIT"
					}
				}
			case "FITID":
				current.FitID = value
			case "NAME":
				current.Name = value
			case "MEMO":
				current.Memo = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ParseOFX scan: %w", err)
	}

	return transactions, nil
}

func parseOFXTag(line string) (tag, value string) {
	// Format: <TAGNAME>value or <TAGNAME>
	if !strings.HasPrefix(line, "<") {
		return "", line
	}
	closeIdx := strings.Index(line, ">")
	if closeIdx < 0 {
		return "", line
	}
	tag = line[1:closeIdx]
	value = strings.TrimSpace(line[closeIdx+1:])
	// Remove closing tag if present (e.g., value</TAGNAME>)
	if idx := strings.Index(value, "</"); idx >= 0 {
		value = value[:idx]
	}
	return tag, value
}

func parseOFXDate(s string) (time.Time, error) {
	// OFX dates: YYYYMMDD or YYYYMMDDHHMMSS or YYYYMMDDHHMMSS.XXX[TZ]
	s = strings.TrimSpace(s)
	// Remove timezone info after [ or .
	for _, sep := range []string{"[", "."} {
		if idx := strings.Index(s, sep); idx >= 0 {
			s = s[:idx]
		}
	}
	formats := []string{
		"20060102150405",
		"20060102",
	}
	for _, f := range formats {
		if len(s) >= len(f) {
			t, err := time.Parse(f, s[:len(f)])
			if err == nil {
				return t, nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse OFX date: %s", s)
}

// ParseCSV parses CSV data using the provided column mapping.
func ParseCSV(data []byte, mapping CSVMapping) ([]OFXTransaction, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ParseCSV read: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	// Skip header row
	dataRows := records[1:]
	if len(dataRows) == 0 {
		return nil, nil
	}

	dateFormat := mapping.DateFormat
	if dateFormat == "" {
		dateFormat = "2006-01-02"
	}

	maxCol := mapping.DateCol
	for _, c := range []int{mapping.AmountCol, mapping.DescriptionCol, mapping.TypeCol} {
		if c > maxCol {
			maxCol = c
		}
	}

	var result []OFXTransaction
	for i, row := range dataRows {
		if len(row) <= maxCol {
			continue
		}

		dateStr := strings.TrimSpace(row[mapping.DateCol])
		date, err := time.Parse(dateFormat, dateStr)
		if err != nil {
			// Try common formats
			for _, f := range []string{"02/01/2006", "01/02/2006", "2006-01-02", "02-01-2006"} {
				if t, e := time.Parse(f, dateStr); e == nil {
					date = t
					err = nil
					break
				}
			}
			if err != nil {
				continue // skip unparseable dates
			}
		}

		amountStr := strings.TrimSpace(row[mapping.AmountCol])
		// Remove currency symbols and spaces
		amountStr = strings.Map(func(r rune) rune {
			if unicode.IsDigit(r) || r == '.' || r == ',' || r == '-' {
				return r
			}
			return -1
		}, amountStr)
		amountStr = strings.ReplaceAll(amountStr, ",", ".")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			continue
		}

		description := ""
		if mapping.DescriptionCol >= 0 && mapping.DescriptionCol < len(row) {
			description = strings.TrimSpace(row[mapping.DescriptionCol])
		}

		txType := "DEBIT"
		if mapping.TypeCol >= 0 && mapping.TypeCol < len(row) {
			t := strings.ToUpper(strings.TrimSpace(row[mapping.TypeCol]))
			if strings.Contains(t, "C") || strings.Contains(t, "CREDIT") || strings.Contains(t, "ENTRADA") {
				txType = "CREDIT"
			}
		} else if amount > 0 {
			txType = "CREDIT"
		}

		fitID := fmt.Sprintf("csv-%d-%s-%.2f", i, dateStr, math.Abs(amount))

		result = append(result, OFXTransaction{
			Type:   txType,
			Date:   date,
			Amount: math.Abs(amount),
			FitID:  fitID,
			Name:   description,
			Memo:   description,
		})
	}

	return result, nil
}

func (uc *importUseCase) ImportOFX(ctx context.Context, userID, accountID uuid.UUID, data []byte) (*entity.ImportResult, error) {
	txs, err := ParseOFX(data)
	if err != nil {
		return nil, fmt.Errorf("importUseCase.ImportOFX parse: %w", err)
	}

	result := &entity.ImportResult{
		Messages: []string{},
	}

	for _, ofxTx := range txs {
		txType := "expense"
		if ofxTx.Type == "CREDIT" {
			txType = "income"
		}

		description := ofxTx.Name
		if description == "" {
			description = ofxTx.Memo
		}

		importID := ofxTx.FitID
		now := time.Now()

		tx := &entity.Transaction{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Type:        txType,
			Amount:      ofxTx.Amount,
			Description: &description,
			Date:        ofxTx.Date,
			ImportID:    &importID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := uc.transactionRepo.Create(ctx, tx); err != nil {
			errMsg := err.Error()
			// Check for duplicate import_id (unique constraint)
			if strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "23505") {
				result.Duplicates++
				result.Messages = append(result.Messages, fmt.Sprintf("Duplicata ignorada: %s (%s)", description, ofxTx.FitID))
			} else {
				result.Errors++
				result.Messages = append(result.Messages, fmt.Sprintf("Erro ao importar: %s — %v", description, err))
			}
			continue
		}

		result.Imported++
	}

	return result, nil
}

func (uc *importUseCase) ImportCSV(ctx context.Context, userID, accountID uuid.UUID, data []byte, mapping CSVMapping) (*entity.ImportResult, error) {
	txs, err := ParseCSV(data, mapping)
	if err != nil {
		return nil, fmt.Errorf("importUseCase.ImportCSV parse: %w", err)
	}

	result := &entity.ImportResult{
		Messages: []string{},
	}

	for _, csvTx := range txs {
		txType := "expense"
		if csvTx.Type == "CREDIT" {
			txType = "income"
		}

		description := csvTx.Name
		importID := csvTx.FitID
		now := time.Now()

		tx := &entity.Transaction{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Type:        txType,
			Amount:      csvTx.Amount,
			Description: &description,
			Date:        csvTx.Date,
			ImportID:    &importID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := uc.transactionRepo.Create(ctx, tx); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "23505") {
				result.Duplicates++
			} else {
				result.Errors++
				result.Messages = append(result.Messages, fmt.Sprintf("Erro ao importar linha: %v", err))
			}
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (uc *importUseCase) PreviewCSV(ctx context.Context, data []byte) ([]CSVPreviewRow, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("importUseCase.PreviewCSV read: %w", err)
	}

	if len(records) == 0 {
		return []CSVPreviewRow{}, nil
	}

	// Skip header, take up to 5 data rows
	dataRows := records[1:]
	if len(dataRows) > 5 {
		dataRows = dataRows[:5]
	}

	result := make([]CSVPreviewRow, 0, len(dataRows))
	for _, row := range dataRows {
		if len(row) == 0 {
			continue
		}
		preview := CSVPreviewRow{}
		if len(row) > 0 {
			preview.Date = strings.TrimSpace(row[0])
		}
		if len(row) > 1 {
			amountStr := strings.TrimSpace(row[1])
			amountStr = strings.Map(func(r rune) rune {
				if unicode.IsDigit(r) || r == '.' || r == ',' || r == '-' {
					return r
				}
				return -1
			}, amountStr)
			amountStr = strings.ReplaceAll(amountStr, ",", ".")
			if v, err := strconv.ParseFloat(amountStr, 64); err == nil {
				preview.Amount = math.Abs(v)
			}
		}
		if len(row) > 2 {
			preview.Description = strings.TrimSpace(row[2])
		}
		if len(row) > 3 {
			preview.Type = strings.TrimSpace(row[3])
		}
		result = append(result, preview)
	}

	return result, nil
}
