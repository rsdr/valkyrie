package storagetest

import (
	"reflect"
	"testing"
	"time"

	radio "github.com/R-a-dio/valkyrie"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/arbitrary"
	"github.com/leanovate/gopter/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// OneOff generates a single value from the gopter.Gen given
func OneOff[T any](gen gopter.Gen) T {
	const maxOneOffTries = 100

	pars := gopter.DefaultGenParameters()
	var res *gopter.GenResult

	for i := 0; i < maxOneOffTries; i++ {
		res = gen(pars)
		if res.Result != nil {
			break
		}
	}
	if res.Result == nil {
		panic("didn't get a non-nil value from gopter.Gen after max tries")
	}

	return res.Result.(T)
}

func LimitString(size int) func(string) bool {
	return func(s string) bool {
		return len(s) < size
	}
}

func Positive(id radio.UserID) bool {
	return id > 0
}

func genUser() gopter.Gen {
	arbitraries := arbitrary.DefaultArbitraries()

	return gen.Struct(reflect.TypeOf(radio.User{}), map[string]gopter.Gen{
		"ID":              arbitraries.GenForType(reflect.TypeOf(radio.UserID(0))).SuchThat(Positive),
		"Username":        gen.AlphaString().SuchThat(LimitString(50)),
		"Password":        gen.AlphaString().SuchThat(LimitString(120)),
		"Email":           gen.RegexMatch(`\w+@\w+\.\w{2,25}`).SuchThat(LimitString(255)),
		"IP":              gen.RegexMatch(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
		"UpdatedAt":       genTimePtr(),
		"CreatedAt":       genTime(),
		"UserPermissions": genUserPermissions(),
	})
}

// generates a radio.UserPermissions with random permissions generated by genUserPermission
func genUserPermissions() gopter.Gen {
	g := gen.MapOf(genUserPermission(), gen.Bool())
	return gopter.Gen(func(gp *gopter.GenParameters) *gopter.GenResult {
		res := g(gp)
		actual := make(map[radio.UserPermission]struct{})
		m := res.Result.(map[radio.UserPermission]bool)
		for perm, ok := range m {
			if ok {
				actual[perm] = struct{}{}
			}
		}
		res.Result = actual
		res.ResultType = reflect.TypeOf(res.Result)
		return res
	}).WithShrinker(nil)
}

// generates one of the radio.UserPermission
func genUserPermission() gopter.Gen {
	all := radio.AllUserPermissions()
	in := make([]any, len(all))
	for i := 0; i < len(all); i++ {
		in[i] = all[i]
	}
	return gen.OneConstOf(in...)
}

func genTime() gopter.Gen {
	return gen.TimeRange(ourStartTime, time.Hour*24*3500).WithShrinker(nil)
}

// genTimePtr is genTime but returning *time.Time
func genTimePtr() gopter.Gen {
	g := gen.TimeRange(ourStartTime, time.Hour*24*3500).WithShrinker(nil)
	return func(gp *gopter.GenParameters) *gopter.GenResult {
		res := g(gp)
		t := res.Result.(time.Time)
		res.Result = &t
		res.ResultType = reflect.TypeOf(res.Result)
		return res
	}
}

func (suite *Suite) TestUserCreate(t *testing.T) {
	us := suite.Storage(t).User(suite.ctx)

	user := testUser
	// Create should not be creating a DJ even if one is present
	// so add one here so we can see if it has been added later
	user.DJ = testDJ

	uid, err := us.Create(user)
	require.NoError(t, err, "expected no error")
	require.NotZero(t, uid, "expected new user id back")

	other, err := us.Get(user.Username)
	require.NoError(t, err, "expected no error")
	require.NotNil(t, other, "expected user back")
	require.Zero(t, other.DJ, "expected no DJ")

	assert.Equal(t, user.Username, other.Username)
	assert.Equal(t, user.Password, other.Password)
	assert.Equal(t, user.Email, other.Email)
	assert.Equal(t, user.UserPermissions, other.UserPermissions)
}

func (suite *Suite) TestUserCreateDJ(t *testing.T) {
	us := suite.Storage(t).User(suite.ctx)

	user := testUser
	user.DJ = testDJ

	uid, err := us.Create(user)
	require.NoError(t, err, "expected no error")
	require.NotZero(t, uid, "expected new user id back")
	user.ID = uid

	djid, err := us.CreateDJ(user, user.DJ)
	require.NoError(t, err, "expected no error")
	require.NotZero(t, djid, "expected new dj id back")
	user.DJ.ID = djid

	other, err := us.Get(user.Username)
	require.NoError(t, err, "expected no error")
	require.NotNil(t, other, "expected user back")

	assert.Equal(t, user.Username, other.Username)
	assert.Equal(t, user.Password, other.Password)
	assert.Equal(t, user.Email, other.Email)
	assert.Equal(t, user.UserPermissions, other.UserPermissions)
	assert.Equal(t, user.DJ, other.DJ)
}

func (suite *Suite) TestUserUpdate(t *testing.T) {
	us := suite.Storage(t).User(suite.ctx)

	user := testUser
	user.DJ = testDJ

	uid, err := us.Create(user)
	require.NoError(t, err, "expected no error")
	require.NotZero(t, uid, "expected new user id back")
	user.ID = uid

	djid, err := us.CreateDJ(user, user.DJ)
	require.NoError(t, err, "expected no error")
	require.NotZero(t, djid, "expected new dj id back")
	user.DJ.ID = djid

	other, err := us.Get(user.Username)
	require.NoError(t, err, "expected no error")
	require.NotNil(t, other, "expected user back")

	assert.Equal(t, user.Username, other.Username)
	assert.Equal(t, user.Password, other.Password)
	assert.Equal(t, user.Email, other.Email)
	assert.Equal(t, user.UserPermissions, other.UserPermissions)
	assert.Equal(t, user.DJ, other.DJ)

	user.Email = "otherexample@example.com"
	user.DJ.Role = "dev"

	updated, err := us.Update(user)
	require.NoError(t, err, "expected no error")
	require.NotZero(t, updated, "expected user back")

	assert.Equal(t, user.Email, updated.Email)
	assert.Equal(t, user.DJ.Role, updated.DJ.Role)

	updatedGet, err := us.Get(user.Username)
	require.NoError(t, err, "expected no error")
	require.NotNil(t, updatedGet, "expected user back")

	assert.Equal(t, user.Email, updatedGet.Email)
	assert.Equal(t, user.DJ.Role, updatedGet.DJ.Role)
}

var testUser = radio.User{
	Username: "me",
	Password: "not a real password",
	Email:    "example@example.com",
	IP:       "127.0.0.1",
	UserPermissions: radio.UserPermissions{
		radio.PermAdmin:  struct{}{},
		radio.PermActive: struct{}{},
	},
}

var testDJ = radio.DJ{
	Name:     "testing dj",
	Regex:    "test(ing)? dj",
	Text:     "We are testing here",
	Image:    "none",
	Visible:  true,
	Priority: 500,
	Role:     "staff",
	CSS:      "unused",
	Color:    "also unused",
}
