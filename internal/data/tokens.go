package data

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)

const ScopeActivation = "activation"
const ScopeAuthentication = "authentication"

type contextKey string

const UserKey = contextKey("user")

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) GenerateToken(userID int64, duration time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(duration),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func (m *TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`
	_, err := m.DB.Exec(query, token.Hash, token.UserID, token.Expiry, token.Scope)
	return err
}

func (m *TokenModel) GetForToken(scope, plaintext string) (*User, error) {
	hash := sha256.Sum256([]byte(plaintext))
	query := `
		SELECT users.user_id, users.email, users.password, users.role, users.activated, users.created_at
		FROM tokens
		INNER JOIN users ON tokens.user_id = users.user_id
		WHERE tokens.hash = $1 AND tokens.scope = $2 AND tokens.expiry > $3`
	row := m.DB.QueryRow(query, hash[:], scope, time.Now())

	var user User
	err := row.Scan(&user.UserID, &user.Email, &user.Password, &user.Role, &user.Activated, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `DELETE FROM tokens WHERE scope = $1 AND user_id = $2`
	_, err := m.DB.Exec(query, scope, userID)
	return err
}
