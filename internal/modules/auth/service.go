package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	secret []byte
	db     *sqlx.DB
}

func NewService(secret string, db *sqlx.DB) *Service {
	return &Service{secret: []byte(secret), db: db}
}

type User struct {
	ID              string  `db:"id" json:"id"`
	Email           string  `db:"email" json:"email"`
	PasswordHash    string  `db:"password_hash" json:"-"`
	Active          bool    `db:"active" json:"active"`
	ActivationToken *string `db:"activation_token" json:"-"`
	ResetToken      *string `db:"reset_token" json:"-"`
}

func (s *Service) Register(email, password string) (*User, error) {
	id := uuid.NewString()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	token := uuid.NewString()
	u := &User{ID: id, Email: email, PasswordHash: string(hash), Active: false, ActivationToken: &token}
	_, err := s.db.NamedExec(`INSERT INTO users (id,email,password_hash,active,activation_token) VALUES (:id,:email,:password_hash,:active,:activation_token)`, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Activate(token string) error {
	res, err := s.db.Exec(`UPDATE users SET active=true, activation_token=NULL WHERE activation_token=$1`, token)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("invalid token")
	}
	return nil
}

func (s *Service) Authenticate(email, password string) (*User, error) {
	var u User
	if err := s.db.Get(&u, "SELECT id,email,password_hash,active FROM users WHERE email=$1", email); err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}
	if !u.Active {
		return nil, errors.New("account not active")
	}
	return &u, nil
}

func (s *Service) ForgotPassword(email string) (string, error) {
	token := uuid.NewString()
	res, err := s.db.Exec(`UPDATE users SET reset_token=$1 WHERE email=$2`, token, email)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", errors.New("email not found")
	}
	return token, nil
}

func (s *Service) ResetPassword(token, newPassword string) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	res, err := s.db.Exec(`UPDATE users SET password_hash=$1, reset_token=NULL WHERE reset_token=$2`, string(hash), token)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("invalid token")
	}
	return nil
}

func (s *Service) GenerateToken(userID string, exp time.Duration) (string, error) {
	claims := jwt.MapClaims{"sub": userID, "exp": time.Now().Add(exp).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) ParseToken(tok string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) { return s.secret, nil })
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
