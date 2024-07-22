package storage

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		case "login":
			u.Login = values[i].(string)
		case "password":
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
		case "user_id":
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			p.UserID = id.(string)
		case "name":
			p.Name = values[i].(string)
		case "login":
			p.Login = values[i].(string)
		case "password":
			p.Password = values[i].(string)
		case "meta":
			p.Meta = values[i].(string)
		case "updateat":
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
		case "user_id":
			uuid := pgtype.UUID{
				Bytes: values[i].([16]byte),
				Valid: true,
			}
			id, err := uuid.Value()

			if err != nil {
				return err
			}

			f.UserID = id.(string)
		case "name":
			f.Name = values[i].(string)
		case "pathtofile":
			f.PathToFile = values[i].(string)
		case "meta":
			f.Meta = values[i].(string)
		case "updateat":
			f.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}
