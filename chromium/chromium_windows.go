package chromium

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

func init() {
	browsers.AddBrowser(&ChromiumBasedBrowser{name: "chrome",
		databases: func() []string {
			dbs := make([]string, 0)
			localappd := os.Getenv("LOCALAPPDATA")
			for _, dir := range []string{
				"Google\\Chrome",
				"Google\\Chrome SxS",
				"Chromium",
			} {
				dbs = append(dbs, databasesFromUserData(path.Join(localappd, path.Join(dir, "User Data")))...)
			}
			return dbs
		}})
	browsers.AddBrowser(&ChromiumBasedBrowser{name: "yandex",
		databases: func() []string {
			return databasesFromUserData(path.Join(os.Getenv("LOCALAPPDATA"),
				"Yandex\\YandexBrowser\\User Data\\Default\\Login Data"))
		}})
	browsers.AddBrowser(&ChromiumBasedBrowser{name: "vivaldi",
		databases: func() []string {
			return databasesFromUserData(path.Join(os.Getenv("LOCALAPPDATA"),
				"Vivaldi\\User Data\\Default\\Login Data"))
		}})
	browsers.AddBrowser(&ChromiumBasedBrowser{name: "opera",
		databases: func() []string {
			return databasesFromUserData(path.Join(os.Getenv("LOCALAPPDATA"),
				"Opera Software\\Opera Stable\\Login Data"))
		}})
}
