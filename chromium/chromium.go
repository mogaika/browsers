package chromium

import (
	"database/sql"
	"io"
	"io/ioutil"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mogaika/browsers"
)

func loginInfoFromRow(p *browsers.Password, rows *sql.Rows) bool {
	var rawPasswd []byte
	if err := rows.Scan(
		&p.OriginUrl, &p.ActionUrl,
		&p.UsernameElement, &p.Username,
		&p.PasswordElement, &rawPasswd); err != nil {
		return false
	}
	p.Password = string(dectyptData(rawPasswd))
	return true
}

func loadSavedPasswords(pathToDbFile string) ([]browsers.Password, error) {
	db, err := sql.Open("sqlite3", pathToDbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT origin_url, action_url, username_element, username_value, password_element, password_value from logins")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	passwds := make([]browsers.Password, 0)
	for rows.Next() {
		var p browsers.Password
		if loginInfoFromRow(&p, rows) {
			passwds = append(passwds, p)
		}
	}

	return passwds, nil
}

func SavedPasswords(dbpath string) ([]browsers.Password, error) {
	tempf, err := ioutil.TempFile(os.TempDir(), "_ourdb")
	if err != nil {
		return nil, err
	}
	defer func() {
		tempf.Close()
		os.Remove(tempf.Name())
	}()

	origf, err := os.Open(dbpath)
	if err != nil {
		return nil, err
	}
	defer origf.Close()

	_, err = io.Copy(tempf, origf)
	if err != nil {
		return nil, err
	}

	return loadSavedPasswords(tempf.Name())
}

type ChromiumBasedBrowser struct {
	name      string
	databases func() []string
}

func (cb *ChromiumBasedBrowser) Name() string {
	return cb.name
}

func (cbb *ChromiumBasedBrowser) SavedPasswords() ([]browsers.Password, error) {
	var result []browsers.Password
	var lasterr error = nil
	for _, path := range cbb.databases() {
		if passwds, err := SavedPasswords(path); err == nil {
			if result == nil {
				result = passwds
			} else {
				result = append(result, passwds...)
			}
		} else {
			lasterr = err
		}
	}

	if result == nil {
		return nil, lasterr
	} else {
		return result, nil
	}
}

func databasesFromProfileDir(profileDir string) []string {
	return []string{path.Join(profileDir, "Web Data"),
		path.Join(profileDir, "Login Data")}
}

func databasesFromUserData(userData string) []string {
	dbs := make([]string, 0)
	dbs = append(dbs, databasesFromProfileDir(path.Join(userData, "Default"))...)

	dirs, err := ioutil.ReadDir(userData)
	if err == nil {
		for _, dir := range dirs {
			if dir.IsDir() {
				dbs = append(dbs, databasesFromProfileDir(path.Join(userData, dir.Name()))...)
			}
		}
	}

	return dbs
}
