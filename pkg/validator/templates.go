package validator

// ErrorTemplates maps validation tags to html templates.
var ErrorTemplates = map[string]string{
	"alpha":                   "{{.Namespace}} can only contain alphabetic characters",
	"alphanum":                "{{.Namespace}} can only contain alphanumeric characters",
	"alphanumunicode":         "{{.Namespace}} can only contain unicode alphanumeric characters",
	"alphaunicode":            "{{.Namespace}} can only contain unicode alphabetic characters",
	"ascii":                   "{{.Namespace}} must contain only ascii characters",
	"base64":                  "{{.Namespace}} must be a valid Base64 string",
	"base64url":               "{{.Namespace}} must be a valid Base64 URL string",
	"btc_addr":                "{{.Namespace}} must be a valid Bitcoin address",
	"btc_addr_bech32":         "{{.Namespace}} must be a valid bech32 Bitcoin address",
	"cidr":                    "{{.Namespace}} must contain a valid CIDR notation",
	"cidrv4":                  "{{.Namespace}} must contain a valid CIDR notation for an IPv4 address",
	"cidrv6":                  "{{.Namespace}} must contain a valid CIDR notation for an IPv6 address",
	"contains":                "{{.Namespace}} must contain the text '{{.Param}}'",
	"containsany":             "{{.Namespace}} must contain at least one of the following characters '{{.Param}}'",
	"containsrune":            "{{.Namespace}} must contain the following '{{.Param}}'",
	"datauri":                 "{{.Namespace}} must contain a valid Data URI",
	"datetime":                "{{.Namespace}} does not match the {{.Param}} format",
	"dir":                     "{{.Namespace}} must be a valid directory",
	"e164":                    "{{.Namespace}} must be a valid E.164 formatted phone number",
	"email":                   "{{.Namespace}} must be a valid email address",
	"endsnotwith":             "{{.Namespace}} must not end with {{.Param}}",
	"endswith":                "{{.Namespace}} must end with {{.Param}}",
	"eq":                      "{{.Namespace}} must be equal to {{.Param}}",
	"eqcsfield":               "{{.Namespace}} must be equal to {{.Param}}",
	"eqfield":                 "{{.Namespace}} must be equal to {{.Param}}",
	"eth_addr":                "{{.Namespace}} must be a valid Ethereum address",
	"excluded_with":           "{{.Namespace}} must not be present or it must be empty",
	"excluded_with_all":       "{{.Namespace}} must not be present or it must be empty",
	"excluded_without":        "{{.Namespace}} must not be present or it must be empty",
	"excluded_without_all":    "{{.Namespace}} must not be present or it must be empty",
	"excludes":                "{{.Namespace}} cannot contain the text '{{.Param}}'",
	"excludesall":             "{{.Namespace}} cannot contain any of the following characters '{{.Param}}'",
	"excludesrune":            "{{.Namespace}} cannot contain the following '{{.Param}}'",
	"fieldcontains":           "{{.Namespace}} must contain the field {{.Param}}",
	"fieldexcludes":           "{{.Namespace}} must not contain the field {{.Param}}",
	"file":                    "{{.Namespace}} must be a valid file path",
	"fqdn":                    "{{.Namespace}} must be a Fully Qualified Domain Name (FQDN)",
	"gt":                      "{{.Namespace}} must be greater than {{.Param}}",
	"gtcsfield":               "{{.Namespace}} must be greater than {{.Param}}",
	"gte":                     "{{.Namespace}} must be greater or equal {{.Param}}",
	"gtecsfield":              "{{.Namespace}} must be greater or equal {{.Param}}",
	"gtefield":                "{{.Namespace}} must be greater or equal {{.Param}}",
	"gtfield":                 "{{.Namespace}} must be greater than {{.Param}}",
	"hexadecimal":             "{{.Namespace}} must be a valid hexadecimal",
	"hexcolor":                "{{.Namespace}} must be a valid HEX color",
	"hostname":                "{{.Namespace}} must be a valid hostname as per RFC 952",
	"hostname_port":           "{{.Namespace}} must be in the format DNS:PORT",
	"hostname_rfc1123":        "{{.Namespace}} must be a valid hostname as per RFC 1123",
	"hsl":                     "{{.Namespace}} must be a valid HSL color",
	"hsla":                    "{{.Namespace}} must be a valid HSLA color",
	"html":                    "{{.Namespace}} must be valid HTML",
	"html_encoded":            "{{.Namespace}} must be HTML-encoded",
	"ip":                      "{{.Namespace}} must be a valid IP address",
	"ip4_addr":                "{{.Namespace}} must be a resolvable IPv4 address",
	"ip6_addr":                "{{.Namespace}} must be a resolvable IPv6 address",
	"ip_addr":                 "{{.Namespace}} must be a resolvable IP address",
	"ipv4":                    "{{.Namespace}} must be a valid IPv4 address",
	"ipv6":                    "{{.Namespace}} must be a valid IPv6 address",
	"isbn":                    "{{.Namespace}} must be a valid ISBN number",
	"isbn10":                  "{{.Namespace}} must be a valid ISBN-10 number",
	"isbn13":                  "{{.Namespace}} must be a valid ISBN-13 number",
	"iscolor":                 "{{.Namespace}} must be a valid color",
	"isdefault":               "{{.Namespace}} must not be present or it must be empty",
	"iso3166_1_alpha2":        "{{.Namespace}} must be a valid iso3166-1 alpha-2 country code",
	"iso3166_1_alpha3":        "{{.Namespace}} must be a valid iso3166-1 alpha-3 country code",
	"iso3166_1_alpha_numeric": "{{.Namespace}} must be a valid iso3166-1 alpha-numeric country code",
	"json":                    "{{.Namespace}} must be a valid json string",
	"latitude":                "{{.Namespace}} must contain valid latitude coordinates",
	"len":                     "{{.Namespace}} must be {{.Param}} in length",
	"longitude":               "{{.Namespace}} must contain a valid longitude coordinates",
	"lowercase":               "{{.Namespace}} must be a lowercase string",
	"lt":                      "{{.Namespace}} must be less than {{.Param}}",
	"ltcsfield":               "{{.Namespace}} must be less than {{.Param}}",
	"lte":                     "{{.Namespace}} must be less or equal {{.Param}}",
	"ltecsfield":              "{{.Namespace}} must be less or equal {{.Param}}",
	"ltefield":                "{{.Namespace}} must be less or equal {{.Param}}",
	"ltfield":                 "{{.Namespace}} must be less than {{.Param}}",
	"mac":                     "{{.Namespace}} must contain a valid MAC address",
	"max":                     "{{.Namespace}} must have a maximum size/length/value of {{.Param}}",
	"min":                     "{{.Namespace}} must have a minimum size/length/value of {{.Param}}",
	"multibyte":               "{{.Namespace}} must contain multibyte characters",
	"ne":                      "{{.Namespace}} must be different than {{.Param}}",
	"necsfield":               "{{.Namespace}} must be different than {{.Param}}",
	"nefield":                 "{{.Namespace}} must be different than {{.Param}}",
	"number":                  "{{.Namespace}} must be a valid number",
	"numeric":                 "{{.Namespace}} must be a valid numeric value",
	"oneof":                   "{{.Namespace}} must be one of {{.Param}}",
	"printascii":              "{{.Namespace}} must contain only printable ascii characters",
	"required":                "{{.Namespace}} is required",
	"required_if":             "{{.Namespace}} is required",
	"required_unless":         "{{.Namespace}} is required",
	"required_with":           "{{.Namespace}} is required",
	"required_with_all":       "{{.Namespace}} is required",
	"required_without":        "{{.Namespace}} is required",
	"required_without_all":    "{{.Namespace}} is required",
	"rgb":                     "{{.Namespace}} must be a valid RGB color",
	"rgba":                    "{{.Namespace}} must be a valid RGBA color",
	"ssn":                     "{{.Namespace}} must be a valid SSN number",
	"startsnotwith":           "{{.Namespace}} must not start with {{.Param}}",
	"startswith":              "{{.Namespace}} must start with {{.Param}}",
	"tcp4_addr":               "{{.Namespace}} must be a valid IPv4 TCP address",
	"tcp6_addr":               "{{.Namespace}} must be a valid IPv6 TCP address",
	"tcp_addr":                "{{.Namespace}} must be a valid TCP address",
	"timezone":                "{{.Namespace}} must be a valid time zone string",
	"udp4_addr":               "{{.Namespace}} must be a valid IPv4 UDP address",
	"udp6_addr":               "{{.Namespace}} must be a valid IPv6 UDP address",
	"udp_addr":                "{{.Namespace}} must be a valid UDP address",
	"unique":                  "{{.Namespace}} must contain unique values",
	"unix_addr":               "{{.Namespace}} must be a resolvable UNIX address",
	"uppercase":               "{{.Namespace}} must be an uppercase string",
	"uri":                     "{{.Namespace}} must be a valid URI",
	"url":                     "{{.Namespace}} must be a valid URL",
	"url_encoded":             "{{.Namespace}} must be URL-encoded",
	"urn_rfc2141":             "{{.Namespace}} must be a valid URN as per RFC 2141",
	"uuid":                    "{{.Namespace}} must be a valid UUID",
	"uuid3":                   "{{.Namespace}} must be a valid version 3 UUID",
	"uuid3_rfc4122":           "{{.Namespace}} must be a valid version 3 UUID as per RFC 4122",
	"uuid4":                   "{{.Namespace}} must be a valid version 4 UUID",
	"uuid4_rfc4122":           "{{.Namespace}} must be a valid version 4 UUID as per RFC 4122",
	"uuid5":                   "{{.Namespace}} must be a valid version 5 UUID",
	"uuid5_rfc4122":           "{{.Namespace}} must be a valid version 5 UUID as per RFC 4122",
	"uuid_rfc4122":            "{{.Namespace}} must be a valid UUID as per RFC 4122",
}
