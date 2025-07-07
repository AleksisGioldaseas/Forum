package database

import (
	"errors"
	"fmt"
	"html/template"

	"forum/utils"
	"html"
	"net/mail"
	"slices"
	"strings"
)

var (
	SPECIAL_CHARACTERS = tableize("~`! @#$%^&*()-_+={}[]|\\;:\"<>,./?")
	CATEGORIES         = []string{}
	VOTEACTIONS        = []string{"like", "dislike", "neutral"}
)

// Validates username, will return error if the argument is not valid, the error will explain why
func (db *DataBase) ValidateUsername(userName string) error {
	if len(userName) > db.Limits.MaxUsername {
		return errors.New("username too long")
	}

	if len(userName) < db.Limits.MinUsername {
		return errors.New("username too short")
	}

	if !utils.IsAlphaNumeric(userName) {
		return errors.New("contains invalid characters")
	}

	return nil
}

// Validates email, will return error if the argument is not valid, the error will explain why
func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email) //<- apparently this is barely sufficient
	if err != nil {
		fmt.Println(email)
		return err
	}

	return nil
}

func (db *DataBase) ValidateTitle(title string) error {

	if len(title) > db.Limits.MaxTitle {
		return errors.New("title too long")
	}

	if len(title) < db.Limits.MinTitle {
		return errors.New("title too short")
	}

	return nil
}

func (db *DataBase) ValidatePostBody(body string) error {
	if len(body) > db.Limits.MaxPostBody {
		return errors.New("post body too long")
	}

	if len(body) < db.Limits.MinBody {
		return errors.New("post body too short")
	}

	return nil
}

func (db *DataBase) ValidateReport(body string) error {
	if len(body) > db.Limits.MaxReportBody {
		return errors.New("report too long")
	}

	if len(body) < db.Limits.MinBody {
		return errors.New("report too short")
	}
	return nil
}

func (db *DataBase) ValidateCommentBody(body string) error {
	if len(body) > db.Limits.MaxCommentBody {
		return errors.New("title too long")
	}

	if len(body) < db.Limits.MinBody {
		return errors.New("title too short")
	}

	return nil
}

// Validates password, will return error if the argument is not valid, the error will explain why
func (db *DataBase) ValidatePassword(password, passwordRepeat string) error {

	if password != passwordRepeat {
		return errors.New("passwords do not match")
	}

	if len(password) > db.Limits.MaxPass {
		return errors.New("password is too long")
	}

	if len(password) < db.Limits.MinPass {
		return errors.New("password is too short")
	}

	for _, r := range password {
		if r < 32 || r > 126 {
			return errors.New("invalid characters in password")
		}
	}

	hasLower := false
	hasUpper := false
	hasSpecialChar := false
	hasNumber := false
	for _, r := range password {
		if r >= 'a' && r <= 'z' {
			hasLower = true
		}

		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		}

		if r >= '0' && r <= '9' {
			hasNumber = true
		}

		if SPECIAL_CHARACTERS[r] {
			hasSpecialChar = true
		}

		if hasLower && hasUpper && hasSpecialChar && hasNumber {
			break
		}
	}

	if !hasLower {
		return errors.New("password must contain at least 1 lower-cased character")
	}

	if !hasUpper {
		return errors.New("password must contain at least 1 upper-cased character")
	}

	if !hasSpecialChar {
		return errors.New("password must contain at least 1 special character")
	}

	if !hasNumber {
		return errors.New("password must contain at least 1 number")
	}

	return nil
}

func (db *DataBase) ValidateBio(bio string) error {
	if len(bio) > db.Limits.MaxBio {
		return errors.New("bio text too long")
	}
	return nil
}

func (db *DataBase) ValidateCategories(categories []string) error {
	if len(categories) > db.Limits.MaxCategories {
		return errors.New("too many categories")
	}

	if len(categories) == 0 {
		return errors.New("categories required")
	}

	return nil
}

func Sanitize(str string) string {
	return html.EscapeString(strings.TrimSpace(str))
}

// turns a string into a array-map (behaves kinda like a map, but it's an array). Used for quick lookup if a character is contained in the collection of characters that are allowed
func tableize(characters string) [127]bool {
	table := [127]bool{}

	for _, r := range characters {
		table[int(r)] = true
	}
	return table
}

func VoteAction(action string) error {
	if !slices.Contains(VOTEACTIONS, action) {
		return errors.New("unknown vote action")
	}
	return nil
}

func BasicHTMLSanitize(input string) template.HTML {
	textBroke := strings.ReplaceAll(input, "\n", "<br>")

	// escape the HTML first
	escapedText := template.HTMLEscapeString(textBroke)

	// unescape formatting tags to allow users expressive freedom (pws ta lew)
	replacements := map[string]string{
		"&lt;b&gt;": "<b>", "&lt;/b&gt;": "</b>", // bold
		"&lt;i&gt;": "<i>", "&lt;/i&gt;": "</i>", // italics
		"&lt;u&gt;": "<u>", "&lt;/u&gt;": "</u>", // underlined
		"&lt;br&gt;": "<br>", "&lt;/br&gt;": "</br>", // linebreak
		"&lt;p&gt;": "<p>", "&lt;/p&gt;": "</p>", // paragraph
		"&lt;ul&gt;": "<ul>", "&lt;/ul&gt;": "</ul>", // unordered list
		"&lt;ol&gt;": "<ol>", "&lt;/ol&gt;": "</ol>", // ordered list
		"&lt;li&gt;": "<li>", "&lt;/li&gt;": "</li>", // list item
	}

	for old, new := range replacements {
		escapedText = strings.ReplaceAll(escapedText, old, new)
	}

	return template.HTML(escapedText)
}
