package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/financeos/api/internal/domain/entity"
	domainrepo "github.com/financeos/api/internal/domain/repository"
	"github.com/financeos/api/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Fake TransactionRepository for import tests ---

type fakeImportTransactionRepo struct {
	created   []*entity.Transaction
	importIDs map[string]bool
}

func newFakeImportTransactionRepo() *fakeImportTransactionRepo {
	return &fakeImportTransactionRepo{
		created:   []*entity.Transaction{},
		importIDs: make(map[string]bool),
	}
}

func (r *fakeImportTransactionRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	if tx.ImportID != nil && r.importIDs[*tx.ImportID] {
		return fmt.Errorf("duplicate key value violates unique constraint: 23505")
	}
	if tx.ImportID != nil {
		r.importIDs[*tx.ImportID] = true
	}
	r.created = append(r.created, tx)
	return nil
}

func (r *fakeImportTransactionRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error) {
	return nil, nil
}

func (r *fakeImportTransactionRepo) FindByUserID(ctx context.Context, userID uuid.UUID, filter domainrepo.TransactionFilter) ([]*entity.Transaction, int, error) {
	return r.created, len(r.created), nil
}

func (r *fakeImportTransactionRepo) Update(ctx context.Context, tx *entity.Transaction) error {
	return nil
}

func (r *fakeImportTransactionRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}

func (r *fakeImportTransactionRepo) GetSummary(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domainrepo.TransactionSummary, error) {
	return &domainrepo.TransactionSummary{}, nil
}

func (r *fakeImportTransactionRepo) CreateTransfer(ctx context.Context, debit, credit *entity.Transaction) error {
	return nil
}

func (r *fakeImportTransactionRepo) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, delta float64) error {
	return nil
}

// --- OFX Parse Tests ---

func TestParseOFX_ValidData(t *testing.T) {
	ofxData := []byte(`
OFXHEADER:100
DATA:OFXSGML

<OFX>
<BANKMSGSRSV1>
<STMTTRNRS>
<STMTRS>
<BANKTRANLIST>
<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20260115
<TRNAMT>-150.00
<FITID>20260115001
<NAME>Supermercado ABC
<MEMO>Compras do mês
</STMTTRN>
<STMTTRN>
<TRNTYPE>CREDIT
<DTPOSTED>20260120
<TRNAMT>3000.00
<FITID>20260120001
<NAME>Salário
<MEMO>Pagamento mensal
</STMTTRN>
</BANKTRANLIST>
</STMTRS>
</STMTTRNRS>
</BANKMSGSRSV1>
</OFX>
`)

	txs, err := usecase.ParseOFX(ofxData)
	require.NoError(t, err)
	require.Len(t, txs, 2)

	// First transaction: DEBIT
	assert.Equal(t, "DEBIT", txs[0].Type)
	assert.Equal(t, 150.0, txs[0].Amount)
	assert.Equal(t, "20260115001", txs[0].FitID)
	assert.Equal(t, "Supermercado ABC", txs[0].Name)
	assert.Equal(t, 2026, txs[0].Date.Year())
	assert.Equal(t, 1, int(txs[0].Date.Month()))
	assert.Equal(t, 15, txs[0].Date.Day())

	// Second transaction: CREDIT
	assert.Equal(t, "CREDIT", txs[1].Type)
	assert.Equal(t, 3000.0, txs[1].Amount)
	assert.Equal(t, "20260120001", txs[1].FitID)
	assert.Equal(t, "Salário", txs[1].Name)
}

func TestParseOFX_EmptyData(t *testing.T) {
	txs, err := usecase.ParseOFX([]byte(""))
	require.NoError(t, err)
	assert.Empty(t, txs)
}

func TestParseOFX_Deduplication(t *testing.T) {
	repo := newFakeImportTransactionRepo()
	uc := usecase.NewImportUseCase(repo)

	userID := uuid.New()
	accountID := uuid.New()

	ofxData := []byte(`
<OFX>
<STMTTRN>
<TRNTYPE>DEBIT
<DTPOSTED>20260101
<TRNAMT>-100.00
<FITID>UNIQUE-001
<NAME>Loja XYZ
</STMTTRN>
</OFX>
`)

	// First import
	result1, err := uc.ImportOFX(context.Background(), userID, accountID, ofxData)
	require.NoError(t, err)
	assert.Equal(t, 1, result1.Imported)
	assert.Equal(t, 0, result1.Duplicates)

	// Second import — same FitID should be duplicate
	result2, err := uc.ImportOFX(context.Background(), userID, accountID, ofxData)
	require.NoError(t, err)
	assert.Equal(t, 0, result2.Imported)
	assert.Equal(t, 1, result2.Duplicates)
}

// --- CSV Parse Tests ---

func TestParseCSV_BasicMapping(t *testing.T) {
	csvData := []byte(`data,valor,descricao,tipo
2026-01-15,150.00,Supermercado,D
2026-01-20,3000.00,Salário,C
2026-01-25,50.50,Farmácia,D
`)

	mapping := usecase.CSVMapping{
		DateCol:        0,
		AmountCol:      1,
		DescriptionCol: 2,
		TypeCol:        3,
		DateFormat:     "2006-01-02",
	}

	txs, err := usecase.ParseCSV(csvData, mapping)
	require.NoError(t, err)
	require.Len(t, txs, 3)

	assert.Equal(t, "DEBIT", txs[0].Type)
	assert.Equal(t, 150.0, txs[0].Amount)
	assert.Equal(t, "Supermercado", txs[0].Name)

	assert.Equal(t, "CREDIT", txs[1].Type)
	assert.Equal(t, 3000.0, txs[1].Amount)
	assert.Equal(t, "Salário", txs[1].Name)

	assert.Equal(t, "DEBIT", txs[2].Type)
	assert.Equal(t, 50.5, txs[2].Amount)
}

func TestParseCSV_SkipsInvalidRows(t *testing.T) {
	csvData := []byte(`date,amount,desc
not-a-date,100.00,Test
2026-02-01,not-a-number,Test2
2026-02-10,200.00,Valid
`)

	mapping := usecase.CSVMapping{
		DateCol:        0,
		AmountCol:      1,
		DescriptionCol: 2,
		TypeCol:        -1,
		DateFormat:     "2006-01-02",
	}

	txs, err := usecase.ParseCSV(csvData, mapping)
	require.NoError(t, err)
	// Only the valid row should be parsed
	require.Len(t, txs, 1)
	assert.Equal(t, 200.0, txs[0].Amount)
}

func TestImportCSV_Basic(t *testing.T) {
	repo := newFakeImportTransactionRepo()
	uc := usecase.NewImportUseCase(repo)

	userID := uuid.New()
	accountID := uuid.New()

	csvData := []byte(`date,amount,description,type
2026-01-10,250.00,Conta de luz,D
2026-01-15,5000.00,Salario,C
`)

	mapping := usecase.CSVMapping{
		DateCol:        0,
		AmountCol:      1,
		DescriptionCol: 2,
		TypeCol:        3,
		DateFormat:     "2006-01-02",
	}

	result, err := uc.ImportCSV(context.Background(), userID, accountID, csvData, mapping)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Imported)
	assert.Equal(t, 0, result.Duplicates)
	assert.Equal(t, 0, result.Errors)
}

func TestPreviewCSV(t *testing.T) {
	repo := newFakeImportTransactionRepo()
	uc := usecase.NewImportUseCase(repo)

	csvData := []byte(`date,amount,description,type
2026-01-01,100.00,Row1,D
2026-01-02,200.00,Row2,C
2026-01-03,300.00,Row3,D
2026-01-04,400.00,Row4,C
2026-01-05,500.00,Row5,D
2026-01-06,600.00,Row6,C
`)

	rows, err := uc.PreviewCSV(context.Background(), csvData)
	require.NoError(t, err)
	// Should return only 5 rows max
	assert.Len(t, rows, 5)
	assert.Equal(t, "2026-01-01", rows[0].Date)
	assert.Equal(t, 100.0, rows[0].Amount)
	assert.Equal(t, "Row1", rows[0].Description)
}
