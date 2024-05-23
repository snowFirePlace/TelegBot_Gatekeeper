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
	q := `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, idUserTeleg TEXT,username TEXT, fio TEXT NOT NULL, branch TEXT NOT NULL, unit TEXT NOT NULL, phone TEXT NOT NULL)`
	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create users table: %w", err)
	}

	q = `CREATE TABLE IF NOT EXISTS admins (id INTEGER, date INTEGER NOT NULL, creater INTEGER, FOREIGN KEY(id) REFERENCES users(id), FOREIGN KEY(creater) REFERENCES users(id))`
	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create admins table: %w", err)
	}

	q = `CREATE TABLE IF NOT EXISTS logs (id INTEGER PRIMARY KEY AUTOINCREMENT, date INTEGER, event TEXT)`
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

func (s *Storage) AddUser(ctx context.Context, idUser int64, fio, branch, unit, phone string) error {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, fio)

	if err != nil {
		return fmt.Errorf("Error in Add User: %w", err)
	}
	if exist {
		return fmt.Errorf("Пользователь уже добавлен")
	}
	q := `INSERT INTO users (fio, branch, unit, phone) VALUES (?,?,?,?)`
	if _, err := s.db.ExecContext(ctx, q, fio, branch, unit, phone); err != nil {
		return fmt.Errorf("Error in Add User: %w", err)
	}
	if err := s.log(ctx, fmt.Sprintf("User %d add new user: %s %s %s %s", idUser, fio, branch, unit, phone)); err != nil {
		return err
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
		return "", fmt.Errorf("Error in show admins: %w", err)
	}

	for _, v := range s.Admins {
		str = str + fmt.Sprintf("%s\n", v.FIO)
	}
	return str, nil
}
func (s *Storage) ShowUsers(ctx context.Context) (str string, err error) {
	q := `SELECT fio, branch, unit, phone FROM users`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return "", fmt.Errorf("Error in show users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var fio, branch, unit, phone string
		err = rows.Scan(&fio, &branch, &unit, &phone)
		if err != nil {
			return "", fmt.Errorf("Error in show users: %w", err)
		}
		str = str + fmt.Sprintf("%s, %s, %s, %s\n", fio, branch, unit, phone)
	}
	return str, nil
}
func (s *Storage) DelUser(ctx context.Context, idUser int64, fio string) (idDelUser int64, err error) {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM admins,users where admins.id = users.id and users.fio = ?`, fio)
	if err != nil {
		return 0, fmt.Errorf("Error in delete user: %w", err)
	}
	if exist {
		return 0, fmt.Errorf("Не возможно удалить пользователя с админской ролью")
	}
	q := `Select idUserTeleg from users where fio = ?`
	if err := s.db.QueryRowContext(ctx, q, fio).Scan(&idDelUser); err != nil {
		return 0, fmt.Errorf("Error in delete user: %w", err)
	}

	q = `DELETE FROM users WHERE fio = ?`
	if _, err := s.db.ExecContext(ctx, q, fio); err != nil {
		return 0, fmt.Errorf("Error in delete user: %w", err)
	}

	if err := s.log(ctx, fmt.Sprintf("User %d deleted: %s", idUser, fio)); err != nil {
		return 0, err
	}
	return idDelUser, nil
}
func (s *Storage) AddAdmin(ctx context.Context, idUser int64, fio string, idTeleg int64) error {

	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM admins,users where admins.id = users.id and users.fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("Error in add admin: %w", err)
	}
	if exist {
		return fmt.Errorf("Пользователь уже имеет админскую роль")
	}

	exist, err = s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("Error in add admin: %w", err)
	}
	if !exist {
		return fmt.Errorf("Пользователь не найден")
	}

	q := `INSERT INTO admins (id, date, idUserTeleg) VALUES ((select id from users where fio = ?),?,(Select id from users where idUserTeleg = ?))`

	if _, err := s.db.ExecContext(ctx, q, 1, time.Now().Unix(), strconv.FormatInt(idTeleg, 10)); err != nil {
		return fmt.Errorf("Error in add admin: %w", err)
	}

	if err := s.log(ctx, fmt.Sprintf("User %s add admin: %s", fio, strconv.FormatInt(idTeleg, 10))); err != nil {
		return err
	}
	return nil
}
func (s *Storage) DelAdmin(ctx context.Context, idUser int64, fio string) error {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM admins,users where admins.id = users.id and users.fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("Error in delete admin: %w", err)
	}
	if !exist {
		return fmt.Errorf("Пользователь не найден")
	}
	q := `Delete from admins,users where admins.id = users.id and users.fio = ?`

	if _, err := s.db.ExecContext(ctx, q, fio); err != nil {
		return fmt.Errorf("Error in delete admin: %w", err)
	}
	if err := s.log(ctx, fmt.Sprintf("User %d delete admin: %s", idUser, fio)); err != nil {
		return err
	}

	return nil
}
func (s *Storage) Registration(ctx context.Context, fio, phone string, id int64, username string) error {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, fio)
	if err != nil {
		return fmt.Errorf("Error in Registration: %w", err)
	}
	if !exist {
		return fmt.Errorf("Доступа запрещен")
	}
	exist, err = s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ? and idUserTeleg = ?`, fio, id)
	if err != nil {
		return fmt.Errorf("Error in Registration: %w", err)
	}
	if exist {
		return fmt.Errorf("Пользователь уже зарегистрирован")
	}
	exist, err = s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ? and phone = ?`, fio, phone)
	if err != nil {
		return fmt.Errorf("Error in Registration: %w", err)
		// return fmt.Errorf("can't add user: %w", err)
	}
	if exist {
		q := `Update users set idUserTeleg = ?, username = ? where fio = ? and phone = ?`
		if _, err := s.db.ExecContext(ctx, q, id, username, fio, phone); err != nil {
			return fmt.Errorf("can't update user data: %w", err)
		}
		if err := s.log(ctx, fmt.Sprintf("User %d updated: %s", id, fio)); err != nil {
			return err
		}
		return nil
	}
	q := `select phone from users where fio = ?`

	rows, err := s.db.QueryContext(ctx, q, fio)
	if err != nil {
		return fmt.Errorf("Error in Registration: %w", err)
	}
	defer rows.Close()
	var p string

	for rows.Next() {
		if err := rows.Scan(&p); err != nil {
			return fmt.Errorf("Error in Registration: %w", err)
		}
	}
	return fmt.Errorf("Для регистрации для пользователя %s указан номер телефона: %s\nПовторите попытку, указав верный номер телефона.", fio, hidePhone(p))
}
func (s *Storage) log(ctx context.Context, event string) error {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO logs (date, event) VALUES (?, ?)`, time.Now().Unix(), event); err != nil {
		return fmt.Errorf("Error in log: %w", err)
	}
	return nil
}

// Todo Move this func to another package
func hidePhone(phone string) string {
	return fmt.Sprintf("+7********%s", phone[len(phone)-2:])
}
