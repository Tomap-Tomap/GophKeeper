package storage

const (
	queryInsertUser = `
	INSERT INTO users (login, password)
	VALUES ($1, $2)
	RETURNING id, login, password;
	`

	queryInsertSalt = `
	INSERT INTO salts (login, salt)
	VALUES ($1, $2)
	RETURNING salt;
	`

	querySelectUser = `
	SELECT u.id, u.login, u.password, s.salt
	FROM users u, salts s
	WHERE u.login = $1 AND s.login = $2;
	`
)

const (
	queryInsertPassword = `
	INSERT INTO passwords (user_id, name, login, password, meta)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *;
	`

	queryUpdatePassword = `
	UPDATE passwords
	SET user_id = $1,
		name = $2,
		login = $3,
		password = $4,
		meta = $5
	WHERE id = $6
	RETURNING *;
	`

	querySelectPassword = `
	SELECT *
	FROM passwords
	WHERE id = $1 AND user_id = $2;
	`

	querySelectPasswords = `
	SELECT *
	FROM passwords
	WHERE user_id = $1;
	`

	queryDeletePassword = `
	DELETE FROM passwords WHERE id = $1 AND user_id = $2 RETURNING *;
	`
)

const (
	queryInsertFile = `
	INSERT INTO files (user_id, name, pathtofile, meta) VALUES ($1, $2, $3, $4)
	RETURNING *
	`

	queryUpdateFile = `
		WITH t AS (
			UPDATE files
			SET user_id = $1,
				name = $2,
				pathtofile = $3,
				meta = $4
			WHERE id = $5
			RETURNING *
		)
		SELECT t.id, t.user_id, t.name, files.pathtofile, t.meta, t.updateat FROM t
		INNER JOIN files ON t.id = files.id;
	`

	querySelectFile = `
	SELECT *
	FROM files
	WHERE id = $1 AND user_id = $2;
	`

	querySelectFiles = `
	SELECT *
	FROM files
	WHERE user_id = $1;
	`

	queryDeleteFile = `
	DELETE FROM files WHERE id = $1 AND user_id = $2 RETURNING *;
	`
)

const (
	queryInsertBank = `
	INSERT INTO banks (user_id, name, cardnumber, exp, cvc, owner, meta)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING *
	`

	queryUpdateBank = `
	UPDATE banks
	SET user_id = $1,
		name = $2,
		cardnumber = $3,
		exp = $4,
		cvc = $5,
		owner = $6,
		meta = $7
	WHERE id = $8
	RETURNING *
	`

	querySelectBank = `
	SELECT *
	FROM banks
	WHERE id = $1 AND user_id = $2;
	`

	querySelectBanks = `
	SELECT *
	FROM banks
	WHERE user_id = $1;
	`

	queryDeleteBank = `
	DELETE FROM banks WHERE id = $1 AND user_id = $2 RETURNING *;
	`
)

const (
	queryInsertText = `
	INSERT INTO texts (user_id, name, text, meta) VALUES ($1, $2, $3, $4)
	RETURNING *
	`

	queryUpdateText = `
	UPDATE texts
	SET user_id = $1,
		name = $2,
		text = $3,
		meta = $4
	WHERE id = $5
	RETURNING *
	`

	querySelectText = `
	SELECT *
	FROM texts
	WHERE id = $1 AND user_id = $2;
	`

	querySelectTexts = `
	SELECT *
	FROM texts
	WHERE user_id = $1;
	`

	queryDeleteText = `
	DELETE FROM texts WHERE id = $1 AND user_id = $2 RETURNING *;
	`
)
