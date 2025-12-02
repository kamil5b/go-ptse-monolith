package sql

import (
	"context"
	"time"

	"go-modular-monolith/internal/modules/auth/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const authDriverName = "AuthPostgreSQL"

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) StartContext(ctx context.Context) context.Context {
	tx := r.db.MustBeginTx(ctx, nil)
	return context.WithValue(ctx, authDriverName, tx)
}

func (r *SQLRepository) DeferErrorContext(ctx context.Context, err error) {
	tx := r.getTxFromContext(ctx)
	if tx != nil {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}
}

func (r *SQLRepository) getTxFromContext(ctx context.Context) *sqlx.Tx {
	txVal := ctx.Value(authDriverName)
	tx, ok := txVal.(*sqlx.Tx)
	if !ok {
		return nil
	}
	return tx
}

// Credential operations

func (r *SQLRepository) CreateCredential(ctx context.Context, cred *domain.Credential) error {
	query := `INSERT INTO auth_credentials 
		(id, user_id, username, email, password_hash, is_active, created_at) 
		VALUES (:id, :user_id, :username, :email, :password_hash, :is_active, :created_at)`

	tx := r.getTxFromContext(ctx)
	if cred.ID == "" {
		cred.ID = uuid.NewString()
	}
	cred.CreatedAt = time.Now().UTC()
	cred.IsActive = true

	if tx != nil {
		_, err := tx.NamedExec(query, cred)
		return err
	}
	_, err := r.db.NamedExec(query, cred)
	return err
}

func (r *SQLRepository) GetCredentialByUsername(ctx context.Context, username string) (*domain.Credential, error) {
	var cred domain.Credential
	tx := r.getTxFromContext(ctx)
	query := `SELECT id, user_id, username, email, password_hash, is_active, last_login_at, created_at, updated_at, deleted_at 
		FROM auth_credentials WHERE username = $1 AND deleted_at IS NULL`

	if tx != nil {
		if err := tx.Get(&cred, query, username); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&cred, query, username); err != nil {
			return nil, err
		}
	}
	return &cred, nil
}

func (r *SQLRepository) GetCredentialByEmail(ctx context.Context, email string) (*domain.Credential, error) {
	var cred domain.Credential
	tx := r.getTxFromContext(ctx)
	query := `SELECT id, user_id, username, email, password_hash, is_active, last_login_at, created_at, updated_at, deleted_at 
		FROM auth_credentials WHERE email = $1 AND deleted_at IS NULL`

	if tx != nil {
		if err := tx.Get(&cred, query, email); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&cred, query, email); err != nil {
			return nil, err
		}
	}
	return &cred, nil
}

func (r *SQLRepository) GetCredentialByUserID(ctx context.Context, userID string) (*domain.Credential, error) {
	var cred domain.Credential
	tx := r.getTxFromContext(ctx)
	query := `SELECT id, user_id, username, email, password_hash, is_active, last_login_at, created_at, updated_at, deleted_at 
		FROM auth_credentials WHERE user_id = $1 AND deleted_at IS NULL`

	if tx != nil {
		if err := tx.Get(&cred, query, userID); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&cred, query, userID); err != nil {
			return nil, err
		}
	}
	return &cred, nil
}

func (r *SQLRepository) UpdateCredential(ctx context.Context, cred *domain.Credential) error {
	now := time.Now().UTC()
	cred.UpdatedAt = &now
	tx := r.getTxFromContext(ctx)
	query := `UPDATE auth_credentials SET 
		username = :username, email = :email, is_active = :is_active, updated_at = :updated_at 
		WHERE id = :id AND deleted_at IS NULL`

	if tx != nil {
		_, err := tx.NamedExec(query, cred)
		return err
	}
	_, err := r.db.NamedExec(query, cred)
	return err
}

func (r *SQLRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	now := time.Now().UTC()
	tx := r.getTxFromContext(ctx)
	query := `UPDATE auth_credentials SET password_hash = $1, updated_at = $2 WHERE user_id = $3 AND deleted_at IS NULL`

	if tx != nil {
		_, err := tx.Exec(query, passwordHash, now, userID)
		return err
	}
	_, err := r.db.Exec(query, passwordHash, now, userID)
	return err
}

func (r *SQLRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	tx := r.getTxFromContext(ctx)
	query := `UPDATE auth_credentials SET last_login_at = $1, updated_at = $1 WHERE user_id = $2 AND deleted_at IS NULL`

	if tx != nil {
		_, err := tx.Exec(query, now, userID)
		return err
	}
	_, err := r.db.Exec(query, now, userID)
	return err
}

// Session operations

func (r *SQLRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO auth_sessions 
		(id, user_id, token, expires_at, created_at, user_agent, ip_address) 
		VALUES (:id, :user_id, :token, :expires_at, :created_at, :user_agent, :ip_address)`

	tx := r.getTxFromContext(ctx)
	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	session.CreatedAt = time.Now().UTC()

	if tx != nil {
		_, err := tx.NamedExec(query, session)
		return err
	}
	_, err := r.db.NamedExec(query, session)
	return err
}

func (r *SQLRepository) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	var session domain.Session
	tx := r.getTxFromContext(ctx)
	query := `SELECT id, user_id, token, expires_at, created_at, updated_at, revoked_at, user_agent, ip_address 
		FROM auth_sessions WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()`

	if tx != nil {
		if err := tx.Get(&session, query, token); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&session, query, token); err != nil {
			return nil, err
		}
	}
	return &session, nil
}

func (r *SQLRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	tx := r.getTxFromContext(ctx)
	query := `SELECT id, user_id, token, expires_at, created_at, updated_at, revoked_at, user_agent, ip_address 
		FROM auth_sessions WHERE id = $1`

	if tx != nil {
		if err := tx.Get(&session, query, id); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Get(&session, query, id); err != nil {
			return nil, err
		}
	}
	return &session, nil
}

func (r *SQLRepository) GetSessionsByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	var sessions []domain.Session
	tx := r.getTxFromContext(ctx)
	query := `SELECT id, user_id, token, expires_at, created_at, updated_at, revoked_at, user_agent, ip_address 
		FROM auth_sessions WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW() ORDER BY created_at DESC`

	if tx != nil {
		if err := tx.Select(&sessions, query, userID); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Select(&sessions, query, userID); err != nil {
			return nil, err
		}
	}
	return sessions, nil
}

func (r *SQLRepository) RevokeSession(ctx context.Context, sessionID string) error {
	now := time.Now().UTC()
	tx := r.getTxFromContext(ctx)
	query := `UPDATE auth_sessions SET revoked_at = $1, updated_at = $1 WHERE id = $2`

	if tx != nil {
		_, err := tx.Exec(query, now, sessionID)
		return err
	}
	_, err := r.db.Exec(query, now, sessionID)
	return err
}

func (r *SQLRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	tx := r.getTxFromContext(ctx)
	query := `UPDATE auth_sessions SET revoked_at = $1, updated_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`

	if tx != nil {
		_, err := tx.Exec(query, now, userID)
		return err
	}
	_, err := r.db.Exec(query, now, userID)
	return err
}

func (r *SQLRepository) DeleteExpiredSessions(ctx context.Context) error {
	tx := r.getTxFromContext(ctx)
	query := `DELETE FROM auth_sessions WHERE expires_at < NOW()`

	if tx != nil {
		_, err := tx.Exec(query)
		return err
	}
	_, err := r.db.Exec(query)
	return err
}
