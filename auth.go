package cosmosdb

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

type AuthorizationPayload struct {
	Verb         string
	ResourceType string
	ResourceLink string
	Date         string
}

// stringToSign constructs the string to be signed from an `AuthorizationPayload`
// struct. The generated string only works with the addressing by user ids, as
// we use in this package. Addressing with self links requires different capitalization.
func stringToSign(p AuthorizationPayload) string {
	return strings.ToLower(p.Verb) + "\n" +
		strings.ToLower(p.ResourceType) + "\n" +
		p.ResourceLink + "\n" +
		strings.ToLower(p.Date) + "\n" +
		"" + "\n"
}

func makeSignedPayload(verb, link, date, key string) (string, error) {
	if strings.HasPrefix(link, "/") == true {
		link = link[1:]
	}

	rLink, rType := resourceTypeFromLink(verb, link)

	pl := AuthorizationPayload{
		Verb:         verb,
		ResourceType: rType,
		ResourceLink: rLink,
		Date:         date,
	}

	s := stringToSign(pl)
	fmt.Printf("payload to sign: %s\n", s)

	return authorize(s, key)
}

func makeAuthHeader(sPayload string) string {
	masterToken := "master"
	tokenVersion := "1.0"
	return url.QueryEscape(
		"type=" + masterToken + "&ver=" + tokenVersion + "&sig=" + sPayload,
	)
}

func authorize(str, key string) (string, error) {
	var ret string
	enc := base64.StdEncoding
	salt, err := enc.DecodeString(key)
	if err != nil {
		return ret, err
	}
	hmac := hmac.New(sha256.New, salt)
	hmac.Write([]byte(str))
	b := hmac.Sum(nil)

	ret = enc.EncodeToString(b)
	return ret, nil
}
