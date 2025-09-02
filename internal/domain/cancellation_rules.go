package domain

import "time"

type CancellationRuleType string

const (
	CancellationRuleTypeGeneral    CancellationRuleType = "general"
	CancellationRuleTypeRefund     CancellationRuleType = "refund"
	CancellationRuleTypeConditions CancellationRuleType = "conditions"
)

type CancellationRule struct {
	ID           int                  `json:"id" db:"id"`
	RuleType     CancellationRuleType `json:"rule_type" db:"rule_type"`
	Title        string               `json:"title" db:"title"`
	Content      string               `json:"content" db:"content"`
	IsActive     bool                 `json:"is_active" db:"is_active"`
	DisplayOrder int                  `json:"display_order" db:"display_order"`
	CreatedAt    time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at" db:"updated_at"`
}

type CancellationRuleRepository interface {
	GetAll() ([]*CancellationRule, error)
	GetByType(ruleType CancellationRuleType) ([]*CancellationRule, error)
	GetByID(id int) (*CancellationRule, error)
	Create(rule *CancellationRule) error
	Update(rule *CancellationRule) error
	Delete(id int) error
}

type CancellationRuleUseCase interface {
	GetActiveCancellationRules() ([]*CancellationRule, error)
	GetCancellationRulesByType(ruleType CancellationRuleType) ([]*CancellationRule, error)
}
