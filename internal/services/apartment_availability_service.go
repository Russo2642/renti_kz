package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type ApartmentAvailabilityService struct {
	db            *sql.DB
	apartmentRepo domain.ApartmentRepository
}

func NewApartmentAvailabilityService(
	db *sql.DB,
	apartmentRepo domain.ApartmentRepository,
) *ApartmentAvailabilityService {
	return &ApartmentAvailabilityService{
		db:            db,
		apartmentRepo: apartmentRepo,
	}
}

func (s *ApartmentAvailabilityService) RecalculateApartmentAvailability(apartmentID int) error {
	isFree, err := s.calculateIsFree(apartmentID)
	if err != nil {
		return fmt.Errorf("failed to calculate availability for apartment %d: %w", apartmentID, err)
	}

	err = s.apartmentRepo.UpdateIsFree(apartmentID, isFree)
	if err != nil {
		return fmt.Errorf("failed to update is_free for apartment %d: %w", apartmentID, err)
	}

	log.Printf("🏠 Квартира %d: is_free = %t", apartmentID, isFree)
	return nil
}

func (s *ApartmentAvailabilityService) calculateIsFree(apartmentID int) (bool, error) {
	now := time.Now()

	query := `
		SELECT COUNT(*) 
		FROM bookings 
		WHERE apartment_id = $1 
		AND (
			-- Активные прямо сейчас
			(status = 'active' AND start_date <= $2 AND end_date > $2)
			OR
			-- Подтвержденные/ожидающие в ближайшие 2 часа (увеличено окно)
			(status IN ('approved', 'pending', 'awaiting_payment') 
			 AND start_date BETWEEN $2 AND $2 + INTERVAL '2 hours')
			OR
			-- Недавно созданные бронирования с началом в ближайшие 2 часа
			(status = 'created' 
			 AND created_at > $2 - INTERVAL '30 minutes'
			 AND start_date BETWEEN $2 AND $2 + INTERVAL '2 hours')
		)`

	var conflictCount int
	err := s.db.QueryRow(query, apartmentID, now).Scan(&conflictCount)
	if err != nil {
		return false, fmt.Errorf("failed to check apartment availability: %w", err)
	}

	return conflictCount == 0, nil
}

func (s *ApartmentAvailabilityService) RecalculateAllApartments() error {
	start := time.Now()

	query := `
		SELECT DISTINCT apartment_id 
		FROM bookings 
		WHERE status IN ('active', 'approved', 'pending', 'awaiting_payment', 'created')
		AND (
			-- Активные сейчас
			(status = 'active' AND start_date <= NOW() AND end_date > NOW()) OR
			-- Начинающиеся в ближайшие 2 часа
			(start_date BETWEEN NOW() AND NOW() + INTERVAL '2 hours') OR
			-- Завершающиеся в ближайшие 2 часа
			(end_date BETWEEN NOW() - INTERVAL '2 hours' AND NOW() + INTERVAL '2 hours')
		)
		UNION
		-- Добавляем квартиры которые могут стать свободными
		SELECT DISTINCT apartment_id
		FROM apartments 
		WHERE is_free = false`

	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to get apartments for recalculation: %w", err)
	}
	defer rows.Close()

	var apartmentIDs []int
	for rows.Next() {
		var apartmentID int
		if err := rows.Scan(&apartmentID); err != nil {
			continue
		}
		apartmentIDs = append(apartmentIDs, apartmentID)
	}

	if len(apartmentIDs) > 10 {
		err = s.RecalculateMultipleApartments(apartmentIDs)
		if err != nil {
			log.Printf("❌ Ошибка батчевого пересчета, fallback на по одной: %v", err)
			updated := 0
			for _, apartmentID := range apartmentIDs {
				if err := s.RecalculateApartmentAvailability(apartmentID); err != nil {
					log.Printf("❌ Ошибка пересчёта квартиры %d: %v", apartmentID, err)
					continue
				}
				updated++
			}
			log.Printf("🔄 Fallback пересчет: %d из %d квартир за %v", updated, len(apartmentIDs), time.Since(start))
		} else {
			log.Printf("🚀 Батчевый пересчет: %d квартир за %v", len(apartmentIDs), time.Since(start))
		}
	} else {
		updated := 0
		for _, apartmentID := range apartmentIDs {
			if err := s.RecalculateApartmentAvailability(apartmentID); err != nil {
				log.Printf("❌ Ошибка пересчёта квартиры %d: %v", apartmentID, err)
				continue
			}
			updated++
		}
		log.Printf("🔄 Обычный пересчет: %d квартир за %v", updated, time.Since(start))
	}

	return nil
}

func (s *ApartmentAvailabilityService) CleanupExpiredCreatedBookings() error {
	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)

	query := `
		SELECT DISTINCT apartment_id 
		FROM bookings 
		WHERE status = 'created' 
		AND created_at <= $1`

	rows, err := s.db.Query(query, thirtyMinutesAgo)
	if err != nil {
		return fmt.Errorf("failed to find apartments with expired created bookings: %w", err)
	}
	defer rows.Close()

	var apartmentIDs []int
	for rows.Next() {
		var apartmentID int
		if err := rows.Scan(&apartmentID); err != nil {
			continue
		}
		apartmentIDs = append(apartmentIDs, apartmentID)
	}

	err = s.RecalculateMultipleApartments(apartmentIDs)
	if err != nil {
		log.Printf("❌ Ошибка батчевого пересчета: %v", err)
		for _, apartmentID := range apartmentIDs {
			s.RecalculateApartmentAvailability(apartmentID)
		}
	}

	log.Printf("🧹 Очищено expired created бронирований для %d квартир", len(apartmentIDs))
	return nil
}

func (s *ApartmentAvailabilityService) RecalculateMultipleApartments(apartmentIDs []int) error {
	if len(apartmentIDs) == 0 {
		return nil
	}

	start := time.Now()

	apartmentStatusMap, err := s.calculateMultipleIsFree(apartmentIDs)
	if err != nil {
		return fmt.Errorf("failed to calculate multiple apartment statuses: %w", err)
	}

	err = s.apartmentRepo.UpdateMultipleIsFree(apartmentStatusMap)
	if err != nil {
		return fmt.Errorf("failed to batch update is_free statuses: %w", err)
	}

	for apartmentID, isFree := range apartmentStatusMap {
		log.Printf("🏠 Квартира %d: is_free = %t", apartmentID, isFree)
	}

	log.Printf("🚀 Батчевый пересчет %d квартир завершен за %v", len(apartmentIDs), time.Since(start))
	return nil
}

func (s *ApartmentAvailabilityService) calculateMultipleIsFree(apartmentIDs []int) (map[int]bool, error) {
	if len(apartmentIDs) == 0 {
		return make(map[int]bool), nil
	}

	now := time.Now()

	idList := make([]interface{}, len(apartmentIDs))
	for i, id := range apartmentIDs {
		idList[i] = id
	}

	query := `
		SELECT DISTINCT apartment_id 
		FROM bookings 
		WHERE apartment_id = ANY($1)
		AND (
			-- Активные прямо сейчас
			(status = 'active' AND start_date <= $2 AND end_date > $2)
			OR
			-- Подтвержденные/ожидающие в ближайшие 2 часа
			(status IN ('approved', 'pending', 'awaiting_payment') 
			 AND start_date BETWEEN $2 AND $2 + INTERVAL '2 hours')
			OR
			-- Недавно созданные бронирования с началом в ближайшие 2 часа
			(status = 'created' 
			 AND created_at > $2 - INTERVAL '30 minutes'
			 AND start_date BETWEEN $2 AND $2 + INTERVAL '2 hours')
		)`

	apartmentIDsArray := "{" + fmt.Sprintf("%d", apartmentIDs[0])
	for i := 1; i < len(apartmentIDs); i++ {
		apartmentIDsArray += fmt.Sprintf(",%d", apartmentIDs[i])
	}
	apartmentIDsArray += "}"

	rows, err := s.db.Query(query, apartmentIDsArray, now)
	if err != nil {
		return nil, fmt.Errorf("failed to check multiple apartments availability: %w", err)
	}
	defer rows.Close()

	occupiedApartments := make(map[int]bool)
	for rows.Next() {
		var apartmentID int
		if err := rows.Scan(&apartmentID); err != nil {
			continue
		}
		occupiedApartments[apartmentID] = true
	}

	result := make(map[int]bool, len(apartmentIDs))
	for _, apartmentID := range apartmentIDs {
		result[apartmentID] = !occupiedApartments[apartmentID]
	}

	return result, nil
}
