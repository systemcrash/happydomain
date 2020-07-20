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

import SourceSettingsApi from '@/services/SourceSettingsApi'

export default {
  data () {
    return {
      form: null,
      nextIsWorking: false,
      previousIsWorking: false,
      settings: null,
      state: 0
    }
  },

  computed: {
    isLoading () {
      return this.form == null || this.sourceSpecs == null
    }
  },

  mounted () {
    this.resetSettings()
    this.updateSourceSettingsForm()
  },

  methods: {
    loadState (toState, recallid, cbSuccess, cbFail) {
      SourceSettingsApi.getSourceSettings(this.sourceSpecsSelected, toState, this.settings, recallid)
        .then(
          response => {
            this.previousIsWorking = false
            this.nextIsWorking = false
            if (response.data.form) {
              this.form = response.data.form
              this.state = toState
              if (response.data.redirect && window.location.pathname !== response.data.redirect) {
                this.$router.push(response.data.redirect)
              } else if (cbSuccess) {
                cbSuccess(toState)
              }
            } else if (response.data.Source) {
              this.$root.$bvToast.toast(
                'Done', {
                  title: (response.data.Source._comment ? response.data.Source._comment : 'Your new source') + ' has been ' + (this.settings._id ? 'updated' : 'added') + '.',
                  autoHideDelay: 5000,
                  variant: 'success',
                  toaster: 'b-toaster-content-right'
                }
              )
              this.state = toState
              if (response.data.redirect && window.location.pathname !== response.data.redirect) {
                this.$router.push(response.data.redirect)
              } else if (cbSuccess) {
                cbSuccess(toState, response.data.Source)
              } else {
                this.$router.push('/sources/' + encodeURIComponent(response.data.Source._id) + '/domains')
              }
            }
          },
          error => {
            this.previousIsWorking = false
            this.nextIsWorking = false
            this.$root.$bvToast.toast(
              error.response.data.errmsg, {
                title: 'Something went wrong during source configuration validation',
                autoHideDelay: 5000,
                variant: 'danger',
                toaster: 'b-toaster-content-right'
              }
            )
            if (cbFail) {
              cbFail(error.response.data)
            }
          })
    },

    resetSettings () {
      this.settings = {
        Source: {},
        _comment: '',
        redirect: null
      }
    },

    updateSourceSettingsForm () {
      if (this.sourceSpecsSelected && this.state >= 0) {
        this.loadState(this.state, this.$route.query.recall)
      }
    }
  }
}
