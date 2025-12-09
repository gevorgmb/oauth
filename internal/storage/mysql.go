package storage

import (
	"database/sql"
	"errors"
	"log"
	"oauth/internal/dto"
	"oauth/internal/entity"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct {
	db *sql.DB
}

func NewMySQL(conData *dto.MySQLConnectionDto) (*MySQL, error) {
	db, err := sql.Open("mysql", conData.Dns)
	if err != nil {
		return nil, err
	}

	//ToDo: Move constants to ENV file
	db.SetConnMaxLifetime(time.Duration(conData.ConnectionMaxLifetime) * time.Minute)
	db.SetMaxOpenConns(conData.MaxOpenConnections)
	db.SetMaxIdleConns(conData.MaxIdleConnections)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}
	return &MySQL{db: db}, nil
}

func (s *MySQL) Close() error { return s.db.Close() }

func migrate(db *sql.DB) error {
	_, err := db.Exec(`DROP TABLE IF EXISTS users;`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		full_name VARCHAR(255),
		phone VARCHAR(50),
		birthday DATETIME,
		role varchar(50) NOT NULL DEFAULT 'user',
		password_hash VARCHAR(255) NOT NULL,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	return err
}

func (s *MySQL) AddUser(user *entity.User) error {
	dateLayout := "2006-01-02"
	_, err := s.db.Exec(
		`INSERT INTO users (email, full_name, password_hash, phone, birthday) VALUES (?, ?, ?, ?, ?)`,
		user.Email,
		user.FullName,
		user.PasswordHash,
		user.Phone,
		user.Birthday.Format(dateLayout),
	)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (s *MySQL) GetUser(email string) (*entity.User, error) {
	row := s.db.QueryRow(`SELECT email, full_name, password_hash, phone, birthday, role, created FROM users WHERE email = ?`, email)
	u := &entity.User{}
	if err := row.Scan(&u.Email, &u.FullName, &u.PasswordHash, &u.Phone, &u.Birthday, &u.Role, &u.Created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}
	return u, nil
}

func (s *MySQL) GetUserList(limit int64, offset int64) ([]entity.User, int64, error) {
	queryString := "SELECT id, email, full_name, phone, birthday, created, role FROM users"
	condition := ""
	//ToDo: add condition from filters
	// fetchQuery := queryString + condition + " LIMIT ? OFFSET ?"
	// fmt.Println(fetchQuery)
	rows, err := s.db.Query(
		queryString+condition+" LIMIT ? OFFSET ?",
		limit,
		offset,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.Id, &u.Email, &u.FullName, &u.Phone, &u.Birthday, &u.Created, &u.Role); err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	countRow := s.db.QueryRow("SELECT count(*) AS c FROM users " + condition)

	var count int64

	if err := countRow.Scan(&count); err != nil {
		log.Fatal(err)
	}

	return users, count, nil
}

func (s *MySQL) DeleteUser(id int64) error {
	_, err := s.db.Query("DELETE FROM users WHERE id = ?", id)
	return err
}
