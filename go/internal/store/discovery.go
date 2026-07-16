package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var validDiscoveryCards = map[string]bool{
	"day0": true, "day2": true, "day4": true, "day6": true,
}

type DiscoveryProgress struct {
	UserID         string    `json:"userId"`
	StartedAt      time.Time `json:"startedAt"`
	CompletedCards []string  `json:"completedCards"`
	StreakDays     int       `json:"streakDays"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (s *Store) GetDiscoveryProgress(ctx context.Context, userID string) (DiscoveryProgress, error) {
	var p DiscoveryProgress
	var cardsJSON []byte
	err := s.pool.QueryRow(ctx, `
		SELECT user_id::text, started_at, completed_cards, streak_days, updated_at
		FROM discovery.progress WHERE user_id = $1`, userID,
	).Scan(&p.UserID, &p.StartedAt, &cardsJSON, &p.StreakDays, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return DiscoveryProgress{
			UserID: userID, StartedAt: time.Now(), CompletedCards: []string{}, StreakDays: 0,
		}, nil
	}
	if err != nil {
		return DiscoveryProgress{}, err
	}
	if err := json.Unmarshal(cardsJSON, &p.CompletedCards); err != nil {
		p.CompletedCards = []string{}
	}
	return p, nil
}

func (s *Store) CompleteDiscoveryCard(ctx context.Context, userID, cardKey string) (DiscoveryProgress, error) {
	if !validDiscoveryCards[cardKey] {
		return DiscoveryProgress{}, errors.New("invalid_card_key")
	}
	existing, err := s.GetDiscoveryProgress(ctx, userID)
	if err != nil {
		return DiscoveryProgress{}, err
	}
	for _, c := range existing.CompletedCards {
		if c == cardKey {
			return existing, nil
		}
	}
	existing.CompletedCards = append(existing.CompletedCards, cardKey)
	cardsJSON, err := json.Marshal(existing.CompletedCards)
	if err != nil {
		return DiscoveryProgress{}, err
	}
	now := time.Now()
	streak := existing.StreakDays
	if existing.UserID == userID && !existing.UpdatedAt.IsZero() {
		lastDay := existing.UpdatedAt.Truncate(24 * time.Hour)
		today := now.Truncate(24 * time.Hour)
		diff := int(today.Sub(lastDay).Hours() / 24)
		switch {
		case diff == 0:
			// same day — keep streak
		case diff == 1:
			streak++
		default:
			streak = 1
		}
	} else {
		streak = 1
	}
	var p DiscoveryProgress
	err = s.pool.QueryRow(ctx, `
		INSERT INTO discovery.progress (user_id, started_at, completed_cards, streak_days, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			completed_cards = EXCLUDED.completed_cards,
			streak_days = EXCLUDED.streak_days,
			updated_at = EXCLUDED.updated_at
		RETURNING user_id::text, started_at, completed_cards, streak_days, updated_at`,
		userID, existing.StartedAt, cardsJSON, streak, now,
	).Scan(&p.UserID, &p.StartedAt, &cardsJSON, &p.StreakDays, &p.UpdatedAt)
	if err != nil {
		return DiscoveryProgress{}, err
	}
	if err := json.Unmarshal(cardsJSON, &p.CompletedCards); err != nil {
		p.CompletedCards = existing.CompletedCards
	}
	return p, nil
}
