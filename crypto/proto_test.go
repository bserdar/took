package crypto

import (
	"testing"
)

func TestProto(t *testing.T) {
	server, err := InitServer("password")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	rp := NewRequestProcessor(server, nil, "test")

	req := DataRequest{Data: "test"}
	var rsp DataResponse
	rp.Encrypt(req, &rsp)
	req.Data = rsp.Data
	rp.Decrypt(req, &rsp)
	if rsp.Data != "test" {
		t.Errorf("Decrypt error %s != %s", rsp.Data, req.Data)
	}
}

func TestLoginEmul(t *testing.T) {
	server, err := InitServer("password")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	rp := NewRequestProcessor(server, nil, "test")

	req := DataRequest{Data: "test"}
	var rsp DataResponse
	rp.Encrypt(req, &rsp)

	a, _ := server.GetAuthKey()
	server2, _ := NewServer("password", a)
	rp2 := NewRequestProcessor(server2, nil, "test")

	req.Data = rsp.Data
	rp2.Decrypt(req, &rsp)
	if rsp.Data != "test" {
		t.Errorf("Decrypt error %s != %s", rsp.Data, req.Data)
	}
}

func TestLoginEmul2(t *testing.T) {
	server2, _ := NewServer("123", "RV0JOcxpIe7cKrFluYpqaG/QIzx036Ea3CNW8wdrOQI8WrEnTJeAAaAGa7w1sejmCF+MqjT0uWidZb1/O6T4Ryx75ZmslQ4iRXflzWk442wD1r5gbD0vv1hBg/H5b3m9h/zFKAor7pCBzdi2O6YQOA==")
	rp2 := NewRequestProcessor(server2, nil, "test")

	req := DataRequest{}
	req.Data = "31kkFfT22SpohcRmgvzXjCMLsqp+NtKctwRcJA8aY4iZvJitw36wxyjt3boDTjG1pQS1oWVouGFSXfvQ7Ecqcn90CxbfGhxdD4I/4BeBaKArbk90K0RgtHHApuWQ7NT9y0fU5r1RCZY1RjDBPYxI8tLvPtKhEudq4Del0If4Ae4rHHiGAbZO6KhmMFPeuXn5uFPG3ne8R+QUvaDmotGXcKT8+x51yRGbsTp2XBW4y6mCpdFdbxkHIIv2RSt6p1fthd/PQZ6MFl4EKkT11MediJqM5hhhiWGpIGrhaeQ8elGn0vcdkdoD"
	var rsp DataResponse
	rp2.Decrypt(req, &rsp)
	if rsp.Data != `{"additionalscopes":[],"authapi":"","callbackurl":"http://localhost","clientid":"04dd1960","clientsecret":"ff7b40cee9ab8770706f2f5fcb00ab8e","form":null,"insecure":false,"passwordgrant":true,"profile":"rhqa","tokenapi":"","url":""}` {
		t.Errorf("Decrypt error %s != %s", rsp.Data, req.Data)
	}
}
