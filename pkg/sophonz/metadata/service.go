package metadata

import "strings"

// Service is a row of sophonz_metadata.v_postgres_service.
type Service struct {
	ID        string `ch:"id"`
	Namespace string `ch:"serviceNamespace"`
	Name      string `ch:"serviceName"`
	Type      string `ch:"type"`
	Key       string `ch:"key"`
}

// ServiceKeyMap maps service ID to Service.
type ServiceKeyMap map[string]Service

// ConvertToServiceMap builds the service lookup keyed by ID, skipping
// incomplete entries.
func ConvertToServiceMap(services []Service) ServiceKeyMap {
	serviceKeyMap := make(ServiceKeyMap)
	for _, service := range services {
		if service.Name != "" && service.Namespace != "" {
			serviceKeyMap[service.ID] = service
		}
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
