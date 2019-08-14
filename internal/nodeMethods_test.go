package internal

import (
	"testing"
)

func TestNodeMethods(t *testing.T) {
	str := []struct{ quest, answ string }{
		{"192.168.1.4              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", "192.168.1.4"},
		{"192.168.11.4              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", "192.168.11.4"},
		{"192.168.1.0              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", ""},
		{"392.168.1.0              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", ""},
		{"127.168.1              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", ""},
		{"192.168.1.1              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", ""},
		{"192.168.1.255              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", ""},
		{"0.168.1.255              ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", ""},
		{"168.1.            ether   00:24:01:d2:b5:c5   C                     wlx503eaa8f92a6", ""},
	}
	for _, v := range str {
		s := extractIP(v.quest)
		if !isNotRouter(s) {
			s = ""
		}
		if s != v.answ {
			t.Log("Result should be "+v.answ+" but got: ", s)
			t.Fail()
		}
	}

	//check getLanIPs() function
	_, err := getLanIPs()
	if err != nil {
		t.Log("Error shoud be nil but it is:", err)
		t.Fail()
	}
}
