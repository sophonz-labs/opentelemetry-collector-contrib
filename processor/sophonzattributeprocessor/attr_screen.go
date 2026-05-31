// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzattributeprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor"

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"

	sophonzmetadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/metadata"
	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

func (p *SOPHONZAttributeProcessor) setScreenAttribute(serviceNamespace pcommon.Value, attrs pcommon.Map) {
	var screenType string
	if st, ok := attrs.Get(sophonzsemconv.AppScreenType); ok {
		screenType = strings.ToUpper(st.Str())
		attrs.Remove(sophonzsemconv.AppScreenType)
		attrs.PutStr(sophonzsemconv.AppScreenType, screenType)
	}

	if p.MetadataManager == nil {
		return
	}
	screens, ok := p.MetadataManager.Screen.Load().(sophonzmetadata.ServiceScreenMap)
	if !ok || screens == nil {
		return
	}

	if screenName, ok := attrs.Get(sophonzsemconv.AppScreenName); ok {
		key := sophonzmetadata.ServiceScreenKey{
			ServiceNamespace: serviceNamespace.Str(),
			ScreenName:       screenName.Str(),
			ScreenType:       screenType,
		}

		if val, ok := screens[key]; ok {
			attrs.PutInt(sophonzsemconv.AppScreenID, int64(val.ScreenID))
			attrs.PutInt(sophonzsemconv.AppScreenGroupID, int64(val.ScreenGroupID))
			attrs.PutStr(sophonzsemconv.AppScreenGroupName, val.ScreenGroupName)
		}
	}
}
