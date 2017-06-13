package models

import (
	"time"
	"github.com/s-gv/orangeforum/models/db"
	"math/rand"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"encoding/hex"
	"database/sql"
)

const (
	VoteUp = 1
	VoteDown = 2
	VoteFlag = 3
)

const ModelVersion = 1

const (
	ForumName string = "forum_name"
	HeaderMsg string = "header_msg"
	SignupDisabled string = "signup_disabled"
	GroupCreationDisabled string = "group_creation_disabled"
	ImageUploadEnabled string = "image_upload_enabled"
	FileUploadEnabled string = "file_upload_enabled"
	AllowGroupSubscription string = "allow_group_subscription"
	AllowTopicSubscription string = "allow_topic_subscription"
	DataDir string = "data_dir"
	DefaultFromMail string = "default_from_mail"
	SMTPHost string = "smtp_host"
	SMTPPort string = "smtp_port"
	SMTPUser string = "smtp_user"
	SMTPPass string = "smtp_pass"
)

var ErrIncorrectPasswd = errors.New("Incorrect username/password.")
var ErrUserNotFound = errors.New("Username not found.")
var ErrUserAlreadyExists = errors.New("Username already exists.")

type User struct {
	ID int
	Username string
	PasswdHash string
	Email string
	About string
	Karma int
	IsBanned bool
	IsSuperAdmin bool
	CreatedDate time.Time
	UpdatedDate time.Time
}

type Group struct {
	ID int
	Name string
	Desc string
	IsSticky string
	IsPrivate string
	IsClosed string
	CreatedDate time.Time
	UpdatedDate time.Time
}

type Topic struct {
	ID int
	Content string
	AuthorID int
	GroupID int
	IsDeleted bool
	IsSticky bool
	IsClosed bool
	NumComments int
	Upvotes int
	Downvotes int
	Flagvotes int
	CreatedDate time.Time
	UpdatedDate time.Time
}

type Comment struct {
	ID int
	Content string
	AuthorID int
	TopicID int
	ParentID int
	IsDeleted bool
	IsSticky bool
	Upvotes int
	Downvotes int
	Flagvotes int
	CreatedDate time.Time
	UpdatedDate time.Time
}

type Mod struct {
	ID int
	UserID int
	GroupID int
	CreatedDate time.Time
}

type Admin struct {
	ID int
	UserID int
	GroupID int
	CreatedDate time.Time
}

type TopicVote struct {
	ID int
	UserID int
	TopicID int
	VoteType int
	CreatedDate time.Time
}

type CommentVote struct {
	ID int
	UserID int
	CommentID int
	VoteType int
	CreatedDate time.Time
}

type TopicSubscription struct {
	ID int
	UserID int
	TopicID int
	CreatedDate time.Time
}

type GroupSubscription struct {
	ID int
	UserID int
	GroupID int
	CreatedDate time.Time
}

type ExtraNote struct {
	ID int
	Name string
	Content string
	URL string
	CreatedDate time.Time
	UpdatedDate time.Time
}

type CommonData struct {
	CSRF string
	Msg string
	UserName string
	Karma int
	ForumName string
	ExtraNotesShort []ExtraNote
}

func createUser(userName string, passwd string, email string, isSuperAdmin bool) error {
	if passwdHash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost); err == nil {
		r := db.QueryRow(`SELECT username FROM users WHERE username=?;`, userName)
		var tmp string
		if err := r.Scan(&tmp); err == sql.ErrNoRows {
			db.Exec(`INSERT INTO users(username, passwdhash, email, is_superadmin) VALUES(?, ?, ?, ?);`,
				userName, hex.EncodeToString(passwdHash), email, isSuperAdmin)
		} else {
			return ErrUserAlreadyExists
		}
	} else {
		return err
	}
	return nil
}

func CreateUser(userName string, passwd string, email string) error {
	return createUser(userName, passwd, email, false)
}

func CreateSuperUser(userName string, passwd string) error {
	return createUser(userName, passwd, "", true)
}

func ReadUserAbout(userName string) string {
	r := db.QueryRow(`SELECT about FROM users WHERE username=?; `, userName)
	var about string
	if err := r.Scan(&about); err == nil {
		return about
	}
	return ""
}

func ReadUserKarma(userName string) int {
	r := db.QueryRow(`SELECT karma FROM users WHERE username=?; `, userName)
	var karma int
	if err := r.Scan(&karma); err == nil {
		return karma
	}
	return 0
}

func ReadUserEmail(userName string) string {
	r := db.QueryRow(`SELECT email FROM users WHERE username=?;`, userName)
	var email string
	if err := r.Scan(&email); err == nil {
		return email
	}
	return ""
}

func ReadUserNameByToken(resetToken string) (string, error) {
	if len(resetToken) > 0 {
		r := db.QueryRow(`SELECT username, reset_token_date FROM users WHERE reset_token=?;`, resetToken)
		var userName string
		var rDate int64
		if err := r.Scan(&userName, &rDate); err == nil {
			resetDate := time.Unix(rDate, 0)
			if resetDate.After(time.Now().Add(-48*time.Hour)) {
				return userName, nil
			}
		}
	}
	return "", errors.New("Invalid/Expired reset token.")
}

func ReadUserIDByName(userName string) (int, error) {
	r := db.QueryRow(`SELECT id FROM users WHERE username=?;`, userName)
	var id int
	if err := r.Scan(&id); err == nil {
		return id, nil
	}
	return 0, errors.New("User not found.")
}

func UpdateUserPasswd(userName string, passwd string) error {
	if passwdHash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost); err == nil {
		db.Exec(`UPDATE users SET passwdhash=?, reset_token="", reset_token_date=0 WHERE username=?`, hex.EncodeToString(passwdHash), userName)
	} else {
		return err
	}
	return nil
}

func UpdateUserProfile(userName string, email string, about string) {
	db.Exec(`UPDATE users SET email=?, about=?, updated_date=? WHERE username=?`, email, about, int64(time.Now().Unix()), userName)
}

func CreateResetToken(userName string) string {
	resetToken := RandSeq(64)
	db.Exec(`UPDATE users SET reset_token=?, reset_token_date=? WHERE username=?;`, resetToken, int64(time.Now().Unix()), userName)
	return resetToken
}

func ProbeUser(userName string) bool {
	r := db.QueryRow(`SELECT username FROM users WHERE username=?;`, userName)
	var tmp string
	if err := r.Scan(&tmp); err == sql.ErrNoRows {
		return false
	}
	return true
}

func CreateGroup(name string, desc string, headerMsg string) {
	now := time.Now().Unix()
	db.Exec(`INSERT INTO groups(name, desc, header_msg, created_date, updated_date) VALUES(?, ?, ?, ?, ?);`, name, desc, headerMsg, now, now)
}

func ReadGroupIDByName(name string) string {
	r := db.QueryRow(`SELECT id FROM groups WHERE name=?;`, name)
	var id string
	if err := r.Scan(&id); err == nil {
		return id
	}
	return ""
}

func ReadGroupDesc(groupID string) string {
	r := db.QueryRow(`SELECT desc FROM groups WHERE id=?;`, groupID)
	var desc string
	if err := r.Scan(&desc); err == nil {
		return desc
	}
	return ""
}

func ReadGroupHeaderMsg(groupID string) string {
	r := db.QueryRow(`SELECT header_msg FROM groups WHERE id=?;`, groupID)
	var headerMsg string
	if err := r.Scan(&headerMsg); err == nil {
		return headerMsg
	}
	return ""
}

func ReadGroupName(groupID string) string {
	r := db.QueryRow(`SELECT name FROM groups WHERE id=?;`, groupID)
	var name string
	if err := r.Scan(&name); err == nil {
		return name
	}
	return ""
}

func ReadGroupIsDeleted(groupID string) bool {
	r := db.QueryRow(`SELECT is_closed FROM groups WHERE id=?;`, groupID)
	var isDeleted string
	if err := r.Scan(&isDeleted); err == nil {
		return isDeleted == "1"
	}
	return false
}

func UpdateGroup(groupID string, name string, desc string, headerMsg string) {
	db.Exec(`UPDATE groups SET name=?, desc=?, header_msg=?, updated_date=? WHERE id=?`, name, desc, headerMsg, int(time.Now().Unix()), groupID)
}

func DeleteGroup(groupID string) {
	db.Exec(`UPDATE groups SET is_closed=1 WHERE id=?;`, groupID)
}

func UndeleteGroup(groupID string) {
	db.Exec(`UPDATE groups SET is_closed=0 WHERE id=?;`, groupID)
}

func CreateMod(userName string, groupID string) {
	if uid, err := ReadUserIDByName(userName); err == nil {
		db.Exec(`INSERT INTO mods(userid, groupid, created_date) VALUES(?, ?, ?);`, uid, groupID, int64(time.Now().Unix()))
	}
}

func ReadMods(groupID string) []string {
	rows := db.Query(`SELECT users.username FROM users INNER JOIN mods ON users.id=mods.userid WHERE mods.groupid=?;`, groupID)
	var mods []string
	for rows.Next() {
		var mod string
		rows.Scan(&mod)
		mods = append(mods, mod)
	}
	return mods
}

func DeleteMods(groupID string) {
	db.Exec(`DELETE FROM admins WHERE groupid=?;`, groupID)
}


func CreateAdmin(userName string, groupID string) {
	if uid, err := ReadUserIDByName(userName); err == nil {
		db.Exec(`INSERT INTO admins(userid, groupid, created_date) VALUES(?, ?, ?);`, uid, groupID, int64(time.Now().Unix()))
	}
}

func ReadAdmins(groupID string) []string {
	rows := db.Query(`SELECT users.username FROM users INNER JOIN admins ON users.id=admins.userid WHERE admins.groupid=?;`, groupID)
	var admins []string
	for rows.Next() {
		var admin string
		rows.Scan(&admin)
		admins = append(admins, admin)
	}
	return admins
}

func IsUserGroupAdmin(userID string, groupID string) bool {
	r := db.QueryRow(`SELECT id FROM admins WHERE userid=? AND groupid=?`, userID, groupID)
	var tmp string
	if err := r.Scan(&tmp); err == nil {
		return true
	}
	return false
}

func DeleteAdmins(groupID string) {
	db.Exec(`DELETE FROM mods WHERE groupid=?;`, groupID)
}



func CreateExtraNote(name string, URL string, content string) {
	now := time.Now()
	db.Exec(`INSERT INTO extranotes(name, URL, content, created_date, updated_date) VALUES(?, ?, ?, ?, ?);`, name, URL, content, int64(now.Unix()), int64(now.Unix()))
}

func ReadExtraNotes() []ExtraNote {
	rows := db.Query(`SELECT id, name, URL, content FROM extranotes;`)
	var extraNotes []ExtraNote
	for rows.Next() {
		var extraNote ExtraNote
		rows.Scan(&extraNote.ID, &extraNote.Name, &extraNote.URL, &extraNote.Content)
		extraNotes = append(extraNotes, extraNote)
	}
	return extraNotes
}

func ReadExtraNote(id string) (ExtraNote, error) {
	r := db.QueryRow(`SELECT name, URL, content, created_date, updated_date FROM extranotes WHERE id=?;`, id)
	var e ExtraNote
	var cDate int64
	var uDate int64
	if err := r.Scan(&e.Name, &e.URL, &e.Content, &cDate, &uDate); err == nil {
		e.CreatedDate = time.Unix(cDate, 0)
		e.UpdatedDate = time.Unix(uDate, 0)
		return e, nil
	}
	return ExtraNote{}, errors.New("No note with that ID found")
}

func ReadExtraNotesShort() []ExtraNote {
	rows := db.Query(`SELECT id, name FROM extranotes;`)
	var extraNotes []ExtraNote
	for rows.Next() {
		var extraNote ExtraNote
		rows.Scan(&extraNote.ID, &extraNote.Name)
		extraNotes = append(extraNotes, extraNote)
	}
	return extraNotes
}

func UpdateExtraNote(id string, name string, URL string, content string) {
	now := time.Now()
	db.Exec(`UPDATE extranotes SET name=?, URL=?, content=?, updated_date=? WHERE id=?;`, name, URL, content, int64(now.Unix()), id)
}

func DeleteExtraNote(id string) {
	db.Exec(`DELETE FROM extranotes WHERE id=?;`, id)
}

func ReadCommonData(sess Session) CommonData {
	userName := ""
	karma := 0
	if sess.UserID.Valid {
		r := db.QueryRow(`SELECT username, karma FROM users WHERE id=?;`, sess.UserID)
		r.Scan(&userName, &karma)
	}
	return CommonData{
		CSRF:sess.CSRFToken,
		Msg:sess.FlashMsg(),
		UserName:userName,
		Karma:karma,
		ForumName:Config(ForumName),
		ExtraNotesShort:ReadExtraNotesShort(),
	}
}

func RandSeq(n int) string {
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func WriteConfig(key string, val string) {
	db.Exec(`INSERT OR REPLACE INTO configs(key, val) values(?, ?);`, key, val)
}


func Config(key string) string {
	row := db.QueryRow(`SELECT val FROM configs WHERE key=?;`, key)
	var val string
	if err := row.Scan(&val); err == nil {
		return val
	}
	return "0"
}

func ConfigAllVals() map[string]interface{} {
	vals := map[string]interface{}{
		"forum_name": Config(ForumName),
		"header_msg": Config(HeaderMsg),
		"signup_disabled": Config(SignupDisabled) == "1",
		"group_creation_disabled": Config(GroupCreationDisabled) == "1",
		"image_upload_enabled": Config(ImageUploadEnabled) == "1",
		"file_upload_enabled": Config(FileUploadEnabled) == "1",
		"allow_group_subscription": Config(AllowGroupSubscription) == "1",
		"allow_topic_subscription": Config(AllowTopicSubscription) == "1",
		"data_dir": Config(DataDir),
		"default_from_mail": Config(DefaultFromMail),
		"smtp_host": Config(SMTPHost),
		"smtp_port": Config(SMTPPort),
		"smtp_user": Config(SMTPUser),
		"smtp_pass": Config(SMTPPass),
	}
	return vals
}

func ConfigCommonVals() map[string]string {
	vals := map[string]string{
		"forum_name": Config(ForumName),
	}
	return vals
}

func NumUsers() int64 {
	r := db.QueryRow(`SELECT MAX(_ROWID_) FROM users LIMIT 1;`)
	var n sql.NullInt64
	if err := r.Scan(&n); err == nil {
		return n.Int64
	}
	return 0
}

func NumGroups() int64 {
	r := db.QueryRow(`SELECT MAX(_ROWID_) FROM groups LIMIT 1;`)
	var n sql.NullInt64
	if err := r.Scan(&n); err == nil {
		return n.Int64
	}
	return 0
}

func NumTopics() int64 {
	r := db.QueryRow(`SELECT MAX(_ROWID_) FROM topics LIMIT 1;`)
	var n sql.NullInt64
	if err := r.Scan(&n); err == nil {
		return n.Int64
	}
	return 0
}

func NumComments() int64 {
	r := db.QueryRow(`SELECT MAX(_ROWID_) FROM comments LIMIT 1;`)
	var n sql.NullInt64
	if err := r.Scan(&n); err == nil {
		return n.Int64
	}
	return 0
}

func Migrate() {
	db.CreateTables()

	WriteConfig("version", "1");
	WriteConfig(HeaderMsg, "")
	WriteConfig(ForumName, "Orange Forum")
	WriteConfig(SignupDisabled, "0")
	WriteConfig(GroupCreationDisabled, "0")
	WriteConfig(FileUploadEnabled, "0")
	WriteConfig(ImageUploadEnabled, "0")
	WriteConfig(AllowGroupSubscription, "0")
	WriteConfig(AllowTopicSubscription, "0")
	WriteConfig(DataDir, "")
	WriteConfig(DefaultFromMail, "admin@example.com")
	WriteConfig(SMTPHost, "")
	WriteConfig(SMTPPort, "25")
	WriteConfig(SMTPUser, "")
	WriteConfig(SMTPPass, "")
}

func IsMigrationNeeded() bool {
	dbver := db.DBVersion()
	return dbver != ModelVersion
}