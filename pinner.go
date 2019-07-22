package gokhttp

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"strings"
)

func (p *SSLPinner) GetPinsForHost(hostname string) (*SSLPin, error) {
	if pin, ok := p.SSLPins[hostname]; ok {
		return &pin, nil
	}
	return nil, errors.New("host not found")
}

func (p *SSLPinner) AddPin(hostname string, skipCA bool, pins ...string) error {
	pinO := SSLPin{Hostname: hostname, SkipCA: skipCA, Pins: []string{}}
	for _, pin := range pins {
		step := strings.Split(pin, "\\")
		if pinO.Algorithm == "" {
			pinO.Algorithm = step[0]
		}
		if step[0] != pinO.Algorithm {
			return errors.New("unmatched algorithm")
		}
		pinO.Pins = append(pinO.Pins, step[1])
	}
	p.SSLPins[hostname] = pinO
	return nil
}

func MakeDialer(pinner SSLPinner) Dialer {
	return func(network, addr string) (net.Conn, error) {
		hostname := strings.Split(addr, ":")[0]
		pins, err := pinner.GetPinsForHost(hostname)
		if err != nil {
			if err.Error() != "host not found" {
				panic(err)
			}
			return tls.Dial(network, addr, nil)
		} else {
			var acquiredPins []string
			acquiredPins = append(acquiredPins, "Chain for "+hostname+":")
			c, err := tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: pins.SkipCA})
			if err != nil {
				return c, err
			}
			connState := c.ConnectionState()
			keyPinValid := false
			for _, peerCert := range connState.PeerCertificates {
				var hash []byte
				switch pins.Algorithm {
				case "sha256":
					result := sha256.Sum256(peerCert.RawSubjectPublicKeyInfo)
					hash = result[:]
				case "sha1":
					result := sha1.Sum(peerCert.RawSubjectPublicKeyInfo)
					hash = result[:]
				default:
					panic("Unsupported algorithm")
				}
				acquiredPin := base64.StdEncoding.EncodeToString(hash)
				var pinMessage []string
				if len(peerCert.Subject.Country) > 0 {
					pinMessage = append(pinMessage, "C="+strings.Join(peerCert.Subject.Country, " "))
				}
				if len(peerCert.Subject.Province) > 0 {
					pinMessage = append(pinMessage, "ST="+strings.Join(peerCert.Subject.Province, " "))
				}
				if len(peerCert.Subject.Locality) > 0 {
					pinMessage = append(pinMessage, "L="+strings.Join(peerCert.Subject.Locality, " "))
				}
				if len(peerCert.Subject.Organization) > 0 {
					pinMessage = append(pinMessage, "O="+strings.Join(peerCert.Subject.Organization, " "))
				}
				if len(peerCert.Subject.OrganizationalUnit) > 0 {
					pinMessage = append(pinMessage, "OU="+strings.Join(peerCert.Subject.OrganizationalUnit, " "))
				}
				if peerCert.Subject.CommonName != "" {
					pinMessage = append(pinMessage, "CN="+peerCert.Subject.CommonName)
				}
				acquiredPins = append(acquiredPins, strings.Join(pinMessage, "/")+":\n\t"+acquiredPin+"\t(valid from "+peerCert.NotBefore.String()+" until "+peerCert.NotAfter.String()+")")
				for _, pin := range pins.Pins {
					if pin == acquiredPin {
						keyPinValid = true
						break
					}
				}
				if keyPinValid {
					break
				}
			}
			if keyPinValid == false {
				return c, errors.New("Insecure connection detected\n" + strings.Join(acquiredPins, "\n"))
			}
			return c, nil
		}
	}
}

func GetSSLPinner() SSLPinner {
	return SSLPinner{SSLPins: map[string]SSLPin{}}
}
