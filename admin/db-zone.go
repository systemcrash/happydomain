// Copyright or © or Copr. happyDNS (2020)
//
// contact@happydns.org
//
// This software is a computer program whose purpose is to provide a modern
// interface to interact with DNS systems.
//
// This software is governed by the CeCILL license under French law and abiding
// by the rules of distribution of free software.  You can use, modify and/or
// redistribute the software under the terms of the CeCILL license as
// circulated by CEA, CNRS and INRIA at the following URL
// "http://www.cecill.info".
//
// As a counterpart to the access to the source code and rights to copy, modify
// and redistribute granted by the license, users are provided only with a
// limited warranty and the software's author, the holder of the economic
// rights, and the successive licensors have only limited liability.
//
// In this respect, the user's attention is drawn to the risks associated with
// loading, using, modifying and/or developing or reproducing the software by
// the user in light of its specific status of free software, that may mean
// that it is complicated to manipulate, and that also therefore means that it
// is reserved for developers and experienced professionals having in-depth
// computer knowledge. Users are therefore encouraged to load and test the
// software's suitability as regards their requirements in conditions enabling
// the security of their systems and/or data to be ensured and, more generally,
// to use and operate it in the same conditions as regards security.
//
// The fact that you are presently reading this means that you have had
// knowledge of the CeCILL license and that you accept its terms.

package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"git.happydns.org/happydns/api"
	"git.happydns.org/happydns/config"
	"git.happydns.org/happydns/model"
	"git.happydns.org/happydns/storage"
)

func init() {
	router.GET("/api/users/:userid/domains/:domain/zones", api.ApiHandler(userHandler(domainHandler(getUserDomainZones))))
	router.PUT("/api/users/:userid/domains/:domain/zones", api.ApiHandler(userHandler(domainHandler(updateUserDomainZones))))
	router.POST("/api/users/:userid/domains/:domain/zones", api.ApiHandler(userHandler(domainHandler(newUserDomainZone))))

	router.GET("/api/users/:userid/domains/:domain/zones/:zoneid", api.ApiHandler(zoneHandler(getZone)))
	router.PUT("/api/users/:userid/domains/:domain/zones/:zoneid", api.ApiHandler(zoneHandler(updateZone)))
	router.DELETE("/api/users/:userid/domains/:domain/zones/:zoneid", api.ApiHandler(deleteZone))

	router.GET("/api/zones/:zoneid", api.ApiHandler(zoneHandler(getZone)))
	router.PUT("/api/zones/:zoneid", api.ApiHandler(zoneHandler(updateZone)))
	router.DELETE("/api/zones/:zoneid", api.ApiHandler(deleteZone))
}

func getUserDomainZones(_ *config.Options, domain *happydns.Domain, _ httprouter.Params, _ io.Reader) api.Response {
	return api.NewAPIResponse(domain.ZoneHistory, nil)
}

func updateUserDomainZones(_ *config.Options, domain *happydns.Domain, _ httprouter.Params, body io.Reader) api.Response {
	err := json.NewDecoder(body).Decode(&domain.ZoneHistory)
	if err != nil {
		return api.NewAPIErrorResponse(http.StatusBadRequest, fmt.Errorf("Something is wrong in received data: %w", err))
	}

	return api.NewAPIResponse(domain, storage.MainStore.UpdateDomain(domain))
}

func newUserDomainZone(_ *config.Options, domain *happydns.Domain, _ httprouter.Params, body io.Reader) api.Response {
	uz := &happydns.Zone{}
	err := json.NewDecoder(body).Decode(&uz)
	if err != nil {
		return api.NewAPIErrorResponse(http.StatusBadRequest, fmt.Errorf("Something is wrong in received data: %w", err))
	}
	uz.Id = 0

	return api.NewAPIResponse(uz, storage.MainStore.CreateZone(uz))
}

func zoneHandler(f func(*config.Options, *happydns.Zone, httprouter.Params, io.Reader) api.Response) func(*config.Options, httprouter.Params, io.Reader) api.Response {
	return func(opts *config.Options, ps httprouter.Params, body io.Reader) api.Response {
		zoneid, err := strconv.ParseInt(ps.ByName("zoneid"), 10, 64)
		if err != nil {
			return api.NewAPIErrorResponse(http.StatusNotFound, err)
		} else {
			zone, err := storage.MainStore.GetZone(zoneid)
			if err != nil {
				return api.NewAPIErrorResponse(http.StatusNotFound, err)
			} else {
				return f(opts, zone, ps, body)
			}
		}
	}
}

func getZone(_ *config.Options, zone *happydns.Zone, _ httprouter.Params, _ io.Reader) api.Response {
	return api.NewAPIResponse(zone, nil)
}

func updateZone(_ *config.Options, zone *happydns.Zone, _ httprouter.Params, body io.Reader) api.Response {
	uz := &happydns.Zone{}
	err := json.NewDecoder(body).Decode(&uz)
	if err != nil {
		return api.NewAPIErrorResponse(http.StatusBadRequest, fmt.Errorf("Something is wrong in received data: %w", err))
	}
	uz.Id = zone.Id

	return api.NewAPIResponse(uz, storage.MainStore.UpdateZone(uz))
}

func deleteZone(opts *config.Options, ps httprouter.Params, body io.Reader) api.Response {
	zoneid, err := strconv.ParseInt(ps.ByName("zoneid"), 10, 64)
	if err != nil {
		return api.NewAPIErrorResponse(http.StatusNotFound, err)
	} else {
		return api.NewAPIResponse(true, storage.MainStore.DeleteZone(&happydns.Zone{Id: zoneid}))
	}
}
