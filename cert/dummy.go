package cert

import "v/model"

// DummyCertManager 是一个空实现的证书管理器
type DummyCertManager struct{}

// 实现必要的接口方法
func (m *DummyCertManager) GetCertificate(domain string) (*model.Certificate, error) {
	return &model.Certificate{
		Domain:   domain,
		CertFile: "/dummy/path/cert.pem",
		KeyFile:  "/dummy/path/key.pem",
	}, nil
}

func (m *DummyCertManager) CreateCertificate(domain string) (*model.Certificate, error) {
	return &model.Certificate{
		Domain:   domain,
		CertFile: "/dummy/path/cert.pem",
		KeyFile:  "/dummy/path/key.pem",
	}, nil
}

func (m *DummyCertManager) RenewCertificate(domain string) error {
	return nil
}

func (m *DummyCertManager) DeleteCertificate(domain string) error {
	return nil
}

func (m *DummyCertManager) ListCertificates() []*model.Certificate {
	return []*model.Certificate{}
}
