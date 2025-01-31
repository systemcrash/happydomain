// Copyright or © or Copr. happyDNS (2020)
//
// contact@happydomain.org
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

package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"

	"git.happydns.org/happydomain/config"
	"git.happydns.org/happydomain/model"
	"git.happydns.org/happydomain/storage"
)

func declareDomainsRoutes(cfg *config.Options, router *gin.RouterGroup) {
	router.GET("/domains", GetDomains)
	router.POST("/domains", addDomain)

	apiDomainsRoutes := router.Group("/domains/:domain")
	apiDomainsRoutes.Use(DomainHandler)

	apiDomainsRoutes.GET("", GetDomain)
	apiDomainsRoutes.PUT("", UpdateDomain)
	apiDomainsRoutes.DELETE("", delDomain)

	declareZonesRoutes(cfg, apiDomainsRoutes)
}

func GetDomains(c *gin.Context) {
	user := myUser(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errmsg": "User not defined"})
		return
	}

	if domains, err := storage.MainStore.GetDomains(user); err != nil {
		log.Printf("%s: An error occurs when trying to GetDomains: %s", c.ClientIP(), err.Error())
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"errmsg": err})
	} else if len(domains) > 0 {
		c.JSON(http.StatusOK, domains)
	} else {
		c.JSON(http.StatusOK, []happydns.Domain{})
	}
}

func addDomain(c *gin.Context) {
	var uz happydns.Domain
	err := c.ShouldBindJSON(&uz)
	if err != nil {
		log.Printf("%s sends invalid Domain JSON: %s", c.ClientIP(), err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": fmt.Sprintf("Something is wrong in received data: %s", err.Error())})
		return
	}

	if len(uz.DomainName) <= 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": "The given domain is invalid."})
		return
	}

	uz.DomainName = dns.Fqdn(uz.DomainName)

	if _, ok := dns.IsDomainName(uz.DomainName); !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": fmt.Sprintf("%q is not a valid domain name.", uz.DomainName)})
		return
	}

	user := c.MustGet("LoggedUser").(*happydns.User)

	provider, err := storage.MainStore.GetProvider(user, uz.IdProvider)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": fmt.Sprintf("Unable to find the provider.")})
		return
	}

	if storage.MainStore.DomainExists(uz.DomainName) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": "This domain has already been imported."})
		return

	} else if err := provider.DomainExists(uz.DomainName); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": err.Error()})
		return
	} else if err := storage.MainStore.CreateDomain(user, &uz); err != nil {
		log.Printf("%s was unable to CreateDomain: %s", c.ClientIP(), err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errmsg": "Sorry, we are unable to create your domain now."})
		return
	} else {
		c.JSON(http.StatusOK, uz)
	}
}

func DomainHandler(c *gin.Context) {
	// Get a valid user
	user := myUser(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errmsg": "User not defined."})
		return
	}

	dnid, err := happydns.NewIdentifierFromString(c.Param("domain"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": fmt.Sprintf("Invalid domain identifier: %s", err.Error())})
		return
	}

	domain, err := storage.MainStore.GetDomain(user, dnid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"errmsg": "Domain not found"})
		return
	}

	// If source is provided, check that the domain is a parent of the source
	var source *happydns.SourceMeta
	if src, exists := c.Get("source"); exists {
		source = &src.(*happydns.SourceCombined).SourceMeta
	} else if src, exists := c.Get("sourcemeta"); exists {
		source = src.(*happydns.SourceMeta)
	}
	if source != nil && !source.Id.Equals(domain.IdProvider) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"errmsg": "Domain not found (not child of source)"})
		return
	}

	c.Set("domain", domain)

	c.Next()
}

type apiDomain struct {
	Id          happydns.Identifier `json:"id"`
	IdUser      happydns.Identifier `json:"id_owner"`
	IdProvider  happydns.Identifier `json:"id_provider"`
	DomainName  string              `json:"domain"`
	ZoneHistory []happydns.ZoneMeta `json:"zone_history"`
	Group       string              `json:"group,omitempty"`
}

func GetDomain(c *gin.Context) {
	domain := c.MustGet("domain").(*happydns.Domain)
	ret := &apiDomain{
		Id:          domain.Id,
		IdUser:      domain.IdUser,
		IdProvider:  domain.IdProvider,
		DomainName:  domain.DomainName,
		ZoneHistory: []happydns.ZoneMeta{},
		Group:       domain.Group,
	}

	for _, zm := range domain.ZoneHistory {
		zoneMeta, err := storage.MainStore.GetZoneMeta(zm)

		if err != nil {
			log.Println("%s: An error occurs in getDomain, when retrieving a meta history: %s", c.ClientIP(), err.Error())
		} else {
			ret.ZoneHistory = append(ret.ZoneHistory, *zoneMeta)
		}
	}

	c.JSON(http.StatusOK, ret)
}

func UpdateDomain(c *gin.Context) {
	old := c.MustGet("domain").(*happydns.Domain)

	var domain apiDomain
	err := c.ShouldBindJSON(&domain)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": err.Error()})
		return
	}

	if !old.Id.Equals(domain.Id) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errmsg": "You cannot change the domain reserved ID"})
		return
	}

	old.Group = domain.Group

	err = storage.MainStore.UpdateDomain(old)
	if err != nil {
		log.Printf("%s: Unable to UpdateDomain in UpdateDomain: %s", c.ClientIP(), err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errmsg": "Sorry, we are currently unable to update your domain. Please retry later."})
		return
	}

	c.JSON(http.StatusOK, old)
}

func delDomain(c *gin.Context) {
	if err := storage.MainStore.DeleteDomain(c.MustGet("domain").(*happydns.Domain)); err != nil {
		log.Printf("%s was unable to DeleteDomain: %s", c.ClientIP(), err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errmsg": fmt.Sprintf("Unable to delete your domain: %s", err.Error())})
		return
	}

	c.JSON(http.StatusNoContent, true)
}
