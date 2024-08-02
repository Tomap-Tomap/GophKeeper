-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION "uuid-ossp";

CREATE FUNCTION gen_uuid() RETURNS trigger AS $gen_uuid$
BEGIN
    NEW.id := uuid_generate_v4();
    RETURN NEW;
END;
$gen_uuid$ LANGUAGE plpgsql;

CREATE FUNCTION gen_update_at() RETURNS trigger AS $gen_update_at$
BEGIN
    NEW.updateAt := current_timestamp;
    RETURN NEW;
END; 
$gen_update_at$ LANGUAGE plpgsql;

CREATE TABLE users (
    id UUID NOT NULL,
    login VARCHAR(150) UNIQUE NOT NULL CHECK (login <> ''),
    password CHAR(64) NOT NULL CHECK (password <> ''),
    updateAt TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY(id)
);
CREATE TRIGGER gen_uuid BEFORE INSERT ON users
    FOR EACH ROW EXECUTE PROCEDURE gen_uuid();
CREATE TRIGGER gen_update_at BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE gen_update_at();    

CREATE TABLE salts (
    login CHAR(64) UNIQUE NOT NULL CHECK (login <> ''),
    salt VARCHAR(150) NOT NULL CHECK (salt <> '')
);
CREATE INDEX salts_login_idx ON salts (login);

CREATE TABLE passwords (
    id UUID NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    name VARCHAR(150),
    login VARCHAR(150),
    password VARCHAR(150),
    meta TEXT,
    updateAt TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY(id)
);
CREATE TRIGGER gen_uuid BEFORE INSERT ON passwords
    FOR EACH ROW EXECUTE PROCEDURE gen_uuid();
CREATE TRIGGER gen_update_at BEFORE INSERT OR UPDATE ON passwords
    FOR EACH ROW EXECUTE PROCEDURE gen_update_at(); 

CREATE TABLE files (
    id UUID NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    name VARCHAR(150),
    pathToFile VARCHAR(150),
    meta TEXT,
    updateAt TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY(id)
);
CREATE TRIGGER gen_uuid BEFORE INSERT ON files
    FOR EACH ROW EXECUTE PROCEDURE gen_uuid();
CREATE TRIGGER gen_update_at BEFORE INSERT OR UPDATE ON files
    FOR EACH ROW EXECUTE PROCEDURE gen_update_at(); 

CREATE TABLE banks (
    id UUID NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    name VARCHAR(150),
    banksData VARCHAR(150),
    meta TEXT,
    updateAt TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY(id)
);
CREATE TRIGGER gen_uuid BEFORE INSERT ON banks
    FOR EACH ROW EXECUTE PROCEDURE gen_uuid();
CREATE TRIGGER gen_update_at BEFORE INSERT OR UPDATE ON banks
    FOR EACH ROW EXECUTE PROCEDURE gen_update_at();   

CREATE TABLE texts (
    id UUID NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    name VARCHAR(150),
    text TEXT,
    meta TEXT,
    updateAt TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY(id)
);
CREATE TRIGGER gen_uuid BEFORE INSERT ON texts
    FOR EACH ROW EXECUTE PROCEDURE gen_uuid();
CREATE TRIGGER gen_update_at BEFORE INSERT OR UPDATE ON texts
    FOR EACH ROW EXECUTE PROCEDURE gen_update_at();         

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX salts_login_idx;

DROP TRIGGER gen_uuid ON users;
DROP TRIGGER gen_update_at ON users;
DROP TRIGGER gen_uuid ON passwords;
DROP TRIGGER gen_update_at ON passwords;
DROP TRIGGER gen_uuid ON files;
DROP TRIGGER gen_update_at ON files;
DROP TRIGGER gen_uuid ON banks;
DROP TRIGGER gen_update_at ON banks;
DROP TRIGGER gen_uuid ON texts;
DROP TRIGGER gen_update_at ON texts;

DROP FUNCTION gen_uuid;
DROP FUNCTION gen_update_at;

DROP TABLE passwords;
DROP TABLE files;
DROP TABLE banks;
DROP TABLE texts;
DROP TABLE users;
DROP TABLE salts;

DROP EXTENSION "uuid-ossp";
-- +goose StatementEnd
