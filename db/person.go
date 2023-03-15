package db

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-msvc/errors"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jansemmelink/events/email"
)

type Person struct {
	ID           *string `db:"id,omitempty"`
	NatID        string  `db:"nat_id" doc:"South African national ID is 13 digits long."`
	Name         string  `db:"first_name" doc:"First name"`
	Surname      string  `db:"last_name" doc:"Surname"`
	Dob          string  `db:"dob" doc:"Date of birth CCYY-MM-DD"`
	Gender       string  `db:"gender" doc:"Gender must be M or F"`
	Phone        *string `db:"phone,omitempty"`
	Email        *string `db:"email,omitempty"`
	PasswordHash *string `db:"password_hash,omitempty" doc:"Hashed password"`
}

//identifier must have one element with name id, nat_id, phone or email
//as they are the unique identifiers on a person record
func GetPerson(identifier map[string]string) (*Person, error) {
	if len(identifier) != 1 {
		return nil, errors.Errorf("invalid identifier %+v requiring id,nat_it,phone or email")
	}
	var n string
	for n = range identifier { //empty for just to get the map key
	}
	switch n {
	case "id":
	case "nat_id":
	case "phone":
	case "email":
	default:
		return nil, errors.Errorf("invalid identifier %+v requiring id,nat_it,phone or email")
	}
	var p Person
	if err := NamedGet(&p,
		"SELECT id,nat_id,first_name,last_name,dob,gender,phone,email,password_hash FROM persons WHERE "+n+"=:"+n,
		identifier,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //not found
		}
		return nil, errors.Wrapf(err, "failed to query person from db")
	}
	return &p, nil //found
} //GetPerson()

const emailPattern = `[a-z]([a-z0-9.-]*[a-z0-9])*@([a-z][a-z0-9-]*\.)*[a-z]+`
const phonePattern = `[+]?[0-9]{9,15}`
const natIdPattern = `[0-9]{13}`
const passwordSymbols = "!@#$%^&*;:\"'(){}[]/\\|?~,.<>+=_-"

var emailRegex = regexp.MustCompile("^" + emailPattern + "$")
var phoneRegex = regexp.MustCompile("^" + phonePattern + "$")
var natIdRegex = regexp.MustCompile("^" + natIdPattern + "$")

func ValidatePassword(pw string) error {
	nrDigits := 0
	nrSymbols := 0
	nrUpper := 0
	nrLower := 0
	if len(pw) < 8 {
		return errors.Errorf("too short")
	}
	if len(pw) > 60 {
		return errors.Errorf("too long")
	}
	for _, c := range pw {
		if unicode.IsDigit(c) {
			nrDigits++
			continue
		}
		if unicode.IsLower(c) {
			nrLower++
			continue
		}
		if unicode.IsUpper(c) {
			nrUpper++
			continue
		}
		if strings.Contains(passwordSymbols, string(c)) {
			nrSymbols++
			continue
		}
		return errors.Errorf("invalid character (%s)", string(c))
	}
	if nrDigits < 1 {
		return errors.Errorf("no digits")
	}
	if nrLower < 1 {
		return errors.Errorf("no lowercase letters")
	}
	if nrUpper < 1 {
		return errors.Errorf("no uppercase letters")
	}
	if nrSymbols < 1 {
		return errors.Errorf("no valid symbols")
	}
	return nil
}

type RegisterRequest struct {
	Name    string `json:"first_name" doc:"Name"`
	Surname string `json:"last_name" doc:"Surname"`
	NatID   string `json:"national_id" doc:"South African national ID is 13 digits long."`
	Dob     string `json:"dob" doc:"Date of birth CCYY-MM-DD"`
	Gender  string `json:"gender" doc:"Gender must be M or F"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}

func (req *RegisterRequest) Validate() error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errors.Errorf("missing name")
	}
	req.Surname = strings.TrimSpace(req.Surname)
	if req.Name == "" {
		return errors.Errorf("missing surname")
	}
	req.NatID = strings.TrimSpace(req.NatID)           //trim outer spaces
	req.NatID = strings.ReplaceAll(req.NatID, " ", "") //remove inner spaces
	if !natIdRegex.MatchString(req.NatID) {
		return errors.Errorf("invalid national_id")
	}
	if _, err := time.ParseInLocation("2006-01-02", req.Dob, time.Now().Location()); err != nil {
		return errors.Errorf("invalid dob \"%s\", expecting for example 1980-07-30", req.Dob)
	}
	switch req.Gender {
	case "M", "F":
	default:
		return errors.Errorf("invalid gender \"%s\", expecting M or F", req.Gender)
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email != "" && !emailRegex.MatchString(req.Email) {
		return errors.Errorf("invalid email \"%s\"", req.Email)
	}
	req.Phone = strings.TrimSpace(req.Phone)           //trim outer spaces
	req.Phone = strings.ReplaceAll(req.Phone, " ", "") //remove inner spaces
	if req.Phone != "" && !phoneRegex.MatchString(req.Phone) {
		return errors.Errorf("invalid phone \"%s\" expecting digits with optional leading plus", req.Phone)
	}
	if req.Phone == "" && req.Email == "" {
		return errors.Errorf("missing phone and email, one is required")
	}
	return nil
} //RegisterRequest.Validate()

func AddPerson(req RegisterRequest) (id string, err error) {
	if err := req.Validate(); err != nil {
		return "", errors.Wrapf(err, "invalid request")
	}

	//only when user has an email address can the user activate
	//the account to login
	tpw := ""
	txp := time.Now()
	if req.Email != "" {
		//generate temp password for activation link in email
		//this is not same format as full password because we keep
		//it simple to send on a URL, also uniq then we can search
		//on it without using the person.id publically
		tpw = uuid.New().String()
		txp = time.Now().Add(time.Hour)
	}

	newId := uuid.New().String()
	_, err = db.NamedExec(
		"INSERT INTO `persons` SET first_name=:first_name,last_name=:last_name,nat_id=:nat_id,dob=:dob,gender=:gender,phone=:phone,email=:email,tpw=:tpw,tpx=:txp",
		map[string]interface{}{
			"id":         newId,
			"first_name": req.Name,
			"last_name":  req.Surname,
			"nat_id":     req.NatID,
			"dob":        req.Dob,
			"gender":     req.Gender,
			"phone":      req.Phone,
			"email":      req.Email,
			"tpw":        tpw,
			"txp":        txp,
		})
	if err != nil {
		fmt.Printf("Failed to insert id=%s: %+v\n", newId, err)
		if sqlError, ok := err.(*mysql.MySQLError); ok {
			switch sqlError.Number {
			case 1062:
				//duplicate
				if strings.Contains(err.Error(), "email") {
					return "", errors.Errorf("email \"%s\" is already registered", req.Email)
				}
				if strings.Contains(err.Error(), "phone") {
					return "", errors.Errorf("phone \"%s\" is already registered", req.Phone)
				}
				if strings.Contains(err.Error(), "nat_id") {
					return "", errors.Errorf("national ID \"%s\" is already registered", req.NatID)
				}
			default: //fallthrough
			}
		}
		return "", errors.Wrapf(err, "failed to register")
	}
	fmt.Printf("Inserted id=%s\n", newId)

	if tpw != "" {
		//send account activation email
		activateLink := fmt.Sprintf("http://localhost:3000/activate?tpw=%s", tpw)
		msg := email.Message{
			From:        email.Email{Addr: "no-reply@events.net", Name: "Events"},
			To:          []email.Email{{Addr: req.Email, Name: req.Name + " " + req.Surname}},
			Subject:     "Account Registration",
			ContentType: "text/html",
		}
		msg.Content = "<H1>Account Registration</H1>"
		msg.Content += "<P>Your email address was registered in the Events system.</P>"
		msg.Content += "<P>Click <a href=\"" + activateLink + "\">here</a> to activate your account.</P>"
		msg.Content += "<P>This link is only active until " + txp.Format("2006-01-02 15:04:05 Z") + "</P>"
		if err := email.Send(msg); err != nil {
			log.Errorf("failed to send email: %+v", err) //log and do not return the underlying fault to the user
			return "", errors.Errorf("failed to send activation message to %s", req.Email)
		}
	} //if must send activation email
	return newId, nil
} //AddPerson()

type AuthActivateRequest struct {
	Tpw         string `json:"tpw" doc:"Temp password from activateLink sent in email to the person."`
	NewPassword string `json:"new_password" doc:"New password that user wants to use. It will be checked against the rules."`
}

func (req AuthActivateRequest) Validate() error {
	if req.Tpw == "" {
		return errors.Errorc(http.StatusBadRequest, "missing tpw")
	}
	if err := ValidatePassword(req.NewPassword); err != nil {
		return errors.Errorc(http.StatusBadRequest, "invalid new password: "+err.Error())
	}
	return nil
}

type PersonToActivate struct {
	ID    string  `db:"id"`
	NatID string  `db:"nat_id"`
	Tpx   SqlTime `db:"tpx"` //time when temp password (tpw) expires
}

func AuthActivate(req AuthActivateRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	var record PersonToActivate
	if err := NamedGet(&record,
		"SELECT `id`,`nat_id`,`tpx` FROM `persons` WHERE `tpw`=:tpw",
		map[string]interface{}{
			"tpw": req.Tpw,
		},
	); err != nil {
		log.Errorf("failed to get PersonToActivate(tpw=%s): %+v", req.Tpw, err)
		return errors.Errorf("personal record not found to activate")
	}
	if time.Time(record.Tpx).Before(time.Now()) {
		log.Errorf("activation expired on (%T)%+v", record, record)
		return errors.Errorf("activation link expired at %v", record.Tpx)
	}

	//this is a valid activation
	//delete the temp password and set the hashed password
	//so the user will be able to login
	pwh := HashPassword(record.NatID, req.NewPassword)
	if _, err := db.NamedExec(
		"UPDATE `persons` SET `tpw`=NULL,`tpx`=NULL,`password_hash`=:pwh WHERE `id`=:id",
		map[string]interface{}{
			"id":  record.ID,
			"pwh": pwh,
		},
	); err != nil {
		return errors.Wrapf(err, "failed to set new password")
	}
	return nil
}

type AuthResetRequest struct {
	Email string `json:"email"`
}

func (req AuthResetRequest) Validate() error {
	if req.Email == "" {
		return errors.Errorc(http.StatusBadRequest, "missing email")
	}
	return nil
}

func AuthReset(req AuthResetRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	//generate temp password for activation link in email
	//this is not same format as full password because we keep
	//it simple to send on a URL, also uniq then we can search
	//on it without using the person.id publically
	tpw := uuid.New().String()
	txp := time.Now().Add(time.Hour)
	result, err := db.NamedExec(
		"UPDATE `persons` SET `tpw`=:tpw,`tpx`=:txp WHERE `email`=:email",
		map[string]interface{}{
			"email": req.Email,
			"tpw":   tpw,
			"txp":   txp,
		})
	if err != nil {
		return errors.Wrapf(err, "failed to update tpw and tpx for email=%s: %+v\n", req.Email, err)
	}
	if nr, _ := result.RowsAffected(); nr < 1 {
		return errors.Errorc(http.StatusNotFound, "email not found")
	}
	//send password reset email
	activateLink := fmt.Sprintf("http://localhost:3000/activate?tpw=%s", tpw)
	msg := email.Message{
		From:        email.Email{Addr: "no-reply@events.net", Name: "Events"},
		To:          []email.Email{{Addr: req.Email, Name: ""}},
		Subject:     "Password Reset",
		ContentType: "text/html",
	}
	msg.Content = "<H1>Password Reset</H1>"
	msg.Content += "<P>We received a request to reset your password for the Events system. If you did not send the request, please delete and ignore this email and keep using your current password.</P>"
	msg.Content += "<P>Click <a href=\"" + activateLink + "\">here</a> to set a new password for your Events account.</P>"
	msg.Content += "<P>This link is only active until " + txp.Format("2006-01-02 15:04:05 Z") + "</P>"
	if err := email.Send(msg); err != nil {
		log.Errorf("failed to send email: %+v", err) //log and do not return the underlying fault to the user
		return errors.Errorf("failed to send reset message to %s", req.Email)
	}
	return nil
} //AuthReset()

type LoginRequest struct {
	Username string `json:"username" doc:"Username could be either the national ID, phone or email."`
	Password string `json:"password"`
}

func (req LoginRequest) Validate() error {
	if req.Username == "" {
		return errors.Errorf("missing username")
	}
	if req.Password == "" {
		return errors.Errorf("missing password")
	}
	return nil
}

func Login(req LoginRequest) (*Person, error) {
	//nat_id and phone nr could have same value for different people
	//so we may find multiple persons matching the username, one as id
	//and another as phone nr...
	var persons []Person
	if err := NamedSelect(
		&persons,
		"SELECT id,nat_id,last_name,first_name,dob,gender,phone,email,password_hash"+
			" FROM `persons`"+
			" WHERE (`password_hash` IS NOT NULL) AND (`nat_id`=:username OR `email`=:username OR `phone`=:username)",
		map[string]interface{}{
			"username": req.Username,
		}); err != nil {
		log.Errorf("Failed to search users for login: %+v", err)
		return nil, errors.Errorf("unknown username %s", req.Username)
	}
	if len(persons) == 0 {
		return nil, errors.Errorf("unknown username %s", req.Username)
	}
	for _, p := range persons {
		pwh := HashPassword(p.NatID, req.Password)
		if pwh == *p.PasswordHash {
			return &p, nil
		}
	}
	return nil, errors.Errorf("wrong password")
} //Login

func HashPassword(natID, pw string) string {
	h := sha1.New()
	s := natID + pw
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
