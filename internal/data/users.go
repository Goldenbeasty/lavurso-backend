package data

import (
	"context"
	"crypto/sha256"
	"errors"
	"regexp"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdministrator = "admin"
	RoleTeacher       = "teacher"
	RoleParent        = "parent"
	RoleStudent       = "student"
)

var (
	ErrEmailAlreadyExists  = errors.New("an user with specified email already exists")
	ErrNoSuchUser          = errors.New("no such user")
	ErrNoSuchUsers         = errors.New("no such users")
	ErrNoSuchStudents      = errors.New("no such students")
	ErrNotAStudent         = errors.New("not a student")
	ErrNoSuchParentForUser = errors.New("no such parent set for child")
	ErrNotAParent          = errors.New("not a parent")
)

var EmailRegex = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

type User struct {
	ID          *int       `json:"id"`
	Name        *string    `json:"name,omitempty"`
	Email       *string    `json:"email,omitempty"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	IdCode      *int64     `json:"id_code,omitempty"`
	BirthDate   *Date      `json:"birth_date,omitempty"`
	Password    Password   `json:"-"`
	Role        *string    `json:"role,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	Active      *bool      `json:"active,omitempty"`
	Archived    *bool      `json:"archived,omitempty"`
	Class       *Class     `json:"class,omitempty"`
	Marks       []*Mark    `json:"marks,omitempty"`
}

type Password struct {
	Hashed    []byte
	Plaintext string
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m UserModel) HashPassword(plaintext string) ([]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return nil, err
	}
	return hashed, nil
}

// DATABASE

func (m UserModel) AllUsers() ([]*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.password, u.role, u.class_id, c.name, u.created_at, u.active, u.archived
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		user.BirthDate = new(Date)
		user.Class = new(Class)

		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PhoneNumber,
			&user.IdCode,
			&user.BirthDate.Time,
			&user.Password.Hashed,
			&user.Role,
			&user.Class.ID,
			&user.Class.Name,
			&user.CreatedAt,
			&user.Active,
			&user.Archived,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil

}

func (m UserModel) SearchUser(name string) ([]*User, error) {
	query := `SELECT id, name, role
	FROM users
	WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1))
	ORDER BY name ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) GetUserByID(userID int) (*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.password, u.role, u.class_id, c.name, u.created_at, u.active, u.archived
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	WHERE u.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	user.BirthDate = new(Date)
	user.Class = new(Class)

	err := m.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.IdCode,
		&user.BirthDate.Time,
		&user.Password.Hashed,
		&user.Role,
		&user.Class.ID,
		&user.Class.Name,
		&user.CreatedAt,
		&user.Active,
		&user.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetUsersByRole(role string) ([]*User, error) {
	query := `SELECT id, name, role
	FROM users
	WHERE role = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, role)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil

}

func (m UserModel) GetUserByIDMinimal(userID int) (*User, error) {
	query := `SELECT id, name, role
	FROM users
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := m.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Role,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) InsertUser(u *User) error {
	stmt := `INSERT INTO users
	(name, email, phone_number, id_code, birth_date, password, role, class_id, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt,
		u.Name,
		u.Email,
		u.PhoneNumber,
		u.IdCode,
		u.BirthDate.Time,
		u.Password.Hashed,
		u.Role,
		u.Class.ID,
		u.CreatedAt).Scan(&u.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrEmailAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (m UserModel) UpdateUser(u *User) error {
	stmt := `UPDATE users SET (name, email, phone_number, id_code, birth_date, password, role, class_id, created_at, active) =
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	WHERE id = $11`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt,
		u.Name,
		u.Email,
		u.PhoneNumber,
		u.IdCode,
		u.BirthDate.Time,
		u.Password.Hashed,
		u.Role,
		u.Class.ID,
		u.CreatedAt,
		u.Active,
		u.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrEmailAlreadyExists
		} else {
			return err
		}
	}
	return nil

}

func (m UserModel) GetAllUserIDs() ([]int, error) {
	query := `SELECT
	array(SELECT id	FROM users)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRow(ctx, query).Scan(&ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m UserModel) GetAllStudentIDs() ([]int, error) {
	query := `SELECT
	array(SELECT id	FROM users WHERE role = 'student')`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRow(ctx, query).Scan(&ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m UserModel) GetUserBySessionToken(plaintextToken string) (*User, *int, error) {
	hash := sha256.Sum256([]byte(plaintextToken))

	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.password, u.role, u.class_id, c.name, u.created_at, u.active, u.archived, s.id
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	INNER JOIN sessions s
	ON u.id = s.user_id
	WHERE s.token_hash = $1
	AND s.expires > $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	user.BirthDate = new(Date)
	user.Class = new(Class)
	var sessionID int

	err := m.DB.QueryRow(ctx, query, hash[:], time.Now().UTC()).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.IdCode,
		&user.BirthDate.Time,
		&user.Password.Hashed,
		&user.Role,
		&user.Class.ID,
		&user.Class.Name,
		&user.CreatedAt,
		&user.Active,
		&user.Archived,
		&sessionID,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, nil, ErrInvalidToken
		default:
			return nil, nil, err
		}
	}

	return &user, &sessionID, nil
}

func (m UserModel) GetUserByEmail(email string) (*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.password, u.role, u.class_id, c.name, u.created_at, u.active, u.archived
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	WHERE u.email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	user.BirthDate = new(Date)
	user.Class = new(Class)

	err := m.DB.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.IdCode,
		&user.BirthDate.Time,
		&user.Password.Hashed,
		&user.Role,
		&user.Class.ID,
		&user.Class.Name,
		&user.CreatedAt,
		&user.Active,
		&user.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) AddParentToChild(parentID, childID int) error {
	stmt := `INSERT INTO parents_children
	(parent_id, child_id)
	VALUES
	($1, $2)
	ON CONFLICT DO NOTHING`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, parentID, childID)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) RemoveParentFromChild(parentID, childID int) error {
	stmt := `DELETE FROM parents_children
	WHERE parent_id = $1 and child_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, parentID, childID)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) IsParentOfChild(parentID, childID int) (bool, error) {
	query := `SELECT COUNT(1) FROM parents_children
	WHERE parent_id = $1 and child_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRow(ctx, query, parentID, childID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (m UserModel) GetStudentByID(userID int) (*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.role, u.class_id, c.name
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	WHERE u.id = $1 and u.role = 'student'`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	user.BirthDate = new(Date)
	user.Class = new(Class)

	err := m.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.IdCode,
		&user.BirthDate.Time,
		&user.Role,
		&user.Class.ID,
		&user.Class.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetParentsForChild(childID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.role
	FROM users u
	INNER JOIN parents_children pc
	ON u.id = pc.parent_id
	WHERE pc.child_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, childID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		user.BirthDate = new(Date)

		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PhoneNumber,
			&user.IdCode,
			&user.BirthDate.Time,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) GetChildrenForParent(parentID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.role
	FROM users u
	INNER JOIN parents_children pc
	ON u.id = pc.child_id
	WHERE pc.parent_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
