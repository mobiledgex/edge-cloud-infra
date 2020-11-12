package ormctl

import (
	"fmt"
	"io/ioutil"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

func GetUserCommand() *cobra.Command {
	cmds := []*cli.Command{&cli.Command{
		Use:            "create",
		RequiredArgs:   "name email",
		OptionalArgs:   "nickname familyname givenname callbackurl otptype",
		AliasArgs:      "name=user.name email=user.email nickname=user.nickname familyname=user.familyname givenname=user.givenname password=user.passhash callbackurl=verify.callbackurl otptype=user.totptype",
		PasswordArg:    "user.passhash",
		VerifyPassword: true,
		ReqData:        &ormapi.CreateUser{},
		Run:            runRest("/usercreate"),
	}, &cli.Command{
		Use:          "delete",
		RequiredArgs: "name",
		ReqData:      &ormapi.User{},
		Run:          runRest("/auth/user/delete"),
	}, &cli.Command{
		Use:          "show",
		ReqData:      &ormapi.Organization{},
		OptionalArgs: "orgname",
		AliasArgs:    "orgname=name",
		ReplyData:    &[]ormapi.User{},
		Run:          runRest("/auth/user/show"),
	}, &cli.Command{
		Use:       "current",
		ReplyData: &ormapi.User{},
		Run:       runRest("/auth/user/current"),
	}, &cli.Command{
		Use:            "newpass",
		PasswordArg:    "password",
		VerifyPassword: true,
		ReqData:        &ormapi.NewPassword{},
		Run:            runRest("/auth/user/newpass"),
	}, &cli.Command{
		Use:          "resendverify",
		RequiredArgs: "email",
		ReqData:      &ormapi.EmailRequest{},
		Run:          runRest("/resendverify"),
	}, &cli.Command{
		Use:          "verifyemail",
		RequiredArgs: "token",
		ReqData:      &ormapi.Token{},
		Run:          runRest("/verifyemail"),
	}, &cli.Command{
		Use:          "passwordresetrequest",
		RequiredArgs: "email",
		ReqData:      &ormapi.EmailRequest{},
		Run:          runRest("/passwordresetrequest"),
	}, &cli.Command{
		Use:            "passwordreset",
		RequiredArgs:   "token",
		PasswordArg:    "password",
		VerifyPassword: true,
		ReqData:        &ormapi.PasswordReset{},
		Run:            runRest("/passwordreset"),
	}, &cli.Command{
		Use:          "restricteduserupdate",
		OptionalArgs: "name email emailverified familyname givenname nickname locked",
		ReqData:      &ormapi.User{},
		Run:          runRest("/auth/restricted/user/update"),
	}, &cli.Command{
		Use:          "disableotp",
		RequiredArgs: "name",
		ReqData:      &ormapi.User{},
		Run:          runRest("/auth/user/disable/otp"),
	}, &cli.Command{
		Use:          "resetotp",
		RequiredArgs: "name",
		OptionalArgs: "emailotp",
		AliasArgs:    "emailotp=user.emailtotp",
		ReqData:      &ormapi.User{},
		Run:          runRest("/auth/user/reset/otp"),
	}, &cli.Command{
		Use:          "createapikey",
		RequiredArgs: "name",
		ReqData:      &ormapi.UserApiKey{},
		Run:          runRest("/auth/user/create/apikey"),
	}, &cli.Command{
		Use:          "deleteapikey",
		RequiredArgs: "name",
		ReqData:      &ormapi.UserApiKey{},
		Run:          runRest("/auth/user/delete/apikey"),
	}, &cli.Command{
		Use:          "showapikey",
		ReqData:      &ormapi.UserApiKey{},
		OptionalArgs: "name",
		ReplyData:    &[]ormapi.UserApiKey{},
		Run:          runRest("/auth/user/show/apikey"),
	}}
	return cli.GenGroup("user", "manage users", cmds)
}

func GetLoginCmd() *cobra.Command {
	cmd := cli.Command{
		Use:          "login",
		RequiredArgs: "name",
		OptionalArgs: "otp apikey",
		Run:          runLogin,
	}
	return cmd.GenCmd()
}

func runLogin(c *cli.Command, args []string) error {
	input := cli.Input{
		RequiredArgs: []string{"name"},
		PasswordArg:  "password",
		ApiKeyArg:    "apikey",
		AliasArgs:    []string{"name=username", "otp=totp", "apikey=apikey"},
	}
	login := ormapi.UserLogin{}
	_, err := input.ParseArgs(args, &login)
	if err != nil {
		return err
	}
	token, err := client.DoLogin(getUri(), login.Username, login.Password, login.TOTP, login.ApiKey)
	if err != nil {
		return err
	}

	if cli.Parsable {
		fmt.Printf("%s\n", token)
		return nil
	}

	fmt.Println("login successful")
	err = ioutil.WriteFile(getTokenFile(), []byte(token), 0600)
	if err != nil {
		fmt.Printf("warning, cannot save token file %s, %v\n", getTokenFile(), err)
		fmt.Printf("token: %s\n", token)
	} else {
		fmt.Printf("token saved to %s\n", getTokenFile())
	}
	return nil
}
