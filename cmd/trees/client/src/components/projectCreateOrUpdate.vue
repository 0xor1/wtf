<template>
  <div>
    <div v-if="loading">
      loading...
    </div>
    <div v-else>
      <h1>project {{isCreate? 'create': 'update'}}</h1>
      <span>
        <input ref="name" v-model="name" placeholder="name" @keyup="validate" @keydown.enter="ok">
        <label> name</label>
        <p v-if="nameErr.length > 0" class="err">{{nameErr}}</p>
      </span>
      <span>
        <input id="public" type="checkbox" v-model="isPublic" placeholder="isPublic" @keydown.enter="ok">
        <label for="public"> public</label>
      </span>
      <span>
        <select v-model="currencyCode">
          <option v-for="currency in currencies" v-bind:value="currency" v-bind:key="currency">
            {{currency}}
          </option>
        </select>
        <label> currency code</label>
      </span>
      <span>
        <input v-model.number="hoursPerDay" :min="0" :max="24" type="number" placeholder="hours per day" @blur="validate" @keyup="validate" @keydown.enter="ok">
        <label> hours per day</label>
      </span>
      <span>
        <input v-model.number="daysPerWeek" :min="0" :max="7" type="number" placeholder="days per week" @blur="validate" @keyup="validate" @keydown.enter="ok">
        <label> days per week</label>
      </span>
      <span class="date">
        <datepicker :monday-first="true" v-model="startOn" placeholder="start on" @closed="validate"></datepicker>
        <label> start on</label>
      </span>
      <span class="date">
        <datepicker :monday-first="true" v-model="endOn" placeholder="end on" @closed="validate"></datepicker>
        <label> end on</label>
      </span>
      <button @click="ok">{{isCreate? 'create': 'update'}}</button>
      <button @click="close">close</button>
      <span v-if="err.length > 0" class="err">{{err}}</span>
    </div>
  </div>
</template>

<script>
  import datepicker from 'vuejs-datepicker';
  export default {
    name: 'projectCreateOrUpdate',
    components: {datepicker},
    props: {
      project: Object
    },
    computed: {
      isCreate(){
        return this.project == null
      }
    },
    data: function() {
      return this.initState()
    },
    methods: {
      initState (){
        return {
          loading: true,
          name: "",
          nameErr: true,
          isPublic: false,
          currencyCode: "USD",
          hoursPerDay: null,
          daysPerWeek: null,
          startOn: null,
          endOn: null,
          err: "",
          currencies: [
            "AED",
            "AFN",
            "ALL",
            "AMD",
            "ANG",
            "AOA",
            "ARS",
            "AUD",
            "AWG",
            "AZN",
            "BAM",
            "BBD",
            "BDT",
            "BGN",
            "BHD",
            "BIF",
            "BMD",
            "BND",
            "BOB",
            "BOV",
            "BRL",
            "BSD",
            "BTN",
            "BWP",
            "BYN",
            "BZD",
            "CAD",
            "CDF",
            "CHE",
            "CHF",
            "CHW",
            "CLF",
            "CLP",
            "CNY",
            "COP",
            "COU",
            "CRC",
            "CUC",
            "CUP",
            "CVE",
            "CZK",
            "DJF",
            "DKK",
            "DOP",
            "DZD",
            "EGP",
            "ERN",
            "ETB",
            "EUR",
            "FJD",
            "FKP",
            "GBP",
            "GEL",
            "GHS",
            "GIP",
            "GMD",
            "GNF",
            "GTQ",
            "GYD",
            "HKD",
            "HNL",
            "HRK",
            "HTG",
            "HUF",
            "IDR",
            "ILS",
            "INR",
            "IQD",
            "IRR",
            "ISK",
            "JMD",
            "JOD",
            "JPY",
            "KES",
            "KGS",
            "KHR",
            "KMF",
            "KPW",
            "KRW",
            "KWD",
            "KYD",
            "KZT",
            "LAK",
            "LBP",
            "LKR",
            "LRD",
            "LSL",
            "LYD",
            "MAD",
            "MDL",
            "MGA",
            "MKD",
            "MMK",
            "MNT",
            "MOP",
            "MRU",
            "MUR",
            "MVR",
            "MWK",
            "MXN",
            "MXV",
            "MYR",
            "MZN",
            "NAD",
            "NGN",
            "NIO",
            "NOK",
            "NPR",
            "NZD",
            "OMR",
            "PAB",
            "PEN",
            "PGK",
            "PHP",
            "PKR",
            "PLN",
            "PYG",
            "QAR",
            "RON",
            "RSD",
            "RUB",
            "RWF",
            "SAR",
            "SBD",
            "SCR",
            "SDG",
            "SEK",
            "SGD",
            "SHP",
            "SLL",
            "SOS",
            "SRD",
            "SSP",
            "STN",
            "SVC",
            "SYP",
            "SZL",
            "THB",
            "TJS",
            "TMT",
            "TND",
            "TOP",
            "TRY",
            "TTD",
            "TWD",
            "TZS",
            "UAH",
            "UGX",
            "USD",
            "USN",
            "UYI",
            "UYU",
            "UYW",
            "UZS",
            "VES",
            "VND",
            "VUV",
            "WST",
            "XAF",
            "XAG",
            "XAU",
            "XBA",
            "XBB",
            "XBC",
            "XBD",
            "XCD",
            "XDR",
            "XOF",
            "XPD",
            "XPF",
            "XPT",
            "XSU",
            "XTS",
            "XUA",
            "XXX",
            "YER",
            "ZAR",
            "ZMW",
            "ZWL"
          ]
        }
      },
      init(){
        this.$u.copyProps(this.initState(), this)
        if (!this.isCreate) {
          this.$api.user.me().then((me)=>{
            if (me.id !== this.$u.rtr.host()) {
              this.$u.rtr.goHome()
              return
            }
            this.name = this.project.name
            this.isPublic = this.project.isPublic
            this.currencyCode = this.project.currencyCode
            this.hoursPerDay = this.project.hoursPerDay
            this.daysPerWeek = this.project.daysPerWeek
            if (this.project.startOn != null) {
              this.startOn = new Date(this.project.startOn)
            }
            if (this.project.endOn != null) {
              this.endOn = new Date(this.project.endOn)
            }
            this.loading = false
            this.$nextTick(()=>{
              this.$refs.name.focus()
            })
          })
        } else {
          this.loading = false
          this.$nextTick(()=>{
            this.$refs.name.focus()
          })
        }
      },
      validate(){
        if (this.name.length < 1 || this.name.length > 250) {
            this.nameErr = "name must be 1 - 250 characters"
        } else {
            this.nameErr = ""
        }
        if (this.hoursPerDay != null) {
          if (this.hoursPerDay > 24) {
            this.hoursPerDay = 24
          }
          if (this.hoursPerDay < 1) {
            this.hoursPerDay = null
          }
        }
        if (this.daysPerWeek != null) { 
          if (this.daysPerWeek > 7) {
            this.daysPerWeek = 7
          }
          if (this.daysPerWeek < 1) {
            this.daysPerWeek = null
          }
        }
        if (this.startOn != null) {
            this.startOn.setHours(0, 0, 0, 0)
        }
        if (this.endOn != null) {
            this.endOn.setHours(0, 0, 0, 0)
        }
        if (this.startOn != null && 
          this.endOn != null &&
          this.startOn.getTime() >= this.endOn.getTime()) {
            this.endOn.setDate(this.startOn.getDate()+1)
        }
        return this.nameErr.length === 0
      },
      ok(){
        if (this.validate()) {
          if (this.isCreate) {
            this.$api.project.create({
              name: this.name, 
              isPublic: this.isPublic, 
              currencyCode: this.currencyCode,
              hoursPerDay: this.hoursPerDay,
              daysPerWeek: this.daysPerWeek, 
              startOn: this.startOn,
              endOn: this.endOn
            }).then((p)=>{
              this.$u.rtr.goto(`/host/${p.host}/project/${p.id}/task/${p.id}`)
            })
          } else {
            this.$api.project.updateOne({
              id: this.project.id, 
              name: {v: this.name},
              isPublic: {v: this.isPublic},
              currencyCode: {v: this.currencyCode},
              hoursPerDay: {v: this.hoursPerDay},
              daysPerWeek: {v: this.daysPerWeek},
              startOn: {v: this.startOn},
              endOn: {v: this.endOn}
            }).then((p)=>{
              this.$u.copyProps(p, this.project)
              this.close()
            })
          }
        }
      },
      close(){
        this.$emit('close')
      },
      handleEsc(e){
        if (e.key == "Escape") {
          this.close()
        }
      }
    },
    mounted(){
      this.init()
      window.addEventListener('keydown', this.handleEsc)
    },
    destroyed(){
      window.removeEventListener('keydown', this.handleEsc)
    },
    watch: {
      $route () {
        this.init()
      }
    }
  }
</script>

<style scoped lang="scss">
div > div {
  & > * {
    display: block;
    margin-bottom: 5px;
  }
  button, a{
    display: inline;
    margin-right: 15px;
  }
  input[type="number"] {
    width: 10pc;
  }
  p {
    margin: 0.3pc 0;
  }
  span.date {
    &> * {
      display: inline-block;
    }
    label{
      margin-left: 0.8pc;
    }
  }
}
</style>