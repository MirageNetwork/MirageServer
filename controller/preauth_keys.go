package controller

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	ErrPreAuthKeyNotFound          = Error("AuthKey not found")
	ErrPreAuthKeyExpired           = Error("AuthKey expired")
	ErrSingleUseAuthKeyHasBeenUsed = Error("AuthKey has already been used")
	ErrUserMismatch                = Error("user mismatch")
	ErrPreAuthKeyACLTagInvalid     = Error("AuthKey tag is invalid")
)

// PreAuthKey describes a pre-authorization key usable in a particular user.
type PreAuthKey struct {
	ID        uint64 `gorm:"primary_key"`
	Key       string
	UserID    int64
	User      User
	Reusable  bool
	Ephemeral bool `gorm:"default:false"`
	Used      bool `gorm:"default:false"`
	ACLTags   StringList

	CreatedAt  *time.Time
	Expiration *time.Time
}

/*
// PreAuthKeyACLTag describes an autmatic tag applied to a node when registered with the associated PreAuthKey.
type PreAuthKeyACLTag struct {
	ID           uint64 `gorm:"primary_key"`
	PreAuthKeyID uint64
	Tag          string
}
*/

// CreatePreAuthKey creates a new PreAuthKey in a user, and returns it.
func (h *Mirage) CreatePreAuthKey(
	user *User,
	reusable bool,
	ephemeral bool,
	expiration *time.Time,
	aclTags []string,
) (*PreAuthKey, error) {

	for _, tag := range aclTags {
		if !strings.HasPrefix(tag, "tag:") {
			return nil, fmt.Errorf("%w: '%s' did not begin with 'tag:'", ErrPreAuthKeyACLTagInvalid, tag)
		}
	}

	now := time.Now().UTC()
	kstr, err := h.generateKey()
	if err != nil {
		return nil, err
	}

	key := PreAuthKey{
		Key:        kstr,
		UserID:     user.ID,
		User:       *user,
		Reusable:   reusable,
		Ephemeral:  ephemeral,
		CreatedAt:  &now,
		Expiration: expiration,
	}

	err = h.db.Transaction(func(db *gorm.DB) error {
		if len(aclTags) > 0 {
			tagSet := map[string]struct{}{}
			tagSlice := []string{}

			for _, tag := range aclTags {
				if _, ok := tagSet[tag]; !ok {
					tagSlice = append(tagSlice, tag)
					tagSet[tag] = struct{}{}
				}
			}
			key.ACLTags = StringList(tagSlice)
		}
		if err := db.Save(&key).Error; err != nil {
			return fmt.Errorf("failed to create key in the database: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &key, nil
}

// ListPreAuthKeys returns the list of PreAuthKeys for a user.
func (h *Mirage) ListPreAuthKeys(userID int64) ([]PreAuthKey, error) {
	keys := []PreAuthKey{}
	if err := h.db.Preload("User").Where(&PreAuthKey{UserID: userID}).Find(&keys).Error; err != nil {
		return nil, err
	}

	return keys, nil
}

// GetPreAuthKey returns a PreAuthKey for a given key.
func (h *Mirage) GetPreAuthKey(user string, key string) (*PreAuthKey, error) {
	pak, err := h.checkKeyValidity(key)
	if err != nil {
		return nil, err
	}

	if pak.User.Name != user {
		return nil, ErrUserMismatch
	}

	return pak, nil
}

// DestroyPreAuthKey destroys a preauthkey. Returns error if the PreAuthKey
// does not exist.
func (h *Mirage) DestroyPreAuthKey(pak PreAuthKey) error {
	return h.db.Transaction(func(db *gorm.DB) error {
		if result := db.Unscoped().Delete(pak); result.Error != nil {
			return result.Error
		}

		return nil
	})
}

// MarkExpirePreAuthKey marks a PreAuthKey as expired.
func (h *Mirage) ExpirePreAuthKey(k *PreAuthKey) error {
	if err := h.db.Model(&k).Update("Expiration", time.Now()).Error; err != nil {
		return err
	}

	return nil
}

// UsePreAuthKey marks a PreAuthKey as used.
func (h *Mirage) UsePreAuthKey(k *PreAuthKey) error {
	k.Used = true
	if err := h.db.Save(k).Error; err != nil {
		return fmt.Errorf("failed to update key used status in the database: %w", err)
	}

	return nil
}

// checkKeyValidity does the heavy lifting for validation of the PreAuthKey coming from a node
// If returns no error and a PreAuthKey, it can be used.
func (h *Mirage) checkKeyValidity(k string) (*PreAuthKey, error) {
	pak := PreAuthKey{}
	if result := h.db.Preload("User").First(&pak, "key = ?", k); errors.Is(
		result.Error,
		gorm.ErrRecordNotFound,
	) {
		return nil, ErrPreAuthKeyNotFound
	}

	if pak.Expiration != nil && pak.Expiration.Before(time.Now()) {
		return nil, ErrPreAuthKeyExpired
	}

	if pak.Reusable { // cgao6: 依据TS逻辑，自熄并不影响是否可重用|| pak.Ephemeral { // we don't need to check if has been used before
		return &pak, nil
	}

	machines := []Machine{}
	if err := h.db.Preload("AuthKey").Where(&Machine{AuthKeyID: uint(pak.ID)}).Find(&machines).Error; err != nil {
		return nil, err
	}

	if len(machines) != 0 || pak.Used {
		return nil, ErrSingleUseAuthKeyHasBeenUsed
	}

	return &pak, nil
}

func (h *Mirage) generateKey() (string, error) {
	size := 24
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (key *PreAuthKey) GetAclTags() []string {
	aclTags := make([]string, len(key.ACLTags))
	copy(aclTags, key.ACLTags)
	return aclTags
}
