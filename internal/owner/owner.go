package owner

import (
	"crypto/sha256"
	"dataStore/internal/algorithm/encoding"
	"encoding/json"
	"fmt"
	"time"
)

type (
	Token struct {
		valid  bool
		Bucket string `json:"bucket"`
		Expire int64  `json:"expire"`
		Owner  *Owner `json:"owner"`
	}

	Owner struct {
		Name        string       `json:"name"`
		Email       string       `json:"email"`
		Permissions *Permissions `json:"permission"`
	}

	Permissions struct {
		Admin  bool `json:"admin"`
		Edit   bool `json:"edit"`
		Delete bool `json:"delete"`
	}
)

func NewToken(user, pass, email string, expire int64) (*Token, error) {
	t := &Token{
		Expire: expire,
		Owner: &Owner{
			Name:        user,
			Email:       email,
			Permissions: &Permissions{},
		},
	}

	// search for the user in the database to get the permissions and set them

	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(data)
	h.Write([]byte(user + pass))

	t.Bucket = fmt.Sprintf("%x-%x-%x-%x-%x", h.Sum(nil)[:4], h.Sum(nil)[4:6], h.Sum(nil)[6:8], h.Sum(nil)[8:10], h.Sum(nil)[10:16])

	return t, nil
}

func (t *Token) Encode() (string, error) {
	return encoding.SerializeStructCBOR(t)
}

func (t *Token) Decode(data string) error {
	if err := encoding.DeserializeStructCBOR(data, t); err != nil {
		return err
	}

	t.valid = true
	return nil
}

func (t *Token) IsValid() bool {
	if t.Bucket == "" {
		t.valid = false
	}

	if t.Expire < time.Now().Unix() {
		t.valid = false
	}

	//if t.Owner == nil {
	//	t.valid = false
	//}

	t.valid = true

	return t.valid
}

func (t *Token) GetBucket() string {
	return t.Bucket
}

func (t *Token) GetExpire() int64 {
	return t.Expire
}

func (t *Token) GetOwner() *Owner {
	return t.Owner
}

func (o *Owner) GetName() string {
	return o.Name
}

func (o *Owner) GetEmail() string {
	return o.Email
}

func (o *Owner) GetPermissions() *Permissions {
	return o.Permissions
}

func (p *Permissions) GetAdmin() bool {
	return p.Admin
}

func (p *Permissions) GetEdit() bool {
	return p.Edit
}

func (p *Permissions) GetDelete() bool {
	return p.Delete
}

func (p *Permissions) GetPermissions() *Permissions {
	return p
}

func (p *Permissions) SetAdmin(admin bool) {
	p.Admin = admin
}

func (p *Permissions) SetEdit(edit bool) {
	p.Edit = edit
}

func (p *Permissions) SetDelete(delete bool) {
	p.Delete = delete
}

func (o *Owner) SetName(name string) {
	o.Name = name
}

func (o *Owner) SetEmail(email string) {
	o.Email = email
}

func (o *Owner) SetPermissions(permissions *Permissions) {
	o.Permissions = permissions
}
