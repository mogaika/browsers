package browsers

/* Saved password interface */
type Password interface {
	/* Url, where password been saved */
	Urls() (form, target string)
	Username() (element, username string)
	Password() (element, password string)
}

type Browser interface {
	Name() string
	/* Load all saved passwords from browser store */
	SavedPasswords() ([]Password, error)
}

var browsers = make(map[string]Browser)

func AddBrowser(b Browser) {
	if _, already := browsers[b.Name()]; already {
		panic(b)
	}
	browsers[b.Name()] = b
}

func Browsers() map[string]Browser {
	return browsers
}

/* Load all saved passwords from all browsers */
func SavedPasswords() (passwds []Password, errs map[string]error) {
	for bname, b := range browsers {
		if currentPasswds, err := b.SavedPasswords(); err == nil {
			if currentPasswds != nil {
				if passwds == nil {
					passwds = make([]Password, 0)
				}
				passwds = append(passwds, currentPasswds...)
			}
		} else {
			if errs == nil {
				errs = make(map[string]error)
			}
			errs[bname] = err
		}
	}
	return
}
