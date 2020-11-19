package orm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	math "math"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/rbac"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/trustelem/zxcvbn"
)

var BadAuthDelay = 3 * time.Second
var NoOTP = ""
var NoApiKeyId = ""
var NoApiKey = ""
var OTPLen otp.Digits = otp.DigitsSix
var OTPExpirationTime = uint(2 * 60) // seconds
var OTPExpirationTimeStr = "2 minutes"

// Init admin creates the admin user and adds the admin role.
func InitAdmin(ctx context.Context, superuser, superpass string) error {
	log.SpanLog(ctx, log.DebugLevelApi, "init admin")

	// create superuser if it doesn't exist
	passhash, salt, iter := NewPasshash(superpass)
	super := ormapi.User{
		Name:          superuser,
		Email:         superuser + "@mobiledgex.net",
		EmailVerified: true,
		Passhash:      passhash,
		Salt:          salt,
		Iter:          iter,
		GivenName:     superuser,
		FamilyName:    superuser,
		Nickname:      superuser,
	}

	db := loggedDB(ctx)
	err := db.FirstOrCreate(&super, &ormapi.User{Name: superuser}).Error
	if err != nil {
		return err
	}

	// set role of superuser to admin manager
	err = enforcer.AddGroupingPolicy(ctx, super.Name, RoleAdminManager)
	if err != nil {
		return err
	}
	return nil
}

func Login(c echo.Context) error {
	ctx := GetContext(c)
	login := ormapi.UserLogin{}
	if err := c.Bind(&login); err != nil {
		return bindErr(c, err)
	}
	if login.Username == "" {
		return c.JSON(http.StatusBadRequest, Msg("Username not specified"))
	}

	if login.Password != "" && login.ApiKey != "" {
		return c.JSON(http.StatusBadRequest, Msg("Please specify either password or apikeyid/apikey"))
	}
	if login.ApiKey != "" && login.ApiKeyId == "" {
		return c.JSON(http.StatusBadRequest, Msg("Missing apikeyid"))
	}
	user := ormapi.User{}
	lookup := ormapi.User{Name: login.Username}
	db := loggedDB(ctx)
	res := db.Where(&lookup).First(&user)
	if res.RecordNotFound() {
		// try look-up by email
		lookup.Name = ""
		lookup.Email = login.Username
		res = db.Where(&lookup).First(&user)
	}
	err := res.Error
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "user lookup failed", "lookup", lookup, "err", err)
		time.Sleep(BadAuthDelay)
		return c.JSON(http.StatusBadRequest, Msg("Invalid username or password"))
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("username", user.Name)
	span.SetTag("email", user.Email)

	if login.ApiKey == "" {
		matches, err := PasswordMatches(login.Password, user.Passhash, user.Salt, user.Iter)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "password matches err", "err", err)
		}
		if !matches || err != nil {
			time.Sleep(BadAuthDelay)
			return c.JSON(http.StatusBadRequest, Msg("Invalid username or password"))
		}
	} else {
		// Validate apiKey
		apiKeys, err := UnmarshalApiKeys(user.ApiKeys)
		if err != nil {
			return err
		}
		apiKey, ok := apiKeys[login.ApiKeyId]
		if !ok {
			time.Sleep(BadAuthDelay)
			return c.JSON(http.StatusBadRequest, Msg("Invalid token id"))
		}
		matches, err := PasswordMatches(login.ApiKey, apiKey.ApiKeyHash, apiKey.ApiKeySalt, apiKey.ApiKeyIter)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "api key matches err", "err", err)
		}
		if !matches || err != nil {
			time.Sleep(BadAuthDelay)
			return c.JSON(http.StatusBadRequest, Msg("Invalid username or apikey"))
		}

	}
	if user.Locked {
		return c.JSON(http.StatusBadRequest, Msg("Account is locked, please contact MobiledgeX support"))
	}
	if !getSkipVerifyEmail(ctx, nil) && !user.EmailVerified {
		return c.JSON(http.StatusBadRequest, Msg("Email not verified yet"))
	}

	if login.Password != "" && user.PassCrackTimeSec == 0 {
		calcPasswordStrength(ctx, &user, login.Password)
		isAdmin, err := isUserAdmin(ctx, user.Name)
		if err != nil {
			return setReply(c, err, nil)
		}
		err = checkPasswordStrength(ctx, &user, nil, isAdmin)
		if err != nil {
			if isAdmin {
				time.Sleep(BadAuthDelay)
				return c.JSON(http.StatusBadRequest, Msg("Existing password for Admin too weak, please update first"))
			} else {
				// log warning for now
				log.SpanLog(ctx, log.DebugLevelApi, "user password strength check failure", "user", user.Name, "err", err)
			}
		}
		// save password strength
		err = db.Model(&user).Updates(&user).Error
		if err != nil {
			return setReply(c, dbErr(err), nil)
		}
	}

	if login.Password != "" && user.TOTPSharedKey != "" {
		if login.TOTP == "" {
			// Send OTP over email
			otp, err := totp.GenerateCode(user.TOTPSharedKey, time.Now().UTC())
			if err != nil {
				return setReply(c, err, nil)
			}
			err = sendOTPEmail(ctx, user.Name, user.Email, otp, OTPExpirationTimeStr)
			if err != nil {
				// log and ignore
				log.SpanLog(ctx, log.DebugLevelApi, "failed to send otp email", "err", err)
			}
			return c.JSON(http.StatusNetworkAuthenticationRequired, Msg("Missing OTP\nPlease use two factor authenticator app on "+
				"your phone to get OTP. We have also sent OTP to your registered email address"))
		}
		valid := totp.Validate(login.TOTP, user.TOTPSharedKey)
		if !valid {
			return c.JSON(http.StatusBadRequest, Msg("Invalid OTP"))
		}
	}

	cookie, err := GenerateCookie(&user)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to generate cookie", "err", err)
		return c.JSON(http.StatusBadRequest, Msg("Failed to generate cookie"))
	}
	return c.JSON(http.StatusOK, M{"token": cookie})
}

func RefreshAuthCookie(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)
	if claims.FirstIssuedAt == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to generate cookie as issued time is missing")
		return c.JSON(http.StatusBadRequest, Msg("Failed to refresh auth cookie"))
	}
	// refresh auth cookie only if it was issued within 30 days
	if time.Unix(claims.FirstIssuedAt, 0).AddDate(0, 0, 30).Unix() < time.Now().Unix() {
		return c.JSON(http.StatusUnauthorized, Msg("expired jwt"))
	}
	claims.StandardClaims.IssuedAt = time.Now().Unix()
	claims.StandardClaims.ExpiresAt = time.Now().AddDate(0, 0, 1).Unix()
	cookie, err := Jwks.GenerateCookie(claims)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to generate cookie", "err", err)
		return c.JSON(http.StatusBadRequest, Msg("Failed to generate cookie"))
	}
	return c.JSON(http.StatusOK, M{"token": cookie})
}

func GenerateTOTPQR(accountName string) (string, []byte, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "MobiledgeX",
		AccountName: accountName,
		Period:      OTPExpirationTime,
		Digits:      OTPLen,
	})
	if err != nil {
		return "", nil, err
	}

	// Convert TOTP key into a QR code encoded as a PNG image.
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return "", nil, err
	}
	err = png.Encode(&buf, img)
	if err != nil {
		return "", nil, err
	}

	return key.Secret(), buf.Bytes(), nil
}

func CreateUser(c echo.Context) error {
	ctx := GetContext(c)
	createuser := ormapi.CreateUser{}
	if err := c.Bind(&createuser); err != nil {
		return bindErr(c, err)
	}
	user := createuser.User
	if user.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("Name not specified"))
	}
	err := ValidName(user.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Msg(err.Error()))
	}
	if !util.ValidEmail(user.Email) {
		return c.JSON(http.StatusBadRequest, Msg("Invalid email address"))
	}
	if err := ValidPassword(user.Passhash); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid password, "+
			err.Error()))
	}
	orgT, err := GetAllOrgs(ctx)
	if err == nil {
		for orgName, _ := range orgT {
			if strings.ToLower(user.Name) == strings.ToLower(orgName) {
				return c.JSON(http.StatusBadRequest, Msg("user name cannot be same as org name"))
			}
		}
	}

	config, err := getConfig(ctx)
	if err != nil {
		return err
	}
	calcPasswordStrength(ctx, &user, user.Passhash)
	// check password strength (new users are never admins)
	isAdmin := false
	err = checkPasswordStrength(ctx, &user, config, isAdmin)
	if err != nil {
		return c.JSON(http.StatusBadRequest, MsgErr(err))
	}

	if !getSkipVerifyEmail(ctx, config) {
		// real email will be filled in later
		createuser.Verify.Email = "dummy@dummy.com"
		err := ValidEmailRequest(c, &createuser.Verify)
		if err != nil {
			return c.JSON(http.StatusBadRequest, MsgErr(err))
		}
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("username", user.Name)
	span.SetTag("email", user.Email)

	user.Locked = false
	if config.LockNewAccounts {
		user.Locked = true
	}
	user.EmailVerified = false

	userResponse := ormapi.UserResponse{}
	if user.EnableTOTP {
		totpKey, totpQR, err := GenerateTOTPQR(user.Email)
		if err != nil {
			return setReply(c, fmt.Errorf("Failed to setup 2FA: %v", err), nil)
		}
		user.TOTPSharedKey = totpKey

		userResponse.TOTPSharedKey = totpKey
		userResponse.TOTPQRImage = totpQR
	}

	// password should be passed through in Passhash field.
	user.Passhash, user.Salt, user.Iter = NewPasshash(user.Passhash)
	db := loggedDB(ctx)
	if err := db.Create(&user).Error; err != nil {
		//check specifically for duplicate username and/or emails
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_pkey\"" {
			return setReply(c, fmt.Errorf("Username with name %s (case-insensitive) already exists", user.Name), nil)
		}
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return setReply(c, fmt.Errorf("Email already in use"), nil)
		}

		return setReply(c, dbErr(err), nil)
	}
	createuser.Verify.Email = user.Email
	err = sendVerifyEmail(ctx, user.Name, &createuser.Verify)
	if err != nil {
		db.Delete(&user)
		return err
	}

	gitlabCreateLDAPUser(ctx, &user)
	artifactoryCreateUser(ctx, &user)

	if user.Locked {
		msg := fmt.Sprintf("Locked account created for user %s, email %s", user.Name, user.Email)
		// just log in case of error
		senderr := sendNotify(ctx, config.NotifyEmailAddress,
			"Locked account created", msg)
		if senderr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "failed to send notify of new locked account", "err", senderr)
		}
	}

	if user.TOTPSharedKey != "" {
		userResponse.Message = "User created with two factor authentication enabled. " +
			"Please use the following text code with the two factor authentication app on your " +
			"phone to set it up"
	} else {
		userResponse.Message = "user created"
	}
	return c.JSON(http.StatusOK, &userResponse)
}

func ResendVerify(c echo.Context) error {
	ctx := GetContext(c)

	req := ormapi.EmailRequest{}
	if err := c.Bind(&req); err != nil {
		return bindErr(c, err)
	}
	if err := ValidEmailRequest(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, MsgErr(err))
	}
	return sendVerifyEmail(ctx, "MobiledgeX user", &req)
}

func VerifyEmail(c echo.Context) error {
	ctx := GetContext(c)
	tok := ormapi.Token{}
	if err := c.Bind(&tok); err != nil {
		return bindErr(c, err)
	}
	claims := EmailClaims{}
	token, err := Jwks.VerifyCookie(tok.Token, &claims)
	if err != nil || !token.Valid {
		return &echo.HTTPError{
			Code:     http.StatusUnauthorized,
			Message:  "invalid or expired token",
			Internal: err,
		}
	}
	user := ormapi.User{Email: claims.Email}
	db := loggedDB(ctx)
	err = db.Where(&user).First(&user).Error
	if err != nil {
		// user got deleted in the meantime?
		return nil
	}
	span := log.SpanFromContext(ctx)
	span.SetTag("username", user.Name)

	user.EmailVerified = true
	if err := db.Model(&user).Updates(&user).Error; err != nil {
		return setReply(c, dbErr(err), nil)
	}
	return c.JSON(http.StatusOK, Msg("email verified, thank you"))
}

func DeleteUser(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)

	user := ormapi.User{}
	if err := c.Bind(&user); err != nil {
		return bindErr(c, err)
	}
	if user.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("User Name not specified"))
	}
	// Only user themself or super-user can delete user.
	if user.Name != claims.Username {
		if err := authorized(ctx, claims.Username, "", ResourceUsers, ActionManage); err != nil {
			return err
		}
	}
	if user.Name == Superuser {
		return c.JSON(http.StatusBadRequest, Msg("Cannot delete superuser"))
	}

	// delete role mappings
	groups, err := enforcer.GetGroupingPolicy()
	if err != nil {
		return dbErr(err)
	}
	// check role mappings first before deleting
	// need to make sure we are not deleting the last manager from an org or deleting the last AdminManager
	managerCounts := make(map[string]int)
	var userOrgs []string // orgs for which the user is a manager of
	for _, grp := range groups {
		if len(grp) < 2 {
			continue
		}
		strs := strings.Split(grp[0], "::")
		if grp[1] == RoleAdminManager || grp[1] == RoleDeveloperManager || grp[1] == RoleOperatorManager {
			org := ""
			username := grp[0]
			if len(strs) == 2 {
				org = strs[0]
				username = strs[1]
			}
			managerCounts[org] = managerCounts[org] + 1
			if username == user.Name {
				userOrgs = append(userOrgs, org)
			}
		}
	}
	for _, org := range userOrgs {
		if managerCounts[org] < 2 {
			if org == "" {
				err = fmt.Errorf("Error: Cannot delete the last remaining AdminManager")
			} else {
				err = fmt.Errorf("Error: Cannot delete the last remaining manager for the org %s", org)
			}
			return setReply(c, err, nil)
		}
	}
	for _, grp := range groups {
		if len(grp) < 2 {
			continue
		}
		strs := strings.Split(grp[0], "::")
		if grp[0] == user.Name || (len(strs) == 2 && strs[1] == user.Name) {
			err := enforcer.RemoveGroupingPolicy(ctx, grp[0], grp[1])
			if err != nil {
				return dbErr(err)
			}
		}
	}
	// delete user
	db := loggedDB(ctx)
	err = db.Delete(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	gitlabDeleteLDAPUser(ctx, user.Name)
	artifactoryDeleteUser(ctx, user.Name)

	return c.JSON(http.StatusOK, Msg("user deleted"))
}

// Show current user info
func CurrentUser(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	user := ormapi.User{Name: claims.Username}
	db := loggedDB(ctx)
	err = db.Where(&user).First(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	user.Passhash = ""
	user.Salt = ""
	user.Iter = 0
	user.TOTPSharedKey = ""
	user.ApiKeys = nil
	return c.JSON(http.StatusOK, user)
}

// Show users by Organization
func ShowUser(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	filter := ormapi.Organization{}
	if c.Request().ContentLength > 0 {
		if err := c.Bind(&filter); err != nil {
			return bindErr(c, err)
		}
	}
	users := []ormapi.User{}
	if err := authorized(ctx, claims.Username, filter.Name, ResourceUsers, ActionView); err != nil {
		if filter.Name == "" && c.Request().ContentLength == 0 {
			// user probably forgot to specify orgname
			return c.JSON(http.StatusBadRequest, Msg("No organization name specified"))
		}
		return err
	}
	// if filter ID is 0, show all users (super user only)
	db := loggedDB(ctx)
	if filter.Name == "" {
		err = db.Find(&users).Error
		if err != nil {
			return setReply(c, dbErr(err), nil)
		}
	} else {
		groupings, err := enforcer.GetGroupingPolicy()
		if err != nil {
			return dbErr(err)
		}
		for _, grp := range groupings {
			if len(grp) < 2 {
				continue
			}
			orguser := strings.Split(grp[0], "::")
			if len(orguser) > 1 && orguser[0] == filter.Name {
				user := ormapi.User{}
				user.Name = orguser[1]
				err = db.Where(&user).First(&user).Error
				if err != nil {
					return setReply(c, dbErr(err), nil)
				}
				users = append(users, user)
			}
		}
	}
	for ii, _ := range users {
		// don't show auth/private info
		users[ii].Passhash = ""
		users[ii].TOTPSharedKey = ""
		users[ii].Salt = ""
		users[ii].Iter = 0
		users[ii].ApiKeys = nil
	}
	return c.JSON(http.StatusOK, users)
}

func NewPassword(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	in := ormapi.NewPassword{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	return setPassword(c, claims.Username, in.Password)
}

func setPassword(c echo.Context, username, password string) error {
	ctx := GetContext(c)
	if err := ValidPassword(password); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid password, "+
			err.Error()))
	}
	user := ormapi.User{Name: username}
	db := loggedDB(ctx)
	err := db.Where(&user).First(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}

	calcPasswordStrength(ctx, &user, password)
	// check password strength
	isAdmin, err := isUserAdmin(ctx, user.Name)
	if err != nil {
		return setReply(c, err, nil)
	}
	err = checkPasswordStrength(ctx, &user, nil, isAdmin)
	if err != nil {
		return c.JSON(http.StatusBadRequest, MsgErr(err))
	}

	user.Passhash, user.Salt, user.Iter = NewPasshash(password)
	if err := db.Model(&user).Updates(&user).Error; err != nil {
		return setReply(c, dbErr(err), nil)
	}
	return c.JSON(http.StatusOK, Msg("password updated"))
}

func calcPasswordStrength(ctx context.Context, user *ormapi.User, password string) {
	pwscore := zxcvbn.PasswordStrength(password, []string{})
	user.PassCrackTimeSec = pwscore.Guesses / float64(BruteForceGuessesPerSecond)
}

func checkPasswordStrength(ctx context.Context, user *ormapi.User, config *ormapi.Config, isAdmin bool) error {
	if config == nil {
		var err error
		config, err = getConfig(ctx)
		if err != nil {
			return err
		}
	}
	minCrackTime := config.PasswordMinCrackTimeSec
	if isAdmin {
		minCrackTime = config.AdminPasswordMinCrackTimeSec
	}
	if user.PassCrackTimeSec < minCrackTime {
		return fmt.Errorf("Password too weak, requires crack time %s but is %s. Please increase length or complexity", secDisplayTime(minCrackTime), secDisplayTime(user.PassCrackTimeSec))
	}
	return nil
}

func PasswordResetRequest(c echo.Context) error {
	ctx := GetContext(c)
	req := ormapi.EmailRequest{}
	if err := c.Bind(&req); err != nil {
		return bindErr(c, err)
	}
	if err := ValidEmailRequest(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, MsgErr(err))
	}
	noreply, err := getNoreply(ctx)
	if err != nil {
		return err
	}

	tmpl := passwordResetNoneTmpl
	arg := emailTmplArg{
		From:    noreply.Email,
		Email:   req.Email,
		OS:      req.OperatingSystem,
		Browser: req.Browser,
		IP:      req.ClientIP,
	}
	// To ensure we do not leak user accounts, we do not
	// return an error if the user is not found. Instead, we always
	// send an email to the account specified, but the contents
	// of the email are different if the user was not found.
	user := ormapi.User{Email: req.Email}
	db := loggedDB(ctx)
	res := db.Where(&user).First(&user)
	if !res.RecordNotFound() && res.Error == nil {
		info := EmailClaims{
			StandardClaims: jwt.StandardClaims{
				IssuedAt: time.Now().Unix(),
				// 1 hour
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
			Email:    req.Email,
			Username: user.Name,
		}
		cookie, err := Jwks.GenerateCookie(&info)
		if err != nil {
			return err
		}
		if req.CallbackURL != "" {
			arg.URL = req.CallbackURL + "?token=" + cookie
		}
		arg.Name = user.Name
		arg.Token = cookie
		tmpl = passwordResetTmpl
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, &arg); err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "send password reset email",
		"from", noreply.Email, "to", req.Email)
	return sendEmail(noreply, req.Email, &buf)
}

func PasswordReset(c echo.Context) error {
	pw := ormapi.PasswordReset{}
	if err := c.Bind(&pw); err != nil {
		return bindErr(c, err)
	}
	claims := EmailClaims{}
	token, err := Jwks.VerifyCookie(pw.Token, &claims)
	if err != nil || !token.Valid {
		return &echo.HTTPError{
			Code:     http.StatusUnauthorized,
			Message:  "invalid or expired token",
			Internal: err,
		}
	}
	ctx := GetContext(c)
	span := log.SpanFromContext(ctx)
	span.SetTag("username", claims.Username)
	return setPassword(c, claims.Username, pw.Password)
}

func RestrictedUserUpdate(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	// Only admin user allowed to update user data.
	if err := authorized(ctx, claims.Username, "", ResourceUsers, ActionManage); err != nil {
		return err
	}
	// Pull json directly so we can unmarshal twice.
	// First time is to do lookup, second time is to apply
	// modified fields.
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return bindErr(c, err)
	}
	in := ormapi.User{}
	err = json.Unmarshal(body, &in)
	if err != nil {
		return bindErr(c, err)
	}
	// in may contain other fields, but can only specify
	// name and email for where clause.
	lookup := ormapi.User{
		Name:  in.Name,
		Email: in.Email,
	}
	user := ormapi.User{}
	db := loggedDB(ctx)
	res := db.Where(&lookup).First(&user)
	if res.RecordNotFound() {
		return c.JSON(http.StatusBadRequest, Msg("user not found"))
	}
	if res.Error != nil {
		return dbErr(res.Error)
	}
	saveuser := user
	// apply specified fields
	err = json.Unmarshal(body, &user)
	if err != nil {
		return bindErr(c, err)
	}
	// cannot update password or invariant fields
	user.Passhash = saveuser.Passhash
	user.Salt = saveuser.Salt
	user.Iter = saveuser.Iter
	user.CreatedAt = saveuser.CreatedAt
	user.Name = saveuser.Name
	user.Email = saveuser.Email

	err = db.Save(&user).Error
	if err != nil {
		return dbErr(err)
	}
	return nil
}

// modified version of go-zxcvbn scoring.displayTime that more closely
// matches the java script version of the lib.
func secDisplayTime(seconds float64) string {
	formater := "%.1f %s"
	minute := float64(60)
	hour := minute * float64(60)
	day := hour * float64(24)
	month := day * float64(31)
	year := month * float64(12)
	century := year * float64(100)

	if seconds < 1 {
		return "less than a second"
	} else if seconds < minute {
		return fmt.Sprintf(formater, (1 + math.Ceil(seconds)), "seconds")
	} else if seconds < hour {
		return fmt.Sprintf(formater, (1 + math.Ceil(seconds/minute)), "minutes")
	} else if seconds < day {
		return fmt.Sprintf(formater, (1 + math.Ceil(seconds/hour)), "hours")
	} else if seconds < month {
		return fmt.Sprintf(formater, (1 + math.Ceil(seconds/day)), "days")
	} else if seconds < year {
		return fmt.Sprintf(formater, (1 + math.Ceil(seconds/month)), "months")
	} else if seconds < century {
		return fmt.Sprintf(formater, (1 + math.Ceil(seconds/century)), "years")
	} else {
		return "centuries"
	}
}

func isUserAdmin(ctx context.Context, username string) (bool, error) {
	// it doesn't matter what the resource/action is here, as long
	// as it's part of one of the admin roles.
	authOrgs, err := enforcer.GetAuthorizedOrgs(ctx, username, ResourceConfig, ActionView)
	if err != nil {
		return false, dbErr(err)
	}
	if _, found := authOrgs[""]; found {
		return true, nil
	}
	return false, nil
}

func UpdateUser(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}

	cuser := ormapi.CreateUser{}
	cuser.User.Name = claims.Username
	user := &cuser.User
	db := loggedDB(ctx)
	res := db.Where(user).First(user)
	if res.RecordNotFound() {
		return c.JSON(http.StatusBadRequest, Msg("User not found"))
	}
	old := *user

	// read args onto existing data, will overwrite only specified fields
	if err := c.Bind(&cuser); err != nil {
		return bindErr(c, err)
	}
	// check for fields that are not allowed to change
	if old.Name != user.Name {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change username"))
	}
	if old.EmailVerified != user.EmailVerified {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change emailverified"))
	}
	if old.Passhash != user.Passhash {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change passhash"))
	}
	if old.Salt != user.Salt {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change salt"))
	}
	if old.Iter != user.Iter {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change iter"))
	}
	if !old.CreatedAt.Equal(user.CreatedAt) {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change createdat"))
	}
	if !old.UpdatedAt.Equal(user.UpdatedAt) {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change updatedat"))
	}
	if old.Locked != user.Locked {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change locked"))
	}
	if old.PassCrackTimeSec != user.PassCrackTimeSec {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change passcracktimesec"))
	}

	// if email changed, need to verify
	sendVerify := false
	if old.Email != user.Email {
		config, err := getConfig(ctx)
		if err != nil {
			return err
		}
		if !getSkipVerifyEmail(ctx, config) {
			err := ValidEmailRequest(c, &cuser.Verify)
			if err != nil {
				return c.JSON(http.StatusBadRequest, MsgErr(err))
			}
			sendVerify = true
		}
		user.EmailVerified = false
	}

	userResponse := ormapi.UserResponse{}
	otpChanged := false
	if old.EnableTOTP != user.EnableTOTP {
		if user.EnableTOTP {
			totpKey, totpQR, err := GenerateTOTPQR(user.Email)
			if err != nil {
				return setReply(c, fmt.Errorf("Failed to setup 2FA: %v", err), nil)
			}
			user.TOTPSharedKey = totpKey

			userResponse.TOTPSharedKey = totpKey
			userResponse.TOTPQRImage = totpQR
		} else {
			user.TOTPSharedKey = NoOTP
		}
		otpChanged = true
	}

	err = db.Save(user).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"email_pkey") {
			return setReply(c, fmt.Errorf("Email %s already in use", user.Email), nil)
		}
		return setReply(c, dbErr(err), nil)
	}

	if sendVerify {
		err = sendVerifyEmail(ctx, user.Name, &cuser.Verify)
		if err != nil {
			undoErr := db.Save(&old).Error
			if undoErr != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "undo update user failed", "user", claims.Username, "err", undoErr)
			}
			return setReply(c, fmt.Errorf("Failed to send verification email to %s, %v", user.Email, err), nil)
		}
	}

	if otpChanged && user.TOTPSharedKey != "" {
		userResponse.Message = "User updated\nEnabled two factor authentication. " +
			"Please use the following text code with the two factor authentication app on your " +
			"phone to set it up"
	} else {
		userResponse.Message = "user updated"
	}
	return c.JSON(http.StatusOK, &userResponse)
}

func MarshalApiKeys(apiKeys map[string]ormapi.ApiKey) ([]byte, error) {
	out, err := json.Marshal(apiKeys)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal api keys: %v", err)
	}
	return out, err
}

func UnmarshalApiKeys(data []byte) (map[string]ormapi.ApiKey, error) {
	apiKeys := make(map[string]ormapi.ApiKey)
	if len(data) == 0 {
		return apiKeys, nil
	}
	if err := json.Unmarshal(data, &apiKeys); err != nil {
		return nil, fmt.Errorf("Failed to decode api keys: %v", err)
	}
	return apiKeys, nil
}

func CreateUserApiKey(c echo.Context) error {
	ctx := GetContext(c)
	db := loggedDB(ctx)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	apiKeyObj := ormapi.UserApiKey{}
	if err := c.Bind(&apiKeyObj); err != nil {
		return bindErr(c, err)
	}
	user := ormapi.User{Name: claims.Username}
	err = db.Where(&user).First(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	// fetch existing apikeys
	apiKeys, err := UnmarshalApiKeys(user.ApiKeys)
	if err != nil {
		return err
	}
	groupings, err := enforcer.GetGroupingPolicy()
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	authz, err := newShowAuthz(ctx, "", claims.Username, ResourceUsers, ActionView)
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}

	var userRole *ormapi.Role
	for ii, _ := range groupings {
		role := parseRole(groupings[ii])
		if role == nil {
			continue
		}
		if !authz.Ok(role.Org) {
			continue
		}
		if role.Username != claims.Username {
			continue
		}
		if role.Org == apiKeyObj.Org {
			userRole = role
			break
		}
	}
	if userRole == nil {
		return c.JSON(http.StatusBadRequest, Msg("User has no permissions to access org "+apiKeyObj.Org))
	}
	if apiKeyObj.Role == "" {
		apiKeyObj.Role = userRole.Role
	} else {
		org := ormapi.Organization{}
		org.Name = userRole.Org
		err := db.Where(&org).First(&org).Error
		if err != nil {
			return setReply(c, dbErr(err), nil)
		}
		orgRoles := []string{}
		switch org.Type {
		case OrgTypeDeveloper:
			orgRoles = DeveloperRoles
		case OrgTypeOperator:
			orgRoles = OperatorRoles
		case OrgTypeAdmin:
			orgRoles = AdminRoles
		default:
			return c.JSON(http.StatusBadRequest, Msg("Invalid org type"))
		}
		// ensure api key role is a valid org role
		roles := make(map[string]int)
		for ii, val := range orgRoles {
			roles[val] = ii
		}
		apiKeyRoleIndex, ok := roles[apiKeyObj.Role]
		if !ok {
			return c.JSON(http.StatusBadRequest, Msg("Invalid role, please specify appropriate role"))
		}
		// Get current role index
		roleIndex, ok := roles[userRole.Role]
		if !ok {
			return c.JSON(http.StatusBadRequest, Msg("Invalid user role"))
		}
		validRoles := strings.Join(orgRoles[roleIndex:], ",")
		if apiKeyRoleIndex < roleIndex {
			return c.JSON(http.StatusBadRequest, Msg("User has no permissions to set this role, valid roles user can set are "+validRoles))
		}

		apiKeyApiKeyId := uuid.New().String()
		apiKey := uuid.New().String()
		psub := rbac.GetCasbinGroup(userRole.Org, apiKeyApiKeyId)
		err = enforcer.AddGroupingPolicy(ctx, psub, apiKeyObj.Role)
		if err != nil {
			return setReply(c, dbErr(err), nil)
		}
		apiKeyHash, apiKeySalt, apiKeyIter := NewPasshash(apiKey)
		apiKeys[apiKeyApiKeyId] = ormapi.ApiKey{
			ApiKeyOrg:  userRole.Org,
			ApiKeyRole: apiKeyObj.Role,
			ApiKeyHash: apiKeyHash,
			ApiKeySalt: apiKeySalt,
			ApiKeyIter: apiKeyIter,
			ApiKeyDesc: apiKeyObj.Description,
		}
		out, err := MarshalApiKeys(apiKeys)
		if err != nil {
			return err
		}
		user.ApiKeys = out
		if err := db.Model(&user).Updates(&user).Error; err != nil {
			return setReply(c, dbErr(err), nil)
		}
		apiKeyOut := ormapi.UserApiKey{}
		apiKeyOut.ApiKeyId = apiKeyApiKeyId
		apiKeyOut.ApiKey = apiKey
		return c.JSON(http.StatusOK, &apiKeyOut)
	}
	return nil
}

func DeleteUserApiKey(c echo.Context) error {
	ctx := GetContext(c)
	db := loggedDB(ctx)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	apiKeyObj := ormapi.UserApiKey{}
	if err := c.Bind(&apiKeyObj); err != nil {
		return bindErr(c, err)
	}
	user := ormapi.User{Name: claims.Username}
	err = db.Where(&user).First(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	// fetch existing apikeys
	apiKeys, err := UnmarshalApiKeys(user.ApiKeys)
	if err != nil {
		return err
	}
	userApiKey, ok := apiKeys[apiKeyObj.ApiKeyId]
	if !ok {
		return c.JSON(http.StatusBadRequest, Msg(fmt.Sprintf("API Key with ID %s doesn't exist", apiKeyObj.ApiKeyId)))
	}

	psub := rbac.GetCasbinGroup(userApiKey.ApiKeyOrg, apiKeyObj.ApiKeyId)
	found, err := enforcer.HasGroupingPolicy(psub, userApiKey.ApiKeyRole)
	if err != nil {
		return dbErr(err)
	}
	if !found {
		return c.JSON(http.StatusBadRequest, Msg("Missing grouping policy"))
	}
	err = enforcer.RemoveGroupingPolicy(ctx, psub, userApiKey.ApiKeyRole)
	if err != nil {
		return dbErr(err)
	}

	delete(apiKeys, apiKeyObj.ApiKeyId)
	out, err := MarshalApiKeys(apiKeys)
	if err != nil {
		return err
	}
	user.ApiKeys = out
	if err := db.Model(&user).Updates(&user).Error; err != nil {
		return setReply(c, dbErr(err), nil)
	}

	return c.JSON(http.StatusOK, Msg("deleted API Key successfully"))
}

func ShowUserApiKey(c echo.Context) error {
	ctx := GetContext(c)
	db := loggedDB(ctx)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	user := ormapi.User{Name: claims.Username}
	err = db.Where(&user).First(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	// fetch apikeys
	userApiKeys := []ormapi.UserApiKey{}
	apiKeys, err := UnmarshalApiKeys(user.ApiKeys)
	if err != nil {
		return err
	}
	for apiKeyId, keyObj := range apiKeys {
		userApiKeys = append(userApiKeys, ormapi.UserApiKey{
			ApiKeyId:    apiKeyId,
			Description: keyObj.ApiKeyDesc,
			Org:         keyObj.ApiKeyOrg,
			Role:        keyObj.ApiKeyRole,
		})
	}
	return c.JSON(http.StatusOK, &userApiKeys)
}
