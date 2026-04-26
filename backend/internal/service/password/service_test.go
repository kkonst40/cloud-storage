package password

import "testing"

func TestCreateHashedPassword(t *testing.T) {
	pwd := "very-strong-secret-7"
	pwdHash, _ := GeneratePwdHash(pwd)

	if pwdHash == "" {
		t.Errorf("hashed password empty")
	}

	if !VerifyPwd(pwd, pwdHash) {
		t.Errorf("hashed password is not valid")
	}
}
