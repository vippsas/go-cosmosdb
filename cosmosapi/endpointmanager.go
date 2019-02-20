package cosmosapi

import (
	"context"
	"net/url"
	"strings"
	"sync"
)

// globalEndpointsManager implements the logic for endpoint management for geo-replicated
// database accounts. Ported from the python cosmos client
type globalEndpointsManager struct {
	Client                  *Client
	DefaultEndpoint         string
	EnableEndpointDiscovery bool
	PreferredLocations      []string

	mu                         sync.RWMutex
	isEndpointCacheInitialized bool
	readEndpoint               string
	writeEndpoint              string
}

// ReadEndpoint gets the current read endpoint from the endpoint cache.
func (m *globalEndpointsManager) ReadEndpoint(ctx context.Context) (string, error) {
	err := m.initializeEndpointsList(ctx)
	if err != nil {
		return "", err
	}
	return m.readEndpoint, nil
}

// WriteEndpoint gets the current read endpoint from the endpoint cache.
func (m *globalEndpointsManager) WriteEndpoint(ctx context.Context) (string, error) {
	err := m.initializeEndpointsList(ctx)
	if err != nil {
		return "", err
	}
	return m.writeEndpoint, nil
}

func (m *globalEndpointsManager) RefreshEndpointsList(ctx context.Context) error {
	if !m.isEndpointCacheInitialized {
		return m.initializeEndpointsList(ctx)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	we, re, err := m.getPreferredEndpoints(ctx)
	m.writeEndpoint = we
	m.readEndpoint = re
	return err
}

func (m *globalEndpointsManager) initializeEndpointsList(ctx context.Context) error {
	if m.isEndpointCacheInitialized {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isEndpointCacheInitialized {
		return nil
	}
	we, re, err := m.getPreferredEndpoints(ctx)
	if err != nil {
		return err
	}
	m.writeEndpoint = we
	m.readEndpoint = re
	m.isEndpointCacheInitialized = true
	return nil
}

func (m *globalEndpointsManager) getPreferredEndpoints(ctx context.Context) (string, string, error) {
	//from the geo-replicated database account and then updating the locations cache.
	// We skip the refreshing if EnableEndpointDiscovery is set to False
	if !m.EnableEndpointDiscovery {
		return "", "", nil
	}
	database_account, err := m.getDatabaseAccount(ctx)
	var writableLocations []DatabaseLocation
	var readableLocations []DatabaseLocation
	if err == nil {
		writableLocations = database_account.WritableLocations
		readableLocations = database_account.ReadableLocations
	} else if err != ErrNotFound {
		return "", "", err
	}

	// Read and Write endpoints will be initialized to default endpoint if we were not able to get the database account info
	writeEndpoint, readEndpoint := m.updateLocationsCache(writableLocations, readableLocations)
	return writeEndpoint, readEndpoint, nil

}

// getDatabaseAccount gets the database account first by using the default endpoint, and if that fails
// use the endpoints for the preferred locations in the order they are specified to get
// the database account.
func (m *globalEndpointsManager) getDatabaseAccount(ctx context.Context) (DatabaseAccount, error) {
	dbAcc, err := m.Client.GetDatabaseAccount(ctx)
	if err != nil {
		for _, locationName := range m.PreferredLocations {
			locationalEndpoint, locErr := m.getLocationalEndpoint(m.DefaultEndpoint, locationName)
			if locErr != nil {
				return dbAcc, locErr
			}
			dbAcc, err = m.Client.getDatabaseAccountCustomURL(ctx, locationalEndpoint)
			// If for any reason(non-globaldb related), we are not able to get the database account from the above call to GetDatabaseAccount,
			// we would try to get this information from any of the preferred locations that the user might have specified(by creating a locational endpoint)
			// and keeping eating the error until we get the database account and return None at the end, if we are not able to get that info from any endpoints
			if err == nil {
				break
			}
		}
	}
	return dbAcc, err
}

func (m *globalEndpointsManager) getLocationalEndpoint(defaultEndpoint string, locationName string) (string, error) {
	// For default_endpoint like 'https://contoso.documents.azure.com:443/' parse it to generate URL format
	// This default_endpoint should be global endpoint(and cannot be a locational endpoint) and we agreed to document that
	endpointUrl, err := url.Parse(defaultEndpoint)
	if err != nil {
		return "", err
	}
	// hostname attribute in endpoint_url will return 'contoso.documents.azure.com'
	if endpointUrl.Hostname() != "" {
		hostnameParts := strings.Split(strings.ToLower(endpointUrl.Hostname()), ".")
		if len(hostnameParts) > 0 {
			// global_database_account_name will return 'contoso'
			globalDatabaseAccountName := hostnameParts[0]

			// Prepare the locational_database_account_name as contoso-EastUS for location_name 'East US'
			locationalDatabaseAccountName := globalDatabaseAccountName + "-" + strings.Replace(locationName, " ", "", -1)

			// Replace 'contoso' with 'contoso-EastUS' and return locational_endpoint as https://contoso-EastUS.documents.azure.com:443/
			locationalEndpoint := strings.Replace(strings.ToLower(defaultEndpoint), globalDatabaseAccountName, locationalDatabaseAccountName, 1)
			return locationalEndpoint, nil

		}
	}
	return "", nil
}

func (m *globalEndpointsManager) updateLocationsCache(writableLocations, readableLocations []DatabaseLocation) (writeEndpoint, readEndpoint string) {
	// Use the default endpoint as Read and Write endpoints if EnableEndpointDiscovery
	// is set to False.
	if !m.EnableEndpointDiscovery {
		writeEndpoint = m.DefaultEndpoint
		readEndpoint = m.DefaultEndpoint
		return
	}

	// Use the default endpoint as Write endpoint if there are no writable locations, or
	// first writable location as Write endpoint if there are writable locations
	if len(writableLocations) == 0 {
		writeEndpoint = m.DefaultEndpoint
	} else {
		writeEndpoint = writableLocations[0].DatabaseAccountEndpoint
	}

	// Use the Write endpoint as Read endpoint if there are no readable locations
	if len(readableLocations) == 0 {
		readEndpoint = writeEndpoint
	} else {
		// Use the writable location as Read endpoint if there are no preferred locations or
		// none of the preferred locations are in read or write locations
		readEndpoint = writeEndpoint
		if len(m.PreferredLocations) == 0 {
			return
		}

		for _, preferredLocation := range m.PreferredLocations {
			// Use the first readable location as Read endpoint from the preferred locations
			for _, readLocation := range readableLocations {
				if readLocation.Name == preferredLocation {
					readEndpoint = readLocation.DatabaseAccountEndpoint
					return
				}
			}
			// Else, use the first writable location as Read endpoint from the preferred locations
			for _, writeLocation := range writableLocations {
				if writeLocation.Name == preferredLocation {
					readEndpoint = writeLocation.DatabaseAccountEndpoint
					return
				}
			}
		}
	}
	return writeEndpoint, readEndpoint
}
