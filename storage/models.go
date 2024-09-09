package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	fieldID         = "id"
	fieldSalt       = "salt"
	fieldUpdateAt   = "updateat"
	fieldLogin      = "login"
	fieldDataName   = "name"
	fieldMeta       = "meta"
	fieldUserID     = "user_id"
	fieldPassword   = "password"
	fieldPathToFile = "pathtofile"
	fieldCardNumber = "cardnumber"
	fieldCVC        = "cvc"
	fieldOwner      = "owner"
	fieldExp        = "exp"
)

// User represents a user data structure.
type User struct {
	ID       string
	Login    string
	Password string
	Salt     string
}

// ScanRow scans the user data from the provided rows.
func (u *User) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		fieldName := strings.ToLower(rows.FieldDescriptions()[i].Name)
		switch fieldName {
		case fieldID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			u.ID = id
		case fieldLogin:
			u.Login = values[i].(string)
		case fieldPassword:
			u.Password = strings.TrimSpace(values[i].(string))
		case fieldSalt:
			u.Salt = strings.TrimSpace(values[i].(string))
		}
	}

	return nil
}

// Password represents a password data structure.
type Password struct {
	ID       string
	UserID   string
	Name     string
	Login    string
	Password string
	Meta     string
	UpdateAt time.Time
}

// ScanRow scans the password data from the provided rows.
func (p *Password) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		fieldName := strings.ToLower(rows.FieldDescriptions()[i].Name)
		switch fieldName {
		case fieldID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}
			p.ID = id
		case fieldUserID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			p.UserID = id
		case fieldDataName:
			p.Name = values[i].(string)
		case fieldLogin:
			p.Login = values[i].(string)
		case fieldPassword:
			p.Password = values[i].(string)
		case fieldMeta:
			p.Meta = values[i].(string)
		case fieldUpdateAt:
			p.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}

// File represents a file data structure.
type File struct {
	ID         string
	UserID     string
	Name       string
	PathToFile string
	Meta       string
	UpdateAt   time.Time
}

// ScanRow scans the file data from the provided rows.
func (f *File) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		fieldName := strings.ToLower(rows.FieldDescriptions()[i].Name)
		switch fieldName {
		case fieldID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			f.ID = id
		case fieldUserID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			f.UserID = id
		case fieldDataName:
			f.Name = values[i].(string)
		case fieldPathToFile:
			f.PathToFile = values[i].(string)
		case fieldMeta:
			f.Meta = values[i].(string)
		case fieldUpdateAt:
			f.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}

// Bank represents a card data structure.
type Bank struct {
	ID         string
	UserID     string
	Name       string
	CardNumber string
	CVC        string
	Owner      string
	Meta       string
	Exp        string
	UpdateAt   time.Time
}

// ScanRow scans the bank data from the provided rows.
func (f *Bank) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		fieldName := strings.ToLower(rows.FieldDescriptions()[i].Name)
		switch fieldName {
		case fieldID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			f.ID = id
		case fieldUserID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			f.UserID = id
		case fieldDataName:
			f.Name = values[i].(string)
		case fieldCardNumber:
			f.CardNumber = values[i].(string)
		case fieldCVC:
			f.CVC = values[i].(string)
		case fieldOwner:
			f.Owner = values[i].(string)
		case fieldMeta:
			f.Meta = values[i].(string)
		case fieldExp:
			f.Exp = values[i].(string)
		case fieldUpdateAt:
			f.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}

// Text represents a text data structure.
type Text struct {
	ID       string
	UserID   string
	Name     string
	Text     string
	Meta     string
	UpdateAt time.Time
}

// ScanRow scans the text data from the provided rows.
func (f *Text) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		fieldName := strings.ToLower(rows.FieldDescriptions()[i].Name)
		switch fieldName {
		case fieldID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			f.ID = id
		case fieldUserID:
			id, err := convertUUIDToString(values[i])

			if err != nil {
				return err
			}

			f.UserID = id
		case fieldDataName:
			f.Name = values[i].(string)
		case "text":
			f.Text = values[i].(string)
		case fieldMeta:
			f.Meta = values[i].(string)
		case fieldUpdateAt:
			f.UpdateAt = values[i].(time.Time)
		}
	}

	return nil
}

func convertUUIDToString(value any) (string, error) {
	v, ok := value.([16]byte)

	if !ok {
		return "", fmt.Errorf("cannot convert value to byte")
	}

	uuid := pgtype.UUID{
		Bytes: v,
		Valid: true,
	}
	id, err := uuid.Value()

	if err != nil {
		return "", err
	}

	result, ok := id.(string)

	if !ok {
		return "", fmt.Errorf("cannot convert driver.Value to string")
	}

	return result, nil
}
