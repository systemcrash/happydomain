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

package happydns

import ()

const (
	FIELDHINT_HIDE = iota
	FIELDHINT_TOOLTIP
	FIELDHINT_FOCUSED
	FIELDHINT_ALWAYS
)

const (
	ZONEVIEW_GRID = iota
	ZONEVIEW_LIST
	ZONEVIEW_RECORDS
)

// UserSettings represents the settings for an account.
type UserSettings struct {
	// Language saves the locale defined by the user.
	Language string `json:"language,omitempty"`

	// Newsletter indicates wether the user wants to receive the newsletter or not.
	Newsletter bool `json:"newsletter,omitempty"`

	// FieldHint stores the way form hints are displayed.
	FieldHint int `json:"fieldhint"`

	// ZoneView keeps the view of the zone wanted by the user.
	ZoneView int `json:"zoneview"`
}

func DefaultUserSettings() *UserSettings {
	return &UserSettings{
		Language:   "en",
		Newsletter: false,
		FieldHint:  FIELDHINT_FOCUSED,
		ZoneView:   ZONEVIEW_GRID,
	}
}
