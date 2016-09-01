package oscmd

import "testing"

func osInit(cmd Oscmd) []string {
	res := cmd.OpenFW(1111, "tcp")
	res = append(res, cmd.OpenFW(2222, "tcp")...)
	return res
}

func TestRHEL(t *testing.T) {
	var os Oscmd
	os = RedHat{}
	t.Logf("testing %T %+v", os, os)
	res := osInit(os)
	t.Logf("res %+v", res)
	if len(res) != 4 {
		t.Fail()
	}
}

func TestCentOS(t *testing.T) {
	var os Oscmd
	os = CentOS{}
	t.Logf("testing %T %+v", os, os)
	res := osInit(os)
	t.Logf("res %+v", res)
	if len(res) != 4 {
		t.Fail()
	}
}

func TestUbuntu(t *testing.T) {
	var os Oscmd
	os = Ubuntu{}
	t.Logf("testing %T %+v", os, os)
	res := osInit(os)
	t.Logf("res %+v", res)
	if len(res) != 2 {
		t.Fail()
	}
}

func TestDebian(t *testing.T) {
	var os Oscmd
	os = Debian{}
	t.Logf("testing %T %+v", os, os)
	res := osInit(os)
	t.Logf("res %+v", res)
	if len(res) != 2 {
		t.Fail()
	}
}
