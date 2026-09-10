package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"nabutilivanie/internal/domain"
	"time"
)

var ErrNotFound = errors.New("not found")

type Store struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Store { return &Store{db} }
func scanTarget(r pgx.Row) (t domain.Target, err error) {
	err = r.Scan(&t.ID, &t.Name, &t.PhotoURL, &t.Description, &t.BottleCount, &t.TotalBottlesReceived, &t.IsLocked, &t.LockUntil, &t.CreatedAt)
	return
}
func (s *Store) CreateTarget(ctx context.Context, name, photo, desc string) (domain.Target, error) {
	return scanTarget(s.db.QueryRow(ctx, `INSERT INTO targets(name,photo_url,description) VALUES($1,$2,$3) RETURNING id::text,name,photo_url,description,bottle_count,total_bottles_received,is_locked,lock_until,created_at`, name, photo, desc))
}
func (s *Store) Leaderboard(ctx context.Context, limit int) ([]domain.Target, error) {
	rows, e := s.db.Query(ctx, `SELECT id::text,name,photo_url,description,bottle_count,total_bottles_received,is_locked,lock_until,created_at FROM targets ORDER BY bottle_count DESC,created_at DESC LIMIT $1`, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Target{}
	for rows.Next() {
		t, e := scanTarget(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) Search(ctx context.Context, q string, limit int) ([]domain.Target, error) {
	rows, e := s.db.Query(ctx, `SELECT id::text,name,photo_url,description,bottle_count,total_bottles_received,is_locked,lock_until,created_at FROM targets WHERE to_tsvector('simple',name || ' ' || description) @@ websearch_to_tsquery('simple',$1) ORDER BY bottle_count DESC LIMIT $2`, q, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Target{}
	for rows.Next() {
		t, e := scanTarget(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) Player(ctx context.Context, tid int64) (domain.Player, error) {
	var p domain.Player
	e := s.db.QueryRow(ctx, `INSERT INTO players(telegram_id) VALUES($1) ON CONFLICT(telegram_id) DO UPDATE SET telegram_id=EXCLUDED.telegram_id RETURNING id::text,telegram_id,bottle_balance,last_passive_claim_at`, tid).Scan(&p.ID, &p.TelegramID, &p.BottleBalance, &p.LastPassiveClaimAt)
	return p, e
}
func (s *Store) Credit(ctx context.Context, tid, bottles int64) (domain.Player, error) {
	var p domain.Player
	e := s.db.QueryRow(ctx, `UPDATE players SET bottle_balance=bottle_balance+$2 WHERE telegram_id=$1 RETURNING id::text,telegram_id,bottle_balance,last_passive_claim_at`, tid, bottles).Scan(&p.ID, &p.TelegramID, &p.BottleBalance, &p.LastPassiveClaimAt)
	return p, e
}
func (s *Store) ClaimPassive(ctx context.Context, tid int64) (domain.Player, int64, error) {
	tx, e := s.db.Begin(ctx)
	if e != nil {
		return domain.Player{}, 0, e
	}
	defer tx.Rollback(ctx)
	var p domain.Player
	e = tx.QueryRow(ctx, `SELECT id::text,telegram_id,bottle_balance,last_passive_claim_at FROM players WHERE telegram_id=$1 FOR UPDATE`, tid).Scan(&p.ID, &p.TelegramID, &p.BottleBalance, &p.LastPassiveClaimAt)
	if e != nil {
		return p, 0, e
	}
	earned := int64(time.Since(p.LastPassiveClaimAt).Hours())
	if earned > 0 {
		e = tx.QueryRow(ctx, `UPDATE players SET bottle_balance=bottle_balance+$2,last_passive_claim_at=last_passive_claim_at+($2 * INTERVAL '1 hour') WHERE id=$1 RETURNING bottle_balance,last_passive_claim_at`, p.ID, earned).Scan(&p.BottleBalance, &p.LastPassiveClaimAt)
		if e != nil {
			return p, 0, e
		}
	}
	return p, earned, tx.Commit(ctx)
}
func (s *Store) Nabutilit(ctx context.Context, tid int64, target, comment string) (domain.Target, error) {
	tx, e := s.db.Begin(ctx)
	if e != nil {
		return domain.Target{}, e
	}
	defer tx.Rollback(ctx)
	tag, e := tx.Exec(ctx, `UPDATE players SET bottle_balance=bottle_balance-1000 WHERE telegram_id=$1 AND bottle_balance>=1000`, tid)
	if e != nil {
		return domain.Target{}, e
	}
	if tag.RowsAffected() != 1 {
		return domain.Target{}, errors.New("insufficient bottles")
	}
	var t domain.Target
	t, e = scanTarget(tx.QueryRow(ctx, `UPDATE targets SET bottle_count=bottle_count+1000,total_bottles_received=total_bottles_received+1000 WHERE id=$1 AND (NOT is_locked OR lock_until<=NOW()) RETURNING id::text,name,photo_url,description,bottle_count,total_bottles_received,is_locked,lock_until,created_at`, target))
	if errors.Is(e, pgx.ErrNoRows) {
		return t, errors.New("target is locked or missing")
	}
	if e != nil {
		return t, e
	}
	_, e = tx.Exec(ctx, `INSERT INTO nabutilations(player_id,target_id,comment) SELECT id,$2,$3 FROM players WHERE telegram_id=$1`, tid, target, comment)
	if e != nil {
		return t, e
	}
	return t, tx.Commit(ctx)
}
func (s *Store) ProcessPayment(ctx context.Context, provider, external string, tid, bottles int64, payload any) (bool, domain.Player, error) {
	b, e := json.Marshal(payload)
	if e != nil {
		return false, domain.Player{}, e
	}
	tx, e := s.db.Begin(ctx)
	if e != nil {
		return false, domain.Player{}, e
	}
	defer tx.Rollback(ctx)
	var p domain.Player
	e = tx.QueryRow(ctx, `INSERT INTO players(telegram_id) VALUES($1) ON CONFLICT(telegram_id) DO UPDATE SET telegram_id=EXCLUDED.telegram_id RETURNING id::text,telegram_id,bottle_balance,last_passive_claim_at`, tid).Scan(&p.ID, &p.TelegramID, &p.BottleBalance, &p.LastPassiveClaimAt)
	if e != nil {
		return false, p, e
	}
	tag, e := tx.Exec(ctx, `INSERT INTO payment_events(provider,external_id,player_id,bottles,payload) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, provider, external, p.ID, bottles, b)
	if e != nil {
		return false, p, e
	}
	if tag.RowsAffected() == 0 {
		return false, p, tx.Commit(ctx)
	}
	e = tx.QueryRow(ctx, `UPDATE players SET bottle_balance=bottle_balance+$2 WHERE id=$1 RETURNING bottle_balance`, p.ID, bottles).Scan(&p.BottleBalance)
	if e != nil {
		return false, p, e
	}
	return true, p, tx.Commit(ctx)
}
