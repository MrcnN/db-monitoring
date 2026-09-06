package advisor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dbplatform/api/internal/crypto"
	"github.com/dbplatform/api/internal/database"
	"github.com/dbplatform/api/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	dbSvc     *database.Service
	encryptor *crypto.Encryptor
}

func NewService(dbSvc *database.Service, encryptor *crypto.Encryptor) *Service {
	return &Service{
		dbSvc:     dbSvc,
		encryptor: encryptor,
	}
}

type ExplainRequest struct {
	Query string `json:"query"`
}

type AdvisorResult struct {
	Plan           interface{} `json:"plan"`
	Recommendations []string    `json:"recommendations"`
}

// PostgreSQL EXPLAIN JSON structure
type PGExplainPlan struct {
	Plan PGExplainNode `json:"Plan"`
}

type PGExplainNode struct {
	NodeType     string          `json:"Node Type"`
	RelationName string          `json:"Relation Name,omitempty"`
	Filter       string          `json:"Filter,omitempty"`
	Plans        []PGExplainNode `json:"Plans,omitempty"`
}

func (s *Service) ExplainQuery(ctx context.Context, dbID uuid.UUID, query string) (*AdvisorResult, error) {
	// 1. Get database target
	target, err := s.dbSvc.GetTarget(ctx, dbID)
	if err != nil {
		return nil, err
	}

	if target.Type != database.TypePostgreSQL {
		return nil, errors.NewValidation("Query advisor is currently only supported for PostgreSQL")
	}

	// 2. Prevent dangerous queries (basic heuristic)
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(upperQuery, "SELECT") && !strings.HasPrefix(upperQuery, "WITH") {
		return nil, errors.NewValidation("Only SELECT queries can be analyzed")
	}

	// 3. Connect to the target DB
	password, err := s.encryptor.Decrypt(target.EncryptedPassword)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		target.Username, password, target.Host, target.Port, target.DatabaseName, target.SSLMode)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, errors.NewInternal(fmt.Errorf("failed to connect to target db: %w", err))
	}
	defer conn.Close(ctx)

	// 4. Execute EXPLAIN
	explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", query)
	var planJSON string
	err = conn.QueryRow(ctx, explainQuery).Scan(&planJSON)
	if err != nil {
		return nil, errors.NewInternal(fmt.Errorf("failed to execute EXPLAIN: %w", err))
	}

	// 5. Parse and Analyze
	var plans []PGExplainPlan
	if err := json.Unmarshal([]byte(planJSON), &plans); err != nil {
		return nil, errors.NewInternal(fmt.Errorf("failed to parse EXPLAIN output: %w", err))
	}

	recommendations := []string{}
	if len(plans) > 0 {
		analyzePlanNode(plans[0].Plan, &recommendations)
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Query looks well optimized. No obvious bottlenecks detected.")
	}

	// Convert parsed JSON into generic interface for frontend rendering
	var rawPlan interface{}
	json.Unmarshal([]byte(planJSON), &rawPlan)

	return &AdvisorResult{
		Plan:            rawPlan,
		Recommendations: recommendations,
	}, nil
}

func analyzePlanNode(node PGExplainNode, recommendations *[]string) {
	if node.NodeType == "Seq Scan" {
		rec := fmt.Sprintf("Sequential Scan detected on table '%s'.", node.RelationName)
		if node.Filter != "" {
			rec += fmt.Sprintf(" Consider adding an index covering the filter: %s", node.Filter)
		} else {
			rec += " If this table is large, consider adding indexes on frequently queried columns or adding a LIMIT."
		}
		*recommendations = append(*recommendations, rec)
	}

	if node.NodeType == "Hash Join" || node.NodeType == "Nested Loop" {
		// Just a heuristic check for joins
		if node.RelationName != "" {
			*recommendations = append(*recommendations, fmt.Sprintf("Ensure foreign keys are indexed for join operations involving '%s'.", node.RelationName))
		}
	}

	for _, child := range node.Plans {
		analyzePlanNode(child, recommendations)
	}
}
