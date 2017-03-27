package chrome

/*
#cgo LDFLAGS: -lcrypt32
#include <windows.h>
#include <wincrypt.h>

BOOL decryptData(char *pIn, int inLen, char **ppOut, int *pOutLen) {
	DATA_BLOB input, output;
	input.cbData = inLen;
	input.pbData = pIn;
	BOOL result = CryptUnprotectData(&input, 0, 0, 0, 0, 0, &output) == TRUE;
	if (ppOut) { *ppOut = output.pbData; }
	if (pOutLen) { *pOutLen = (int)output.cbData; }
	return result;
}
*/
import "C"

import (
	"database/sql"
	"io"
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mogaika/browsers"
)

func dectyptData(dataIn []byte) []byte {
	var pbData *C.char
	var cbData C.int
	if C.decryptData((*C.char)(unsafe.Pointer(&dataIn[0])), C.int(len(dataIn)), &pbData, &cbData) == C.FALSE {
		panic(syscall.GetLastError())
	}
	return C.GoBytes(unsafe.Pointer(pbData), cbData)
}

type ChromePassword struct {
	OriginUrl, ActionUrl           string
	UsernameElement, UsernameValue string
	PasswordElement, PasswordValue string
	SignonRealm                    string
	DateCreated                    int64
}

func (p *ChromePassword) Urls() (form, target string) {
	return p.OriginUrl, p.ActionUrl
}
func (p *ChromePassword) Username() (element, username string) {
	return p.UsernameElement, p.UsernameValue
}
func (p *ChromePassword) Password() (element, password string) {
	return p.PasswordElement, p.PasswordValue
}

func (info *ChromePassword) loginInfoFromRow(rows *sql.Rows) bool {
	var rawPasswd []byte
	if err := rows.Scan(
		&info.OriginUrl, &info.ActionUrl,
		&info.UsernameElement, &info.UsernameValue,
		&info.PasswordElement, &rawPasswd,
		&info.SignonRealm, &info.DateCreated); err != nil {
		return false
	}
	info.PasswordValue = string(dectyptData(rawPasswd))
	return true
}

func loadSavedPasswords(pathToDbFile string) ([]*ChromePassword, error) {
	db, err := sql.Open("sqlite3", pathToDbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT origin_url, action_url, username_element, username_value, password_element, password_value, signon_realm, date_created from logins")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	passwds := make([]*ChromePassword, 0)
	for rows.Next() {
		if p := new(ChromePassword); p.loginInfoFromRow(rows) {
			passwds = append(passwds, p)
		}
	}

	return passwds, nil
}

type ChromeBrowser struct{}

func (*ChromeBrowser) Name() string {
	return "chrome"
}

func SavedPasswords() ([]*ChromePassword, error) {
	tempf, err := ioutil.TempFile(os.TempDir(), "_ourdb")
	if err != nil {
		return nil, err
	}
	defer func() {
		tempf.Close()
		os.Remove(tempf.Name())
	}()

	dataPath := path.Join(os.Getenv("LOCALAPPDATA"), `Google\Chrome\User Data\Default\Login Data`)
	origf, err := os.Open(dataPath)
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

func (cb *ChromeBrowser) SavedPasswords() ([]browsers.Password, error) {
	passwds, err := SavedPasswords()
	if err != nil {
		return nil, err
	} else {
		bases := make([]browsers.Password, len(passwds))
		for i := range passwds {
			bases[i] = passwds[i]
		}
		return bases, nil
	}
}

func init() {
	browsers.AddBrowser(&ChromeBrowser{})
}
