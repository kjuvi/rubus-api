package models

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"golang.org/x/crypto/bcrypt"
)

// User is the `User` model in the database
type User struct {
	ID           int64  `json:"id" orm:"pk;auto"`
	Username     string `json:"username" orm:"unique"`
	Email        string `json:"email" orm:"unique"`
	PasswordHash string `json:"-"`
}

// NewUser is the model sent to create a new `User`
type NewUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func init() {
	orm.RegisterModel(new(User))
}

// Bind transforms the given payload into a `User`, with some validations
func (u *User) Bind(requestBody []byte) *JSONError {
	var newUser NewUser
	if err := json.Unmarshal(requestBody, &newUser); err != nil {
		return NewBadRequestError()
	}

	// fields validation
	if len(newUser.Username) == 0 {
		return &JSONError{
			Status: http.StatusBadRequest,
			Error:  "username is required",
		}
	}

	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !re.MatchString(newUser.Email) {
		return &JSONError{
			Status: http.StatusBadRequest,
			Error:  "email address is not valid",
		}
	}

	if len(newUser.Password) < 8 {
		return &JSONError{
			Status: http.StatusBadRequest,
			Error:  "password should be at least 8 characters",
		}
	}

	cost, _ := beego.AppConfig.Int("hashcost")
	bytes, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), cost)

	// `NewUser` --> `User`
	u.Username = newUser.Username
	u.Email = newUser.Email
	u.PasswordHash = string(bytes)

	return nil
}

// AddUser inserts a new `User` into the database
func AddUser(u *User) *JSONError {
	o := orm.NewOrm()

	if _, err := o.Insert(u); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return &JSONError{
				Status: http.StatusConflict,
				Error:  "username or email already exists",
			}
		}
		return NewInternalServerError()
	}

	return nil
}

// GetUser returns the `User` with the given `uid` from the database
func GetUser(uid int64) (*User, *JSONError) {
	o := orm.NewOrm()

	user := User{ID: uid}
	if err := o.Read(&user); err != nil {
		if err == orm.ErrNoRows {
			return nil, &JSONError{
				Status: http.StatusNotFound,
				Error:  "user does not exists",
			}
		}
		return nil, NewInternalServerError()
	}

	return &user, nil
}

// GetAllUsers returns all the `User`3 from the database
func GetAllUsers() (users []*User, jsonErr *JSONError) {
	o := orm.NewOrm()

	if _, err := o.QueryTable("user").All(&users); err != nil {
		return nil, NewInternalServerError()
	}

	return users, nil
}

// UpdateUser modifies the `User` with the given `uid` in the database
func UpdateUser(uid int64, uu *User) (u *User, jsonErr *JSONError) {
	o := orm.NewOrm()

	u, jsonErr = GetUser(uid)
	if jsonErr != nil {
		return nil, jsonErr
	}

	if uu.Username != "" {
		u.Username = uu.Username
	}
	if uu.Email != "" {
		u.Email = uu.Email
	}
	if uu.PasswordHash != "" {
		u.PasswordHash = uu.PasswordHash
	}

	if _, err := o.Update(u); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			jsonErr.Status = http.StatusConflict
			jsonErr.Error = "username or email already exists"
			return nil, jsonErr
		}
		return nil, NewInternalServerError()
	}

	return u, nil
}

// DeleteUser removes the given Rubus `User` from the database
func DeleteUser(uid int64) *JSONError {
	o := orm.NewOrm()

	user := User{ID: uid}
	uid, err := o.Delete(&user)
	if uid == 0 {
		return &JSONError{
			Status: http.StatusNotFound,
			Error:  "user does not exists",
		}
	}
	if err != nil {
		return NewInternalServerError()
	}

	return nil
}

// Login checks if the given credentials are valid or not
func Login(username, password string) (*int64, bool) {
	o := orm.NewOrm()

	var user User
	err := o.QueryTable("user").Filter("username", username).One(&user)

	if err == orm.ErrNoRows {
		return nil, false
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, false
	}

	return &user.ID, true
}
