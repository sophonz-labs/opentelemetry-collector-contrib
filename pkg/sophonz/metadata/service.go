package metadata

import "strings"

// Service is a row of sophonz_metadata.v_postgres_service.
type Service struct {
	ID        string `ch:"id"`
	Namespace string `ch:"serviceNamespace"`
	Name      string `ch:"serviceName"`
	Type      string `ch:"type"`
	Key       string `ch:"key"`
	// CompanyID is the owning organization — the tenant. It comes from the
	// service's project, so every service of a project shares it.
	CompanyID string `ch:"companyId"`
}

// ServiceKeyMap maps a raw service key (the opaque `sk_...` token the SDK
// sends as the service.key attribute) to the Service it identifies.
type ServiceKeyMap map[string]Service

// ConvertToServiceMap builds the service lookup keyed by the raw service key,
// skipping incomplete entries.
//
// The key is looked up verbatim: it is an opaque token, not something the
// collector can decode, so the map has to be keyed by the value that arrives on
// the wire.
func ConvertToServiceMap(services []Service) ServiceKeyMap {
	serviceKeyMap := make(ServiceKeyMap)
	for _, service := range services {
		if service.Key == "" || service.Name == "" || service.Namespace == "" {
			continue
		}
		// One key can be shared by several services of the same project (two
		// apps shipping the same browser key). They resolve to the same tenant,
		// which is all the lookup is used for, so keeping any one of them is
		// correct — the lowest ID wins so the choice does not flap between
		// refreshes, which return rows in an unspecified order.
		if existing, ok := serviceKeyMap[service.Key]; ok && existing.ID <= service.ID {
			continue
		}
		serviceKeyMap[service.Key] = service
	}
	return serviceKeyMap
}

// ServiceVersion is a row of the service version view.
type ServiceVersion struct {
	ID      int    `json:"id" ch:"id"`
	Version string `json:"version" ch:"version"`
}

// ServiceType is the normalized client platform (formerly sophonz.service.type,
// now surfaced as the client.platform attribute / clientPlatform column).
type ServiceType string

const (
	ServiceTypeUnknown ServiceType = "UNKNOWN"
	ServiceTypeAndroid ServiceType = "ANDROID"
	ServiceTypeIOS     ServiceType = "IOS"
	ServiceTypeWEB     ServiceType = "WEB"
)

// ToServiceType normalizes a raw platform string to a ServiceType.
func ToServiceType(s string) ServiceType {
	switch strings.ToUpper(s) {
	case "ANDROID":
		return ServiceTypeAndroid
	case "IOS":
		return ServiceTypeIOS
	case "WEB":
		return ServiceTypeWEB
	default:
		return ServiceType(strings.ToUpper(s))
	}
}
