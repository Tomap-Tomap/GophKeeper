package storage

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	updateAtField = "updateat"
	loginField    = "login"
	nameField     = "name"
	metaField     = "meta"
	userIDField   = "user_id"
	passwordField = "password"
)

// User прдеставлявет собой структуру данных о пользователе
type User struct {
	ID       string
	Login    string
	Password string
	Salt     string
}

// ScanRow необходим для реализации интерфейса pgx.RowScanner
func (u *User) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "id":
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			u.ID = id.(string)
		case loginField:
			u.Login = values[i].(string)
		case passwordField:
			u.Password = strings.TrimSpace(values[i].(string))
		case "salt":
			u.Salt = strings.TrimSpace(values[i].(string))
		}
	}

	return nil
}

// Password представляет собой структуру данных о сохраненном пароле пользователя
type Password struct {
	ID       string
	UserID   string
	Name     string
	Login    string
	Password string
	Meta     string
	UpdateAt time.Time
}

// ScanRow необходим для реализации интерфейса pgx.RowScanner
func (p *Password) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "id":
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			p.ID = id.(string)
		case userIDField:
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			p.UserID = id.(string)
		case nameField:
			p.Name = values[i].(string)
		case loginField:
			p.Login = values[i].(string)
		case passwordField:
			p.Password = values[i].(string)
		case metaField:
			p.Meta = values[i].(string)
		case updateAtField:
			p.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}

// File представляет собой структуру данных о сохраненном файле пользователя
type File struct {
	ID         string
	UserID     string
	Name       string
	PathToFile string
	Meta       string
	UpdateAt   time.Time
}

// ScanRow необходим для реализации интерфейса pgx.RowScanner
func (f *File) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "id":
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			f.ID = id.(string)
		case userIDField:
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			f.UserID = id.(string)
		case nameField:
			f.Name = values[i].(string)
		case "pathtofile":
			f.PathToFile = values[i].(string)
		case metaField:
			f.Meta = values[i].(string)
		case updateAtField:
			f.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}

// Bank представляет собой структуру данных о сохраненных банковских данных пользователя
type Bank struct {
	ID        string
	UserID    string
	Name      string
	BanksData string
	Meta      string
	UpdateAt  time.Time
}

// ScanRow необходим для реализации интерфейса pgx.RowScanner
func (f *Bank) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "id":
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			f.ID = id.(string)
		case userIDField:
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			f.UserID = id.(string)
		case nameField:
			f.Name = values[i].(string)
		case "banksdata":
			f.BanksData = values[i].(string)
		case metaField:
			f.Meta = values[i].(string)
		case updateAtField:
			f.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}

// Text представляет собой структуру данных о сохраненных текстовых данных
type Text struct {
	ID       string
	UserID   string
	Name     string
	Text     string
	Meta     string
	UpdateAt time.Time
}

// ScanRow необходим для реализации интерфейса pgx.RowScanner
func (f *Text) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "id":
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			f.ID = id.(string)
		case userIDField:
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			f.UserID = id.(string)
		case nameField:
			f.Name = values[i].(string)
		case "text":
			f.Text = values[i].(string)
		case metaField:
			f.Meta = values[i].(string)
		case updateAtField:
			f.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}
