package usecase

import (
	"github.com/russo2642/renti_kz/internal/domain"
)

type cancellationRuleUseCase struct {
	cancellationRuleRepo domain.CancellationRuleRepository
}

func NewCancellationRuleUseCase(
	cancellationRuleRepo domain.CancellationRuleRepository,
) domain.CancellationRuleUseCase {
	return &cancellationRuleUseCase{
		cancellationRuleRepo: cancellationRuleRepo,
	}
}

func (u *cancellationRuleUseCase) GetActiveCancellationRules() ([]*domain.CancellationRule, error) {
	return u.cancellationRuleRepo.GetAll()
}

func (u *cancellationRuleUseCase) GetCancellationRulesByType(ruleType domain.CancellationRuleType) ([]*domain.CancellationRule, error) {
	return u.cancellationRuleRepo.GetByType(ruleType)
}
