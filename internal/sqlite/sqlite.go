package sqlite

import (
	"botTelegram/internal/config"
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db     *sql.DB
	Admins map[int]Admin
}
type Admin struct {
	ID  int64
	FIO string
}

var (
	Admins = make(map[int]Admin)
)

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Init(ctx context.Context, a config.Admin) error {
	q := `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, idUserTeleg TEXT, fio TEXT NOT NULL, branch TEXT NOT NULL, unit TEXT NOT NULL, phone TEXT NOT NULL)`
	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create users table: %w", err)
	}

	q = `CREATE TABLE IF NOT EXISTS admins (id INTEGER, date INTEGER NOT NULL, creater INTEGER, FOREIGN KEY(id) REFERENCES users(id), FOREIGN KEY(creater) REFERENCES users(id))`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create admins table: %w", err)
	}

	q = `CREATE TABLE IF NOT EXISTS logs (id INTEGER PRIMARY KEY AUTOINCREMENT, date INTEGER, events TEXT)`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create admins table: %w", err)
	}

	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, a.FIO)
	if err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	if exist {
		return nil
	}

	q = `INSERT INTO users (idUserTeleg, fio, branch, unit, phone) VALUES (?,?,?,?,?)`
	if _, err := s.db.ExecContext(ctx, q, a.ID, a.FIO, a.Branch, a.Unit, a.Phone); err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	q = `INSERT INTO admins (id, date, creater) VALUES (?,?,?)`
	if _, err := s.db.ExecContext(ctx, q, 1, time.Now().Unix(), 1); err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}

	return nil

}
func (s *Storage) IsExist(ctx context.Context, q string, args ...interface{}) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		return false, fmt.Errorf("can't check exist: %w", err)
	}
	return count > 0, nil
}

func (s *Storage) Add(ctx context.Context, fio, branch, unit, phone string) error {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	if exist {
		return fmt.Errorf("can't add user: user already exists")
	}
	q := `INSERT INTO users (fio, branch, unit, phone) VALUES (?,?,?,?)`
	if _, err := s.db.ExecContext(ctx, q, fio, branch, unit, phone); err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	return nil
}

func (s *Storage) GetAdmins(ctx context.Context) (err error) {
	s.Admins = make(map[int]Admin)

	q := `SELECT idUserTeleg, fio FROM users,admins where users.id = admins.id`
	rows, err := s.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var fio string
		var id string
		err = rows.Scan(&id, &fio)
		if err != nil {
			return err
		}
		idUserTeleg, _ := strconv.ParseInt(id, 10, 64)
		s.Admins[len(s.Admins)] = Admin{ID: idUserTeleg, FIO: fio}
	}
	return nil
}
func (s *Storage) ShowAdmins(ctx context.Context) (str string, err error) {
	if err := s.GetAdmins(ctx); err != nil {
		return "", fmt.Errorf("can't get admins: %w", err)
	}

	for _, v := range s.Admins {
		str = str + fmt.Sprintf("%s\n", v.FIO)
	}
	return str, nil
}
func (s *Storage) Show(ctx context.Context) (str string, err error) {
	q := `SELECT fio, branch, unit, phone FROM users`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return "", fmt.Errorf("can't get users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var fio, branch, unit, phone string
		err = rows.Scan(&fio, &branch, &unit, &phone)
		if err != nil {
			return "", fmt.Errorf("can't get users: %w", err)
		}
		str = str + fmt.Sprintf("%s, %s, %s, %s\n", fio, branch, unit, phone)
	}
	return str, nil
}
func (s *Storage) Delete(ctx context.Context, fio string) error {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM admins,users where admins.id = users.id and users.fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("can't delete user: %w", err)
	}
	if exist {
		return fmt.Errorf("Не возможно удалить пользователя с админской ролью")
	}
	q := `DELETE FROM users WHERE fio = ?`
	if _, err := s.db.ExecContext(ctx, q, fio); err != nil {
		return fmt.Errorf("can't delete user: %w", err)
	}
	return nil
}
func (s *Storage) AddAdmin(ctx context.Context, fio string, idTeleg int64) error {

	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM admins,users where admins.id = users.id and users.fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	if exist {
		return fmt.Errorf("can't add user: user already exists")
	}

	exist, err = s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	if !exist {
		return fmt.Errorf("can't add user: user doesn't exist")
	}

	q := `INSERT INTO admins (id, date, idUserTeleg) VALUES ((select id from users where fio = ?),?,(Select id from users where idUserTeleg = ?))`

	if _, err := s.db.ExecContext(ctx, q, 1, time.Now().Unix(), strconv.FormatInt(idTeleg, 10)); err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	return nil
}
func (s *Storage) DelAdmin(ctx context.Context, fio string) error {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM admins,users where admins.id = users.id and users.fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("can't delete user: %w", err)
	}
	if !exist {
		return fmt.Errorf("can't delete user: user doesn't exist")
	}
	q := `Delete from admins,users where admins.id = users.id and users.fio = ?`

	if _, err := s.db.ExecContext(ctx, q, fio); err != nil {
		return fmt.Errorf("can't delete user: %w", err)
	}
	return nil
}
