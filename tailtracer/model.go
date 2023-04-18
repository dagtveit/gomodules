package tailtracer

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
	"math/rand"
	"time"
)

type Atm struct {
	ID           int64
	Version      string
	Name         string
	StateID      string
	SerialNumber string
	ISPNetwork   string
}

type BackendSystem struct {
	Version       string
	ProcessName   string
	OSType        string
	OSVersion     string
	CloudProvider string
	CloudRegion   string
	Endpoint      string
}

func generateAtm() Atm {
	i := getRandomNumber(1, 2)
	var newAtm Atm

	switch i {
	case 1:
		newAtm = Atm{
			ID:           111,
			Name:         "ATM-111-IL",
			SerialNumber: "atmxph-2022-111",
			Version:      "v1.0",
			ISPNetwork:   "comcast-chicago",
			StateID:      "IL",
		}

	case 2:
		newAtm = Atm{
			ID:           222,
			Name:         "ATM-222-CA",
			SerialNumber: "atmxph-2022-222",
			Version:      "v1.0",
			ISPNetwork:   "comcast-sanfrancisco",
			StateID:      "CA",
		}
	}

	return newAtm
}

func generateBackendSystem() BackendSystem {
	i := getRandomNumber(1, 3)

	newBackend := BackendSystem{
		ProcessName:   "accounts",
		Version:       "v2.5",
		OSType:        "lnx",
		OSVersion:     "4.16.10-300.fc28.x86_64",
		CloudProvider: "amzn",
		CloudRegion:   "us-east-2",
	}

	switch i {
	case 1:
		newBackend.Endpoint = "api/v2.5/balance"
	case 2:
		newBackend.Endpoint = "api/v2.5/deposit"
	case 3:
		newBackend.Endpoint = "api/v2.5/withdrawn"
	}

	return newBackend
}

func getRandomNumber(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	i := (rand.Intn(max-min+1) + min)
	return i
}

func generateTraces(numberOfTraces int) ptrace.Traces {
	traces := ptrace.NewTraces()

	for i := 0; i <= numberOfTraces; i++ {
		newAtm := generateAtm()
		newBackendSystem := generateBackendSystem()

		resourceSpan := traces.ResourceSpans().AppendEmpty()
		atmResource := resourceSpan.Resource()
		fillResourceWithAtm(&atmResource, newAtm)

		resourceSpan = traces.ResourceSpans().AppendEmpty()
		backendResource := resourceSpan.Resource()
		fillResourceWithBackendSystem(&backendResource, newBackendSystem)
	}

	return traces
}

func fillResourceWithAtm(resource *pcommon.Resource, atm Atm) {
	atmAttrs := resource.Attributes()
	atmAttrs.PutInt("atm.id", atm.ID)
	atmAttrs.PutStr("atm.stateid", atm.StateID)
	atmAttrs.PutStr("atm.ispnetwork", atm.ISPNetwork)
	atmAttrs.PutStr("atm.serialnumber", atm.SerialNumber)
	atmAttrs.PutStr(conventions.AttributeServiceName, atm.Name)
	atmAttrs.PutStr(conventions.AttributeServiceVersion, atm.Version)

}

func fillResourceWithBackendSystem(resource *pcommon.Resource, backend BackendSystem) {
	backendAttrs := resource.Attributes()
	var osType, cloudProvider string

	switch {
	case backend.CloudProvider == "amzn":
		cloudProvider = conventions.AttributeCloudProviderAWS
	case backend.OSType == "mcrsft":
		cloudProvider = conventions.AttributeCloudProviderAzure
	case backend.OSType == "gogl":
		cloudProvider = conventions.AttributeCloudProviderGCP
	}

	backendAttrs.PutStr(conventions.AttributeCloudProvider, cloudProvider)
	backendAttrs.PutStr(conventions.AttributeCloudRegion, backend.CloudRegion)

	switch {
	case backend.OSType == "lnx":
		osType = conventions.AttributeOSTypeLinux
	case backend.OSType == "wndws":
		osType = conventions.AttributeOSTypeWindows
	case backend.OSType == "slrs":
		osType = conventions.AttributeOSTypeSolaris
	}

	backendAttrs.PutStr(conventions.AttributeOSType, osType)
	backendAttrs.PutStr(conventions.AttributeOSVersion, backend.OSVersion)

	backendAttrs.PutStr(conventions.AttributeServiceName, backend.ProcessName)
	backendAttrs.PutStr(conventions.AttributeServiceVersion, backend.Version)
}
