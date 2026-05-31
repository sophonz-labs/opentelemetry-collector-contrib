package metadata

// Screen is a row of sophonz_metadata.v_postgres_screen.
type Screen struct {
	ID               int32  `ch:"id"`
	Name             string `ch:"screenName"`
	Type             string `ch:"screenType"`
	GroupID          int32  `ch:"groupID"`
	GroupName        string `ch:"groupName"`
	ServiceNamespace string `ch:"serviceNamespace"`
}

// ServiceScreenKey identifies a screen within a service namespace.
type ServiceScreenKey struct {
	ServiceNamespace string
	ScreenName       string
	ScreenType       string
}

// ServiceScreenValue holds the resolved screen identifiers.
type ServiceScreenValue struct {
	ScreenID        int32
	ScreenGroupID   int32
	ScreenGroupName string
}

// ServiceScreenMap is the lookup table cached by the manager.
type ServiceScreenMap map[ServiceScreenKey]ServiceScreenValue

// ConvertToScreenMap builds the lookup table from raw screen rows, skipping
// incomplete entries.
func ConvertToScreenMap(screens []Screen) ServiceScreenMap {
	serviceScreenMap := make(ServiceScreenMap)
	for _, screen := range screens {
		if screen.ServiceNamespace != "" && screen.Name != "" && screen.Type != "" {
			key, val := screen.toMapData()
			serviceScreenMap[key] = val
		}
	}
	return serviceScreenMap
}

func (s Screen) toMapData() (ServiceScreenKey, ServiceScreenValue) {
	key := ServiceScreenKey{
		ServiceNamespace: s.ServiceNamespace,
		ScreenName:       s.Name,
		ScreenType:       s.Type,
	}
	value := ServiceScreenValue{
		ScreenID:        s.ID,
		ScreenGroupID:   s.GroupID,
		ScreenGroupName: s.GroupName,
	}
	return key, value
}
