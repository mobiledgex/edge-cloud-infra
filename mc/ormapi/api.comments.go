package ormapi

// This is an auto-generated file. DO NOT EDIT directly.

var UserComments = map[string]string{
	"name":            `User name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen`,
	"email":           `User email`,
	"emailverified":   `Email address has been verified`,
	"familyname":      `Family Name`,
	"givenname":       `Given Name`,
	"nickname":        `Nick Name`,
	"locked":          `Account is locked`,
	"enabletotp":      `Enable or disable temporary one-time passwords for the account`,
	"metadata":        `Metadata`,
	"lastlogin":       `Last successful login time`,
	"lastfailedlogin": `Last failed login time`,
	"failedlogins":    `Number of failed login attempts since last successful login`,
}

var CreateUserApiKeyComments = map[string]string{
	"userapikey.id":          `API key ID used as an identifier for API keys`,
	"userapikey.description": `Description of the purpose of this API key`,
	"userapikey.org":         `Org to which API key has permissions to access its objects`,
	"apikey":                 `API key`,
	"permissions:#.role":     `Role defines a collection of permissions, which are resource-action pairs`,
	"permissions:#.resource": `Resource defines a resource to act upon`,
	"permissions:#.action":   `Action defines what type of action can be performed on a resource`,
}

var UserApiKeyComments = map[string]string{
	"id":          `API key ID used as an identifier for API keys`,
	"description": `Description of the purpose of this API key`,
	"org":         `Org to which API key has permissions to access its objects`,
}

var OrganizationComments = map[string]string{
	"name":             `Organization name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen`,
	"type":             `Organization type: "developer" or "operator"`,
	"address":          `Organization address`,
	"phone":            `Organization phone number`,
	"publicimages":     `Images are made available to other organization`,
	"deleteinprogress": `Delete of this organization is in progress`,
	"edgeboxonly":      `Edgebox only operator organization`,
}

var InvoiceRequestComments = map[string]string{
	"name":      `Billing Organization name to retrieve invoices for`,
	"startdate": `Date filter for invoice selection, YYYY-MM-DD format`,
	"enddate":   `Date filter for invoice selection, YYYY-MM-DD format`,
}

var BillingOrganizationComments = map[string]string{
	"name":       `BillingOrganization name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen`,
	"type":       `Organization type: "parent" or "self"`,
	"firstname":  `Billing info first name`,
	"lastname":   `Billing info last name`,
	"email":      `Organization email`,
	"address":    `Organization address`,
	"address2":   `Organization address2`,
	"city":       `Organization city`,
	"country":    `Organization country`,
	"state":      `Organization state`,
	"postalcode": `Organization postal code`,
	"phone":      `Organization phone number`,
	"children":   `Children belonging to this BillingOrganization`,
}

var AccountInfoComments = map[string]string{
	"orgname":        `Billing Organization name to commit`,
	"accountid":      `Account ID given by the billing platform`,
	"subscriptionid": `Subscription ID given by the billing platform`,
}

var PaymentProfileDeletionComments = map[string]string{
	"org": `Billing Organization Name associated with the payment profile`,
	"id":  `Payment Profile Id`,
}

var ControllerComments = map[string]string{
	"region":     `Controller region name`,
	"address":    `Controller API address or URL`,
	"notifyaddr": `Controller notify address or URL`,
	"influxdb":   `InfluxDB address`,
	"dnsregion":  `Unique DNS label for the region`,
	"thanosmetrics": `Thanos Query URL`,
}

var ConfigComments = map[string]string{
	"locknewaccounts":              `Lock new accounts (must be unlocked by admin)`,
	"notifyemailaddress":           `Email to notify when locked account is created`,
	"skipverifyemail":              `Skip email verification for new accounts (testing only)`,
	"passwordmincracktimesec":      `User accounts min password crack time seconds (a measure of strength)`,
	"adminpasswordmincracktimesec": `Admin accounts min password crack time seconds (a measure of strength)`,
	"maxmetricsdatapoints":         `InfluxDB max number of data points returned`,
	"userapikeycreatelimit":        `Max number of API keys a user can create`,
	"billingenable":                `Toggle for enabling billing (primarily for testing purposes)`,
	"disableratelimit":             `Toggle to enable and disable MC API rate limiting`,
	"ratelimitmaxtrackedips":       `Maximum number of IPs tracked per API group for rate limiting at MC`,
	"ratelimitmaxtrackedusers":     `Maximum number of users tracked per API group for rate limiting at MC`,
	"failedloginlockoutthreshold1": `Failed login lockout threshold 1, after this count, lockout time 1 is enabled (default 3)`,
	"failedloginlockouttimesec1":   `Number of seconds to lock account from logging in after threshold 1 is hit (default 60)`,
	"failedloginlockoutthreshold2": `Failed login lockout threshold 2, after this count, lockout time 2 is enabled (default 10)`,
	"failedloginlockouttimesec2":   `Number of seconds to lock account from logging in after threshold 2 is hit (default 300)`,
}

var McRateLimitFlowSettingsComments = map[string]string{
	"flowsettingsname": `Unique name for FlowSettings`,
	"apiname":          `Name of API Path (eg. /api/v1/usercreate)`,
	"reqspersecond":    `Number of requests per second`,
	"burstsize":        `Number of requests allowed at once`,
}

var McRateLimitMaxReqsSettingsComments = map[string]string{
	"maxreqssettingsname": `Unique name for MaxReqsSettings`,
	"apiname":             `Name of API Path (eg. /api/v1/usercreate)`,
	"maxrequests":         `Maximum number of requests for the specified interval`,
}

var McRateLimitSettingsComments = map[string]string{
	"apiname": `Name of API Path (eg. /api/v1/usercreate)`,
}

var OrgCloudletPoolComments = map[string]string{
	"org":             `Developer Organization`,
	"region":          `Region`,
	"cloudletpool":    `Operator's CloudletPool name`,
	"cloudletpoolorg": `Operator's Organization`,
	"type":            `Type is an internal-only field which is either invitation or response`,
	"decision":        `Decision is to either accept or reject an invitation`,
}

var RolePermComments = map[string]string{
	"role":     `Role defines a collection of permissions, which are resource-action pairs`,
	"resource": `Resource defines a resource to act upon`,
	"action":   `Action defines what type of action can be performed on a resource`,
}

var RoleComments = map[string]string{
	"org":      `Organization name`,
	"username": `User name`,
	"role":     `Role which defines the set of permissions`,
}

var OrgCloudletComments = map[string]string{
	"region": `Region name`,
}

var ShowUserComments = map[string]string{
	"user.name":            `User name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen`,
	"user.email":           `User email`,
	"user.emailverified":   `Email address has been verified`,
	"user.familyname":      `Family Name`,
	"user.givenname":       `Given Name`,
	"user.nickname":        `Nick Name`,
	"user.locked":          `Account is locked`,
	"user.enabletotp":      `Enable or disable temporary one-time passwords for the account`,
	"user.metadata":        `Metadata`,
	"user.lastlogin":       `Last successful login time`,
	"user.lastfailedlogin": `Last failed login time`,
	"user.failedlogins":    `Number of failed login attempts since last successful login`,
	"org":                  `Organization name`,
	"role":                 `Role name`,
}

var UserLoginComments = map[string]string{
	"username": `User's name or email address`,
	"password": `User's password`,
}

var CreateUserComments = map[string]string{
	"user.name":            `User name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen`,
	"user.email":           `User email`,
	"user.emailverified":   `Email address has been verified`,
	"user.familyname":      `Family Name`,
	"user.givenname":       `Given Name`,
	"user.nickname":        `Nick Name`,
	"user.locked":          `Account is locked`,
	"user.enabletotp":      `Enable or disable temporary one-time passwords for the account`,
	"user.metadata":        `Metadata`,
	"user.lastlogin":       `Last successful login time`,
	"user.lastfailedlogin": `Last failed login time`,
	"user.failedlogins":    `Number of failed login attempts since last successful login`,
}

var EmailRequestComments = map[string]string{
	"callbackurl": `Callback URL to verify user email`,
}

var PasswordResetComments = map[string]string{
	"token":    `Authentication token`,
	"password": `User's new password`,
}

var TokenComments = map[string]string{
	"token": `Authentication token`,
}

var RegionDataComments = map[string]string{
	"region": `Region name`,
}

var MetricsCommonComments = map[string]string{
	"numsamples": `Display X samples spaced out evenly over start and end times`,
	"limit":      `Display the last X metrics`,
}

var RegionAppInstMetricsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
}

var RegionClusterInstMetricsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
}

var RegionCloudletMetricsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
	"selector":                 `Comma separated list of metrics to view. Available metrics: utilization, network, ipusage`,
}

var RegionClientApiUsageMetricsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
	"method":                   `API call method, one of: FindCloudlet, PlatformFindCloudlet, RegisterClient, VerifyLocation`,
	"dmecloudlet":              `Cloudlet name where DME is running`,
	"dmecloudletorg":           `Operator organization where DME is running`,
	"selector":                 `Comma separated list of metrics to view. Available metrics: utilization, network, ipusage`,
}

var RegionClientAppUsageMetricsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
	"selector":                 `Comma separated list of metrics to view. Available metrics: utilization, network, ipusage`,
	"devicecarrier":            `Device carrier. Can be used for selectors: latency, deviceinfo`,
	"datanetworktype":          `Data network type used by client device. Can be used for selectors: latency`,
	"devicemodel":              `Device model. Can be used for selectors: deviceinfo`,
	"deviceos":                 `Device operating system. Can be used for selectors: deviceinfo`,
	"locationtile":             `Provides the range of GPS coordinates for the location tile/square. Format is: 'LocationUnderLongitude,LocationUnderLatitude_LocationOverLongitude,LocationOverLatitude_LocationTileLength'. LocationUnder are the GPS coordinates of the corner closest to (0,0) of the location tile. LocationOver are the GPS coordinates of the corner farthest from (0,0) of the location tile. LocationTileLength is the length (in kilometers) of one side of the location tile square`,
}

var RegionClientCloudletUsageMetricsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
	"selector":                 `Comma separated list of metrics to view. Available metrics: utilization, network, ipusage`,
	"devicecarrier":            `Device carrier. Can be used for selectors: latency, deviceinfo`,
	"datanetworktype":          `Data network type used by client device. Can be used for selectors: latency`,
	"devicemodel":              `Device model. Can be used for selectors: deviceinfo`,
	"deviceos":                 `Device operating system. Can be used for selectors: deviceinfo`,
	"locationtile":             `Provides the range of GPS coordinates for the location tile/square. Format is: 'LocationUnderLongitude,LocationUnderLatitude_LocationOverLongitude,LocationOverLatitude_LocationTileLength'. LocationUnder are the GPS coordinates of the corner closest to (0,0) of the location tile. LocationOver are the GPS coordinates of the corner farthest from (0,0) of the location tile. LocationTileLength is the length (in kilometers) of one side of the location tile square`,
}

var RegionAppInstEventsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
}

var RegionClusterInstEventsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
}

var RegionCloudletEventsComments = map[string]string{
	"metricscommon.numsamples": `Display X samples spaced out evenly over start and end times`,
	"metricscommon.limit":      `Display the last X metrics`,
	"region":                   `Region name`,
}

var RegionAppInstUsageComments = map[string]string{
	"region":    `Region name`,
	"starttime": `Time to start displaying stats from`,
	"endtime":   `Time up to which to display stats`,
	"vmonly":    `Show only VM-based apps`,
}

var RegionClusterInstUsageComments = map[string]string{
	"region":    `Region name`,
	"starttime": `Time to start displaying stats from`,
	"endtime":   `Time up to which to display stats`,
}

var RegionCloudletPoolUsageComments = map[string]string{
	"region":         `Region name`,
	"starttime":      `Time to start displaying stats from`,
	"endtime":        `Time up to which to display stats`,
	"showvmappsonly": `Show only VM-based apps`,
}

var RegionCloudletPoolUsageRegisterComments = map[string]string{
	"region": `Region name`,
}

var AlertReceiverComments = map[string]string{
	"name":                    `Receiver Name`,
	"type":                    `Receiver type. Eg. email, slack, pagerduty`,
	"severity":                `Alert severity filter`,
	"region":                  `Region for the alert receiver`,
	"user":                    `User that created this receiver`,
	"email":                   `Custom receiving email`,
	"slackchannel":            `Custom slack channel`,
	"slackwebhook":            `Custom slack webhook`,
	"pagerdutyintegrationkey": `PagerDuty integration key`,
	"pagerdutyapiversion":     `PagerDuty API version`,
}

var ReporterComments = map[string]string{
	"name":              `Reporter name. Can only contain letters, digits, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen`,
	"org":               `Organization name`,
	"email":             `Email to send generated reports`,
	"schedule":          `Indicates how often a report should be generated, one of EveryWeek, Every15Days, EveryMonth`,
	"startscheduledate": `Start date (in RFC3339 format with intended timezone) when the report is scheduled to be generated (Default: today)`,
	"nextscheduledate":  `Date when the next report is scheduled to be generated (for internal use only)`,
	"username":          `User name (for internal use only)`,
	"timezone":          `Timezone in which to show the reports, defaults to UTC`,
	"status":            `Last report status`,
}

var DownloadReportComments = map[string]string{
	"org":      `Organization name`,
	"reporter": `Reporter name`,
	"filename": `Name of the report file to be downloaded`,
}

var GenerateReportComments = map[string]string{
	"org":       `Organization name`,
	"starttime": `Absolute time (in RFC3339 format with intended timezone) to start report capture`,
	"endtime":   `Absolute time (in RFC3339 format with intended timezone) to end report capture`,
	"region":    `Region name (for internal use only)`,
	"timezone":  `Timezone in which to show the reports, defaults to UTC`,
}

var FederatorComments = map[string]string{
	"federationid":    `Globally unique string used to indentify a federation with partner federation`,
	"operatorid":      `Globally unique string to identify an operator platform`,
	"countrycode":     `ISO 3166-1 Alpha-2 code for the country where operator platform is located`,
	"federationaddr":  `Federation access point address`,
	"region":          `Region to which this federator is associated with`,
	"mcc":             `Mobile country code of operator sending the request`,
	"mnc":             `List of mobile network codes of operator sending the request`,
	"locatorendpoint": `IP and Port of discovery service URL of operator platform`,
	"revision":        `Revision ID to track object changes. We use jaeger traceID for easy debugging but this can differ with what partner federator uses`,
	"apikey":          `API Key used for authentication (stored in secure storage)`,
}

var FederationComments = map[string]string{
	"federator.federationid":        `Globally unique string used to indentify a federation with partner federation`,
	"federator.operatorid":          `Globally unique string to identify an operator platform`,
	"federator.countrycode":         `ISO 3166-1 Alpha-2 code for the country where operator platform is located`,
	"federator.federationaddr":      `Federation access point address`,
	"federator.region":              `Region to which this federator is associated with`,
	"federator.mcc":                 `Mobile country code of operator sending the request`,
	"federator.mnc":                 `List of mobile network codes of operator sending the request`,
	"federator.locatorendpoint":     `IP and Port of discovery service URL of operator platform`,
	"federator.revision":            `Revision ID to track object changes. We use jaeger traceID for easy debugging but this can differ with what partner federator uses`,
	"federator.apikey":              `API Key used for authentication (stored in secure storage)`,
	"name":                          `Name to uniquely identify a federation`,
	"selffederationid":              `Self federation ID`,
	"selfoperatorid":                `Self operator ID`,
	"partnerrolesharezoneswithself": `Partner shares its zones with self federator as part of federation`,
	"partnerroleaccesstoselfzones":  `Partner is allowed access to self federator zones as part of federation`,
}

var FederatorZoneComments = map[string]string{
	"zoneid":      `Globally unique string used to authenticate operations over federation interface`,
	"operatorid":  `Globally unique string to identify an operator platform`,
	"countrycode": `ISO 3166-1 Alpha-2 code for the country where operator platform is located`,
	"geolocation": `GPS co-ordinates associated with the zone (in decimal format)`,
	"city":        `Comma seperated list of cities under this zone`,
	"state":       `Comma seperated list of states under this zone`,
	"locality":    `Type of locality eg rural, urban etc.`,
	"region":      `Region in which cloudlets reside`,
	"cloudlets":   `List of cloudlets part of this zone`,
	"revision":    `Revision ID to track object changes. We use jaeger traceID for easy debugging but this can differ with what partner federator uses`,
}

var FederatedSelfZoneComments = map[string]string{
	"zoneid":         `Globally unique identifier of the federator zone`,
	"selfoperatorid": `Self operator ID`,
	"federationname": `Name of the Federation`,
	"registered":     `Zone registered by partner federator`,
	"revision":       `Revision ID to track object changes. We use jaeger traceID for easy debugging but this can differ with what partner federator uses`,
}

var FederatedPartnerZoneComments = map[string]string{
	"federatorzone.zoneid":      `Globally unique string used to authenticate operations over federation interface`,
	"federatorzone.operatorid":  `Globally unique string to identify an operator platform`,
	"federatorzone.countrycode": `ISO 3166-1 Alpha-2 code for the country where operator platform is located`,
	"federatorzone.geolocation": `GPS co-ordinates associated with the zone (in decimal format)`,
	"federatorzone.city":        `Comma seperated list of cities under this zone`,
	"federatorzone.state":       `Comma seperated list of states under this zone`,
	"federatorzone.locality":    `Type of locality eg rural, urban etc.`,
	"federatorzone.region":      `Region in which cloudlets reside`,
	"federatorzone.cloudlets":   `List of cloudlets part of this zone`,
	"federatorzone.revision":    `Revision ID to track object changes. We use jaeger traceID for easy debugging but this can differ with what partner federator uses`,
	"selfoperatorid":            `Self operator ID`,
	"federationname":            `Name of the Federation`,
	"registered":                `Zone registered by self federator`,
}

var FederatedZoneRegRequestComments = map[string]string{
	"selfoperatorid": `Self operator ID`,
	"federationname": `Name of the Federation`,
	"zones":          `Partner federator zones to be registered/deregistered`,
}
