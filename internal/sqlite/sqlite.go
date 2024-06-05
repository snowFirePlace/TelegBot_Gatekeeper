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

var (
	Admins = make(map[int]Admin)
)

type Storage struct {
	db     *sql.DB
	Admins map[int]Admin
}
type Admin struct {
	ID  int64
	FIO string
}

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

func (s *Storage) Init(ctx context.Context, admin config.Admin) error {
	// Create users table if it doesn't exist
	const usersTableQuery = `CREATE TABLE IF NOT EXISTS users (` +
		`id INTEGER PRIMARY KEY AUTOINCREMENT, ` +
		`idUserTeleg TEXT, ` +
		`username TEXT, ` +
		`fio TEXT NOT NULL, ` +
		`branch TEXT NOT NULL, ` +
		`unit TEXT NOT NULL, ` +
		`phone TEXT NOT NULL)`
	if _, err := s.db.ExecContext(ctx, usersTableQuery); err != nil {
		return fmt.Errorf("can't create users table: %w", err)
	}

	// Create admins table if it doesn't exist
	const adminsTableQuery = `CREATE TABLE IF NOT EXISTS admins (` +
		`id INTEGER, ` +
		`date INTEGER NOT NULL, ` +
		`creater INTEGER, ` +
		`FOREIGN KEY (id) REFERENCES users(id), ` +
		`FOREIGN KEY (creater) REFERENCES users(id))`
	if _, err := s.db.ExecContext(ctx, adminsTableQuery); err != nil {
		return fmt.Errorf("can't create admins table: %w", err)
	}

	// Create logs table if it doesn't exist
	const logsTableQuery = `CREATE TABLE IF NOT EXISTS logs (` +
		`id INTEGER PRIMARY KEY AUTOINCREMENT, ` +
		`date INTEGER, ` +
		`event TEXT)`
	if _, err := s.db.ExecContext(ctx, logsTableQuery); err != nil {
		return fmt.Errorf("can't create logs table: %w", err)
	}

	// Add admin user if it doesn't exist
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, admin.FIO)
	if err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}
	if exist {
		return nil
	}

	const addUserQuery = `INSERT INTO users (idUserTeleg, fio, branch, unit, phone) VALUES (?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, addUserQuery, admin.ID, admin.FIO, admin.Branch, admin.Unit, admin.Phone); err != nil {
		return fmt.Errorf("can't add user: %w", err)
	}

	const addAdminQuery = `INSERT INTO admins (id, date, creater) VALUES (?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, addAdminQuery, 1, time.Now().Unix(), 1); err != nil {
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
func (s *Storage) ShowUsers(ctx context.Context) (str string, err error) {
	q := `SELECT idUserTeleg, fio, username, branch, unit, phone FROM users Order by fio`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return "", fmt.Errorf("Error in show users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var fio, branch, unit, phone string
		var idUser, username sql.NullString
		err = rows.Scan(&idUser, &fio, &username, &branch, &unit, &phone)
		if err != nil {
			return "", fmt.Errorf("Error in show users: %w", err)
		}

		if username.String == "" && idUser.String == "" {
			str = str + fmt.Sprintf("%s, %s, %s, %s\n", fio, branch, unit, phone)

		} else {
			url := fmt.Sprintf("<a href=\"tg://user?id=%s\">%s</a>", idUser.String, username.String)
			str = str + fmt.Sprintf("%s, %s, %s, %s, %s\n", fio, url, branch, unit, phone)
		}
	}
	return str, nil
}
func (s *Storage) ShowUsersWithID(ctx context.Context) (a [][]string, err error) {
	q := `SELECT idTeleg, username FROM users where idUserTeleg IS NOT NULL AND username IS NOT NULL`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return a, fmt.Errorf("Error in show users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var idUser, username string
		err = rows.Scan(&idUser, &username)
		if err != nil {
			return nil, fmt.Errorf("Error in show users: %w", err)
		}
		a = append(a, []string{idUser, username})
	}
	return a, nil
}
func (s *Storage) DelUser(ctx context.Context, idUser int64, fio string) (idDelUser int64, err error) {
	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM users WHERE fio = ?`, fio)
	if err != nil {
		return 0, fmt.Errorf("Error in delete user: %w", err)
	}
	if !exist {
		return 0, fmt.Errorf("Пользователь не найден")
	}
	exist, err = s.IsExist(ctx, `SELECT COUNT(*) FROM admins, users where admins.id = users.id and users.fio = ?`, fio)
	if err != nil {
		return 0, fmt.Errorf("Error in delete user: %w", err)
	}
	if exist {
		return 0, fmt.Errorf("Не возможно удалить пользователя с админской ролью")
	}
	q := `Select idUserTeleg from users where fio = ?`
	var id sql.NullString
	if err := s.db.QueryRowContext(ctx, q, fio).Scan(&id); err != nil {
		return 0, fmt.Errorf("Error in delete user: %w", err)
	} else {
		if id.Valid {
			idDelUser, _ = strconv.ParseInt(id.String, 10, 64)
		}
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

func (s *Storage) AddAdmin(ctx context.Context, idUser int64, fio string) error {

	exist, err := s.IsExist(ctx, `SELECT COUNT(*) FROM admins, users where admins.id = users.id and users.fio = ?`, fio)
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

	q := `INSERT INTO admins (id, date, creater) VALUES ((select id from users where fio = ?),?,(Select id from users where idUserTeleg = ?))`

	if _, err := s.db.ExecContext(ctx, q, fio, time.Now().Unix(), strconv.FormatInt(idUser, 10)); err != nil {
		return fmt.Errorf("Error in add admin: %w", err)
	}

	if err := s.log(ctx, fmt.Sprintf("User %s add admin: %s", idUser, fio)); err != nil {
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
	return fmt.Errorf("Для регистрации пользователя %s указан номер телефона: %s\nПовторите попытку, указав верный номер телефона.", fio, hidePhone(p))
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
