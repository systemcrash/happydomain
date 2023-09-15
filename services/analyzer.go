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

package svcs

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/miekg/dns"

	"git.happydns.org/happyDomain/model"
	"git.happydns.org/happyDomain/utils"
)

type Analyzer struct {
	origin     string
	zone       models.Records
	services   map[string][]*happydns.ServiceCombined
	defaultTTL uint32
}

func (a *Analyzer) GetOrigin() string {
	return a.origin
}

type AnalyzerRecordFilter struct {
	Prefix       string
	Domain       string
	SubdomainsOf string
	Contains     string
	Type         uint16
	Ttl          uint32
}

func (a *Analyzer) SearchRR(arrs ...AnalyzerRecordFilter) (rrs models.Records) {
	for _, record := range a.zone {
		for _, arr := range arrs {
			if rdtype, ok := dns.StringToType[record.Type]; strings.HasPrefix(record.NameFQDN, arr.Prefix) &&
				strings.HasSuffix(record.NameFQDN, arr.SubdomainsOf) &&
				(arr.Domain == "" || record.NameFQDN == arr.Domain) &&
				(arr.Type == 0 || (ok && rdtype == arr.Type)) &&
				(arr.Ttl == 0 || record.TTL == arr.Ttl) &&
				(arr.Contains == "" || strings.Contains(fmt.Sprintf("%s. %d IN %s %s", record.NameFQDN, record.TTL, record.Type, record.String()), arr.Contains)) {
				rrs = append(rrs, record)
			}
		}
	}

	return
}

func (a *Analyzer) UseRR(rr *models.RecordConfig, domain string, svc happydns.Service) error {
	found := false
	for k, record := range a.zone {
		if record == rr {
			found = true
			a.zone[k] = a.zone[len(a.zone)-1]
			a.zone = a.zone[:len(a.zone)-1]
		}
	}

	if !found {
		return errors.New("Record not found.")
	}

	// svc nil, just drop the record from the zone (probably handle another way)
	if svc == nil {
		return nil
	}

	// Remove origin to get an relative domain here
	domain = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(domain, "."), strings.TrimSuffix(a.origin, ".")), ".")

	for _, service := range a.services[domain] {
		if service.Service == svc {
			service.Comment = svc.GenComment(a.origin)
			service.NbResources = svc.GetNbResources()
			return nil
		}
	}

	hash := sha1.New()
	io.WriteString(hash, rr.String())

	var ttl uint32 = 0
	if rr.TTL != a.defaultTTL {
		ttl = rr.TTL
	}

	a.services[domain] = append(a.services[domain], &happydns.ServiceCombined{
		Service: svc,
		ServiceMeta: happydns.ServiceMeta{
			Id:          hash.Sum(nil),
			Type:        reflect.Indirect(reflect.ValueOf(svc)).Type().String(),
			Domain:      domain,
			Ttl:         ttl,
			Comment:     svc.GenComment(a.origin),
			NbResources: svc.GetNbResources(),
		},
	})

	return nil
}

func getMostUsedTTL(zone models.Records) uint32 {
	ttls := map[uint32]int{}
	for _, rr := range zone {
		ttls[rr.TTL] += 1
	}

	var max uint32 = 0
	for k, v := range ttls {
		if w, ok := ttls[max]; !ok || v > w {
			max = k
		}
	}

	return max
}

func AnalyzeZone(origin string, zone models.Records) (svcs map[string][]*happydns.ServiceCombined, defaultTTL uint32, err error) {
	defaultTTL = getMostUsedTTL(zone)

	a := Analyzer{
		origin:     origin,
		zone:       zone,
		services:   map[string][]*happydns.ServiceCombined{},
		defaultTTL: defaultTTL,
	}

	// Find services between all registered ones
	for _, service := range OrderedServices() {
		if service.Analyzer == nil {
			continue
		}

		if err = service.Analyzer(&a); err != nil {
			return
		}
	}

	svcs = a.services

	// Consider records not used by services as Orphan
	for _, record := range a.zone {
		// Skip DNSSEC records
		if rdtype, ok := dns.StringToType[record.Type]; ok && utils.IsDNSSECType(rdtype) {
			continue
		}
		if record.NameFQDN == "__dnssec."+origin && record.Type == "TXT" {
			continue
		}

		domain := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(record.NameFQDN, "."), strings.TrimSuffix(a.origin, ".")), ".")

		hash := sha1.New()
		io.WriteString(hash, record.String())

		orphan := &Orphan{record.Type, record.String()}
		svcs[domain] = append(svcs[domain], &happydns.ServiceCombined{
			Service: orphan,
			ServiceMeta: happydns.ServiceMeta{
				Id:          hash.Sum(nil),
				Type:        reflect.Indirect(reflect.ValueOf(orphan)).Type().String(),
				Domain:      domain,
				Ttl:         record.TTL,
				NbResources: 1,
				Comment:     orphan.GenComment(a.origin),
			},
		})
	}

	return
}
