package ormapi

import (
	"fmt"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/stretchr/testify/require"
)

type DBExec struct {
	obj  interface{}
	pass bool
}

func StartDB() (*intprocess.Sql, *gorm.DB, error) {
	sqlAddrHost := "127.0.0.1"
	sqlAddrPort := "51001"
	dbUser := "testuser"
	dbName := "mctestdb"
	sql := intprocess.Sql{
		Common: process.Common{
			Name: "sql1",
		},
		DataDir:  "./.postgres",
		HttpAddr: sqlAddrHost + ":" + sqlAddrPort,
		Username: dbUser,
		Dbname:   dbName,
	}
	_, err := os.Stat(sql.DataDir)
	if os.IsNotExist(err) {
		sql.InitDataDir()
	}
	err = sql.StartLocal("")
	if err != nil {
		return nil, nil, fmt.Errorf("local sql start failed: %v", err)
	}

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable", sqlAddrHost, sqlAddrPort, dbUser, dbName))
	if err != nil {
		sql.StopLocal()
		return nil, nil, fmt.Errorf("failed to open gorm object: %v", err)
	}
	return &sql, db, nil
}

func TestFederationGormObjs(t *testing.T) {
	sql, db, err := StartDB()
	require.Nil(t, err, "start sql db")
	defer sql.StopLocal()
	defer db.Close()

	dbObjs := []interface{}{
		&Federator{},
		&Federation{},
		&FederatorZone{},
		&FederatedPartnerZone{},
		&FederatedSelfZone{},
	}

	db.DropTableIfExists(dbObjs...)
	db.LogMode(true)
	db.AutoMigrate(dbObjs...)

	err = InitFederationAPIConstraints(db)
	require.Nil(t, err, "set constraints")

	tests := []DBExec{
		{
			obj:  &Federator{OperatorId: "TDG", CountryCode: "EU", FederationKey: "key1"},
			pass: true,
		},
		{
			obj:  &Federator{OperatorId: "BT", CountryCode: "US", FederationKey: "key2"},
			pass: true,
		},
		{
			obj: &Federation{
				SelfOperatorId: "TDG", SelfCountryCode: "EU",
				Federator: Federator{
					OperatorId: "VOD", CountryCode: "KR", FederationKey: "keyA",
				},
				PartnerRoleShareZonesWithSelf: true,
			},
			pass: true,
		},
		{
			obj: &Federation{
				SelfOperatorId: "BT", SelfCountryCode: "US",
				Federator: Federator{
					OperatorId: "VOD", CountryCode: "KR", FederationKey: "keyB",
				},
			},
			pass: true,
		},
		{
			// embedded struct's primary keys are considered
			obj: &Federation{
				SelfOperatorId: "BT", SelfCountryCode: "US",
				Federator: Federator{
					OperatorId: "VODA", CountryCode: "KR", FederationKey: "keyC",
				},
			},
			pass: true,
		},
		{
			// NOTE: This should fail
			obj: &Federation{
				SelfOperatorId: "BTS", SelfCountryCode: "US",
				Federator: Federator{
					OperatorId: "VODA", CountryCode: "KR", FederationKey: "keyD",
				},
			},
			pass: false,
		},
		{
			obj: &FederatorZone{
				OperatorId: "BT", CountryCode: "US",
				ZoneId:      "Z1",
				GeoLocation: "123,321",
			},
			pass: true,
		},
		{
			obj: &FederatorZone{
				OperatorId: "TDG", CountryCode: "EU",
				ZoneId:      "Z2",
				GeoLocation: "123,321",
			},
			pass: true,
		},
		{
			// NOTE: should fail
			obj: &FederatorZone{
				OperatorId: "BTS", CountryCode: "US",
				ZoneId:      "Z3",
				GeoLocation: "123,321",
			},
			pass: false,
		},
		{
			// NOTE: should fail
			obj: &FederatedPartnerZone{
				SelfOperatorId: "BTS", SelfCountryCode: "US",
				FederatorZone: FederatorZone{
					OperatorId: "VOD", CountryCode: "KR",
					ZoneId:      "Z4",
					GeoLocation: "123,321",
				},
				Registered: true,
			},
			pass: false,
		},
		{
			// NOTE: should fail
			obj: &FederatedPartnerZone{
				SelfOperatorId: "BT", SelfCountryCode: "US",
				FederatorZone: FederatorZone{
					OperatorId: "VODAF", CountryCode: "KR",
					ZoneId:      "Z4",
					GeoLocation: "123,321",
				},
				Registered: true,
			},
			pass: false,
		},
		{
			// NOTE: should fail, as such federation doesn't exist
			obj: &FederatedPartnerZone{
				SelfOperatorId: "TDG", SelfCountryCode: "EU",
				FederatorZone: FederatorZone{
					OperatorId: "VODA", CountryCode: "KR",
					ZoneId:      "Z4",
					GeoLocation: "123,321",
				},
				Registered: true,
			},
			pass: false,
		},
		{
			obj: &FederatedPartnerZone{
				SelfOperatorId: "TDG", SelfCountryCode: "EU",
				FederatorZone: FederatorZone{
					OperatorId: "VOD", CountryCode: "KR",
					ZoneId:      "Z4",
					GeoLocation: "123,321",
				},
				Registered: true,
			},
			pass: true,
		},
	}

	for _, test := range tests {
		err = db.Create(test.obj).Error
		if test.pass {
			require.Nil(t, err, test.obj)
		} else {
			require.NotNil(t, err, test.obj)
		}
		defer db.Delete(test.obj)
	}
}
