package pkg

import (
	sha12 "crypto/sha1"
	"fmt"
	"hash"
	"io"
)

type Crypt struct {
	salt string
	sha1 hash.Hash
}

func NewCrypt(salt string) *Crypt {
	return &Crypt{
		sha1: sha12.New(),
		salt: salt,
	}
}

func (c *Crypt) CryptWithSalt(value string) string {
	_, err := io.WriteString(c.sha1, fmt.Sprintf("%s%s", c.salt, value))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", c.sha1.Sum(nil))
}

func (c *Crypt) Crypt(value string) string {
	_, err := io.WriteString(c.sha1, value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", c.sha1.Sum(nil))
}

func (c *Crypt) Renew() *Crypt {
	c.sha1 = sha12.New()
	return c
}
